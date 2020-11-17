package cluster

import (
	"fmt"
	"net"
	"wordcounter/config"
	"wordcounter/multicast"
)

var sender *multicast.Sender

var multicastAddr *string

func printLog(msgType, topic, msg, senderIP string) {
	if _, ok := multicast.MulticastTopicsHideLog[topic]; ok {
		return
	}
	prefix := "Send"
	suffix := ""
	if msgType == "Listen" {
		prefix = "Received"
		suffix = "from IP " + senderIP
	}
	fmt.Printf("[INFO]>>>>MULTICAST>>>> "+prefix+" multicast topic %s, message: %s; %s\n", topic, msg, suffix)

}

func SendMulticast(topic, msg string) {

	printLog("Send", topic, msg, "")

	if sender == nil {
		envAddr := config.Envs["ENV_MULTICAST_GROUP"]
		multicastAddr = &envAddr
		senderGot := multicast.GetSender(*multicastAddr)
		sender = &senderGot
	}

	sender.Send(topic, msg)
}

func ListenMulticast(topic string, listener func(nodeID string, senderIP string, UDPAddr *net.UDPAddr)) {
	if multicastAddr == nil {
		envAddr := config.Envs["ENV_MULTICAST_GROUP"]
		multicastAddr = &envAddr
	}

	go multicast.Register(
		*multicastAddr,
		func(multicastTopic string, payload string, senderIP string, UDPAddr *net.UDPAddr) {
			if multicastTopic != topic {
				return
			}

			printLog("Listen", topic, payload, senderIP)

			listener(payload, senderIP, UDPAddr)
		},
	)
}
