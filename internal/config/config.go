package config

import (
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type YamlURL struct {
	*url.URL
}

func (j *YamlURL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string

	err := unmarshal(&str)
	if err != nil {
		return fmt.Errorf("config.UnmarshalYAML unmarshal failed: %w", err)
	}

	j.URL, err = url.Parse(str)
	if err != nil {
		return fmt.Errorf("config.UnmarshalYAML url.Parse failed: %w", err)
	}

	return nil
}

func (j *YamlURL) MarshalYAML() (interface{}, error) {
	return j.String(), nil
}

func ParseConfig(filename string, conf any) error {
	confFile, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("config.ParseConfig ReadFile failed: %w", err)
	}

	err = yaml.Unmarshal(confFile, conf)
	if err != nil {
		return fmt.Errorf("config.ParseConfig yaml Unmarshal failed: %w", err)
	}

	return nil
}
