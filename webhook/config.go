package webhook

import (
	"io/ioutil"
	"crypto/sha256"
	"github.com/ghodss/yaml"
	"strings"
)

func LoadConfig(injectFile string) (*Config, error) {
	data, err := ioutil.ReadFile(injectFile)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		log.Warnf("Failed to parse injectFile %s", string(data))
		return nil, err
	}

	log.Infof("New configuration: sha256sum %x", sha256.Sum256(data))
	log.Infof("Template: |\n  %v", strings.Replace(c.Template, "\n", "\n  ", -1))

	return &c, nil
}
