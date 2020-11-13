package rpc

import (
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"wordcounter/config"
)

// RegisterRPC ...
func RegisterRPC(instance interface{}, portName string) {
	// membership := NewMembership()

	rpc.Register(instance)
	rpc.HandleHTTP()

	// listen and serve default HTTP server
	http.ListenAndServe(":"+config.EnvPorts[portName], nil)

}

func CallRPC(remoteHost, remotePortName, methodName string, payload interface{}, reply interface{}) error {

	// get RPC client by dialing at `rpc.DefaultRPCPath` endpoint
	client, _ := rpc.DialHTTP("tcp", remoteHost+":"+config.EnvPorts[remotePortName]) // or `localhost:9000`

	if err := client.Call(methodName, payload, reply); err != nil {
		// if err := client.Call("College.Get", 1, &john); err != nil {
		fmt.Println("Error:1 College.Get()", err)
		return err
	}

	return nil

}

func init() {
	// sample test endpoint
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		io.WriteString(res, "GO RPC SERVER IS ALIVE!")
		io.WriteString(res, "Debug: "+rpc.DefaultDebugPath)
	})
}
