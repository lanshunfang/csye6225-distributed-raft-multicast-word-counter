package multicast

import (
	"net"
)

type Sender struct {
	conn *net.UDPConn
}

func (s *Sender) Send(msg string) {
	s.conn.Write([]byte(msg))
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
//		multicast.Register("239.0.0.1:10000", func (msg string, ip string, source *net.UDPAddr) {
// 			fmt.Print("Received Message: " + msg + " From IP " + ip)
//		})
//		```
func Register(multicastGroup string, callback func(string, string, *net.UDPAddr)) {
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
		callback(string(buffer[:numBytes]), sourceIP, source)

	}
}
