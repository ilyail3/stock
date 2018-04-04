package stock

import (
	"os"
	"encoding/json"
)

type ApplicationConfig interface {
	GetApiKey() string
}

type jsonConfig struct {
	ApiKey string
}

func(j *jsonConfig) GetApiKey() string{
	return j.ApiKey
}

func ReadConfig(fileName string) (ApplicationConfig,error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	conf := &jsonConfig{}
	decoder := json.NewDecoder(f)
	err = decoder.Decode(conf)

	if err != nil {
		return nil, err
	}

	return conf, nil
}