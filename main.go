package main

import (
	"sync"
	"wordcounter/cluster"
	"wordcounter/distributetask"
	"wordcounter/web"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	cluster.StartMembershipService()
	cluster.StartRaftLogService()
	cluster.StartJoinGroupService()
	cluster.StartLeaderElectionService()
	distributetask.StartWordCountService()
	web.ServeWeb()

	wg.Wait()

}
