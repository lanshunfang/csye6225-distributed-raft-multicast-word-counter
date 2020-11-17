package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"wordcounter/config"
)

// RegisterRPC ...
func RegisterType(instance interface{}) {
	// membership := NewMembership()

	rpc.Register(instance)

}

// CallRPC ...
// Call GO RPC in the remote host
func CallRPC(remoteHost, methodName string, payload interface{}, replyPointer interface{}) error {

	// get RPC client by dialing at `rpc.DefaultRPCPath` endpoint

	addr := remoteHost + ":" + config.Envs["ENV_RPC_PORT"]

	client, err := rpc.DialHTTP("tcp", addr) // or `localhost:9000`
	if err != nil {
		fmt.Printf("[ERROR] RPC Client ERR %s\n", err)
		return err
	}

	err = client.Call(methodName, payload, replyPointer)

	if err != nil {

		fmt.Println("[ERROR] RPC Call Error", err)
		return err
	}

	return nil

}

func StartRPCService() {
	fmt.Printf("[INFO] StartRPCService at Port %s", config.Envs["ENV_RPC_PORT"])
	// sample test endpoint
	http.HandleFunc("/rpc", func(res http.ResponseWriter, req *http.Request) {
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
		rpc.HandleHTTP()
		http.ListenAndServe(":"+config.Envs["ENV_RPC_PORT"], nil)
	}()
}
