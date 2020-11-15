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
			fmt.Printf("[INFO] Start to join group. Attempts left %v\n", maxAttempt)
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

func leaderAllowJoinGroup(senderNodeID, senderIP string) {

	if !isIAmLeader() {
		return
	}

	membership := GetMembership()
	addNewMember(membership, senderNodeID, senderIP)
	syncLog()

}

func StartJoinGroupService() {
	ListenMulticast(
		multicast.MulticastTopics["JOIN_GROUP"],
		func(senderNodeID string, senderIP string, UDPAddr *net.UDPAddr) {
			leaderAllowJoinGroup(senderNodeID, senderIP)
		},
	)
	go JoinGroup()
}
