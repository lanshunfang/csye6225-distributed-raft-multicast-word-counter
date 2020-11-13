package config

import (
	"os"
)

var Envs = map[string]string{
	"ENV_MULTICAST_GROUP": "239.0.0.1:10000",
}

var EnvPorts = map[string]string{
	"ENV_PORT_MEMBERSHIP_SYNC": "10001",
	"ENV_PORT_LOGGER_SYNC":     "10002",
	"ENV_PORT_TASK_WORDCOUNT":  "10003",
}

func initEnvs(defaultMap map[string]string) {
	for k, _ := range defaultMap {
		env := os.Getenv(k)
		if env == "" {
			continue
		}

		defaultMap[k] = env

	}
}

func init() {
	initEnvs(EnvPorts)
	initEnvs(Envs)
}
