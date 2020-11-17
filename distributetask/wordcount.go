package distributetask

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"wordcounter/cluster"
	"wordcounter/config"
	"wordcounter/rpc"
	"wordcounter/utils"
)

// WordCount ...
// Word count RPC receiver type
type WordCount struct {
	text string
}

// CountDescriptor ...
// Use for distributed task
type CountDescriptor struct {
	// LogOffset ...
	// The log offset for the wordcounting to execute against
	LogOffset,
	// PayloadByteOffset ...
	// The offset of the payload that the follower should start with
	PayloadByteOffset,
	// Countlength ...
	// The length from the PayloadByteOffset
	Countlength int
}

var wc *WordCount

func (wc *WordCount) countIt(desc CountDescriptor) (int, error) {
	oplog, err := cluster.GetOplogByOffset(desc.LogOffset)
	if err != nil {
		return 0, err
	}

	bytes, err := utils.GetBytes(oplog.Payload)
	if err != nil {
		return 0, err
	}

	lenBytes := len(bytes)

	payloadByteOffset := desc.PayloadByteOffset
	if desc.PayloadByteOffset > lenBytes {
		return 0, nil
	}

	payloadByteEnd := payloadByteOffset + desc.Countlength
	if payloadByteEnd > lenBytes {
		payloadByteEnd = lenBytes
	}

	myStr := string(bytes[payloadByteOffset:payloadByteEnd])
	m1 := regexp.MustCompile(`\s+`)
	myStr = m1.ReplaceAllString(myStr, " ")
	// replace twice so that all tab \t could be replaced
	myStr = m1.ReplaceAllString(myStr, " ")

	return len(strings.Split(myStr, " ")), nil

}

func newWordCount() *WordCount {
	inst := WordCount{}
	return &inst
}

// Count ...
// RPC method for counting the text file in distributed manner
func (wc *WordCount) Count(desc CountDescriptor, replyWordCount *int) error {
	count, err := wc.countIt(desc)
	*replyWordCount = count
	return err
}

func (wc *WordCount) rpcRegister() {
	rpc.RegisterType(wc)
}

// HTTPProxyWordCount ...
// HTTP outlet for client incoming word counting request
func HTTPProxyWordCount(file multipart.File, w http.ResponseWriter, r *http.Request) (int, error) {

	fmt.Println("[INFO] Accepting request to process word counting.")

	logger := cluster.GetLogger()

	content, err := ioutil.ReadAll(file)

	if err != nil {
		return 0, err
	}

	oplog, err := cluster.AppendOplog(logger, &content, config.HTTPRPCList["WordCount.Count"].Name)

	if err != nil {
		return 500, err
	}

	res := runtask(oplog)

	w.Write([]byte(strconv.Itoa(res)))

	return 200, nil

}

func runtask(oplog cluster.Oplog) int {

	membership := cluster.GetMembership()
	countMembers := len(membership.Members)

	fmt.Printf("[INFO] Run task with log offset %v \n", oplog.LogOffset)

	payloadBytes := []byte(fmt.Sprintf("%v", oplog.Payload))
	countPayloadBytes := len(payloadBytes)
	eachShare := countPayloadBytes / countMembers
	offset := 0

	ret := 0

	reportChan := make(chan int)

	// var wg sync.WaitGroup

	cluster.ForEachMember(
		membership,
		func(member cluster.Member, isLeader bool) {

			// wg.Add(1)
			// go func(offset int, wg *sync.WaitGroup, member *cluster.Member) {
			go func(offset int, member *cluster.Member) {
				fmt.Printf("[INFO] Member IP %s is processing task at log offset %v \n", *member.IP, oplog.LogOffset)

				replyWordCount, err := wc.callRPC(
					*member.IP,
					CountDescriptor{
						LogOffset:         oplog.LogOffset,
						Countlength:       eachShare,
						PayloadByteOffset: offset,
					},
				)
				if err != nil {
					fmt.Printf("[INFO] Member IP %s encountered an error on processing task at log offset %v \n", *member.IP, oplog.LogOffset)
					fmt.Print(err)

					replyWordCount = 0
				}

				reportChan <- replyWordCount

			}(offset, &member)

			offset += eachShare
		},
	)

	for i := 0; i < countMembers; i++ {
		ret += <-reportChan
	}

	return ret
}

func (wc *WordCount) callRPC(ip string, countDescriptor CountDescriptor) (int, error) {

	replyWordCount := 0

	err := rpc.CallRPC(
		ip,
		config.HTTPRPCList["WordCount.Count"].Name,
		countDescriptor,
		&replyWordCount,
	)

	if err != nil {
		return 0, err
	}

	return replyWordCount, nil

}

// StartWordCountService ...
// Start word count service that offer distributed word counting in the cluster
func StartWordCountService() {
	fmt.Println("[INFO] StartWordCountService")
	wc = newWordCount()
	wc.rpcRegister()
}
