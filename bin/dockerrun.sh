#!/bin/bash

# DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# cd $DIR

killAllGolangDocker(){
	docker rm $(docker stop $(docker ps -a -q --filter ancestor=golang --format="{{.ID}}"))
}

bringUpContainer(){
	containerCount=$1
	
	killAllGolangDocker

	for (( i=0; i<${containerCount}; i++ ))
	do
		exporsePort="-p 8080:8080"
		if [[ $i -gt 0 ]];then
			exporsePort=""
		fi
		cmd="docker run --rm -d -v $(pwd):/root/src/wordcounter ${exporsePort} --entrypoint=/root/src/wordcounter/bin/build-run.sh golang &"
		echo $cmd
		containerId=$(eval $cmd)
		eval "docker logs -f $containerId" &

		sleep 5
		
	done
}

main(){
	inst=$1
	if [[ -z "$inst" ]];then
		inst=1
	fi
	bringUpContainer $inst
}

main "$@"