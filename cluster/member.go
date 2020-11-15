package cluster

import (

	// "crypto/rand"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"wordcounter/config"
	"wordcounter/rpc"
)

type Member struct {
	ID   string
	Term *int
	IP   *string
}

type MemberList map[string]Member

// MyNodeID ...
// Node ID
var MyNodeID string = strconv.Itoa(rand.Int())
var myLogOffset int

type Membership struct {
	Leader  *Member
	Members map[string]*Member
}

var myMembership *Membership

func GetMembership() *Membership {
	return myMembership
}
func getLeader() *Member {
	return myMembership.Leader
}

func isNodeLeader(nodeID, ip string) bool {
	leader := getLeader()
	if leader.ID == "" || *leader.IP == "" || nodeID == "" || ip == "" {
		return false
	}
	return leader.ID == nodeID && *leader.IP == ip
}

func updateMeAsLeaderForNewTerm(newTerm int, latestVote voterVote) {

	myself := getMyself()

	updateMyIP(latestVote.leaderIP)

	// *myself.IP = latestVote.leaderIP

	*myself.Term = newTerm
	myMembership.Leader = myself
	myMembership.Members[myself.ID] = myself
	for _, v := range myMembership.Members {
		*v.Term = newTerm
	}

	fmt.Printf(
		"\n\n\n[INFO] I have been elected as the new leader. My NodeID: %s; My IP: %s; Current Term: %v\n\n\n",
		myself.ID,
		*myself.IP,
		*myself.Term,
	)

	syncMemberList(myMembership)

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

func (m *Membership) ReportLeaderIP(payload interface{}, replyLeaderIP *string) error {
	ip := *m.Leader.IP

	if ip == "" {
		return errors.New("[ERROR] Unable to fetch leader IP. Please retry")
	}
	*replyLeaderIP = ip
	return nil
}

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
	if m.Leader.ID == newMember.ID {
		m.Leader = newMember
	}
	syncMemberList(m)
	return true
}

func (m *Membership) UpdateMembership(newMembership Membership, replyNodeID *string) error {
	m.Members = newMembership.Members
	m.Leader = newMembership.Leader
	*replyNodeID = MyNodeID
	return nil
}

func syncMemberList(m *Membership) {

	if !isIAmLeader() {
		fmt.Println("[WARN] Only allow Leader to sync members")
		return
	}

	for key, value := range m.Members {
		if key == m.Leader.ID {
			continue
		}

		maxattempt := 3
		for {
			if maxattempt <= 0 {
				errMsg := "[ERROR] Fail in perform SyncMemberList to node `" + key + "` with IP: " + *value.IP + ". Max retry reached. Skip."
				fmt.Println(errMsg)
				return
			}

			err := callRPCSyncMembership(m, *value.IP)
			if err == nil {
				break
			}

			fmt.Println("[WARN] Unable to SyncMemberList to node `" + key + "` with IP: " + *value.IP + ". Retry")

			maxattempt--

			time.Sleep(2 * time.Second)
		}

	}

}

func callRPCSyncMembership(m *Membership, ip string) error {
	responseNodeID := ""
	err := rpc.CallRPC(
		ip,
		config.HttpRpcList["Membership.UpdateMembership"].Name,
		m,
		&responseNodeID,
	)

	return err
}

func updateMyIP(IP string) {
	myself := getMyself()
	*myself.IP = IP
}

func newMembership() *Membership {
	members := make(map[string]*Member)
	var myIP string = ""
	var myTerm int = 0
	members[MyNodeID] = &Member{
		ID:   MyNodeID,
		IP:   &myIP,
		Term: &myTerm,
	}
	return &Membership{
		Leader:  &Member{},
		Members: members,
	}
}

func (membership *Membership) rpcRegister() {
	rpc.RegisterType(membership)
}

func StartMembershipService() {
	myMembership = newMembership()
	myMembership.rpcRegister()

}
