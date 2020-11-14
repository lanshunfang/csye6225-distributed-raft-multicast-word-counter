package distributed

import (
	"wordcounter/cluster"
	"wordcounter/distributetask"
)

func main() {

	cluster.StartRaftLogService()
	cluster.StartMembershipService()
	cluster.StartJoinGroupService()
	cluster.StartLeaderElectionService()
	distributetask.StartWordCountService()

}
