package ctors

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-siris/siris/core/errors"
	"github.com/spf13/viper"
)

const (
	envVarName    = "PUSHAAS_ENV"
	configVarName = "PUSHAAS_CONFIG"
)

const defaultEnv = "local"

var envs = map[string]struct{}{
	defaultEnv: {},
	"prod": {},
}

func getEnvVariable() (string, error) {
	env := os.Getenv(envVarName)
	if env == "" {
		fmt.Println("[config] env variable not defined, falling back to default:", defaultEnv)
		env = defaultEnv
		return env, nil
	}

	if _, ok := envs[env]; !ok {
		return "", errors.New(fmt.Sprintf("you passed %s environment variable with an invalid value", envVarName))
	}

	fmt.Println("[config] env:", env)
	return env, nil
}

func setupFromConfigurationFile(config *viper.Viper, env string) error {
	// try to use custom config file, or falls back to file corresponding to env
	filepath := os.Getenv(configVarName)
	if filepath == "" {
		filepath = fmt.Sprintf("./config/%s.yml", env)
	}

	config.SetConfigFile(filepath)
	if err := config.ReadInConfig(); err != nil {
		if env == defaultEnv {
			fmt.Printf("[config] no config file found for default env in %s, using default config from code\n", filepath)
			return nil
		}
		return errors.New(fmt.Sprintf("error loading config file: %s", filepath))
	}

	fmt.Println("[config] loaded config from file:", filepath)
	return nil
}

func setupFromDefaults(config *viper.Viper, env string) {
	config.Set("env", env)

	// api
	config.SetDefault("api.enable_auth", true)
	config.SetDefault("api.basic_auth_user", "tsuru")
	config.SetDefault("api.basic_auth_password", "abc123")
	config.SetDefault("api.statics_path", "./client/build")

	// provisioner
	config.SetDefault("provisioner.provider", "ecs")

	// provisioner - ecs
	config.SetDefault("provisioner.ecs.region", "us-east-1")
	config.SetDefault("provisioner.ecs.cluster", "pushaas-cluster")
	config.SetDefault("provisioner.ecs.logs-group", "/ecs/pushaas")
	config.SetDefault("provisioner.ecs.logs-stream-prefix", "ecs")

	config.SetDefault("provisioner.ecs.image-push-api", "rafaeleyng/push-api:latest") // TODO pass actual tag
	config.SetDefault("provisioner.ecs.image-push-agent", "rafaeleyng/push-agent:latest") // TODO pass actual tag
	config.SetDefault("provisioner.ecs.image-push-stream", "rafaeleyng/push-stream:latest") // TODO pass actual tag

	config.SetDefault("provisioner.ecs.attempts", 10)
	config.SetDefault("provisioner.ecs.interval", time.Duration(5 * time.Second))

	// redis
	config.SetDefault("redis.url", "redis://localhost:6379")
	config.SetDefault("redis.db.instance.prefix", "instance")
	config.SetDefault("redis.db.bind-app.prefix", "bind-app")
	config.SetDefault("redis.db.bind-unit.prefix", "bind-unit")
	config.SetDefault("redis.pubsub.tasks.provision", "provision")
	config.SetDefault("redis.pubsub.tasks.deprovision", "deprovision")

	// server
	config.SetDefault("server.port", "9000")

	// workers
	config.SetDefault("workers.enabled", true)
	config.SetDefault("workers.provision.enabled", true)
}

func setupFromEnvironment(config *viper.Viper) {
	replacer := strings.NewReplacer(".", "__")
	config.SetEnvKeyReplacer(replacer)
	config.SetEnvPrefix("pushaas")
	config.AutomaticEnv()
}

func NewViper() (*viper.Viper, error) {
	env, err := getEnvVariable()
	if err != nil {
		return nil, err
	}

	config := viper.New()
	setupFromDefaults(config, env)
	if err := setupFromConfigurationFile(config, env); err != nil {
		return nil, err
	}
	setupFromEnvironment(config)

	return config, nil
}
