package web

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	Debug bool
	Token string
}

func (a *App) Run() error {
	gin.SetMode(gin.ReleaseMode)
	if a.Debug {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

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

	srv := &http.Server{
		Handler:      r,
		Addr:         ":1914", // `port` in memory of Laughing Owl
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	return srv.ListenAndServe()
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
