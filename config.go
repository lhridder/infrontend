package infrontend

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

var GlobalConfig Config

const Version = "0.1-beta"

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

type pageData struct {
	Year    string
	Version string
	Title   string
	Script  string
	User    User
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
