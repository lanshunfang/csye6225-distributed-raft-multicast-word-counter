package distributed

import (
	"wordcounter/cluster"
)

func main() {

	// if os.Getenv("ENV_MULTICAST_GROUP") == "" {
	// 	os.Setenv("ENV_MULTICAST_GROUP", defaultmulticast)
	// }

	cluster.LoadLog()
	cluster.JoinGroup()
	cluster.ElectMeIfLeaderDie()

}
