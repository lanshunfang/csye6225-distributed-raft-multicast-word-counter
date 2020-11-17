package config

import (
	"os"
)

var Envs = map[string]string{
	"ENV_MULTICAST_GROUP":  "239.0.0.1:10000",
	"ENV_HTTP_STATIC_PORT": "3000",
	"ENV_RPC_PORT":         "8081",
}

// var EnvPorts = map[string]string{

// 	"ENV_PORT_MEMBERSHIP": "8080",

// 	"ENV_PORT_LOGGER":    "10002",
// 	"ENV_PORT_WORDCOUNT": "8081",
// }

type RPCDef struct {
	Name string
	Port string
}

var HttpRpcList = map[string]RPCDef{
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
	// "MembershipForRPC.ReportLeaderIP": {
	// 	Name: "MembershipForRPC.ReportLeaderIP",
	// 	Port: "ENV_PORT_MEMBERSHIP",
	// },

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
	for k, _ := range defaultMap {
		env := os.Getenv(k)
		if env == "" {
			continue
		}

		defaultMap[k] = env

	}
}

func init() {
	// initEnvs(EnvPorts)
	initEnvs(Envs)
}
