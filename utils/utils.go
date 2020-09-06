package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

func FromYaml(body []byte, o interface{}) {
	if err := yaml.Unmarshal(body, o); err != nil {
		log.Fatal(err)
	}
}

func Get(url string) ([]byte, error) {
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

func Json(o interface{}) []byte {
	b, err := json.Marshal(o)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
