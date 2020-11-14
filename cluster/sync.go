package cluster

func syncLog() {
	raftLikeLogger := GetLogger()
	raftLikeLogger.syncOplogs()
}
