package cluster

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"wordcounter/config"
	"wordcounter/osextend"
	"wordcounter/rpc"
)

// Oplog ...
// Operation log
type Oplog struct {
	RPCMethod string
	// The HTTP Request payload for the RPCMethod
	Payload   interface{}
	LogOffset int
	Timestamp string
}

var LogStack = make([]Oplog, 10)

// RaftLikeLogger ...
// Raft log for High Availability cluster
type RaftLikeLogger struct{}

var raftLikeLogger *RaftLikeLogger

var logfilePath string = "./oplog.gob"

var logfileLock sync.Mutex

var oplogSyncing map[string]bool = map[string]bool{}

func getLatestOplog() (Oplog, error) {
	lenLogstack := len(LogStack)
	if lenLogstack < 1 {
		return Oplog{}, errors.New("[WARN] There are no any logs found")
	}

	maxOffset := 0
	for i := 0; i < lenLogstack; i++ {
		if LogStack[i].LogOffset > maxOffset {
			maxOffset = LogStack[i].LogOffset
		}
	}

	return GetOplogByOffset(maxOffset)
}

func (l *RaftLikeLogger) getCachedLatestOplog() int {
	return *myLogOffset
}

// LeaderAppendLog ...
// Append log from leader node then sync to all the followers
func (l *RaftLikeLogger) LeaderAppendLog(log Oplog, replyLog *Oplog) error {

	appendOplog(l, log)
	syncLog()
	*replyLog = log
	return nil

}

func appendOplog(l *RaftLikeLogger, log Oplog) {
	LogStack = append(LogStack, log)
	commitLog(l)
}

// AppendOplog ...
// Append oplog to log pipe
func AppendOplog(l *RaftLikeLogger, payload *[]byte, RPCMethod string) (Oplog, error) {

	logOffset := 0

	lastLog, err := getLatestOplog()

	if err == nil {
		logOffset = lastLog.LogOffset + 1
	}

	log := Oplog{
		RPCMethod: RPCMethod,
		Payload:   *payload,
		Timestamp: time.Now().String(),
		LogOffset: logOffset,
	}

	replyOplog := Oplog{}
	leaderIP := GetLeaderIP()
	err = rpc.CallRPC(
		leaderIP,
		config.HTTPRPCList["RaftLikeLogger.LeaderAppendLog"].Name,
		log,
		&replyOplog,
	)

	return replyOplog, err
}

func commitLog(l *RaftLikeLogger) {
	logfileLock.Lock()
	defer logfileLock.Unlock()

	dataFile, err := os.OpenFile(
		logfilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	defer dataFile.Close()
	if err != nil {
		fmt.Printf("[ERROR] Open/Create log failed. File path %s", logfilePath)
	}
	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(&LogStack)
	updateMemberLogOffset()
}
func (l *RaftLikeLogger) loadLog() {
	// open data file
	isFileExist := osextend.FileExists(logfilePath)
	if !isFileExist {
		fmt.Printf("[WARN] Log loading skipped. Reason: log file doesn't exist at path %s\n", logfilePath)
		return
	}

	dataFile, err := os.Open(logfilePath)
	defer dataFile.Close()

	if err != nil {

		fmt.Printf("[WARN] Unable to load log file at path %s", logfilePath)
		fmt.Println(err)
		return
	}

	dataDecoder := gob.NewDecoder(dataFile)

	err = dataDecoder.Decode(&LogStack)

	if err != nil {
		fmt.Println(err)
		return
	}

	updateMemberLogOffset()

	fmt.Printf("[INFO] Loaded log data from file %s\n", logfilePath)
}

// GetOplogByOffset ...
// Get oplog by its offset
func GetOplogByOffset(offset int) (Oplog, error) {

	for lastIdx := len(LogStack) - 1; lastIdx > 0; lastIdx-- {
		cursor := LogStack[lastIdx]

		if cursor.LogOffset == offset {
			return cursor, nil
		}
	}
	return Oplog{}, errors.New("[ERROR] Unable to find a valid Oplog matching offset" + strconv.Itoa(offset))
}

func (l *RaftLikeLogger) isSyncing(member Member) bool {
	state, ok := oplogSyncing[*member.IP]
	return ok && state == true
}

func (l *RaftLikeLogger) updateSyncing(member Member, state bool) {
	oplogSyncing[*member.IP] = state
}

func (l *RaftLikeLogger) syncOplogs() {

	if !isIAmLeader() {
		fmt.Println("[WARN] Only allow Leader to sync members")
		return
	}

	if len(LogStack) < 1 {
		fmt.Println("[INFO] No log data found")
		return
	}

	fmt.Printf("[INFO] Start to sync logs to all members from %s \n", GetMyIP())

	membership := GetMembership()

	var wg sync.WaitGroup

	ForEachMember(
		membership,
		func(member Member, isLeader bool) {

			fmt.Printf("[INFO] Work on member %s\n", *member.IP)

			if isLeader {
				return
			}

			l.syncLogForMember(member, &wg)

		},
	)

	wg.Wait()

}

func (l *RaftLikeLogger) syncLogForMember(member Member, wg *sync.WaitGroup) {

	maxattempt := 3
	lenLogData := len(LogStack)

	if lenLogData < 1 {
		fmt.Printf("[WARN] Unable to find any log to sync in %s\n", GetMyIP())

		return
	}

	oplog, err := getLatestOplog()

	if err != nil {
		fmt.Printf("[ERROR] Unable to find log at offset %v to sync in %s\n", lenLogData-1, GetMyIP())

		return
	}

	errMsg := "[WARN] Unable to syncOplogs to node `" + member.ID + "` with IP: " + *member.IP

	validLogOffset := oplog.LogOffset

	wg.Add(1)

	go func() {

		if l.isSyncing(member) {
			fmt.Printf("[WARN] Already syncing with member %s", *member.IP)
			wg.Done()
			return
		}

		l.updateSyncing(member, true)

		for {

			if maxattempt < 0 {
				fmt.Println(errMsg)

				break
			}

			oplog, err := GetOplogByOffset(validLogOffset)

			if err != nil {
				fmt.Printf("[WARN] The log offset %v is invalid\n", validLogOffset)
				break
			}

			fmt.Printf("[INFO] Start to sync log at offset %v with member %s\n", oplog.LogOffset, *member.IP)

			err = l.callRPCSyncLog(oplog, *member.IP, &validLogOffset)
			if err == nil {
				if validLogOffset == l.getCachedLatestOplog()+1 {
					break
				} else {
					validLogOffset++
					continue
				}

			} else {

				fmt.Println(errMsg + ". Retry. Max attempts left: " + strconv.Itoa(maxattempt))

				maxattempt--

				time.Sleep(100 * time.Millisecond)

			}

		}

		l.updateSyncing(member, false)

		wg.Done()

	}()

	wg.Wait()

}

// FollowerAppendLog ...
// RPC method for all followers to add their log
// The RPC method Will be called from leader node
func (l *RaftLikeLogger) FollowerAppendLog(oplog Oplog, replyValidOffset *int) error {
	myself := getMyself()
	if isIAmLeader() {
		return fmt.Errorf(
			"[ERROR] I am leader. Do not accept new log from RPC. My NodeId %s, my IP %s", myself.ID, *myself.IP,
		)

	}

	_, err := GetOplogByOffset(oplog.LogOffset)

	if err == nil {
		fmt.Printf(
			"[WARN] I have the log already. MyID: %s, MyIP: %s\n", myself.ID, *myself.IP,
		)
		*replyValidOffset = oplog.LogOffset + 1
		return nil
	}

	acceptableMaxLogoffset := l.getCachedLatestOplog() + 1
	if oplog.LogOffset > acceptableMaxLogoffset {

		*replyValidOffset = acceptableMaxLogoffset

		fmt.Printf(
			"[WARN] I don't have the latest log. Please send me log from logoffset. MyID: %s, MyIP: %s", myself.ID, *myself.IP,
		)

	} else {
		if oplog.LogOffset < acceptableMaxLogoffset {
			fmt.Printf("[WARN] The log offset in leader may be smaller than mine. MyID: %s, MyIP: %s", myself.ID, *myself.IP)

		}
		fmt.Printf("[INFO] Start to append log into my log stack. MyIP: %s. Leader's IP %s\n", *myself.IP, GetLeaderIP())

		appendOplog(l, oplog)
	}

	return nil
}

func updateMemberLogOffset() {
	lastLog, err := getLatestOplog()
	if err == nil {
		*myLogOffset = lastLog.LogOffset
	} else {
		fmt.Printf("[ERROR] Unable to update my log offset.")
	}

}

// GetLogger ...
// Get logger instance
func GetLogger() *RaftLikeLogger {
	return raftLikeLogger
}

func newLogger() *RaftLikeLogger {
	raftLikeLogger = &RaftLikeLogger{}
	return raftLikeLogger
}

func (l *RaftLikeLogger) rpcRegister() {
	rpc.RegisterType(l)
}

func (l *RaftLikeLogger) callRPCSyncLog(oplog Oplog, ip string, replyValidOffset *int) error {

	err := rpc.CallRPC(
		ip,
		config.HTTPRPCList["RaftLikeLogger.FollowerAppendLog"].Name,
		oplog,
		&replyValidOffset,
	)

	return err
}

// StartRaftLogService ...
// We offter a raft safe log so that when a node die,
func StartRaftLogService() {
	fmt.Println("[INFO] StartRaftLogService")
	raftLikeLogger = newLogger()
	raftLikeLogger.rpcRegister()
	raftLikeLogger.loadLog()
}
