package ctors

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	envVarName    = "PUSHAAS_ENV"
	configVarName = "PUSHAAS_CONFIG"
)

func setupDefaults(config *viper.Viper) {
	/*
		mongodb
	*/
	config.SetDefault("mongodb.database", "pushaas")

	/*
		server
	*/
	config.SetDefault("server.port", "9000")
}

func NewViper() (*viper.Viper, error) {
	env := os.Getenv(envVarName)
	if env == "" {
		return nil, fmt.Errorf("you forgot to pass the %s environment variable", envVarName)
	}
	fmt.Println("env:", env)

	config := viper.New()
	setupDefaults(config)

	filepath := os.Getenv(configVarName)
	if filepath == "" {
		filepath = fmt.Sprintf("./config/%s.yml", env)
	}

	config.SetConfigFile(filepath)
	err := config.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %s", filepath)
	}
	fmt.Println("loading config from:", filepath)

	replacer := strings.NewReplacer(".", "__")
	config.SetEnvKeyReplacer(replacer)
	config.SetEnvPrefix("pushaas")
	config.AutomaticEnv()

	config.Set("env", env)

	return config, nil
}
