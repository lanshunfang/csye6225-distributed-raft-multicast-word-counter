package config

import (
	"os"
)

// Envs ...
// Env variables that could be override by environtable variables
var Envs = map[string]string{
	"ENV_MULTICAST_GROUP":  "239.0.0.1:10000",
	"ENV_HTTP_STATIC_PORT": "3000",
	"ENV_RPC_PORT":         "8081",
}

// RPCDef RPC definition
// Port is normally the RPC default server port
type RPCDef struct {
	Name string
	Port string
}

// HTTPRPCList ...
// All the RPC endpoints for initialization and RPC call
var HTTPRPCList = map[string]RPCDef{
	"WordCount": {
		Name: "WordCount",
		Port: "ENV_PORT_WORDCOUNT",
	},
	"WordCount.Count": {
		Name: "WordCount.Count",
		Port: "ENV_PORT_WORDCOUNT",
	},
	"WordCount.HTTPHandle": {
		Name: "WordCount.HTTPHandle",
		Port: "ENV_PORT_WORDCOUNT",
	},

	"Membership": {
		Name: "Membership",
		Port: "ENV_PORT_MEMBERSHIP",
	},
	"Membership.UpdateMembership": {
		Name: "Membership.UpdateMembership",
		Port: "ENV_PORT_MEMBERSHIP",
	},

	"RaftLikeLogger": {
		Name: "RaftLikeLogger",
		Port: "ENV_PORT_LOGGER",
	},
	"RaftLikeLogger.FollowerAppendLog": {
		Name: "RaftLikeLogger.FollowerAppendLog",
		Port: "ENV_PORT_LOGGER",
	},
	"RaftLikeLogger.LeaderAppendLog": {
		Name: "RaftLikeLogger.LeaderAppendLog",
		Port: "ENV_PORT_LOGGER",
	},
}

func initEnvs(defaultMap map[string]string) {
	for k := range defaultMap {
		env := os.Getenv(k)
		if env == "" {
			continue
		}

		defaultMap[k] = env

	}
}

func init() {
	initEnvs(Envs)
}
