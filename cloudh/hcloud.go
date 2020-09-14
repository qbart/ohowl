package cloudh

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qbart/ohowl/tea"
)

const (
	hcloudMetadataApiBase = "http://169.254.169.254/hetzner/v1/metadata"
)

type ServerMetadata struct {
	Hostname   string `yaml:"hostname"`
	PublicIpv4 string `yaml:"public-ipv4"`
	InstanceID string `yaml:"instance-id"`
}

type ServerMetadataPrivateNetwork struct {
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
type ServerMetadataPrivateNetworks = []ServerMetadataPrivateNetwork

type Metadata struct {
	ID          string `json:"id,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	PrivateIpv4 string `json:"ip,omitempty"`
	PublicIpv4  string `json:"public_ip,omitempty"`
}

func GetMetadata() (*Metadata, error) {
	var (
		metadata ServerMetadata
		networks ServerMetadataPrivateNetworks
	)

	r := tea.HttpGet(context.TODO(), fmt.Sprint(hcloudMetadataApiBase)).ToYAML(&metadata)
	if r.Err != nil {
		return nil, r.Err
	}
	r = tea.HttpGet(context.TODO(), fmt.Sprint(hcloudMetadataApiBase, "/private-networks")).ToYAML(&networks)
	if r.Err != nil {
		return nil, r.Err
	}

	return &Metadata{
		ID:          metadata.InstanceID,
		Hostname:    metadata.Hostname,
		PrivateIpv4: networks[0].Ip,
		PublicIpv4:  metadata.PublicIpv4,
	}, nil
}

func WaitForIp() bool {
	ch := make(chan bool)
	go func() {
		ready := false

		// wait up to 300s till you receive IP
		for i := 0; i < 30; i++ {
			log.Printf("Check #%d\n", i+1)
			if metadata, err := GetMetadata(); err == nil {
				if len(metadata.PrivateIpv4) > 0 {
					ready = true
					ch <- ready
					close(ch)
					break
				}
			}

			time.Sleep(10 * time.Second)
		}

		if !ready {
			ch <- false
			close(ch)
		}
	}()

	return <-ch
}
