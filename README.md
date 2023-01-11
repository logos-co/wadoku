# wadoku = `wa`ku  + `do`cker + `ku`rtosis runs
This repo houses the code for initial runs to assess overhead of running waku nodes on bare metal, docker or kurtosis. The waku nodes are specifically chosen to be non-full nodes to minimise protocol cross talk.

Bare metal run is done, docker run is done. kurtosis run is more or less done. Will run some sanity check runs, and add the plots soonish .

the `run_waku.sh' takes 4 options: clean | metal | docker | kurtosis. clean will clear out the bianries, and the rest 3 specify the type of run. metal will comiple and run the pub/sub on local machine. docker will build the docker imaeg and run it inside docker in the local machine. kurtosis will build docker and run this docker inside the kurtosis enclave. From build to run to collecting data, everything is automated. Plots will be automated as well.
