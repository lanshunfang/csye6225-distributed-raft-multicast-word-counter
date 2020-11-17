package cluster

import (
	"fmt"
	"net"
	"strconv"
	"time"
	"wordcounter/multicast"
	"wordcounter/utils"
)

var initTime = time.Now()
var lastHeartbeatTime *time.Time = &initTime
var maxTimeout time.Duration = time.Duration(1000*(1+utils.RandWithSeed().Float32())) * time.Millisecond
var heartbeatFrequency time.Duration = 300 * time.Millisecond

// map[term]votedLeaderID
var myVote = make(map[int]string)

// map[term]map[voterIP]voterResult
type voterVote struct {
	leaderID    string
	leaderIP    string
	voterNodeID string
	voterIP     string
	result      int
}
type LeaderElectionResult map[int]map[string]voterVote

var myLeaderElectionResult = make(LeaderElectionResult)

func getMyTerm() int {
	myself := getMyself()
	return *myself.Term
}

func getNewTerm() int {
	return getMyTerm() + 1
}

// Periodically check if I should be leader
func electMeIfLeaderDie() {

	// wait for Join Group done
	time.Sleep(5 * time.Second)

	go func() {
		for {

			if isIAmLeader() {
				return
			}

			max := lastHeartbeatTime.Add(maxTimeout)

			if max.Before(time.Now()) {

				fmt.Print("[INFO] TIME")
				fmt.Print(max)
				fmt.Print(time.Now())

				electMeNow()
				time.Sleep(3 * time.Second)
			} else {
				time.Sleep(300 * time.Microsecond)
			}

		}
	}()

}

func electMeNow() {

	SendMulticast(
		multicast.MulticastTopics["ELECT_ME_AS_LEADER"],
		multicast.JoinFields(
			strconv.Itoa(getNewTerm()),
			MyNodeID,
			strconv.Itoa(raftLikeLogger.getCachedLatestOplog()),
		),
	)
}

func voteLeader(requestNewTermStr, requestLeaderNodeID, requestLeaderLogOffsetStr, requestLeaderIP string) {

	voteDecision := 1

	newTerm, err := strconv.Atoi(requestNewTermStr)

	if isIAmLeader() {
		// if isIAmLeader() {
		// fmt.Printf("  [INFO] Reject request from %s as I am already a leader. My IP: %s\n", requestLeaderIP, *getMyself().IP)
		// electMeNow()
		return
	}

	if err != nil {
		fmt.Printf(
			"[WARN] Invalid requested voting term: %v; Err: %s\n",
			newTerm,
			err,
		)
		return
	}

	requestLeaderLogOffset, err := strconv.Atoi(requestLeaderLogOffsetStr)

	if err != nil {
		fmt.Printf("[WARN] Invalid requested voting LogOffset: %s; Err: %s\n", requestLeaderLogOffsetStr, err)
		return
	}

	votedLeaderID, ok := myVote[newTerm]

	raftLikeLogger := GetLogger()

	isAlreadyVote := ok
	if isAlreadyVote {
		if votedLeaderID != requestLeaderNodeID || requestLeaderLogOffset < raftLikeLogger.getCachedLatestOplog() {
			voteDecision = 0
		}
	} else if requestLeaderNodeID != MyNodeID {
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
		myVote[newTerm] = requestLeaderNodeID
	}

	voteResultSend(newTerm, requestLeaderNodeID, requestLeaderIP, voteDecision)
}

func voteResultSend(
	requestTerm int,
	requestLeaderNodeID string,
	requestLeaderIP string,
	voteResult int,
) {
	SendMulticast(
		multicast.MulticastTopics["VOTE_LEADER"],
		multicast.JoinFields(
			strconv.Itoa(requestTerm),
			requestLeaderNodeID,
			requestLeaderIP,
			strconv.Itoa(voteResult),
			MyNodeID,
		),
	)
}

func leaderSendHeartBeat() {
	go func() {
		for {
			if isIAmLeader() {
				*lastHeartbeatTime = time.Now()
				SendMulticast(multicast.MulticastTopics["LEADER_HEARTBEAT"], MyNodeID)
			}
			time.Sleep(heartbeatFrequency)
		}

	}()

}

func updateVoteResult(newTermStr, leaderID, leaderIP, result, voterNodeID, voterIP string) {
	newTermInt, err := strconv.Atoi(newTermStr)
	if err != nil || newTermInt != getNewTerm() || leaderID != MyNodeID {
		return
	}
	_, ok := myLeaderElectionResult[newTermInt]
	if !ok {
		myLeaderElectionResult[newTermInt] = make(map[string]voterVote)
	}

	resultInt, err := strconv.Atoi(result)

	currentVote := voterVote{
		leaderID:    leaderID,
		leaderIP:    leaderIP,
		voterIP:     voterIP,
		voterNodeID: voterNodeID,
		result:      resultInt,
	}

	myLeaderElectionResult[newTermInt][voterIP] = currentVote

	if hasReceivedMajorityVote(newTermInt) {
		updateMeAsLeaderForNewTerm(newTermInt, currentVote)
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
	*lastHeartbeatTime = time.Now()
	if isIAmLeader() {
		// fmt.Printf("[WARN] I'm leader. I don't accept other heartbeats. My IP: %s; Heartbeat IP: %s \n", *getMyself().IP, ip)

		return
	}
	if !isNodeLeader(leaderNodeID, ip) {
		fmt.Println("[WARN] Receive a leader heartbeat that is not from current leader")
		leader := getLeader()
		fmt.Printf(
			"[WARN] RecvHeatbeatNodeIP %s | currentLeaderNodeIP %s | My IP: %s",
			ip,
			*leader.IP,
			*getMyself().IP,
		)
		return
	}

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
		func(msg string, senderIP string, UDPAddr *net.UDPAddr) {
			msgSplit := multicast.GetFields(msg)
			updateVoteResult(msgSplit[0], msgSplit[1], msgSplit[2], msgSplit[3], msgSplit[4], senderIP)
		},
	)
}
func listenLeaderElection() {
	ListenMulticast(
		multicast.MulticastTopics["ELECT_ME_AS_LEADER"],
		func(msg string, senderIP string, UDPAddr *net.UDPAddr) {
			msgSplit := multicast.GetFields(msg)
			voteLeader(msgSplit[0], msgSplit[1], msgSplit[2], senderIP)
		},
	)
}

// StartLeaderElectionService ...
// Start leader election service;
// Listen the election message and decide if I should participate the campaign
func StartLeaderElectionService() {
	fmt.Println("[INFO] StartLeaderElectionService")
	listenLeaderElection()
	listenLeaderVote()
	listenLeaderHeartbeat()
	leaderSendHeartBeat()
	electMeIfLeaderDie()
}
