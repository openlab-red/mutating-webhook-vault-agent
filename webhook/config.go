package webhook

import (
	"io/ioutil"
	"crypto/sha256"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

func LoadConfig(configFile string) (*Config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"sha": sha256.Sum256(data),
	}).Infof("New configuration: sha256sum")

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
