package multicast

import (
	"fmt"
	"net"
	"strings"
)

var messageTopicPayloadSplitter = "___||___"

// Sender ...
// Multicast sender (UDP)
type Sender struct {
	conn *net.UDPConn
}

// MulticastListener ...
// Multicast listener over UDP
type MulticastListener func(string, string, string, *net.UDPAddr)

// MulticastTopics ...
// The topic for multicast messages
// Sender and listeners are grouped by topics
var MulticastTopics = map[string]string{
	"JOIN_GROUP":         "JOIN_GROUP",
	"ELECT_ME_AS_LEADER": "ELECT_ME_AS_LEADER",
	"VOTE_LEADER":        "VOTE_LEADER",
	"LEADER_HEARTBEAT":   "LEADER_HEARTBEAT",
}

// MulticastTopicsHideLog ...
// Topics that should not show in logs
var MulticastTopicsHideLog = map[string]string{
	"LEADER_HEARTBEAT": "LEADER_HEARTBEAT",
}

const payloadFieldSplitter string = "|"

// JoinFields ...
// Join multiple fields with splitter for multicast messaging
func JoinFields(fields ...string) string {
	return strings.Join(fields, payloadFieldSplitter)
}

// GetFields ...
// The reverse of JoinFields
func GetFields(joinedFields string) []string {
	return strings.Split(joinedFields, payloadFieldSplitter)
}

// Send ...
// Send message into the topic channel
func (s *Sender) Send(topic string, payload string) {
	s.conn.Write([]byte(topic + messageTopicPayloadSplitter + payload))
}

// GetSender ...
// @param multicastGroup
// 		- multicast group, Port in valid IP Range - 224.0.0.0 through 239.255.255.255
// 		- e.g. 239.0.0.1:10000
// @example
//		```golang
//		sender := multicast.GetSender("239.0.0.1:10000")
//		sender.Send("Hi")
//		```
func GetSender(multicastGroup string) Sender {
	addr, err := net.ResolveUDPAddr("udp4", multicastGroup)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		panic(err)
	}

	return Sender{conn}
}

// Register ...
// @param multicastGroup - See #GetSender
// @example
//		```golang
//		multicast.Register("239.0.0.1:10000", func (msg string, senderIP string, UDPAddr *net.UDPAddr) {
// 			fmt.Print("Received Message: " + msg + " From IP " + ip)
//		})
//		```
func Register(multicastGroup string, callback MulticastListener) {
	udpBufferSize := 1024 * 10
	addr, err := net.ResolveUDPAddr("udp4", multicastGroup)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		panic(err)
	}

	conn.SetReadBuffer(udpBufferSize)

	for {

		buffer := make([]byte, udpBufferSize)
		numBytes, source, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		sourceIP := source.IP.String()
		msg := strings.Split(string(buffer[:numBytes]), messageTopicPayloadSplitter)
		msgLen := len(msg)
		_, ok := MulticastTopics[msg[0]]
		if msgLen != 2 || !ok {
			fmt.Println("[WARN] Received unknown message", msg)
			continue
		}
		callback(msg[0], msg[1], sourceIP, source)

	}
}
