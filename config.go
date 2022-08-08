package infrontend

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

var GlobalConfig Config

type Redis struct {
	Host string `yaml:"host"`
	Pass string `yaml:"pass"`
	DB   int    `yaml:"db"`
}

type Config struct {
	Redis  Redis
	Listen string `yaml:"listen"`
	Debug  bool   `yaml:"debug"`
}

func LoadGlobalConfig() error {
	log.Println("Loading config.yml")
	ymlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(ymlFile, &GlobalConfig)
	if err != nil {
		return err
	}
	return nil
}
