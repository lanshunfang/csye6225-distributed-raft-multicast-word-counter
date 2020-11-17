#!/bin/bash

# DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# cd $DIR
# cd ..

echo "[INFO] Building golang package for Linux ARM binary"
# env GOOS=linux GOARCH=arm go build

cp -r /root/src/wordcounter /usr/local/go/src

cd /usr/local/go/src/wordcounter

# binaryTarget="./wordcounter"

go build

chmod +x ./wordcounter

./wordcounter

# if [[ ! -f "$binaryTarget" ]];then
# 	go build
# else
# 	echo "Skipped building as the file $binaryTarget exist"
# fi

# rm -rf /root/tmp
# mkdir -p /root/tmp
# cp $binaryTarget /root/tmp

# chmod +x /root/tmp/wordcounter

# /root/tmp/wordcounter