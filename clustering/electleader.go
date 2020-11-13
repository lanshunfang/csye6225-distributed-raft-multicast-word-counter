package clustering

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
	"wordcounter/multicast"
)

var lastHeartbeatTime time.Time = time.Now()
var maxTimeout time.Duration = time.Duration(200*(1+200*rand.Float32())) * time.Microsecond
var heartbeatFrequency time.Duration = time.Duration(300) * time.Millisecond

func getMyTerm() int {
	myself := getMyself()
	return myself.Term
}

func checkAndElectMyself() {

	now := time.Now()

	if now.Sub(lastHeartbeatTime) < maxTimeout {
		return
	}

	SendMulticast(
		multicast.MulticastTopics["ELECT_LEADER"],
		multicast.JoinFields(
			MyNodeID,
			strconv.Itoa(getMyTerm()),
		),
	)
}

func voteLeader(nodeId, termStr, ip string) {
	voteDecision := true
	requestTerm, err := strconv.Atoi(termStr)
	if err != nil {
		voteDecision = false
	}

	myTerm := getMyTerm()
	if requestTerm < myTerm {
		voteDecision = false
	}

	VoteResponse(nodeId, voteDecision)
}

func VoteResponse(requestNodeId string, voteResult bool) {

}

func leaderSendHeartBeat() {
	go func() {
		for {
			if IsIAmLeader() {
				SendMulticast(multicast.MulticastTopics["LEADER_HEARTBEAT"], MyNodeID)
			}
			time.Sleep(heartbeatFrequency)
		}

	}()

}

func updateLeaderHeartbeat(leaderNodeId, ip string) {
	if !IsLeader(leaderNodeId, ip) {
		fmt.Println("[WARN] Receive a leader heartbeat that is not from current leader")
		leader := getLeader()
		fmt.Printf(
			"[WARN] RecvHeatbeatNodeId %s, RecvHeatbeatNodeIP %s| currentLeaderNodeId %s, currentLeaderNodeIP %s",
			leaderNodeId,
			ip,
			leader.ID,
			leader.IP,
		)
		return
	}

	lastHeartbeatTime = time.Now()
}

func listenLeaderHeartbeat() {
	ListenMulticast(
		multicast.MulticastTopics["LEADER_HEARTBEAT"],
		func(leaderNodeId string, ip string, UDPAddr *net.UDPAddr) {

			updateLeaderHeartbeat(leaderNodeId, ip)
		},
	)
}

func listenLeaderElection() {
	ListenMulticast(
		multicast.MulticastTopics["ELECT_LEADER"],
		func(msg string, ip string, UDPAddr *net.UDPAddr) {
			msgSplit := multicast.GetFields(msg)
			voteLeader(msgSplit[0], msgSplit[1], ip)
		},
	)
}

func init() {
	listenLeaderElection()
	leaderSendHeartBeat()
}
