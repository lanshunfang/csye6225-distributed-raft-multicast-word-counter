package clustering

import (

	// "crypto/rand"

	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"wordcounter/rpc"
)

type Oplog struct {
	RPCMethod string
	Payload   []byte
	LogOffset int
	Timestamp string
}

type RaftLikeLogger struct {
	Logpipe []Oplog
}

var raftLikeLogger RaftLikeLogger

var logfilePath string = "./oplog.gob"

var logfileLock sync.Mutex

func (l *RaftLikeLogger) appendOplog(log Oplog) {
	l.Logpipe = append(l.Logpipe, log)
	l.commitLog()
}
func (l *RaftLikeLogger) AppendOplog(log Oplog) {
	if !IsIAmLeader() {
		fmt.Println("[WARN] Only allow Leader to append log directly")
		return
	}
	l.appendOplog(log)
	l.syncOplogs()
}

func (l *RaftLikeLogger) commitLog() {
	logfileLock.Lock()
	defer logfileLock.Unlock()
	dataFile, err := os.Create(logfilePath)
	if err != nil {
		fmt.Printf("[ERROR] Create log failed. File path %s", logfilePath)
	}
	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(l.Logpipe)
	defer dataFile.Close()
}
func (l *RaftLikeLogger) loadLog() {
	// open data file
	dataFile, err := os.Open(logfilePath)

	if err != nil {

		fmt.Printf("[WARN] Unable to load log file at path %s", logfilePath)
		fmt.Println(err)
		return
	}

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&l.Logpipe)

	if err != nil {
		fmt.Println(err)
		return
	}

	dataFile.Close()

	fmt.Printf("[INFO] Loaded log data from file %s", logfilePath)
}

func (l *RaftLikeLogger) getOplogByOffset(offset int) (Oplog, error) {

	for lastIdx := len(l.Logpipe) - 1; lastIdx > 0; lastIdx-- {
		cursor := l.Logpipe[lastIdx]
		if cursor.LogOffset == offset {
			return cursor, nil
		}
	}
	return Oplog{}, errors.New("[ERROR] Unable to find a valid Oplog matching offset " + strconv.Itoa(offset))
}

func (l *RaftLikeLogger) syncOplogs() {

	if !IsIAmLeader() {
		fmt.Println("[WARN] Only allow Leader to sync members")
		return
	}

	lenLogData := len(l.Logpipe)

	if lenLogData < 1 {
		fmt.Println("[INFO] No log data found")
		return
	}

	ForEachMember(
		func(member Member, isLeader bool) {

			if isLeader {
				return
			}

			maxattempt := lenLogData - 1
			oplog, err := l.getOplogByOffset(lenLogData - 1)

			if err != nil {
				return
			}

			validLogOffset := oplog.LogOffset

			errMsg := "[WARN] Unable to syncOplogs to node `" + member.ID + "` with IP: " + member.IP
			for {

				if maxattempt < 0 {
					fmt.Println(errMsg)

					break
				}

				oplog, err := l.getOplogByOffset(validLogOffset)

				if err != nil {
					return
				}

				err = callRPCSyncLog(oplog, member.IP, &validLogOffset)
				if err == nil {
					break
				}

				fmt.Println(errMsg + ". Retry. Max attempts left: " + strconv.Itoa(maxattempt))

				maxattempt--

				time.Sleep(100 * time.Millisecond)
			}

		},
	)

}

func (l *RaftLikeLogger) AppendLog(oplog Oplog, replyValidOffset *int) error {
	myself := getMyself()
	if IsIAmLeader() {
		*replyValidOffset = -1
		return errors.New(
			fmt.Sprintf("[ERROR] I am leader. Do not accept new log from RPC. My NodeId %s, my IP %s", myself.ID, myself.IP),
		)
	}

	lenLogpipe := len(l.Logpipe)

	if lenLogpipe > 1 {

		acceptableMaxLogoffset := l.Logpipe[lenLogpipe-1].LogOffset + int(1)
		if oplog.LogOffset > acceptableMaxLogoffset {
			*replyValidOffset = acceptableMaxLogoffset
			return errors.New(
				fmt.Sprintf("[WARN] I don't have the latest log. Please send me log from logoffset", myself.ID, myself.IP),
			)
		}

	}

	l.appendOplog(oplog)

	return nil
}

func NewLogger() RaftLikeLogger {
	raftLikeLogger = RaftLikeLogger{}
	return raftLikeLogger
}

func RPCServeLeaderSyncLog() {

	rpc.RegisterRPC(raftLikeLogger, "ENV_PORT_LOGGER_SYNC")
}

func callRPCSyncLog(oplog Oplog, ip string, replyValidOffset *int) error {

	err := rpc.CallRPC(
		ip,
		"ENV_PORT_MEMBERSHIP_SYNC",
		"RaftLikeLogger.AppendLog",
		oplog,
		&replyValidOffset,
	)

	return err
}

func init() {
	raftLikeLogger = NewLogger()
	raftLikeLogger.syncOplogs()
}
