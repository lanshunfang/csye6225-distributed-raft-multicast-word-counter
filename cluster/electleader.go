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
var heartbeatFrequency time.Duration = 300 * time.Millisecond

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

// Periodically check if I should be leader
func electMeIfLeaderDie() {

	go func() {
		for {

			now := time.Now()

			if now.Sub(lastHeartbeatTime) > maxTimeout {

				raftLikeLogger := GetLogger()

				SendMulticast(
					multicast.MulticastTopics["ELECT_ME_AS_LEADER"],
					multicast.JoinFields(
						strconv.Itoa(getNewTerm()),
						MyNodeID,
						strconv.Itoa(raftLikeLogger.getCachedLatestOplog()),
					),
				)
			}

			time.Sleep(300 * time.Microsecond)

		}
	}()

}

func voteLeader(requestNewTermStr, requestLeaderNodeId, requestLeaderLogOffsetStr, ip string) {

	voteDecision := 1

	newTerm, err := strconv.Atoi(requestNewTermStr)

	if err != nil {
		fmt.Printf("[WARN] Invalid requested voting term: %s; Err: %s", termStr, err)
		return
	}

	requestLeaderLogOffset, err := strconv.Atoi(requestLeaderLogOffsetStr)

	if err != nil {
		fmt.Printf("[WARN] Invalid requested voting LogOffset: %s; Err: %s", requestLeaderLogOffsetStr, err)
		return
	}

	votedLeaderID, ok := myVote[newTerm]

	raftLikeLogger := GetLogger()

	isAlreadyVote := ok
	if isAlreadyVote {
		if votedLeaderID != requestLeaderNodeId || requestLeaderLogOffset < raftLikeLogger.getCachedLatestOplog() {
			voteDecision = 0
		}
	} else if requestLeaderNodeId != MyNodeID {
		if err != nil {
			voteDecision = 0
		}

		myTerm := getMyTerm()
		raftLikeLogger := GetLogger()

		if newTerm < myTerm || requestLeaderLogOffset < raftLikeLogger.getCachedLatestOplog() {
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

func updateLeaderHeartbeat(leaderNodeID, ip string) {
	if !IsLeader(leaderNodeID, ip) {
		fmt.Println("[WARN] Receive a leader heartbeat that is not from current leader")
		leader := getLeader()
		fmt.Printf(
			"[WARN] RecvHeatbeatNodeId %s, RecvHeatbeatNodeIP %s| currentLeaderNodeId %s, currentLeaderNodeIP %s",
			leaderNodeID,
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
		func(leaderNodeID string, ip string, UDPAddr *net.UDPAddr) {

			updateLeaderHeartbeat(leaderNodeID, ip)
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
			voteLeader(msgSplit[0], msgSplit[1], msgSplit[2], ip)
		},
	)
}

func StartLeaderElectionService() {
	listenLeaderElection()
	leaderSendHeartBeat()
	electMeIfLeaderDie()
}
