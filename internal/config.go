package replicant

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Images map[string]ImageConfig
}

type ImageConfig struct {
	UpstreamRepository   string `yaml:"upstream"`
	DownstreamRepository string `yaml:"downstream"`
	TagPrefix            string `yaml:"prefix"`
}

func ReadConfig(configFile string) Config {
	file, err := os.Open(configFile)
	if err != nil {
		log.Error(err)
	}

	config := Config{}
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Error(err)
	}

	return config
}
