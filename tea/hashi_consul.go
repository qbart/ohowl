package tea

import (
	consulapi "github.com/hashicorp/consul/api"
)

type Consul struct {
	client *consulapi.Client
}

func NewConsul() (*Consul, error) {
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &Consul{client: client}, nil
}

func (c *Consul) KV() *consulapi.KV {
	return c.client.KV()
}

func (c *Consul) Register(id string, port int, tags []string, meta map[string]string) error {
	reg := consulapi.AgentServiceRegistration{
		ID:   id,
		Name: id,
		Port: port,
		Tags: tags,
		Meta: meta,
		// Check: &consulapi.AgentServiceCheck{
		// 	Name:          "...",
		// 	Interval:      "10s",
		// 	Timeout:       "5s",
		// 	HTTP:          "http://ip:port/health",
		// 	Method:        "GET",
		// 	TLSSkipVerify: false,
		// },
	}
	return c.client.Agent().ServiceRegister(&reg)
}

func (c *Consul) Deregister(id string) error {
	return c.client.Agent().ServiceDeregister(id)
}
