package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/vault/api"
	"github.com/qbart/ohowl/cloudh"
	"github.com/qbart/ohowl/owl"
	"github.com/qbart/ohowl/tea"
)

type App struct {
	Debug             bool
	DnsEmail          string
	DnsToken          string
	AclToken          string
	CertPathPrefix    string
	AccountPathPrefix string

	consul *tea.Consul
	vault  *tea.Vault
}

type TfCreateRequest struct {
	Domains []string `json:"domains,omitempty" binding:"required"`
}

func (a *App) Run() {
	gin.SetMode(gin.ReleaseMode)
	if a.Debug {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.Status(200)
	})

	if a.AclToken == "" {
		log.Printf("[ERROR] No ACL token provided")
		return
	}

	tf := r.Group("/tf")
	{
		tf.Use(OwlAuth(a.AclToken))

		v1 := tf.Group("/v1")
		{
			// C
			v1.PUT("/certificate", func(c *gin.Context) {
				var req TfCreateRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "No domains provided"})
					return
				}
				log.Printf("Issue: %s", strings.Join(req.Domains, ","))
				fs := cloudh.TlsConsulStorage{KV: a.consul.KV()}
				tls := cloudh.AutoTls{
					Config: cloudh.TlsConfig{
						DnsToken:          a.DnsToken,
						Email:             a.DnsEmail,
						Domains:           req.Domains,
						AccountPathPrefix: a.AccountPathPrefix,
						CertPathPrefix:    a.CertPathPrefix,
						Debug:             a.Debug,
					},
					Storage:        &fs,
					AccountStorage: &fs,
				}

				if err := tls.Issue(); err != nil {
					c.String(http.StatusUnprocessableEntity, fmt.Sprintf("Issue error: %v", err))
				} else {
					c.Status(http.StatusOK)
				}
			})
			// R
			v1.GET("/certificate", func(c *gin.Context) {
			})
			// U
			v1.PATCH("/certificate", func(c *gin.Context) {

			})
			// D
			v1.DELETE("/certificate", func(c *gin.Context) {

			})
		}
	}

	consul, consulErr := tea.NewConsul()
	vault, vaultErr := tea.NewVault()
	err := tea.ErrCoalesce(consulErr, vaultErr)
	if err != nil {
		panic(err)
	}
	a.consul = consul
	a.vault = vault

	err = consul.Register("OhOwl", 1914, []string{"OhOwl", "oh", "ops"}, map[string]string{"version": owl.Version})
	if err != nil {
		log.Fatalf("Consul register failed: %v", err)
	}

	//TODO vault
	a.DnsToken = os.Getenv("HCLOUD_DNS_TOKEN")
	a.DnsEmail = os.Getenv("ACME_EMAIL")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":1914", // `port` in memory of Laughing Owl
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed %v", err)
		}
	}()

	tea.SysCallWaitDefault()
	err = consul.Deregister("OhOwl")
	if err != nil {
		log.Printf("Consul deregister failed: %v", err)
	}

	log.Println("Exiting..")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}

func OwlAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var h OwlAuthHeader
		if err := c.ShouldBindHeader(&h); err != nil {
			log.Printf("[ERROR] Failed to bind headers %v", err)
			c.AbortWithError(http.StatusBadRequest, errors.New("Missing required headers"))
			return
		}

		if h.Token != token {
			log.Printf("[WARN] Invalid token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

type OwlAuthHeader struct {
	Token string `header:"OhOwl-api-token"`
	Email string `header:"OhOwl-email"`
}
