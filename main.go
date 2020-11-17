package main

import (
	"sync"
	"wordcounter/cluster"
	"wordcounter/distributetask"
	"wordcounter/rpc"
	"wordcounter/web"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	rpc.StartRPCService()
	cluster.StartMembershipService()
	cluster.StartRaftLogService()
	cluster.StartJoinGroupService()
	cluster.StartLeaderElectionService()
	distributetask.StartWordCountService()
	web.ServeWeb()

	wg.Wait()

}
