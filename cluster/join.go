package cluster

import (
	"fmt"
	"net"
	"time"
	"wordcounter/multicast"
)

// JoinGroup ...
// Join Cluster group
func JoinGroup() {

	maxAttempt := 5

	membership := GetMembership()

	for ; maxAttempt > 0; maxAttempt-- {
		_, ok := membership.Members[MyNodeID]
		if !ok {
			fmt.Printf("[INFO] Start to join group. Attempts left %v", maxAttempt)
			requestJoinGroup()
		} else {
			return
		}

		time.Sleep(5 * time.Second)
	}

	fmt.Print("[ERROR] Unable to join Group after max attempts")

}

func requestJoinGroup() {
	SendMulticast(multicast.MulticastTopics["JOIN_GROUP"], MyNodeID)
}

func leaderAllowJoinGroup(nodeID, ip string) {

	if !isIAmLeader() {
		return
	}

	membership := GetMembership()
	addNewMember(membership, nodeID, ip)
	syncLog()

}

func StartJoinGroupService() {
	ListenMulticast(
		multicast.MulticastTopics["JOIN_GROUP"],
		func(nodeID string, ip string, UDPAddr *net.UDPAddr) {
			leaderAllowJoinGroup(nodeID, ip)
		},
	)
	JoinGroup()
}
