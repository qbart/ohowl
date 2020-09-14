package tea

import (
	"encoding/json"
	"log"
)

func MustJson(o interface{}) []byte {
	b, err := json.Marshal(o)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
