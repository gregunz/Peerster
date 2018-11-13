#!/usr/bin/env bash
# expected file path of script as first argument ($1) and debug as second ($2)
# it sets $projectPath and $outPath variables (+ create dir if not exist)
# it also builds the scripts and cd to projectPath
# it also sets some variables such as colors or ports

scriptPath=$1
filename=$(basename "$scriptPath" ".sh")

projectPath="$GOPATH/src/github.com/gregunz/Peerster"
outPath="$projectPath/tests/scripts/out/$filename"
mkdir -p $outPath

BLUE='\033[0;34m'
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
DEBUG=${2:-"true"}

outputFiles=()

UIPort=12345
gossipPort=5000
gossipName='A'

message=Weather_is_clear
message2=Winter_is_coming

message_c1_1=Weather_is_clear
message_c2_1=Winter_is_coming
message_c1_2=No_clouds_really
message_c2_2=Let\'s_go_skiing
message_c3=Is_anybody_here?

private_msg="hello, this is a private message :P"

cd "$projectPath"
go build
cd client
go build
cd "$projectPath"


if [[ "$DEBUG" == "true" ]] ; then
	echo "everythings set up! debug mode = $DEBUG"
fi
