#!/bin/sh

if [ $1 ]; then
    delay=$1
else
    delay=1
fi

while read LINE; do
   echo ${LINE}
   sleep $delay
done

exit 0