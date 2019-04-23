package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	envVarName = "ENV"
)

var Config *viper.Viper

func setupDefaults() {
	Config.SetDefault("mongodb.database", "pushaas")
}

func SetupConfig() {
	env := os.Getenv(envVarName)
	if env == "" {
		env = "local"
	}

	fmt.Println("env", env)

	Config = viper.New()
	setupDefaults()

	Config.SetConfigFile(fmt.Sprintf("./config/%s.yml", env))
	err := Config.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
