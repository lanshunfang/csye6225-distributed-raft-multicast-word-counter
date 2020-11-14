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

type Oplog struct {
	RPCMethod string
	Payload   interface{}
	LogOffset int
	Timestamp string
}

type RaftLikeLogger struct {
	Logstack []Oplog
}

var raftLikeLogger RaftLikeLogger

var logfilePath string = "./oplog.gob"

var logfileLock sync.Mutex

var oplogSyncing map[string]bool

func (l *RaftLikeLogger) getLatestOplog() (Oplog, error) {
	lenLogstack := len(l.Logstack)
	if lenLogstack < 1 {
		return Oplog{}, errors.New("[WARN] There are no any logs found")
	}

	return l.Logstack[lenLogstack-1], nil
}

func (l *RaftLikeLogger) getCachedLatestOplog() int {
	return myLogOffset
}

func appendOplog(l *RaftLikeLogger, log Oplog) {
	l.Logstack = append(l.Logstack, log)
	commitLog(l)
}

// AppendOplog ...
// Append oplog to log pipe
func AppendOplog(l *RaftLikeLogger, payload *string) (Oplog, error) {
	if !isIAmLeader() {
		errMsg := "[WARN] Only allow Leader to append log directly"
		fmt.Println(errMsg)
		return Oplog{}, errors.New(errMsg)
	}

	logOffset := 0

	lastLog, err := l.getLatestOplog()

	if err == nil {
		logOffset = lastLog.LogOffset + 1
	}

	log := Oplog{
		RPCMethod: config.HttpRpcList["RaftLikeLogger.AppendLog"].Name,
		Payload:   []byte(*payload),
		Timestamp: time.Now().String(),
		LogOffset: logOffset,
	}
	appendOplog(l, log)
	syncLog()
	return log, nil
}

func commitLog(l *RaftLikeLogger) {
	logfileLock.Lock()
	defer logfileLock.Unlock()

	dataFile, err := os.OpenFile(
		logfilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		fmt.Printf("[ERROR] Open/Create log failed. File path %s", logfilePath)
	}
	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(l.Logstack)
	updateMemberLogOffset()
	defer dataFile.Close()
}
func (l *RaftLikeLogger) loadLog() {
	// open data file
	isFileExist := osextend.FileExists(logfilePath)
	if !isFileExist {
		fmt.Printf("[WARN] Log loading skipped. Reason: log file doesn't exist at path %s", logfilePath)
		return
	}

	dataFile, err := os.Open(logfilePath)

	if err != nil {

		fmt.Printf("[WARN] Unable to load log file at path %s", logfilePath)
		fmt.Println(err)
		return
	}

	dataDecoder := gob.NewDecoder(dataFile)

	err = dataDecoder.Decode(&l.Logstack)

	if err != nil {
		fmt.Println(err)
		return
	}

	dataFile.Close()

	updateMemberLogOffset()

	fmt.Printf("[INFO] Loaded log data from file %s", logfilePath)
}

// GetOplogByOffset ...
// Get oplog by its offset
func GetOplogByOffset(l *RaftLikeLogger, offset int) (Oplog, error) {

	for lastIdx := len(l.Logstack) - 1; lastIdx > 0; lastIdx-- {
		cursor := l.Logstack[lastIdx]
		if cursor.LogOffset == offset {
			return cursor, nil
		}
	}
	return Oplog{}, errors.New("[ERROR] Unable to find a valid Oplog matching offset " + strconv.Itoa(offset))
}

func (l *RaftLikeLogger) isSyncing(member Member) bool {
	state, ok := oplogSyncing[member.IP]
	return ok && state == true
}

func (l *RaftLikeLogger) updateSyncing(member Member, state bool) {
	oplogSyncing[member.IP] = state
}

func (l *RaftLikeLogger) syncOplogs() {

	if !isIAmLeader() {
		fmt.Println("[WARN] Only allow Leader to sync members")
		return
	}

	if len(l.Logstack) < 1 {
		fmt.Println("[INFO] No log data found")
		return
	}

	membership := GetMembership()

	var wg sync.WaitGroup

	ForEachMember(
		membership,
		func(member Member, isLeader bool) {

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
	lenLogData := len(l.Logstack)

	if lenLogData < 1 {
		return
	}

	oplog, err := GetOplogByOffset(l, lenLogData-1)

	if err != nil {
		return
	}

	errMsg := "[WARN] Unable to syncOplogs to node `" + member.ID + "` with IP: " + member.IP

	validLogOffset := oplog.LogOffset

	wg.Add(1)

	go func() {

		if l.isSyncing(member) {
			wg.Done()
			return
		}

		l.updateSyncing(member, true)

		for {

			if maxattempt < 0 {
				fmt.Println(errMsg)

				break
			}

			oplog, err := GetOplogByOffset(l, validLogOffset)

			if err != nil {
				fmt.Printf("[WARN] The log offset %s is invalid", validLogOffset)
				break
			}

			err = l.callRPCSyncLog(oplog, member.IP, &validLogOffset)
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

}

func (l *RaftLikeLogger) AppendLog(oplog Oplog, replyValidOffset *int) error {
	myself := getMyself()
	if isIAmLeader() {
		return errors.New(
			fmt.Sprintf("[ERROR] I am leader. Do not accept new log from RPC. My NodeId %s, my IP %s", myself.ID, myself.IP),
		)
	}

	_, err := GetOplogByOffset(l, oplog.LogOffset)

	if err == nil {
		fmt.Printf(
			"[WARN] I have the log already. MyID: %s, MyIP: %s", myself.ID, myself.IP,
		)
		return nil
	}

	acceptableMaxLogoffset := l.getCachedLatestOplog() + 1
	if oplog.LogOffset > acceptableMaxLogoffset {

		*replyValidOffset = acceptableMaxLogoffset

		fmt.Printf(
			"[WARN] I don't have the latest log. Please send me log from logoffset. MyID: %s, MyIP: %s", myself.ID, myself.IP,
		)

	} else {
		if oplog.LogOffset < acceptableMaxLogoffset {
			fmt.Printf("[WARN] The log offset in leader may be smaller than mine. MyID: %s, MyIP: %s", myself.ID, myself.IP)

		}
		appendOplog(l, oplog)
	}

	return nil
}

func updateMemberLogOffset() {
	lastLog, err := raftLikeLogger.getLatestOplog()
	if err == nil {
		myLogOffset = lastLog.LogOffset
	}

}

func GetLogger() RaftLikeLogger {
	return raftLikeLogger
}

func newLogger() RaftLikeLogger {
	raftLikeLogger = RaftLikeLogger{}
	return raftLikeLogger
}

func (l *RaftLikeLogger) rpcRegister() {
	rpc.RegisterType(l)
}

func (l *RaftLikeLogger) callRPCSyncLog(oplog Oplog, ip string, replyValidOffset *int) error {

	err := rpc.CallRPC(
		ip,
		config.HttpRpcList["RaftLikeLogger.AppendLog"].Name,
		oplog,
		replyValidOffset,
	)

	return err
}

// StartRaftLogService ...
// We offter a raft safe log so that when a node die,
func StartRaftLogService() {
	raftLikeLogger = newLogger()
	raftLikeLogger.rpcRegister()
	raftLikeLogger.loadLog()
}
