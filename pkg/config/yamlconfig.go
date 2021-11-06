package config

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type YamlConfig struct {
	File string
}

func LoadYamlConfig(file string, typeName string, elements interface{}) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil
	}
	content, err := ioutil.ReadFile(file)
	if err != nil {
		// element couldn't be loaded
		return err
	}
	structured := make(map[string]interface{})
	err = yaml.Unmarshal(content, &structured)

	if subList, found := structured[typeName]; found && err == nil {
		items, err := yaml.Marshal(subList)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(items, elements)
		if err != nil {
			return err
		}
		err = SaveYamlConfig(file, elements)
		if err != nil {
			return err
		}
		return nil
	}

	bytes := []byte{}
	err = yaml.Unmarshal(content, &bytes)
	if err == nil {
		fmt.Println(string(bytes))
	}

	switch o := elements.(type) {
	case []uint8:
		fmt.Println(string(o))
	}

	err = yaml.Unmarshal(content, elements)
	if err != nil {
		return errors.Wrap(err, "File "+file+" is not a valid yaml")
	}
	return nil
}

func SaveYamlConfig(file string, elements interface{}) error {
	content, err := yaml.Marshal(elements)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, content, 0600)
}

func ChangeKey(file string, key string, value string) (err error) {
	var content []byte
	if _, err := os.Stat(file); err == nil {
		content, err = ioutil.ReadFile(file)
		if err != nil {
			return err
		}
	}
	structured := make(map[string]interface{})
	err = yaml.Unmarshal(content, &structured)
	if err != nil {
		return err
	}
	structured[key] = value

	out, err := yaml.Marshal(structured)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, out, 0644)
}
