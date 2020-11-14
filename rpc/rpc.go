package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"time"
	"wordcounter/config"
)

// RegisterRPC ...
func RegisterType(instance interface{}) {
	// membership := NewMembership()

	rpc.Register(instance)

}

func CallRPC(remoteHost, methodName string, payload interface{}, reply interface{}) error {

	// get RPC client by dialing at `rpc.DefaultRPCPath` endpoint
	client, _ := rpc.DialHTTP("tcp", remoteHost+":"+config.Envs["ENV_RPC_PORT"]) // or `localhost:9000`

	if err := client.Call(methodName, payload, &reply); err != nil {

		fmt.Println("[ERROR] RPC Call Error", err)
		return err
	}

	return nil

}

func init() {
	// sample test endpoint
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		io.WriteString(res, "GO RPC SERVER IS ALIVE!")
		io.WriteString(res, "\nDebug: "+rpc.DefaultDebugPath)
		httpRPCListBytes, err := json.Marshal(config.HttpRpcList)
		if err != nil {
			return
		}
		io.WriteString(res, "\n\nRPC Report: "+string(httpRPCListBytes))
	})

	go func() {
		// wait for register types in other packages
		time.Sleep(2 * time.Second)
		rpc.HandleHTTP()
		http.ListenAndServe(":"+config.Envs["ENV_RPC_PORT"], nil)
	}()

}
