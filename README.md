# wadoku
This repo houses the code for initial runs to assess overhead of running waku nodes on bare metal, docker or kurtosis. The waku nodes are specifically chosen to be non-full nodes to minimise protocol cross talk.

Bare metal run is done, docker run is done. kurtosis run is more or less done. Will run some sanity check runs, and add the plots soonish .

the `run_waku.sh' takes 3 options, metal | docker | kurtosis, for each type of runs. From build to run to collecting data, everything is automated. Plots will be automated as well.
