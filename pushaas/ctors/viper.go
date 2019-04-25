package ctors

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	envVarName = "ENV"
)

func setupDefaults(config *viper.Viper) {
	config.SetDefault("server.port", "9000")

	config.SetDefault("mongodb.database", "pushaas")
}

func NewViper() (*viper.Viper, error) {
	env := os.Getenv(envVarName)
	if env == "" {
		return nil, fmt.Errorf("you forgot to pass the %s environment variable", envVarName)
	}

	fmt.Println("env:", env)

	config := viper.New()
	setupDefaults(config)

	filepath := fmt.Sprintf("./config/%s.yml", env)
	config.SetConfigFile(filepath)
	err := config.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("config file doesn't exist: %s", filepath)
	}

	replacer := strings.NewReplacer(".", "__")
	config.SetEnvKeyReplacer(replacer)
	config.SetEnvPrefix("pushaas")
	config.AutomaticEnv()

	config.Set("env", env)

	return config, nil
}
