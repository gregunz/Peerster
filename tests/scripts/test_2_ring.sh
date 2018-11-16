#!/usr/bin/env bash
# received 'debug' mode (boolean) as first argument ($1) and race mode (boolean) as ($2)

. ./setup_path.sh $0 $1 $2

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipAddr=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &

for i in `seq 1 10`;
do
	outFileName="$outPath/$gossipName.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$gossipName -peers=$peer > $outFileName &
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$gossipName running at UIPort $UIPort and gossipPort $gossipPort"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	gossipName=$(echo "$gossipName" | tr "A-Y" "B-Z")
done

./client/client -UIPort=12349 -msg=$message_c1_1
./client/client -UIPort=12346 -msg=$message_c2_1
./client/client -UIPort=12349 -msg=$message_c1_2
./client/client -UIPort=12346 -msg=$message_c2_2
./client/client -UIPort=12351 -msg=$message_c3

sleep 10

for pid in $(pgrep -f Peerster); do
    kill $pid
    wait $pid 2> /dev/null
done


#testing
failed="F"

echo "${BLUE}###CHECK that client messages arrived${NC}"

if !(grep -q "CLIENT MESSAGE $message_c1_1" "$outPath/E.out") ; then
	failed="T"
	if [[ "$DEBUG" == "true" ]] ; then
		echo "CLIENT MESSAGE $message_c1_1 MISSING in $outPath/E.out"
	fi
fi

if !(grep -q "CLIENT MESSAGE $message_c1_2" "$outPath/E.out") ; then
	failed="T"
	if [[ "$DEBUG" == "true" ]] ; then
		echo "CLIENT MESSAGE $message_c1_2 MISSING in $outPath/E.out"
	fi
fi

if !(grep -q "CLIENT MESSAGE $message_c2_1" "$outPath/B.out") ; then
	failed="T"
	if [[ "$DEBUG" == "true" ]] ; then
		echo "CLIENT MESSAGE $message_c2_1 MISSING in $outPath/B.out"
	fi
fi

if !(grep -q "CLIENT MESSAGE $message_c2_2" "$outPath/B.out") ; then
	failed="T"
	if [[ "$DEBUG" == "true" ]] ; then
		echo "CLIENT MESSAGE $message_c2_2 MISSING in $outPath/B.out"
	fi
fi

if !(grep -q "CLIENT MESSAGE $message_c3" "$outPath/G.out") ; then
	failed="T"
	if [[ "$DEBUG" == "true" ]] ; then
		echo "CLIENT MESSAGE $message_c3 MISSING in $outPath/G.out"
	fi
fi

if [[ "$failed" == "T" ]] ; then
	echo "${RED}***FAILED***${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo "${BLUE}###CHECK rumor messages ${NC}"

gossipPort=5000
for i in `seq 0 9`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
		relayPort=5009
	fi
	nextPort=$((($gossipPort+1)%10+5000))
	msgLine1="RUMOR origin E from 127.0.0.1:[0-9]{4} ID 1 contents $message_c1_1"
	msgLine2="RUMOR origin E from 127.0.0.1:[0-9]{4} ID 2 contents $message_c1_2"
	msgLine3="RUMOR origin B from 127.0.0.1:[0-9]{4} ID 1 contents $message_c2_1"
	msgLine4="RUMOR origin B from 127.0.0.1:[0-9]{4} ID 2 contents $message_c2_2"
	msgLine5="RUMOR origin G from 127.0.0.1:[0-9]{4} ID 1 contents $message_c3"

	if [[ "$gossipPort" != 5004 ]] ; then
		if !(grep -Eq "$msgLine1" "${outputFiles[$i]}") ; then
			if [[ "$DEBUG" == "true" ]] ; then
				echo "$msgLine1 missing in ${outputFiles[$i]}"
			fi
			failed="T"
		fi
		if !(grep -Eq "$msgLine2" "${outputFiles[$i]}") ; then
			if [[ "$DEBUG" == "true" ]] ; then
				echo "$msgLine2 missing in ${outputFiles[$i]}"
			fi
			failed="T"
		fi
	fi

	if [[ "$gossipPort" != 5001 ]] ; then
		if !(grep -Eq "$msgLine3" "${outputFiles[$i]}") ; then
			if [[ "$DEBUG" == "true" ]] ; then
				echo "$msgLine3 missing in ${outputFiles[$i]}"
			fi
			failed="T"
		fi
		if !(grep -Eq "$msgLine4" "${outputFiles[$i]}") ; then
			if [[ "$DEBUG" == "true" ]] ; then
				echo "$msgLine4 missing in ${outputFiles[$i]}"
			fi
			failed="T"
		fi
	fi

	if [[ "$gossipPort" != 5006 ]] ; then
		if !(grep -Eq "$msgLine5" "${outputFiles[$i]}") ; then
			if [[ "$DEBUG" == "true" ]] ; then
				echo "$msgLine5 missing in ${outputFiles[$i]}"
			fi
			failed="T"
		fi
	fi
	gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
	echo "${RED}***FAILED***${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo "${BLUE}###CHECK mongering${NC}"
gossipPort=5000
for i in `seq 0 9`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
		relayPort=5009
	fi
	nextPort=$((($gossipPort+1)%10+5000))

	msgLine1="MONGERING with 127.0.0.1:$relayPort"
	msgLine2="MONGERING with 127.0.0.1:$nextPort"

	if !(grep -q "$msgLine1" "${outputFiles[$i]}") && !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
	echo "${RED}***FAILED***${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
fi


failed="F"
echo "${BLUE}###CHECK status messages ${NC}"
gossipPort=5000
for i in `seq 0 9`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
		relayPort=5009
	fi
	nextPort=$((($gossipPort+1)%10+5000))

	msgLine1="STATUS from 127.0.0.1:$relayPort"
	msgLine2="STATUS from 127.0.0.1:$nextPort"
	msgLine3="peer E nextID 3"
	msgLine4="peer B nextID 3"
	msgLine5="peer G nextID 2"

	if !(grep -q "$msgLine1" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	if !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	if !(grep -q "$msgLine3" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	if !(grep -q "$msgLine4" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	if !(grep -q "$msgLine5" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
	echo "${RED}***FAILED***${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo "${BLUE}###CHECK flipped coin${NC}"
gossipPort=5000
for i in `seq 0 9`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
		relayPort=5009
	fi
	nextPort=$((($gossipPort+1)%10+5000))

	msgLine1="FLIPPED COIN sending rumor to 127.0.0.1:$relayPort"
	msgLine2="FLIPPED COIN sending rumor to 127.0.0.1:$nextPort"

	if !(grep -q "$msgLine1" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	if !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	gossipPort=$(($gossipPort+1))

done

if [[ "$failed" == "T" ]] ; then
	echo "${RED}***FAILED***${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo "${BLUE}###CHECK in sync${NC}"
gossipPort=5000
for i in `seq 0 9`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
		relayPort=5009
	fi
	nextPort=$((($gossipPort+1)%10+5000))

	msgLine1="IN SYNC WITH 127.0.0.1:$relayPort"
	msgLine2="IN SYNC WITH 127.0.0.1:$nextPort"

	if !(grep -q "$msgLine1" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	if !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
	echo "${RED}***FAILED***${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo "${BLUE}###CHECK correct peers${NC}"
gossipPort=5000
for i in `seq 0 9`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
		relayPort=5009
	fi
	nextPort=$((($gossipPort+1)%10+5000))

	peersLine1="127.0.0.1:$relayPort,127.0.0.1:$nextPort"
	peersLine2="127.0.0.1:$nextPort,127.0.0.1:$relayPort"

	if !(grep -q "$peersLine1" "${outputFiles[$i]}") && !(grep -q "$peersLine2" "${outputFiles[$i]}") ; then
		failed="T"
	fi
	gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
	echo "${RED}***FAILED***${NC}"
else
	echo "${GREEN}***PASSED***${NC}"
fi
