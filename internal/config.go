package replicant

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
)

var allowedModes = []string{"highest", "higher", "semver", "all", ""}

type Config struct {
	Mode   string
	Images map[string]*ImageConfig
}

type ImageConfig struct {
	SourceRepository      string `yaml:"source"`
	DestinationRepository string `yaml:"destination"`
	Mode                  string `yaml:"mode"`
	Compatibility         string `yaml:"compatibility"`
	ReplaceTag            bool   `yaml:"replace-tag"`
	PinnedMajor           string `yaml:"pinned-major"` // string, to be able to distinguish between empty and major version 0.
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
	// Validate and/or set default mode.
	if len(config.Mode) == 0 {
		config.Mode = "highest"
	} else if !stringInSlice(config.Mode, allowedModes) {
		log.Fatalf("default mirroring mode %s not recognized", config.Mode)
	}

	// Validate config per image.
	for _, imageConfig := range config.Images {
		// Validate/set Mode
		if imageConfig.Mode == "" {
			imageConfig.Mode = config.Mode
		} else if !stringInSlice(imageConfig.Mode, allowedModes) {
			log.Fatalf("image specific mirroring mode %s not recognized", imageConfig.Mode)
		}

		// PinnedMajor should be an integer.
		if imageConfig.PinnedMajor != "" {
			_, err := strconv.Atoi(imageConfig.PinnedMajor)
			if err != nil {
				log.Fatal(err)
			}
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
