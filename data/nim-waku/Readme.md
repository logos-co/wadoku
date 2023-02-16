The folder houses the data for 10 node runs with varying message loads (7, 10, 20, 40, 80, 160, 320 msgs/sec). Each run is performed *thrice* with the same network data and message rate. Each file in the run folder (numbered 1, 2, 3) houses  files that contain the respective wakunode/docker's resource consumption, sampled at regular intevals (1s).

For instance, 10nodes/160msgpsec/2/nim-waku-373ae64b3935 contains the docker samples for the 2nd run involving 10 nwakunodes/dockers, with 160 messages per sec.

The Go waku runs will be added once WSL figures out a way to talk to go-waku nodes

plots.py takes a one nim-waku sample file in the run folder and converts it a panda df. 
