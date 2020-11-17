#!/bin/bash

# DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# cd $DIR

dockerNetworkName="wordcounter"

createSharedNetwork(){
	isCreated=$(docker network  ls | grep ${dockerNetworkName})
	if [[ -z "$isCreated" ]];then
		docker network create ${dockerNetworkName}
	fi
}

killAllGolangDocker(){
	docker rm $(docker stop $(docker ps -a -q --filter ancestor=golang --format="{{.ID}}"))
}

bringUpContainer(){
	containerCount=$1
	
	killAllGolangDocker

	createSharedNetwork

	for (( i=0; i<${containerCount}; i++ ))
	do
		exporsePort="-p 3000:3000 -p 8080:8080"
		if [[ $i -gt 0 ]];then
			exporsePort=""
		fi
		# cmd="docker run --net=bridge --rm -d -v $(pwd):/root/src/wordcounter ${exporsePort} --entrypoint=/root/src/wordcounter/bin/build-run.sh golang &"
		cmd="docker run --net=${dockerNetworkName} --rm -d -v $(pwd):/root/src/wordcounter ${exporsePort} --entrypoint=/root/src/wordcounter/bin/build-run.sh golang &"
		echo $cmd
		containerId=$(eval $cmd)
		eval "docker logs -f $containerId" &

		sleep 5
		
	done

	docker network inspect $dockerNetworkName
}

main(){
	inst=$1
	if [[ -z "$inst" ]];then
		inst=1
	fi
	bringUpContainer $inst
}

main "$@"