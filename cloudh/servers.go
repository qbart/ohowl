package cloudh

import (
	"github.com/go-resty/resty/v2"
)

const serversApi = "https://api.hetzner.cloud/v1/servers"

type ServerFilter struct {
	ByLabel        LabelSelector
	ExpectedAmount int
}

type Server struct {
	Meta `json:"meta,omitempty"`
}

func GetServers(token string, filter ServerFilter) ([]byte, error) {
	params := make(map[string]string, 0)
	params["label_selector"] = filter.ByLabel.String()

	client := resty.New()
	resp, err := client.R().
		SetQueryParams(params).
		SetAuthToken(token).
		Get(serversApi)

	return resp.Body(), err
}
