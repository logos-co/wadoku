#!/bin/sh

usage(){
  echo "Usage: ./run.sh   < clean | metal | docker | kurtosis>"
  exit 1
}

filtr="filter"
lpush="lightpush"

prefix="waku"
docker_op_dir="/go/bin/out"
enclave="waku-enclave"

content_topic="80fc1f6b30b63bdd0a65df833f1da3fa"
duration="1000s"
iat="300ms"

sleep_time=5

clean(){
  parent=$(pwd)
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
  echo "\t\t.........................."
  echo "\t\tBEGINNING THE $1 RUN..."
  echo "\t\t.........................."
}

end() {
  echo "\t\t   $1 RUN DONE..."
}

metal_run() {
  parent=$(pwd)
  time="$(\date +"%Y-%m-%d-%Hh%Mm%Ss")"
  FTDIR="$(pwd)/data/metal/$time"
  LPDIR="$(pwd)/data/metal/$time"

  cd waku/$filtr
  [ -d $FTDIR ] || mkdir -p $FTDIR
  ofname="$FTDIR/$filtr.out"
  #echo $ofname
  echo -n "starting  $filtr... $pwd "
  ./$prefix-$filtr -o $ofname -d $duration -i $iat > $FTDIR/$filtr.log &  
  echo " done"
  cd ..
  cd lightpush
  sleep $sleep_time
  [ -d $LPDIR ] || mkdir -p $LPDIR
  ofname="$LPDIR/$lpush.out"
  #echo $ofnameelapsed
  echo -n "starting  $lpush... "
  ./$prefix-$lpush -o $ofname -d $duration -i $iat > $LPDIR/$lpush.log &
  echo "done"
  cd $parent
  echo "$(date): Waiting for the metal run to finish in $duration"
  wait
}


docker_run() {
  parent=$(pwd)
  time="$(\date +"%Y-%m-%d-%Hh%Mm%Ss")"
  FTDIR="$(pwd)/data/docker/$time"
  LPDIR="$(pwd)/data/docker/$time"

  [ -d $FTDIR ] || mkdir -p $FTDIR
  ofname="$FTDIR/$filtr.out"
  echo "docker run $filtr $ofname"

  docker rm $prefix-$filtr
  echo "(docker run --name "$prefix-$filtr"  "$prefix-$filtr:alpha" -o /go/bin/out/$filtr.out -d $duration -i $iat > $FTDIR/$filtr.log)"
  docker run --name "$prefix-$filtr"  "$prefix-$filtr:alpha" -o /go/bin/out/$filtr.out -d $duration -i $iat > $FTDIR/$filtr.log &
  echo "$prefix-$filtr is running as $prefix-$filtr"
  sleep $sleep_time
  # docker run '$prefix-$filtr:alpha' -o /go/bin/out/$filtr.out -d $duration -i $iat > $FTDIR/$filtr.log
  #filtr_cid="$(docker container ls | grep '$prefix-$filtr' | awk '{print $1}')"
  #docker run  --mount type=bind,source="$FTDIR",target=/go/bin/out  "$filtr:alpha"  -o /go/bin/out/$filtr.out -d $duration -i $iat > $FTDIR/$filtr.log &
 # docker run -d  --entry-point  --mount type=bind,source="$(pwd)/FTDIR",target=/go/bin/out $filtr:alpha 
  #./filter -o $ofname > filter.log &   
  [ -d $LPDIR ] || mkdir -p $LPDIR
  ofname="$LPDIR/$lpush.out"
  echo "docker run $lpush $ofname"
  docker rm $prefix-$lpush
  docker run --name "$prefix-$lpush"  "$prefix-$lpush:alpha" -o /go/bin/out/$lpush.out -d $duration -i $iat > $LPDIR/$lpush.log &
  echo "$prefix-$filtr is running as $prefix-$lpush"

 # echo "(docker run  '$prefix-$lpush:alpha' -o /go/bin/out/$lpush.out -d $duration -i $iat > $FTDIR/$filtr.log)"
 # docker run '$prefix-$lpush:alpha' -o /go/bin/out/$lpush.out -d $duration -i $iat > $FTDIR/$filtr.log

 # sleep 5
 # lpush_cid="$(docker container ls | grep '$prefix-$filtr' | awk '{print $1}')"
#  echo "$prefix-$lpush is running as $lpush_cid"
  #docker run  --mount type=bind,source="$LPDIR",target=/go/bin/out  "$lpush:alpha"  -o /go/bin/out/$lpush.out -d $duration -i $iat > $LPDIR/$lpush.log &
  cd $parent
  echo "$(date): Waiting for the docker run to finish in $duration"
  status_code="$(docker container wait $prefix-$filtr $prefix-$lpush)"
  echo "Status code of docker run: ${status_code}"
  echo "$(date): copying output files from docker"
  docker cp "$filtr_cid:/go/bin/out/$filtr.out" $FTDIR
  docker cp "$lpush_cid:/go/bin/out/$lpush.out" $LPDIR
 # docker cp "$lpush_cid:/go/bin/out" $LPDIR
}

kurtosis_run() {
  parent=$(pwd)
  time="$(\date +"%Y-%m-%d-%Hh%Mm%Ss")"
  FTDIR="$(pwd)/data/kurtosis/$time"
  LPDIR="$(pwd)/data/kurtosis/$time"
  
  [ -d $FTDIR ] || mkdir -p $FTDIR
  [ -d $LPDIR ] || mkdir -p $LPDIR
  #cd waku
  kurtosis clean -a
  docker rm $prefix-$filtr
  docker rm $prefix-$lpush
  kurtosis run --enclave-id $enclave main.star '{"config":"github.com/logos-co/wadoku/waku/config.json"}' > $FTDIR/kurtosis_output.log
  sleep 5 
  filtr_suffix="$(kurtosis enclave inspect $enclave | grep $prefix-$filtr | cut -f 1 -d ' ')"
  lpush_suffix="$(kurtosis enclave inspect $enclave | grep $prefix-$lpush | cut -f 1 -d ' ')"
  filtr_cid="$enclave--user-service--$filtr_suffix"
  lpush_cid="$enclave--user-service--$lpush_suffix"
  echo "created $filtr_cid, $lpush_cid..."
  echo "$(date): Waiting for the kurtosis run to finish in $duration"
  status_code="$(docker container wait $filtr_cid $lpush_cid)"
  echo "Status code of the kurtosis run: ${status_code}"
  docker cp "$filtr_cid:/go/bin/out/$filtr.out" $FTDIR
  docker cp "$lpush_cid:/go/bin/out/$lpush.out" $LPDIR
  cd $parent
  #filtr_cid="$(docker container ls | grep '$prefix-$filtr' | awk '{print $1}')"
  #lpush_cid="$(docker container ls | grep '$prefix-$lpush' | awk '{print $1}')"
}

echo "$# $1"
[ 1 -eq "$#" ] || usage
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
  start $1 
  kurtosis_run
  end $1 
else
  usage
fi


#[ 'kurtosis' = '$1' ] || run_kurtosis


#docker run -d   --mount type=bind,source="$(pwd)/target",target=/go/bin\  lightpush:alpha 
