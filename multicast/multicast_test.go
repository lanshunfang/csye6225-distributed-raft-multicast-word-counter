package multicast

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSendReceive(T *testing.T) {
	var messageGot string
	var messageSend string = "Hi"
	go Register(
		"239.0.0.1:10000",
		func(msg string, ip string, source *net.UDPAddr) {
			fmt.Print("Received Message: " + msg + " From IP " + ip)
			messageGot = msg
		},
	)

	time.Sleep(2 * time.Second)

	sender := GetSender("239.0.0.1:10000")
	sender.Send(messageSend)

	time.Sleep(5 * time.Second)

	assert.Equal(T, messageSend, messageGot, "The Message ")

}
