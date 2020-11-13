package clustering

import (

	// "crypto/rand"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"wordcounter/rpc"
)

type Member struct {
	ID        string
	Term      int
	IP        string
	LogOffset int
}

type MemberList map[string]Member

// MyNodeID ...
// Node ID
var MyNodeID string = strconv.Itoa(rand.Int())

type Membership struct {
	Leader  Member
	Members map[string]Member
}

var myMembership = &Membership{}

func getMembership() Membership {
	return *myMembership
}
func getLeader() Member {
	return myMembership.Leader
}

func ForEachMember(callback func(member Member, isLeader bool)) {

	for key, value := range myMembership.Members {
		callback(value, key == myMembership.Leader.ID)
	}
}

func IsLeader(nodeID, ip string) bool {
	leader := myMembership.Leader
	if leader.ID == "" || leader.IP == "" || nodeID == "" || ip == "" {
		return false
	}
	return leader.ID == nodeID && leader.IP == ip
}

func getMyself() Member {
	return myMembership.Members[MyNodeID]
}

func IsIAmLeader() bool {
	return myMembership.Leader.ID == MyNodeID
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
	m.Members[nodeID] = Member{IP: ip, ID: nodeID}
	m.SyncMemberList()
	return true
}

func (m *Membership) UpdateMemberList(mList MemberList, replyNodeID *string) error {
	m.Members = mList
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

			err := callRPCSyncMember(m, value.IP)
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

func callRPCSyncMember(m *Membership, ip string) error {
	responseNodeID := ""
	err := rpc.CallRPC(
		ip,
		"ENV_PORT_MEMBERSHIP_SYNC",
		"Membership.UpdateMemberList",
		m.Members,
		&responseNodeID,
	)

	return err
}

func NewMembership() Membership {
	return Membership{
		Leader: Member{},
		// map[nodeId]Member
		Members: make(MemberList),
	}
}

func createRpcServer() {

	membership := NewMembership()
	rpc.RegisterRPC(membership, "ENV_PORT_MEMBERSHIP_SYNC")
}

func init() {
	createRpcServer()
}
