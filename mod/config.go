package mod

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Targets []string          `yaml:"targets"`
	LDFlags map[string]string `yaml:"ldflags"`
}

func LoadConfig(appRoot string) (Config, error) {
	configPath := filepath.Join(appRoot, "buildpack.yml")
	var config struct {
		Go Config `yaml:"go"`
	}
	if _, err := os.Stat(configPath); err == nil {
		yamlFile, err := ioutil.ReadFile(configPath)
		if err != nil {
			return Config{}, err
		}
		err = yaml.Unmarshal(yamlFile, &config)
		if err != nil {
			return Config{}, err
		}
	}

	if buildTarget := os.Getenv("BP_GO_TARGETS"); buildTarget != "" {
		config.Go.Targets = strings.Split(buildTarget, ":")
	}

	return config.Go, nil
}
