package clustering

import (
	"fmt"
	"net"
	"time"
	"wordcounter/multicast"
)

func joinGroup() {

	maxAttempt := 10

	membership := getMembership()

	for ; maxAttempt > 0; maxAttempt-- {
		_, ok := membership.Members[MyNodeID]
		if !ok {
			requestJoinGroup()
		} else {
			return
		}

		time.Sleep(5 * time.Second)
	}

	panic("[ERROR] Unable to join Group after max attempts")

}
func requestJoinGroup() {
	SendMulticast(multicast.MulticastTopics["JOIN_GROUP"], MyNodeID)
}

func leaderAllowJoinGroup(nodeID, ip string) {

	if !IsIAmLeader() {
		return
	}

	membership := getMembership()
	membership.AddNewMember(nodeID, ip)

}

func init() {
	fmt.Println("[INFO] Init clustering")
	ListenMulticast(
		multicast.MulticastTopics["JOIN_GROUP"],
		func(nodeID string, ip string, UDPAddr *net.UDPAddr) {
			leaderAllowJoinGroup(nodeID, ip)
		},
	)

	go joinGroup()
}
