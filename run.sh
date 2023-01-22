#!/bin/sh

cool_off_time=120
sleep_time=$cool_off_time
prefix=$1
time="$(\date +"%Y-%m-%d-%Hh%Mm%Ss")"

run(){
  duration=$1
  echo $time, "starting $duration run with output at $prefix"
  START=$(date +%s)

  sleep $sleep_time
  ./run-waku.sh metal       $duration $time lf >> $prefix-$duration-$time 2>&1
  sleep $sleep_time
  ./run-waku.sh  docker     $duration $time lf >> $prefix-$duration-$time 2>&1
  sleep $sleep_time
  ./run-waku.sh  kurtosis   $duration $time lf >> $prefix-$duration-$time 2>&1

  STOP=$(date +%s)
  DIFF=$(( $STOP - $START ))
  echo $time, "lf: $duration done"

  echo "LF: $duration took $DIFF secs; metal, docker, kurtosis; sleep_time=$sleep_time)"  >> $prefix-$duration-$time 2>&1
  echo "LF: $duration took $DIFF secs; metal, docker, kurtosis; sleep_time=$sleep_time)"

  sleep $sleep_time
  ./run-waku.sh metal       $duration $time ps >> $prefix-$duration-$time 2>&1
  sleep $sleep_time
  ./run-waku.sh  docker     $duration $time ps >> $prefix-$duration-$time 2>&1
  sleep $sleep_time
  ./run-waku.sh  kurtosis   $duration $time ps >> $prefix-$duration-$time 2>&1


  STOP=$(date +%s)
  DIFF=$(( $STOP - $START ))
  echo $time, "ps: $duration done"

  echo "LF+PS: $duration took $DIFF secs; lf + ps; metal, docker, kurtosis; sleep_time=$sleep_time)"  >> $prefix-$duration-$time 2>&1
  echo "LF+PS: $duration took $DIFF secs; lf + ps; metal, docker, kurtosis; sleep_time=$sleep_time)"
}


echo "$time: HERE WO GO!"
run 32s
run 64s
run 128s
run 256s
run 512s
run 1024s
run 2048s
