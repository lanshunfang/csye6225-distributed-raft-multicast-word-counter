package distributetask

import "wordcounter/rpc"

type WordCount struct {
	text string
}

func NewWordCount() WordCount {
	inst := WordCount{}
	return inst
}
func (wc *WordCount) Count(startByteOffset, endByteOffset) int {

}

func (wc *WordCount) RPCServe() {

	rpc.RegisterRPC(wc, "ENV_PORT_TASK_WORDCOUNT")
}

func (wc *WordCount) callRPC(ip string, logOffset int) error {

	err := rpc.CallRPC(
		ip,
		"ENV_PORT_TASK_WORDCOUNT",
		"WordCount.Count",
		oplog,
		&replyValidOffset,
	)

	return err
}

func init() {
	wc := NewWordCount()
	wc.RPCServe()
	raftLikeLogger.syncOplogs()
}
