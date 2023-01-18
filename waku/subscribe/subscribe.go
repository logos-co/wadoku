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
//	"github.com/waku-org/go-waku/waku/v2/protocol/filter"
//	"github.com/waku-org/go-waku/waku/v2/utils"
	"github.com/logos-co/wadoku/waku/common"
	//"crypto/rand"
	//"encoding/hex"
	//"github.com/ethereum/go-ethereum/crypto"
	//"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	//"github.com/waku-org/go-waku/waku/v2/payload"
)

var log = logging.Logger("subscribe")
var pubSubTopic = protocol.DefaultPubsubTopic()
var conf = common.Config{}

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

   tcpEndPoint :=  "0.0.0.0:" + strconv.Itoa(common.StartPort + common.RandInt(0, common.Offset))
	// create the waku node
	hostAddr, _ := net.ResolveTCPAddr("tcp", tcpEndPoint)
	ctx := context.Background()
	wakuNode, err := node.New(ctx,
    //node.WithNTP(),  // don't use NTP, fails at msec granularity
		node.WithWakuRelay(),
		node.WithHostAddress(hostAddr),
		node.WithWakuFilter(false),
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
	err = wakuNode.DialPeerWithMultiAddress(ctx, nodeList[0])
	if err != nil {
		log.Error("could not connect to ", peerID, err)
		panic(err)
	}

	log.Info("STARTING THE SUB NODE ", conf.ContentTopic)
	// start the sub node
	err = wakuNode.Start()
	if err != nil {
	  log.Error("COULD NOT START THE SUB NODE ", conf.ContentTopic)
		panic(err)
	}

	log.Info("SUBSCRIBING TO THE TOPIC ", conf.ContentTopic)
	// Subscribe to our ContentTopic and send Sub Request
	/*cf := filter.ContentFilter{
		Topic:         pubSubTopic.String(),
		ContentTopics: []string{conf.ContentTopic},
	}*/
	sub, err := wakuNode.Relay().Subscribe(ctx)
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
		for env := range sub.C {
			msg := env.Message()

      if msg.ContentTopic != conf.ContentTopic {
        continue
      }
      rbuf := bytes.NewBuffer(msg.Payload)
      var r32 int32 //:= make([]int64, (len(msg.Payload)+7)/8)
      err = binary.Read(rbuf, binary.LittleEndian, &r32)
      if err != nil {
        log.Error("binary.Read failed:", err)
        panic(err)
      }

      rtt := time.Since(time.Unix(0, msg.Timestamp))
      str := fmt.Sprintf("GOT: %d %s %d %d\n", r32, msg, rtt.Microseconds(), rtt.Milliseconds())
      //str := fmt.Sprintf("GOT: %d %s %s %s %s\n", r32, msg, utils.GetUnixEpochFrom(lightNode.Timesource().Now()), msg_delay.Microseconds(), msg_delay.Milliseconds())
			//"Received msg, @", string(msg.ContentTopic), "@", msg.Timestamp, "@", utils.GetUnixEpochFrom(lightNode.Timesource().Now()) )
			log.Info(str)
			if _, err = f.WriteString(str); err != nil {
				panic(err)
			}
		}
    log.Error("Out of the Write loop: Message channel closed (timeout?)!")
		stopC <- struct{}{}
	}()

  <-time.After(conf.Duration)
  log.Error(conf.Duration, " elapsed, closing the node!");

	// shut the nodes down
	wakuNode.Stop()
}