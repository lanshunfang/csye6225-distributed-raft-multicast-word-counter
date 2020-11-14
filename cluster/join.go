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

	maxAttempt := 10

	membership := GetMembership()

	for ; maxAttempt > 0; maxAttempt-- {
		_, ok := membership.Members[MyNodeID]
		if !ok {
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

	if !IsIAmLeader() {
		return
	}

	membership := GetMembership()
	membership.AddNewMember(nodeID, ip)

}

func init() {
	fmt.Println("[INFO] Init cluster")
	ListenMulticast(
		multicast.MulticastTopics["JOIN_GROUP"],
		func(nodeID string, ip string, UDPAddr *net.UDPAddr) {
			leaderAllowJoinGroup(nodeID, ip)
		},
	)
}
