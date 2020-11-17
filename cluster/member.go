package cluster

import (

	// "crypto/rand"

	"fmt"
	"strconv"
	"time"

	"wordcounter/config"
	"wordcounter/rpc"
	"wordcounter/utils"
)

// Member ...
// @ID -  Member ID, random string assigned on node startup
// @Term - The term for the current member. Start from 0 on node init
// 		It should always align to the leader's term
// @IP - Current member's cluster IP.
type Member struct {
	ID   string
	Term *int
	IP   *string
}

// type MemberForRPC struct {
// 	ID   string
// 	Term int
// 	IP   string
// }

// MemberList ...
type MemberList map[string]Member

// MyNodeID ...
// Node ID
var MyNodeID string = strconv.Itoa(utils.RandWithSeed().Int())
var defaultLogOffset = 0
var myLogOffset *int = &defaultLogOffset

// Membership ...
// Holding all member registry
type Membership struct {
	Leader  *Member
	Members map[string]*Member
}

// type MembershipForRPC struct {
// 	Leader  MemberForRPC
// 	Members map[string]MemberForRPC
// }

var myMembership *Membership = &Membership{}

// GetMembership ...
// Return membership
func GetMembership() *Membership {
	return myMembership
}
func getLeader() *Member {
	return myMembership.Leader
}

// GetLeaderIP ...
// Return Leader IP
func GetLeaderIP() string {
	return *myMembership.Leader.IP
}

func isNodeLeader(nodeID, ip string) bool {
	leader := getLeader()

	if nodeID == "" || ip == "" {
		return false
	}

	if leader.ID == "" || *leader.IP == "" {
		return true
	}

	return leader.ID == nodeID && *leader.IP == ip
}

func updateMeAsLeaderForNewTerm(newTerm int, latestVote voterVote) {

	myself := getMyself()

	updateMyIP(latestVote.leaderIP)

	// *myself.IP = latestVote.leaderIP

	myself.Term = &newTerm
	m := GetMembership()
	m.Leader = myself
	m.Members[myself.ID] = myself
	for _, v := range m.Members {
		v.Term = &newTerm
	}

	fmt.Printf(
		"\n\n\n[INFO] I have been elected as the new leader. My NodeID: %s; My IP: %s; Current Term: %v; My LogOffset: %v; We have %v member(s) in the cluster.\n\n\n",
		myself.ID,
		*myself.IP,
		*myself.Term,
		*myLogOffset,
		len(myMembership.Members),
	)

	syncMembershipToFollowers(myMembership)

}

func getMyself() *Member {
	return myMembership.Members[MyNodeID]
}

func isIAmLeader() bool {
	if isLeaderNil() {
		return false
	}
	return myMembership.Leader.ID == MyNodeID
}

func isLeaderNil() bool {
	return myMembership.Leader == nil
}

// func (m *MembershipForRPC) ReportLeaderIP(payload interface{}, replyLeaderIP *string) error {
// 	ip := m.Leader.IP

// 	if ip == "" {
// 		return errors.New("[ERROR] Unable to fetch leader IP. Please retry")
// 	}
// 	*replyLeaderIP = ip
// 	return nil
// }

// ForEachMember ...
// For each member (including leader) in the membership, run the call back
func ForEachMember(m *Membership, callback func(member Member, isLeader bool)) {

	for key, value := range myMembership.Members {
		callback(*value, key == myMembership.Leader.ID)
	}
}

func addNewMember(m *Membership, newMemberNodeID, newMemberIP string) bool {

	if !isIAmLeader() {
		fmt.Println("[WARN] Only allow Leader to add new member")
		return false
	}

	if newMemberNodeID == getMyself().ID {
		fmt.Printf("[INFO] Should not add myself into the member again. My IP: %s\n", *getMyself().IP)
		return true
	}

	fmt.Printf("[INFO] Adding new member. %s to my cluster. My IP: %s\n", newMemberIP, *getMyself().IP)

	for key, node := range m.Members {
		if *node.IP == newMemberIP {
			delete(m.Members, key)
		}
	}
	newMember := &Member{
		ID:   newMemberNodeID,
		IP:   &newMemberIP,
		Term: m.Leader.Term,
	}
	m.Members[newMemberNodeID] = newMember

	return true
}

// UpdateMembership ...
// Update member status
// It's a RPC methods for syncing membership from the leader to followers
func (m *Membership) UpdateMembership(newMembership Membership, replyNodeID *string) error {
	myself := getMyself()
	myTerm := getMyTerm()
	newLeader := newMembership.Leader
	requestLeaderTerm := *newLeader.Term
	if requestLeaderTerm < myTerm {
		errMsg := fmt.Errorf(
			"Reject syncing member from a leader of previous term. My Term: %v; RequestSyncing Leader term: %v; RequestSyncing Leader ID %s;  RequestSyncing Leader IP %s",
			myTerm,
			requestLeaderTerm,
			newLeader.ID,
			*newLeader.IP,
		)
		return errMsg
	}

	fmt.Printf(
		"\n[INFO] Accept update of membership from Leader. My ID: %s; My IP: %s; My Term: %v; Leader's IP: %s; Leader's Term: %v\n",
		myself.ID,
		*myself.IP,
		*myself.Term,
		*newLeader.IP,
		*newLeader.Term,
	)

	// m.Members = newMembership.Members
	// m.Leader = newMembership.Leader

	*myMembership = newMembership

	// newMembershipTranformed := restoreMembershipFromRPC(newMembership)
	// myMembership.Members = newMembershipTranformed.Members
	// myMembership.Leader = newMembershipTranformed.Leader
	*replyNodeID = MyNodeID

	fmt.Printf("[INFO] Follower Membership updated. New Leader IP: %v", *GetMembership().Leader.IP)

	return nil
}

func syncMembershipToFollowers(m *Membership) {

	if !isIAmLeader() {
		fmt.Println("[WARN] Only allow Leader to sync members")
		return
	}

	fmt.Printf("[INFO] Syncing member list from me. My IP: %s \n", *getMyself().IP)

	for key, value := range m.Members {
		if key == m.Leader.ID {
			// fmt.Println("[INFO] No need to sync membership to a leader. Leader NodeID: `" + key + "` with IP: " + *value.IP)
			continue
		}

		maxattempt := 3
		for {
			if maxattempt <= 0 {
				errMsg := "[ERROR] Fail in perform syncMembershipToFollowers to node `" + key + "` with IP: " + *value.IP + ". Max retry reached. Skip."
				fmt.Println(errMsg)
				return
			}

			fmt.Println("[INFO] Try to sync Member list to node `" + key + "` with IP: " + *value.IP)

			err := callRPCSyncMembership(m, *value.IP)
			if err == nil {
				break
			}

			fmt.Println("[WARN] Unable to syncMembershipToFollowers to node `" + key + "` with IP: " + *value.IP + ". Retry")

			maxattempt--

			time.Sleep(2 * time.Second)
		}

	}

}

func callRPCSyncMembership(m *Membership, ip string) error {

	responseNodeID := ""
	err := rpc.CallRPC(
		ip,
		config.HTTPRPCList["Membership.UpdateMembership"].Name,
		m,
		&responseNodeID,
	)

	return err
}

func updateMyIP(IP string) {
	myself := getMyself()
	*myself.IP = IP
}

// func restoreMembershipFromRPC(m MembershipForRPC) *Membership {

// 	var members map[string]*Member = make(map[string]*Member)

// 	for k, v := range m.Members {
// 		members[k] = getMemberFromMemberForRPC(v)
// 	}

// 	return &Membership{
// 		Leader:  getMemberFromMemberForRPC(m.Leader),
// 		Members: members,
// 	}

// }

// func getMemberFromMemberForRPC(m MemberForRPC) *Member {
// 	return &Member{
// 		ID:   m.ID,
// 		IP:   &m.IP,
// 		Term: &m.Term,
// 	}
// }
// func getMemberForRPCCall(m *Member) MemberForRPC {

// 	return MemberForRPC{
// 		IP:   *m.IP,
// 		ID:   m.ID,
// 		Term: *m.Term,
// 	}
// }
// func getMembershipForRPCCall(m *Membership) MembershipForRPC {

// 	allMembers := make(map[string]MemberForRPC, 10)
// 	for k, m := range m.Members {
// 		allMembers[k] = getMemberForRPCCall(m)
// 	}

// 	return MembershipForRPC{Leader: getMemberForRPCCall(m.Leader), Members: allMembers}

// }

func newMembership() Membership {
	members := make(map[string]*Member)
	var myIP string = ""
	var myTerm int = 0
	members[MyNodeID] = &Member{
		ID:   MyNodeID,
		IP:   &myIP,
		Term: &myTerm,
	}
	return Membership{
		Leader:  &Member{},
		Members: members,
	}
}

func (m *Membership) rpcRegister() {
	rpc.RegisterType(m)
}

// StartMembershipService ...
// Start membership service
// The service holds the status of members and leader
func StartMembershipService() {

	*myMembership = newMembership()
	myMembership.rpcRegister()
	fmt.Printf("[INFO] StartMembershipService. MyNodeID: %s\n", getMyself().ID)

}
