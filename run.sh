#!/bin/sh

cool_off_time=120
sleep_time=$cool_off_time
prefix=$1

run(){
  duration=$1
  echo $(date), "starting for $duration, with output at $prefix"
  START=$(date +%s)

  sleep $sleep_time
  ./run-waku.sh metal       $duration >> $prefix-$duration 2>&1
  sleep $sleep_time
  ./run-waku.sh  docker     $duration >> $prefix-$duration 2>&1
  sleep $sleep_time
  ./run-waku.sh  kurtosis   $duration >> $prefix-$duration 2>&1


  STOP=$(date +%s)
  DIFF=$(( $STOP - $START ))
  echo $(date), "$duration done"

  echo "$duration took $DIFF secs; metal, docker, kurtosis; sleep_time=$sleep_time)"  >> $1-$duration 2>&1
  echo "$duration took $DIFF secs; metal, docker, kurtosis; sleep_time=$sleep_time)"
}

echo "HERE WO GO!"
run 32s
run 64s
run 128s
run 256s
run 512s
run 1024s
