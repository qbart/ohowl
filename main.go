package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

const (
	hcloudMetadataApiBase = "http://169.254.169.254/hetzner/v1/metadata"
)

type HetznerServerMetadata struct {
	Hostname   string `yaml:"hostname"`
	PublicIpv4 string `yaml:"public-ipv4"`
	InstanceID string `yaml:"instance-id"`
}

type HetznerServerMetadataPrivateNetwork struct {
	Ip           string   `yaml:"ip"`
	AliasIps     []string `yaml:"alias_ips"`
	InterfaceNum int      `yaml:"interface_num"`
	MacAddress   string   `yaml:"mac_address"`
	NetworkID    string   `yaml:"network_id"`
	NetworkName  string   `yaml:"network_name"`
	Network      string   `yaml:"network"`
	Subnet       string   `yaml:"subnet"`
	Gateway      string   `yaml:"gateway"`
}
type HetznerServerMetadataPrivateNetworks = []HetznerServerMetadataPrivateNetwork

type Metadata struct {
	ID          string `json:"id,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	PrivateIpv4 string `json:"ip,omitempty"`
	PublicIpv4  string `json:"public_ip,omitempty"`
}

var (
	flagAgent bool
)

func init() {
	flag.BoolVar(&flagAgent, "agent", false, "Start HTTP server on 1914 port")
	flag.Parse()
}

func main() {
	if flagAgent {
		r := mux.NewRouter()
		r.HandleFunc("/hetzner/metadata", func(w http.ResponseWriter, r *http.Request) {
			data, err := hcloudMetadata()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(toJson(map[string]interface{}{
					"error": err,
				}))
			} else {
				w.Write(toJson(data))
			}
		}).Methods("GET")

		srv := &http.Server{
			Handler:      r,
			Addr:         "127.0.0.1:1914", // `port` in memory of Laughing Owl
			WriteTimeout: 60 * time.Second,
			ReadTimeout:  60 * time.Second,
		}

		log.Fatal(srv.ListenAndServe())
	} else {
		data, err := hcloudMetadata()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(toJson(data)))
	}
}

func hcloudMetadata() (*Metadata, error) {
	body, err := get(fmt.Sprint(hcloudMetadataApiBase))
	if err != nil {
		return nil, err
	}

	var metadata HetznerServerMetadata
	fromYaml(body, &metadata)

	body, err = get(fmt.Sprint(hcloudMetadataApiBase, "/private-networks"))
	if err != nil {
		return nil, err
	}

	var networks HetznerServerMetadataPrivateNetworks
	fromYaml(body, &networks)

	return &Metadata{
		ID:          metadata.InstanceID,
		Hostname:    metadata.Hostname,
		PrivateIpv4: networks[0].Ip,
		PublicIpv4:  metadata.PublicIpv4,
	}, nil
}

func fromYaml(body []byte, o interface{}) {
	if err := yaml.Unmarshal(body, o); err != nil {
		log.Fatal(err)
	}
}

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func toJson(o interface{}) []byte {
	b, err := json.Marshal(o)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
