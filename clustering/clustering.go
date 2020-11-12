package clustering

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
	"wordcounter/multicast"
)

type clusterMember struct {
	ID        string
	Term      int
	IP        string
	LogOffset int64
}

// MyID ...
// Node ID
var MyID string = strconv.Itoa(rand.Int())

// clusterMemberMap ...
var clusterMemberMap = struct {
	Leader  clusterMember
	Members map[string]clusterMember
}{
	Leader: clusterMember{},
	// map[nodeId]clusterMember
	Members: make(map[string]clusterMember),
}

func joinGroup() {

	maxAttempt := 10

	for ; maxAttempt > 0; maxAttempt-- {
		_, ok := clusterMemberMap.Members[MyID]
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
	SendMulticast(multicast.MulticastTopics["JOIN_GROUP"], MyID)
}

func isIAmLeader() bool {
	return clusterMemberMap.Leader.ID == MyID
}

func leaderAllowJoinGroup(nodeID, ip string) {

	if !isIAmLeader() {
		return
	}

	for key, node := range clusterMemberMap.Members {
		if node.IP == ip {
			delete(clusterMemberMap.Members, key)
		}
	}
	clusterMemberMap.Members[nodeID] = clusterMember{IP: ip, ID: nodeID}
}

func leaderSyncGroup() {
	if !isIAmLeader() {
		return
	}

	SendMulticast(multicast.MulticastTopics["SYNC_GROUP"], MyID)

}

func memberAllowSyncGroup(requestLeaderId, group string) {
	// require go-rpc package
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
