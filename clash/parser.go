package clash

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConfigYAML struct {

	// EC is short for ExternalController
	// eg: 127.0.0.1:9090
	EC string `yaml:"external-controller"`

	// eg: http://127.0.0.1:7890
	MP string `yaml:"mixed-port"`

	// Secret is random UUID
	Secret string `yaml:"secret"`
}

type Config struct {
	ExternalController string
	Proxy              string
	Secret             string
}

var C Config

func ReadConfig() Config {
	var rawYaml ConfigYAML
	dir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	// ~/.config/clash/config.yaml
	path := filepath.Join(dir, ".config", "clash", "config.yaml")

	config, err := ioutil.ReadFile(path)

	if err != nil {
		path = filepath.Join(dir, "scoop", "apps", "clash-verge", "current", ".config",
			"clash-verge", "config.yaml")
		config, err = ioutil.ReadFile(path)

		if err != nil {
			log.Fatal(err)
		}
	}

	err = yaml.Unmarshal(config, &rawYaml)

	if err != nil {
		log.Fatal(err)
	}

	return Config{
		ExternalController: "http://" + rawYaml.EC,
		Secret:             "Bearer " + rawYaml.Secret,
		Proxy:              "http://127.0.0.1:" + rawYaml.MP,
	}

}
