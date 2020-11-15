package main

import (
	"sync"
	"wordcounter/cluster"
	"wordcounter/distributetask"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	cluster.StartMembershipService()
	cluster.StartRaftLogService()
	cluster.StartJoinGroupService()
	cluster.StartLeaderElectionService()
	distributetask.StartWordCountService()

	wg.Wait()

}
