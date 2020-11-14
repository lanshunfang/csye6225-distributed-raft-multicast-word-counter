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
	Term int
	IP   string
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

func IsLeader(nodeID, ip string) bool {
	leader := getLeader()
	if leader.ID == "" || leader.IP == "" || nodeID == "" || ip == "" {
		return false
	}
	return leader.ID == nodeID && leader.IP == ip
}

func updateMeAsLeaderForNewTerm(newTerm int) {
	myself := getMyself()
	myself.Term = newTerm
	myMembership.Leader = myself
	myMembership.Members[myself.ID] = myself
	for _, v := range myMembership.Members {
		v.Term = newTerm
	}

	myMembership.SyncMemberList()

}

func getMyself() *Member {
	return myMembership.Members[MyNodeID]
}

func IsIAmLeader() bool {
	return myMembership.Leader.ID == MyNodeID
}

func (m *Membership) ReportLeaderIP(payload interface{}, replyLeaderIP *string) error {
	ip := m.Leader.IP

	if ip == "" {
		return errors.New("[ERROR] Unable to fetch leader IP. Please retry")
	}
	*replyLeaderIP = ip
	return nil
}

func (m *Membership) ForEachMember(callback func(member Member, isLeader bool)) {

	for key, value := range myMembership.Members {
		callback(*value, key == myMembership.Leader.ID)
	}
}

func (m *Membership) AddNewMember(nodeID, ip string) bool {

	if !IsIAmLeader() {
		fmt.Println("[WARN] Only allow Leader to add new member")
		return false
	}
	for key, node := range m.Members {
		if node.IP == ip {
			delete(m.Members, key)
		}
	}
	m.Members[nodeID] = &Member{IP: ip, ID: nodeID}
	m.SyncMemberList()
	return true
}

func (m *Membership) UpdateMembership(newMembership Membership, replyNodeID *string) error {
	m.Members = newMembership.Members
	m.Leader = newMembership.Leader
	*replyNodeID = MyNodeID
	return nil
}

func (m *Membership) SyncMemberList() error {

	if !IsIAmLeader() {
		fmt.Println("[WARN] Only allow Leader to sync members")
		return nil
	}

	for key, value := range m.Members {
		if key == m.Leader.ID {
			continue
		}

		maxattempt := 3
		for {
			if maxattempt <= 0 {
				errMsg := "[ERROR] Fail in perform SyncMemberList to node `" + key + "` with IP: " + value.IP + ". Max retry reached. Skip."
				fmt.Println(errMsg)
				return errors.New(errMsg)
			}

			err := callRPCSyncMembership(m, value.IP)
			if err == nil {
				break
			}

			fmt.Println("[WARN] Unable to SyncMemberList to node `" + key + "` with IP: " + value.IP + ". Retry")

			maxattempt--

			time.Sleep(2 * time.Second)
		}

	}

	return nil
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

func NewMembership() *Membership {
	return &Membership{}
}

func (membership *Membership) rpcRegister() {
	rpc.RegisterType(membership)
}

func StartMembershipService() {
	myMembership = NewMembership()
	myMembership.rpcRegister()

}
