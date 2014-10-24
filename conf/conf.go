package conf

import (
	"encoding/json"
	"os"
)

func LoadConfig(filename string, v interface{}) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(r)
	err = decoder.Decode(v)
	if err != nil {
		return err
	}
	return nil
}
