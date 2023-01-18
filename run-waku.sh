#!/bin/sh

usage(){
  echo "Usage: ./run.sh   < clean | metal | docker | kurtosis>"
  exit 1
}

# two tests per run
filtr="filter"
lpush="lightpush"

# file/dir locations
prefix="waku"
docker_op_dir="/go/bin/out"
enclave="enclave-waku"
host_output_dir="/home1/mimosa/runs"

# params for the run
ctopic="8fc1f6b30b63bdd0a65df833f1da3fa"  # default ctopic, a md5 
duration="100s"                           # duration of the run
iat="100ms"                               # pub msg inter-arrival-time
loglvl="info"                             # log level

duration=$2
#iat=$3

sleep_time=30
time=""

unique="XYZ"

clean(){
  parent=$(pwd)           # no pushd / popd
  #rm -rf ./data
  cd waku/$filtr
  rm $prefix-$filtr
  cd ..
  cd $lpush
  rm $prefix-$lpush
  cd $parent
}


build_metal() {
  parent=$(pwd)
  cd waku/$filtr
  make
  cd ..
  cd $lpush
  make
  cd $parent
}


build_docker() {
  parent=$(pwd)
  cd waku/$filtr
  make docker
  cd ..
  cd $lpush
  make docker
  cd $parent
  wait
}


start() {
  time="$(\date +"%Y-%m-%d-%Hh%Mm%Ss")"
  ctopic="$duration-$iat-$time-$unique"
  echo "\t\t.............................................."
  echo "\t\tBEGINNING THE $1 RUN @ $time"
  echo "\t\tparams: $ctopic"
  echo "\t\t.............................................."
}


end() {
  echo "\t\t   $1 RUN DONE @ $(\date +"%Y-%m-%d-%Hh%Mm%Ss") "
  echo "\t\t   params: $ctopic"
}


metal_run() {
  parent=$(pwd)
  #time="$(\date +"%Y-%m-%d-%Hh%Mm%Ss")"
  FTDIR="$host_output_dir/data/metal/$ctopic"
  LPDIR="$host_output_dir/data/metal/$ctopic"

  killall -9 $prefix-$filtr $prefix-$lpush
  cd waku/$filtr
  [ -d $FTDIR ] || mkdir -p $FTDIR
  ofname="$FTDIR/$filtr.out"
  echo -n "starting  $filtr... $pwd "
  ./$prefix-$filtr -o $ofname -d $duration -i $iat  -l $loglvl -c $ctopic > $FTDIR/$filtr.log &  
  echo " done"
  cd ..
  cd lightpush

  sleep $sleep_time

  [ -d $LPDIR ] || mkdir -p $LPDIR
  ofname="$LPDIR/$lpush.out"
  echo -n "starting  $lpush... "
  ./$prefix-$lpush -o $ofname -d $duration -i $iat -l $loglvl -c $ctopic > $LPDIR/$lpush.log &
  echo "done"
  cd $parent
  echo "$(date): Waiting for the metal run to finish in $duration"
  wait # wait for runs to complete
}


docker_run() {
  parent=$(pwd)
  #time="$(\date +"%Y-%m-%d-%Hh%Mm%Ss")"
  FTDIR="$host_output_dir/data/docker/$ctopic"
  LPDIR="$host_output_dir/data/docker/$ctopic"
  #FTDIR="$host_output_dir/data/docker/$time"
  #LPDIR="$host_output_dir/data/docker/$time"

  #darg= "--network=host"

   
  [ -d $LPDIR ] || mkdir -p $LPDIR
  ofname="$LPDIR/$lpush.out"
  echo "docker: stopping and removing $prefix-$lpush"
  docker stop $prefix-$lpush
  docker rm $prefix-$lpush
  echo "docker: starting $prefix-$lpush"
  echo " docker run --name "$prefix-$lpush" -d=true "$prefix-$lpush:alpha" -o "$docker_op_dir/$lpush.out" -d $duration -i $iat -l $loglvl -c $ctopic"
  docker run --name "$prefix-$lpush" -d=true "$prefix-$lpush:alpha" -o "$docker_op_dir/$lpush.out" -d $duration -i $iat -l $loglvl -c $ctopic 

  #sleep $sleep_time

  [ -d $FTDIR ] || mkdir -p $FTDIR
  ofname="$FTDIR/$filtr.out"
  echo "docker: stopping and removing $prefix-$filtr"
  docker stop $prefix-$filtr
  docker rm $prefix-$filtr
  echo "docker: starting $prefix-$filtr"
  echo "docker run --name "$prefix-$filtr" -d=true "$prefix-$filtr:alpha" -o "$docker_op_dir/$filtr.out" -d $duration -l $loglvl -i $iat -c $ctopic "
  docker run --name "$prefix-$filtr" -d=true  "$prefix-$filtr:alpha" -o "$docker_op_dir/$filtr.out" -d $duration -l $loglvl -i $iat -c $ctopic 
  echo "$prefix-$filtr is running as $prefix-$filtr"


  # now wait for runs to complete...
  echo "$(date): Waiting for the docker run to finish in $duration"
  status_code="$(docker container wait $prefix-$filtr $prefix-$lpush)"

  # copy the output files
  echo "Status code of docker run: ${status_code}"
  echo "$(date): copying output files from docker"
  docker logs $prefix-$filtr > "$FTDIR/$filtr.log"
  docker logs $prefix-$lpush > "$LPDIR/$lpush.log"
  docker cp "$prefix-$filtr:$docker_op_dir/$filtr.out" $FTDIR
  docker cp "$prefix-$lpush:$docker_op_dir/$lpush.out" $LPDIR
  cd $parent
}


kurtosis_run() {
  parent=$(pwd)
  #time="$(\date +"%Y-%m-%d-%Hh%Mm%Ss")"
  FTDIR="$host_output_dir/data/kurtosis/$ctopic"
  LPDIR="$host_output_dir/data/kurtosis/$ctopic"
  #FTDIR="$host_output_dir/data/kurtosis/$time"
  #LPDIR="$host_output_dir/data/kurtosis/$time"
  
  [ -d $FTDIR ] || mkdir -p $FTDIR
  [ -d $LPDIR ] || mkdir -p $LPDIR
  #cd waku
  kurtosis clean -a
  docker rm $prefix-$filtr
  docker rm $prefix-$lpush

  # generate the config.json
  echo "{
  \"output_file\": \"output.out\",
  \"duration\": \"$duration\",
  \"iat\": \"$iat\",
  \"content_topic\": \"$ctopic\",
  \"log_level\": \"$loglvl\",
  \"log_file\": \"output.log\"
}" > waku/config.json

  kurtosis run --enclave-id $enclave . '{"config":"github.com/logos-co/wadoku/waku/config.json"}' > $FTDIR/kurtosis_output.log
  sleep $sleep_time 
  filtr_suffix="$(kurtosis enclave inspect $enclave | grep $prefix-$filtr | cut -f 1 -d ' ')"
  lpush_suffix="$(kurtosis enclave inspect $enclave | grep $prefix-$lpush | cut -f 1 -d ' ')"

  # now wait for runs to complete...
  filtr_cid="$enclave--user-service--$filtr_suffix"
  lpush_cid="$enclave--user-service--$lpush_suffix"
  echo "created $filtr_cid, $lpush_cid..."
  echo "$(date): Waiting for the kurtosis ($filtr_cid, $lpush_cid ) run to finish in $duration"
  status_code="$(docker container wait $filtr_cid $lpush_cid)"

  # copy the output files
  docker cp "$filtr_cid:/go/bin/out/$filtr.out" $FTDIR
  docker logs $filtr_cid > $FTDIR/$filtr.log
  docker cp "$lpush_cid:/go/bin/out/$lpush.out" $LPDIR
  docker logs $lpush_cid > $LPDIR/$lpush.log
  #kurtosis enclave dump $enclave $FTDIR/kurtosis_dump
  echo "Status code of the kurtosis run: ${status_code}"
  cd $parent
}


echo "$# $1 $2"
[ 2 -eq "$#" ] || usage
[ metal != $1 -a docker != $1 -a kurtosis != $1 -a clean != $1 ] && usage


if [ clean = $1 ]; then
  start $1 
  clean
  end $1 
elif [ metal = $1 ]; then
  build_metal
  start $1 
  metal_run
  end $1 
elif [ docker = $1 ]; then
  build_metal
  build_docker
  start $1 
  docker_run
  end $1 
elif [ kurtosis = $1 ]; then
  build_metal
  build_docker
  clean       # stay under the 4mb limit imposed by gRPC
  start $1 
  kurtosis_run
  end $1 
else
  usage
fi

# debris

# docker run '$prefix-$filtr:alpha' -o /go/bin/out/$filtr.out -d $duration -i $iat > $FTDIR/$filtr.log
  #filtr_cid="$(docker container ls | grep '$prefix-$filtr' | awk '{print $1}')"
  #docker run  --mount type=bind,source="$FTDIR",target=/go/bin/out  "$filtr:alpha"  -o /go/bin/out/$filtr.out -d $duration -i $iat > $FTDIR/$filtr.log &
 # docker run -d  --entry-point  --mount type=bind,source="$(pwd)/FTDIR",target=/go/bin/out $filtr:alpha 
  #./filter -o $ofname > filter.log &

   # echo "(docker run  '$prefix-$lpush:alpha' -o /go/bin/out/$lpush.out -d $duration -i $iat > $FTDIR/$filtr.log)"
 # docker run '$prefix-$lpush:alpha' -o /go/bin/out/$lpush.out -d $duration -i $iat > $FTDIR/$filtr.log

 # sleep 5
 # lpush_cid="$(docker container ls | grep '$prefix-$filtr' | awk '{print $1}')"
#  echo "$prefix-$lpush is running as $lpush_cid"
  #docker run  --mount type=bind,source="$LPDIR",target=/go/bin/out  "$lpush:alpha"  -o /go/bin/out/$lpush.out -d $duration -i $iat > $LPDIR/$lpush.log &

  #filtr_cid="$(docker container ls | grep '$prefix-$filtr' | awk '{print $1}')"
  #lpush_cid="$(docker container ls | grep '$prefix-$lpush' | awk '{print $1}')"
