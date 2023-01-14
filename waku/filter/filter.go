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
	"github.com/waku-org/go-waku/waku/v2/utils"
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
	lightNode, err := node.New(ctx,
		//node.WithWakuRelay(),
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
	err = lightNode.DialPeerWithMultiAddress(ctx, nodeList[0])
	if err != nil {
		log.Error("could not connect to ", peerID, err)
		panic(err)
	}

	log.Info("STARTING THE FILTER NODE ", conf.ContentTopic)
	// start the light node
	err = lightNode.Start()
	if err != nil {
	  log.Error("COULD NOT START THE FILTER NODE ", conf.ContentTopic)
		panic(err)
	}

	log.Info("SUBSCRIBING TO THE TOPIC ", conf.ContentTopic)
	// Subscribe to our ContentTopic and send a FilterRequest
	cf := filter.ContentFilter{
		Topic:         pubSubTopic.String(),
		ContentTopics: []string{conf.ContentTopic},
	}
	_, theFilter, err := lightNode.Filter().Subscribe(ctx, cf)
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

		for env := range theFilter.Chan {
			msg := env.Message()
			log.Info("Light node received msg, ", string(msg.Payload))

      rbuf := bytes.NewBuffer(msg.Payload)
      var r32 int32 //:= make([]int64, (len(msg.Payload)+7)/8)
      err = binary.Read(rbuf, binary.LittleEndian, &r32)
      if err != nil {
        log.Error("binary.Read failed:", err)
        panic(err)
      }

      msg_delay := time.Since(time.Unix(0, msg.Timestamp))
			str := fmt.Sprintf("Received msg, @", r32, "@", string(msg.ContentTopic), "@", msg.Timestamp, "@", utils.GetUnixEpochFrom(lightNode.Timesource().Now()), "@", msg_delay.Microseconds(), "@", msg_delay.Milliseconds() )
			log.Info(str)
			//"Received msg, @", string(msg.ContentTopic), "@", msg.Timestamp, "@", utils.GetUnixEpochFrom(lightNode.Timesource().Now()) )
			if _, err = f.WriteString(str + "\n"); err != nil {
				panic(err)
			}
		}
		log.Info("Message channel closed!")
		stopC <- struct{}{}
	}()

  <-time.After(conf.Duration)

	// shut the nodes down
	lightNode.Stop()
}
