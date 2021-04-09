package config

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

var configFileFormats = map[string]bool {
	".yml": true,
	".yaml": true,
}

type AppConfig struct {
	Protocol        string          `yaml:"protocol"`
	Basepath        string          `yaml:"basepath"`
	Port            int64           `yaml:"port"`
	Host            string          `yaml:"host"`
	AppLogger       LogConfig       `yaml:"logConfig"`
	CouchBaseConfig CouchBaseConfig `yaml:"couchbaseConfig"`
}

type LogConfig struct {
	// Type of logger to use - can be changed to any new logger added
	Logger 		string   `yaml:"logger"`
	Level 		string   `yaml:"level"`
	Encoding 	string   `yaml:"encoding"`
	OutputPaths []string `yaml:"outputPaths"`
	ErrorPaths  []string `yaml:"errorPaths"`
}

type CouchBaseConfig struct {
	Host string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Bucket string `yaml:"bucket"` // We are working with a single bucket in the db instance
	DocumentKey string `yaml:"documentKey"` // This is the json document collection we will work with
	Retries int `yml:"retries"`
}

func ReadConfig(filename string) (*AppConfig, error) {

	var config AppConfig
	fileExtension := filepath.Ext(filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read config file")
	}
	if _, ok := configFileFormats[fileExtension]; ok {
		//only handle yaml configs for now
		if fileExtension == ".yml" || fileExtension == ".yaml" {
			err = yaml.Unmarshal(file, &config)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to unmarshal yaml file")
			}
		}
	} else {
		err := fmt.Errorf("Unrecognized file type: %v\n", fileExtension)
		return nil, errors.Wrap(err, "file type")
	}
	return &config, nil
}
