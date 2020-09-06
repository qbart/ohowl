package cmds

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/qbart/ohowl/cloudh"
	"github.com/qbart/ohowl/utils"
	"github.com/spf13/cobra"
)

var cmdAgent = &cobra.Command{
	Use:   "agent",
	Short: "Start HTTP server on 1914 port",
	Run: func(cmd *cobra.Command, args []string) {
		r := mux.NewRouter()
		r.HandleFunc("/hcloud/metadata", func(w http.ResponseWriter, r *http.Request) {
			data, err := cloudh.GetMetadata()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(utils.Json(map[string]interface{}{
					"error": err,
				}))
			} else {
				w.Write(utils.Json(data))
			}
		}).Methods("GET")

		srv := &http.Server{
			Handler:      r,
			Addr:         "127.0.0.1:1914", // `port` in memory of Laughing Owl
			WriteTimeout: 60 * time.Second,
			ReadTimeout:  60 * time.Second,
		}

		log.Fatal(srv.ListenAndServe())
	},
}
