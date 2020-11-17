package cluster

import (
	"fmt"
	"net"
	"time"
	"wordcounter/config"
	"wordcounter/multicast"
)

// JoinGroup ...
// Join Cluster group
func joinGroup() {

	fmt.Printf("[INFO] Join Group at %s\n", config.Envs["ENV_MULTICAST_GROUP"])

	membership := GetMembership()

	for maxAttempt := 5; maxAttempt > 0; maxAttempt-- {

		if len(membership.Members) < 2 && !isIAmLeader() {
			fmt.Printf("[INFO] Start to join group. Attempts left %v\n", maxAttempt)
			requestJoinGroup()
		} else {
			fmt.Printf("[INFO] Already in the group Or I am a leader now. My IP %s\n", *getMyself().IP)
			return
		}

		time.Sleep(10 * time.Second)
	}

	fmt.Print("[ERROR] Unable to join Group after max attempts")

}

func requestJoinGroup() {
	SendMulticast(multicast.MulticastTopics["JOIN_GROUP"], MyNodeID)
}

func leaderAllowJoinGroup(senderNodeID, senderIP string) {

	fmt.Printf("[INFO] Check If allow to join the group from %s. My IP: %s\n", senderIP, *getMyself().IP)

	if !isIAmLeader() {
		fmt.Printf("[INFO] Only Leader could allow new member join. I am not a leader. My IP: %s\n", *getMyself().IP)
		return
	}

	if senderNodeID == getMyself().ID {
		return
	}

	fmt.Printf("[INFO] Allow a new IP %s to join my group. My IP: %s \n", senderIP, *getMyself().IP)

	membership := GetMembership()

	addNewMember(membership, senderNodeID, senderIP)

	syncMembershipToFollowers(membership)

	syncLog()

}

// StartJoinGroupService ...
// Join cluster via Multicast IP
func StartJoinGroupService() {
	fmt.Println("[INFO] StartJoinGroupService")
	ListenMulticast(
		multicast.MulticastTopics["JOIN_GROUP"],
		func(senderNodeID string, senderIP string, UDPAddr *net.UDPAddr) {
			leaderAllowJoinGroup(senderNodeID, senderIP)
		},
	)
	go func() {
		joinGroup()
	}()
}
