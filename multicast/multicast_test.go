package multicast

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSendReceive(T *testing.T) {
	var messageTopic string = "SomeTopic"
	var messageTopicGot string
	var messageSend string = "Hi"
	var messageGot string
	go Register(
		"239.0.0.1:10000",
		func(topic string, payload string, ip string, UDPAddr *net.UDPAddr) {
			messageTopicGot = topic
			fmt.Print("Received Message: " + payload + " From IP " + ip)
			messageGot = payload
		},
	)

	time.Sleep(2 * time.Second)

	sender := GetSender("239.0.0.1:10000")
	sender.Send(messageTopic, messageSend)

	time.Sleep(5 * time.Second)

	assert.Equal(T, messageTopic, messageTopicGot, "The topic received should be equal to the topic sent")
	assert.Equal(T, messageSend, messageGot, "The message received should be equal to the message sent")

}
