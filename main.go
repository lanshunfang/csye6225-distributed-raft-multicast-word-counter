package main

import (
	"encoding/binary"
	"errors"
	"net"
)

func main() {
	getBroadcastAddr(&net.IPNet{})
	// fs := http.FileServer(http.Dir("./static"))
	// http.Handle("/", fs)

	// log.Println("Listening on :3000...")
	// err := http.ListenAndServe(":3000", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

// https://stackoverflow.com/questions/36166791/how-to-get-broadcast-address-of-ipv4-net-ipnet
func getBroadcastAddr(n *net.IPNet) (net.IP, error) {
	if n.IP.To4() == nil {
		return net.IP{}, errors.New("does not support IPv6 addresses.")
	}
	ip := make(net.IP, len(n.IP.To4()))
	binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(n.IP.To4())|^binary.BigEndian.Uint32(net.IP(n.Mask).To4()))
	return ip, nil
}
