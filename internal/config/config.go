package config

import (
	"fmt"
	"log"
	"os"

	"github.com/pelletier/go-toml"
)

var (
	globalConfig GlobalConfig
)

// Global Configuration
type GlobalConfig struct {
	Server      Server `toml:"server"`
	RedisConfig Redis  `toml:"redis"`
	MQTTConfig MQTT `toml:"mqtt"`
}

// Redis Configuration
type Redis struct {
	Cert        string `toml:"redis_cert"`
	URL         string `toml:"redis_url"`
	IdleTimeout int    `toml:"redis_idle_timeout"`
	DBNum       int    `toml:"redis_db_num"`
}

// server configuration
type Server struct {
	Address      string `toml:"address"`
	ReadTimeOut  int    `toml:"read_time_out"`
	WriteTimeOut int    `toml:"write_time_out"`
}

type MQTT struct {
	MQTTBroker string `toml:"mqtt_broker"`
	Topic string `toml:"topic"`
}

// Setter method for GlobalConfig
func SetConfig(cfg GlobalConfig) {
	globalConfig = cfg
}

// Getter method for GlobalConfig
func GetConfig() GlobalConfig {
	return globalConfig
}

// Loading the values from default.toml and assigning them as part of GlobalConfig struct
func InitGlobalConfig() error {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)
	config, err := toml.LoadFile("config/defaults.toml")
	fmt.Println("Err : ", err)
	if err != nil {
		log.Printf("Error while loading defaults.toml file : %v ", err)
		return err
	}

	var appConfig GlobalConfig
	err = config.Unmarshal(&appConfig)
	if err != nil {
		log.Printf("Error while unmarshalling config : %v", err)
		return err
	}

	SetConfig(appConfig)
	return nil
}
