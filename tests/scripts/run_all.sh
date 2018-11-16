#!/usr/bin/env bash

declare -a scripts=(
"test_1_ring.sh"
"test_2_ring.sh"
"test_file_sharing.sh"
"test_send_private_message.sh"
)

DEBUG=${1:-"false"}
RACE=${2:-"false"}

## now loop through the above array
for s in "${scripts[@]}"
do
    if [[ "$DEBUG" == "true" ]] ; then
        echo $s
    fi
    sh $s $DEBUG $RACE
done
