package cluster

import (
	"net"
	"testing"
	"time"
	"wordcounter/multicast"

	"github.com/stretchr/testify/assert"
)

func TestSendReceive(T *testing.T) {

	nodeIDGot := ""

	ListenMulticast(
		multicast.MulticastTopics["JOIN_GROUP"],
		func(nodeID string, ip string, UDPAddr *net.UDPAddr) {
			nodeIDGot = nodeID
		},
	)

	time.Sleep(4 * time.Second)

	assert.NotEmpty(T, nodeIDGot, "It should request join group with NodeID")

}
