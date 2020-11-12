package clustering

import (
	"fmt"
	"net"
	"wordcounter/config"
	"wordcounter/multicast"
)

var sender *multicast.Sender

var multicastAddr *string

func SendMulticast(topic, msg string) {

	if sender == nil {
		envAddr := config.GetEnvMulticastGroup()
		multicastAddr = &envAddr
		senderGot := multicast.GetSender(*multicastAddr)
		sender = &senderGot
	}

	sender.Send(topic, msg)
}

func ListenMulticast(topic string, listener func(nodeID string, ip string, UDPAddr *net.UDPAddr)) {
	if multicastAddr == nil {
		envAddr := config.GetEnvMulticastGroup()
		multicastAddr = &envAddr
	}

	go multicast.Register(
		*multicastAddr,
		func(topic string, payload string, ip string, UDPAddr *net.UDPAddr) {
			fmt.Println("[INFO]Received topic: " + topic + " From IP " + ip)
			listener(payload, ip, UDPAddr)
		},
	)
}
