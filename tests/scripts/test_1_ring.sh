#!/usr/bin/env bash
# received 'debug' mode (boolean) as first argument ($1) and race mode (boolean) as ($2)

. ./setup_path.sh $0 $1 $2

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipPort=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &

for i in `seq 1 10`;
do
	outFileName="$outPath/$gossipName.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$gossipName -simple -peers=$peer > $outFileName &
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$gossipName running at UIPort $UIPort and gossipPort $gossipPort"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	gossipName=$(echo "$gossipName" | tr "A-Y" "B-Z")
done

./client/client -UIPort=12349 -msg=$message
./client/client -UIPort=12346 -msg=$message2
sleep 3

for pid in $(pgrep -f Peerster); do
    kill $pid
    wait $pid 2> /dev/null
done

#testing
failed="F"
echo "${BLUE}###CHECK client message${NC}"

if !(grep -q "CLIENT MESSAGE $message" "$outPath/E.out") ; then
	failed="T"
fi

if !(grep -q "CLIENT MESSAGE $message2" "$outPath/B.out") ; then
  failed="T"
fi

if [[ "$failed" == "T" ]] ; then
	echo "${RED}FAILED${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
fi

# echo "${outputFiles[@]}"

echo "${BLUE}###CHECK simple message${NC}"

gossipPort=5000
for i in `seq 0 9`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
		relayPort=5009
	fi
	nextPort=$((($gossipPort+1)%10+5000))
	msgLine="SIMPLE MESSAGE origin E from 127.0.0.1:$relayPort contents $message"
	msgLine2="SIMPLE MESSAGE origin B from 127.0.0.1:$relayPort contents $message2"
	peersLine="127.0.0.1:$nextPort,127.0.0.1:$relayPort"
	gossipPort=$(($gossipPort+1))
	if !(grep -q "$msgLine" "${outputFiles[$i]}") ; then
        if [[ "$DEBUG" == "true" ]] ; then
            echo "$msgLine MISSING in ${outputFiles[$i]}"
        fi
   		failed="T"
	fi
	if !(grep -q "$peersLine" "${outputFiles[$i]}") ; then
        if [[ "$DEBUG" == "true" ]] ; then
            echo "$peersLine MISSING in ${outputFiles[$i]}"
        fi
        failed="T"
    fi
	if !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
        if [[ "$DEBUG" == "true" ]] ; then
            echo "$msgLine2 MISSING in ${outputFiles[$i]}"
        fi
        failed="T"
    fi
done

if [[ "$failed" == "T" ]] ; then
    echo "${RED}***FAILED***${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
fi
