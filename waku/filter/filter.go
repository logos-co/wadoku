package  main

import (
	"context"
	"flag"
	"fmt"
	"net"
  "bytes"
 // "math/rand"
  "strconv"
	"encoding/binary"
	"os"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/filter"
//	"github.com/waku-org/go-waku/waku/v2/utils"
	"github.com/logos-co/wadoku/waku/common"
	//"crypto/rand"
	//"encoding/hex"
	//"github.com/ethereum/go-ethereum/crypto"
	//"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	//"github.com/waku-org/go-waku/waku/v2/payload"
)

var log = logging.Logger("filter")
var pubSubTopic = protocol.DefaultPubsubTopic()
var conf = common.Config{}
var nodeType = "filter"
//const dnsDiscoveryUrl = "enrtree://AOGECG2SPND25EEFMAJ5WF3KSGJNSGV356DSTL2YVLLZWIV6SAYBM@prod.waku.nodes.status.im"
//const nameServer = "1.1.1.1" // your local dns provider might be blocking entr


func init() {
	// args
  fmt.Println("Populating CLI params...")
  common.ArgInit(&conf)
}

func main() {

	flag.Parse()

	// setup the log
	lvl, err := logging.LevelFromString(conf.LogLevel)
	if err != nil {
		panic(err)
	}
	logging.SetAllLoggers(lvl)

  tcpEndPoint :=  common.LocalHost +
                      ":" +
                      strconv.Itoa(common.StartPort + common.RandInt(0, common.PortRange))
	// create the waku node
	hostAddr, _ := net.ResolveTCPAddr("tcp", tcpEndPoint)
	ctx := context.Background()
	filterNode, err := node.New(ctx,
		//node.WithWakuRelay(),
		//node.WithNTP(),  // don't use NTP, fails at msec granularity    
		node.WithHostAddress(hostAddr),
		node.WithWakuFilter(false), // we do NOT want a full node
	)
	if err != nil {
		panic(err)
	}

  log.Info("CONFIG : ", conf)
	// find the list of full node fleet peers
  log.Info("attempting DNS discovery with: ", common.DnsDiscoveryUrl)
	nodes, err := dnsdisc.RetrieveNodes(ctx, common.DnsDiscoveryUrl, dnsdisc.WithNameserver(common.NameServer))
	if err != nil {
		panic(err.Error())
	}

	// connect to the first peer
	var nodeList []multiaddr.Multiaddr
	for _, n := range nodes {
		nodeList = append(nodeList, n.Addresses...)
	}
  log.Info("Discovered and connecting to: ", nodeList[0])
	peerID, err := nodeList[0].ValueForProtocol(multiaddr.P_P2P)
	if err != nil {
		log.Error("could not get peerID: ", err)
		panic(err)
	}

    // initialise the exponential back off
  retries, sleepTime := 0, common.ExpBackOffInit
  for {
    // try / retry
	  if err = filterNode.DialPeerWithMultiAddress(ctx, nodeList[0]); err == nil {
      break // success! done
    }
    // failed, back off for sleepTime and retry
    log.Error("could not connect to ", peerID, err,
                " : will retry in ", sleepTime, " retry# ", retries)
    time.Sleep(sleepTime)   // back off
    retries++
    sleepTime *= 2          // exponential : double the next wait time
    // bail out
    if retries > common.ExpBackOffRetries {
      log.Error("Exhausted retries, could not connect to ", peerID, err,
                  "number of retries performed ", retries)
		  panic(err)
    }
  }
  /*
	err = filterNode.DialPeerWithMultiAddress(ctx, nodeList[0])
	if err != nil {
		log.Error("could not connect to ", peerID, err)
		panic(err)
	}*/

	log.Info("Starting the ", nodeType, " node ", conf.ContentTopic)
	// start the light node
	err = filterNode.Start()
	if err != nil {
	  log.Error("Could not start the", nodeType, " node ", conf.ContentTopic)
		panic(err)
	}

	log.Info("Subscribing to the content topic", conf.ContentTopic)
	// Subscribe to our ContentTopic and send a FilterRequest
	cf := filter.ContentFilter{
		Topic:         pubSubTopic.String(),
		ContentTopics: []string{conf.ContentTopic},
	}
	_, theFilter, err := filterNode.Filter().Subscribe(ctx, cf)
	if err != nil {
		panic(err)
	}

	stopC := make(chan struct{})
	go func() {
		f, err := os.OpenFile(conf.Ofname, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
      log.Error("Could not open file: ", err)
			panic(err)
		}
		defer f.Close()

		log.Info("Waiting to receive the message")
		for env := range theFilter.Chan {
			msg := env.Message()

      rbuf := bytes.NewBuffer(msg.Payload)
      var r32 int32 //:= make([]int64, (len(msg.Payload)+7)/8)
      err = binary.Read(rbuf, binary.LittleEndian, &r32)
      if err != nil {
        log.Error("binary.Read failed:", err)
        panic(err)
      }

      msg_delay := time.Since(time.Unix(0, msg.Timestamp))
      str := fmt.Sprintf("GOT : %d, %s, %d, %d, %d\n", r32, msg.ContentTopic, msg.Timestamp, msg_delay.Microseconds(), msg_delay.Milliseconds())
      //str := fmt.Sprintf("GOT: %d %s %s %s %s\n", r32, msg, utils.GetUnixEpochFrom(lightNode.Timesource().Now()), msg_delay.Microseconds(), msg_delay.Milliseconds())
			//"Received msg, @", string(msg.ContentTopic), "@", msg.Timestamp, "@", utils.GetUnixEpochFrom(lightNode.Timesource().Now()) )
			log.Info(str)
			if _, err = f.WriteString(str); err != nil {
				panic(err)
			}
		}
    log.Error("Out of the Write loop: Message channel closed - timeout")
		stopC <- struct{}{}
	}()
  // add extra 20sec + 5% as a grace period to receive as much as possible
  filterWait := conf.Duration +
                    common.InterPubSubDelay * time.Second +
                    conf.Duration/100*common.GraceWait

  log.Info("Will be waiting for ", filterWait,  ", excess ", common.GraceWait, "% from ", conf.Duration)
  <-time.After(filterWait)
  log.Error(conf.Duration, " elapsed, closing the " + nodeType + " node!");

	// shut the nodes down
	filterNode.Stop()
}
