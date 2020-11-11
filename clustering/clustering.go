package clustering

import (
	"fmt"
	"net"
	"wordcounter/config"
	"wordcounter/multicast"
)

type clusterMember struct {
	IsLeader  bool
	Term      int32
	IP        string
	LogOffset int64
}

// clusterMemberMap ...
var clusterMemberMap = struct {
	IsInCluster: bool
	Leader    clusterMember
	Followers map[string]clusterMember
}{
	IsInCluster: false
	Leader: clusterMember{},
	// map[ip]clusterMember
	Followers: make(map[string]clusterMember),
}

func joinGroup() {
	if !clusterMemberMap.IsInCluster {
		requestJoinGroup()
	}
}
func requestJoinGroup() {
	multicastAddr := config.GetEnvMulticastGroup()
	sender := multicast.GetSender("239.0.0.1:10000")
	sender.Send(messageTopic, messageSend)
}

func updateGroup(ip string) {
	clusterMemberMap.Followers[ip] = clusterMember{IP: ip}
}

func listenMulticast() {
	multicastAddr := config.GetEnvMulticastGroup()
	fmt.Println("[INFO] Listen UDP multicast for service discovery in " + multicastAddr)
	go multicast.Register(
		multicastAddr,
		func(topic string, payload string, ip string, source *net.UDPAddr) {
			fmt.Println("[INFO]Received topic: " + topic + " From IP " + ip)
			getMessageHandler(topic, payload)
		},
	)
}

func ListenMulticast(listener *multicast.MulticastListener) {
	
}

func getMessageHandler(topic, payload string) {
	handlerMap := map[string]func(string){
		"JOIN_GROUP": updateGroup,
	}

	handlerMap[topic](payload)
}

func init() {
	fmt.Println("[INFO] Init clustering")
	listenMulticast()
}
