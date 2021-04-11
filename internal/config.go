package replicant

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

var defaultMode = "highest"
var allowedModes = []string{"highest", "higher", "all", "semver"}

type Config struct {
	Images map[string]*ImageConfig
}

type ImageConfig struct {
	UpstreamRepository   string `yaml:"upstream"`
	DownstreamRepository string `yaml:"downstream"`
	Mode                 string `yaml:"mode"`
}

func ReadConfig(configFile string) *Config {
	file, err := os.Open(configFile)
	if err != nil {
		log.Error(err)
	}

	config := &Config{}
	err = yaml.NewDecoder(file).Decode(config)
	if err != nil {
		log.Error(err)
	}

	validateConfig(config)
	return config
}

func validateConfig(config *Config) {
	for _, imageConfig := range config.Images {
		if imageConfig.Mode == "" {
			// No mode specified, use default
			imageConfig.Mode = defaultMode
			continue
		}
		if !stringInSlice(imageConfig.Mode, allowedModes) {
			log.Fatalf("mode %s not recognized", imageConfig.Mode)
		}
	}
}

func stringInSlice(s string, xs []string) bool {
	for _, x := range xs {
		if s == x {
			return true
		}
	}
	return false
}
