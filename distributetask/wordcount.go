package distributetask

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"wordcounter/cluster"
	"wordcounter/config"
	"wordcounter/rpc"
)

type WordCount struct {
	text string
}

type CountDescriptor struct {
	logOffset, payloadByteOffset, countlength int
}

func (wc *WordCount) countIt(desc CountDescriptor) (int, error) {
	logger := cluster.GetLogger()
	oplog, err := cluster.GetOplogByOffset(&logger, desc.logOffset)
	if err != nil {
		return 0, err
	}

	// bytes := []byte(fmt.Sprintf("%v", oplog.Payload.(interface{})))
	bytes := []byte(fmt.Sprintf("%v", oplog.Payload))
	lenBytes := len(bytes)

	if desc.payloadByteOffset > lenBytes-1 {
		return 0, nil
	}

	payloadByteOffset := desc.payloadByteOffset
	payloadByteEnd := payloadByteOffset + desc.countlength
	if payloadByteEnd > lenBytes {
		payloadByteEnd = lenBytes
	}

	myStr := string(bytes[payloadByteOffset:payloadByteEnd])
	// myStr = strings.Replace(myStr, "\t", " ", -1)
	m1 := regexp.MustCompile(`\s+`)
	myStr = m1.ReplaceAllString(myStr, " ")
	// replace twice so that all tab \t could be replaced
	myStr = m1.ReplaceAllString(myStr, " ")

	return len(strings.Split(myStr, " ")), nil

}

func NewWordCount() WordCount {
	inst := WordCount{}
	return inst
}
func (wc *WordCount) Count(desc CountDescriptor, replyWordCount *int) error {
	count, err := wc.countIt(desc)
	*replyWordCount = count
	return err
}

func (wc *WordCount) rpcRegister() {
	rpc.RegisterType(wc)
}

// HTTPHandle ...
// HTTP outlet for client incoming word counting request
func (wc *WordCount) HTTPHandle(text string, reply *int) error {

	logger := cluster.GetLogger()

	oplog, err := cluster.AppendOplog(&logger, &text)

	if err != nil {
		return err
	}

	*reply = runtask(wc, oplog)

	return nil

}

func runtask(wc *WordCount, oplog cluster.Oplog) int {
	membership := cluster.GetMembership()
	countMembers := len(membership.Members)
	payloadBytes := []byte(fmt.Sprintf("%v", oplog.Payload))
	countPayloadBytes := len(payloadBytes)
	eachShare := countPayloadBytes / countMembers
	offset := 0

	ret := 0

	reportChan := make(chan int)

	var wg sync.WaitGroup

	cluster.ForEachMember(
		membership,
		func(member cluster.Member, isLeader bool) {

			wg.Add(1)
			go func(wg *sync.WaitGroup, member *cluster.Member) {
				replyWordCount, err := wc.callRPC(member.IP, offset)
				if err != nil {
					replyWordCount = 0
				}
				reportChan <- replyWordCount
				wg.Done()
			}(&wg, &member)

			offset += eachShare
		},
	)

	wg.Wait()

	for v := range reportChan {
		ret += v
	}

	return ret
}

func (wc *WordCount) callRPC(ip string, logOffset int) (int, error) {

	replyWordCount := 0

	err := rpc.CallRPC(
		ip,
		config.HttpRpcList["WordCount.Count"].Name,
		logOffset,
		&replyWordCount,
	)

	return replyWordCount, err
}

func StartWordCountService() {
	wc := NewWordCount()
	wc.rpcRegister()
}
