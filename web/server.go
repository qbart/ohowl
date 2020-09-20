package web

import (
	"context"
	"errors"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qbart/ohowl/owl"
	"github.com/qbart/ohowl/tea"
)

type App struct {
	Debug  bool
	Token  string
	consul *tea.Consul
	vault  *tea.Vault
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

	if a.Token == "" {
		log.Printf("[WARN] Terraform API turned off - token is empty")
	} else {
		tf := r.Group("/tf")
		{
			tf.Use(OwlAuth(a.Token))

			v1 := tf.Group("/v1")
			{
				// C
				v1.PUT("/certificate", func(c *gin.Context) {

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

	tea.WaitForSignal(syscall.SIGINT, syscall.SIGTERM)
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
