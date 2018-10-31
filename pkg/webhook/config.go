package webhook

import (
	"io/ioutil"
	"crypto/sha256"
	"github.com/ghodss/yaml"
)

func Load(file string, c interface{})  {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		log.Warnf("Failed to parse %s", string(data))
	}

	log.Debugf("New configuration: sha256sum %x", sha256.Sum256(data))
	log.Debugf("Config: %v", c)
}
