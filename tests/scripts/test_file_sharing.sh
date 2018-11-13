#!/usr/bin/env bash
# received 'debug' mode (boolean) as first arguement ($1)

. ./setup_path.sh $0 $1


if [[ "$DEBUG" == "true" ]] ; then
	echo "deleting file if it exists"
	rm -v "$projectPath/_Downloads/hamlet_F.txt"
fi

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipAddr=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &

for i in `seq 1 6`;
do
	outFileName="$outPath/$gossipName.out"
	peerPort=$((($i)+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$gossipName -peers=$peer -rtimer=1> $outFileName &
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$gossipName running at UIPort $UIPort and gossipPort $gossipPort and peer $peer"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	gossipName=$(echo "$gossipName" | tr "A-Y" "B-Z")
done

./client/client -UIPort=12345 -file=hamlet.txt

sleep 4

./client/client -UIPort=12350 -file=hamlet_F.txt -dest=A -request=d0fdefd8f0e7d259b36b237cc967f06320353700e15354048634a6cd8bda4e59

sleep 15

#pkill -f Peerster > /dev/null

echo "${BLUE}###CHECK hamlet file transfered${NC}"

diff_res=$(diff _SharedFiles/hamlet.txt _Downloads/hamlet_F.txt)

if [[ ! -f ./_Downloads/hamlet_F.txt ]]; then
	echo "${RED}***FAILED***${NC}"
else
    if [[ "$diff_res" != "" ]]; then
        echo "${RED}***FAILED***${NC}"
    else
        echo "${GREEN}***PASSED***${NC}"
    fi

#    rm _Downloads/hamlet_F.txt
#    rm _Downloads/.meta/hamlet_F.txt
fi

pkill -f Peerster
