package configuration

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
)
type Configuration struct {
	Collect Host `toml:"collect"`
	Output Host `toml"output"`
}

type Host struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
	UserName string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

func LoadConfiguration(configFile string) (*Configuration, error) {
	config := &Configuration{
		Host{
			Host: "localhost",
			Port: "8086",
			UserName: "root",
			Password: "root",
		},
		Host{
			Host: "localhost",
			Port: "8086",
			UserName: "root",
			Password: "root",
			Database: "influxdb",
		},
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	if _, err2 := toml.Decode(string(data), config); err != nil {
		fmt.Printf("Error: %s\n", string(data))
		return config, err2
	}

	return config, nil
}
