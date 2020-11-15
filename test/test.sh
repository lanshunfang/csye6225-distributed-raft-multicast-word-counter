#!/bin/bash

#
# e.g. ./test.sh iso_8859-1.txt
#
txtPath=$0

gorpcPort=8081

fetchLeaderIP(){

	leaderIP=$(curlGoRpcCall "localhost" "Membership.ReportLeaderIP" "")
	if [[ $? != 0 ]];then
		sleep 2
		fetchLeaderIP
	else
		echo $leaderIP
	fi
}

curlGoRpcCall(){
	ip=$1
	method=$2
	param=$3
	cmd="curl -X CONNECT --url ${ip}:${gorpcPort}/_goRPC_ -d '{\"method\":\"'$method'\",\"params\":[\"$param\"],\"id\": 0}'"
	>&2 echo $cmd
	eval $cmd
	# curl -X CONNECT \
	# 	--url ${ip}:${gorpcPort}/_goRPC_ \ 
	# 	-d '{"method":"'$method'","params":['$param'],"id": 0}'
}

main(){

	if [[ -z "$txtPath" ]];then
		echo "[INFO] e.g. ./test.sh ./iso_8859-1.txt"
		exit
	fi

	leaderIP=$(fetchLeaderIP)
	mytext=$(cat $txtPath)

	curlGoRpcCall "$leaderIP" "WordCount.HTTPHandle" "${mytext}"
}

main "$@"