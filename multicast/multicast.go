package multicast

import (
	"fmt"
	"net"
	"strings"
)

var messageTopicPayloadSplitter = "___||___"

type Sender struct {
	conn *net.UDPConn
}

type MulticastListener func(string, string, string, *net.UDPAddr)

var MulticastTopics = map[string]string{
	"JOIN_GROUP": "JOIN_GROUP",
	"SYNC_GROUP": "SYNC_GROUP",
}

func (s *Sender) Send(topic string, payload string) {
	s.conn.Write([]byte(topic + messageTopicPayloadSplitter + payload))
}

// GetSenderHandler ...
// @param multicastGroup
// 		- Multicast group, Port in valid IP Range - 224.0.0.0 through 239.255.255.255
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
//		multicast.Register("239.0.0.1:10000", func (msg string, ip string, UDPAddr *net.UDPAddr) {
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
