package config

import (
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type YamlURL struct {
	*url.URL
}

func (j *YamlURL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	j.URL, err = url.Parse(s)
	return err
}

func (j *YamlURL) MarshalYAML() (interface{}, error) {
	return j.String(), nil
}

func ParseConfig(filename string, conf any) error {
	confFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(confFile, conf)
	return err
}
