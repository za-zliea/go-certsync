package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type Config interface {
	Generate() interface{}
	Empty(meta interface{}) bool
}

func WriteConfig(configFilePath string, meta interface{}) error {
	if configFilePath == "" {
		return errors.New("-c: path of config file cannot empty")
	}

	configFileContent, configErr := yaml.Marshal(meta)
	if configErr != nil {
		return errors.New(fmt.Sprintf("format config error: %v", configErr))
	}

	configFile, configErr := os.OpenFile(configFilePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	defer configFile.Close()
	if configErr != nil {
		return errors.New(fmt.Sprintf("open config file error: %v", configErr))
	}

	_, configErr = configFile.Write(configFileContent)
	if configErr != nil {
		return errors.New(fmt.Sprintf("write config file error: %v", configErr))
	}

	return nil
}

func ReadConfig(configFilePath string, meta interface{}) error {
	if configFilePath == "" {
		return errors.New("-c: path of config file cannot empty")
	}

	configFile, configErr := os.OpenFile(configFilePath, os.O_RDONLY, 0644)
	defer configFile.Close()
	if configErr != nil {
		return errors.New(fmt.Sprintf("open config file error: %v", configErr))
	}

	configFileContent, configErr := io.ReadAll(configFile)
	if configErr != nil {
		return errors.New(fmt.Sprintf("read config file error: %v", configErr))
	}

	configErr = yaml.Unmarshal(configFileContent, meta)
	if configErr != nil {
		return errors.New(fmt.Sprintf("parse config file error: %v", configErr))
	}

	return nil
}
