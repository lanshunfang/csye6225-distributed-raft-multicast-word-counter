package config

import "os"

var multicastGroup string = "239.0.0.1:10000"

// GetEnvMulticastGroup ...
func GetEnvMulticastGroup() string {
	env := os.Getenv("ENV_MULTICAST_GROUP")
	if env == "" {
		env = multicastGroup
	}
	return env
}

var MulticastTopics = map[string]string{
	"JOIN_GROUP": "JOIN_GROUP",
}
