#!/usr/bin/env bash
# received 'debug' mode (boolean) as first arguement ($1)

. ./setup_path.sh $0 $1

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

sleep 4

./client/client -UIPort=12345 -msg="$private_msg" -dest=F

sleep 2
#pkill -f Peerster


#testing
failed="F"

echo "${BLUE}###CHECK private message only displayed to F${NC}"

private_msg_received="PRIVATE origin A hop-limit 6 contents $private_msg"

if (grep -q "PRIVATE" "$outPath/B.out") ; then
	failed="T"
fi
if (grep -q "PRIVATE" "$outPath/C.out") ; then
	failed="T"
fi
if (grep -q "PRIVATE" "$outPath/D.out") ; then
	failed="T"
fi
if (grep -q "PRIVATE" "$outPath/E.out") ; then
	failed="T"
fi
if !(grep -q "$private_msg_received" "$outPath/F.out") ; then
	failed="T"
fi

if [[ "$failed" == "T" ]] ; then
	echo "${RED}***FAILED***${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
#	rm *.out
fi

pkill -f Peerster
