package replicant

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

var allowedModes = []string{"highest", "higher", "semver", "all", ""}

type Config struct {
	Images map[string]*ImageConfig
}

type ImageConfig struct {
	SourceRepository      string `yaml:"source"`
	DestinationRepository string `yaml:"destination"`
	Mode                  string `yaml:"mode"`
	AllowPrerelease       bool   `yaml:"allow-prerelease"`
	ReplaceTag            bool   `yaml:"replace-tag"`
}

func ReadConfig(configFile string) *Config {
	file, err := os.Open(configFile)
	if err != nil {
		log.Error(err)
	}

	config := &Config{}
	err = yaml.NewDecoder(file).Decode(config) //TODO: use viper for this too?
	if err != nil {
		log.Error(err)
	}

	validateConfig(config)
	return config
}

func validateConfig(config *Config) {
	for _, imageConfig := range config.Images {
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
