#!/bin/bash

# shellcheck disable=SC2164
# shellcheck disable=SC2034
# shellcheck disable=SC2126
# shellcheck disable=SC2006
# shellcheck disable=SC2010
# shellcheck disable=SC2103
# shellcheck disable=SC2050

#go run cleanDB.go
rm -rf et


cd configs

fileNum=`ls -l |grep "^-"|wc -l`

cd ..
#rm SealABCTest

go build SealABCTest.go SealABCConfig.go

path=$(pwd)

for (( i=1; i<=fileNum; i++ )) do
    osascript -e 'tell app "Terminal"
		do script "cd '"${path}"'; ./SealABCTest --config ./configs/config-1-'"${i}"'.json"
	end tell'
	  if [ "$i" == 1 ]
	  then
	    echo 'start'
      sleep 3
    fi

done


