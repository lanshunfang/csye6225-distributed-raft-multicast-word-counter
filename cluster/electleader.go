package cluster

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

// map[term]votedLeaderID
var myVote = make(map[int]string)

// map[term]map[voterIP]voterResult
type voterVote struct {
	voterIP     string
	voterNodeID string
	result      int
}
type LeaderElectionResult map[int]map[string]voterVote

var myLeaderElectionResult = make(LeaderElectionResult)

func getMyTerm() int {
	myself := getMyself()
	return myself.Term
}

func getNewTerm() int {
	return getMyTerm() + 1
}

func checkAndElectMyself() {

	now := time.Now()

	if now.Sub(lastHeartbeatTime) < maxTimeout {
		return
	}

	SendMulticast(
		multicast.MulticastTopics["ELECT_ME_AS_LEADER"],
		multicast.JoinFields(
			strconv.Itoa(getNewTerm()),
			MyNodeID,
		),
	)
}

func voteLeader(requestNewTermStr, requestLeaderNodeId, ip string) {

	voteDecision := 1

	newTerm, err := strconv.Atoi(requestNewTermStr)

	if err != nil {
		fmt.Printf("[WARN] Invalid requested voting term: %s; Err: %s", termStr, err)
		return
	}

	votedLeaderID, ok := myVote[newTerm]

	isAlreadyVote := ok
	if isAlreadyVote {
		if votedLeaderID != requestLeaderNodeId {
			voteDecision = 0
		}
	} else if requestLeaderNodeId != MyNodeID {
		if err != nil {
			voteDecision = 0
		}

		myTerm := getMyTerm()
		if newTerm < myTerm {
			voteDecision = 0
		}
	}

	if voteDecision == 1 {
		myVote[newTerm] = requestLeaderNodeId
	}

	voteResultSend(newTerm, requestLeaderNodeId, voteDecision)
}

func voteResultSend(requestTerm int, requestLeaderNodeId string, voteResult int) {
	SendMulticast(
		multicast.MulticastTopics["VOTE_LEADER"],
		multicast.JoinFields(
			strconv.Itoa(requestTerm),
			requestLeaderNodeId,
			strconv.Itoa(voteResult),
			MyNodeID,
		),
	)
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

func updateVoteResult(newTermStr, leaderID, result, voterNodeID, voterIP string) {
	newTermInt, err := strconv.Atoi(newTermStr)
	if err != nil || newTermInt != getNewTerm() || leaderID != MyNodeID {
		return
	}
	_, ok := myLeaderElectionResult[newTermInt]
	if !ok {
		myLeaderElectionResult[newTermInt] = make(map[string]voterVote)
	}

	resultInt, err := strconv.Atoi(result)

	myLeaderElectionResult[newTermInt][voterIP] = voterVote{
		voterIP:     voterIP,
		voterNodeID: voterNodeID,
		result:      resultInt,
	}

	if hasReceivedMajorityVote(newTermInt) {
		updateMeAsLeaderForNewTerm(newTermInt)
	}

}

func hasReceivedMajorityVote(newTerm int) bool {
	currentVote, ok := myLeaderElectionResult[newTerm]
	positiveVote := 0
	for _, v := range currentVote {
		positiveVote += v.result
	}
	if !ok {
		return false
	}

	return positiveVote > len(GetMembership().Members)/2

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

func listenLeaderVote() {
	ListenMulticast(
		multicast.MulticastTopics["VOTE_LEADER"],
		func(msg string, ip string, UDPAddr *net.UDPAddr) {
			msgSplit := multicast.GetFields(msg)
			updateVoteResult(msgSplit[0], msgSplit[1], msgSplit[2], msgSplit[3], ip)
		},
	)
}
func listenLeaderElection() {
	ListenMulticast(
		multicast.MulticastTopics["ELECT_ME_AS_LEADER"],
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
