package  main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/filter"
	"github.com/waku-org/go-waku/waku/v2/utils"
	//"crypto/rand"
	//"encoding/hex"
	//"github.com/ethereum/go-ethereum/crypto"
	//"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	//"github.com/waku-org/go-waku/waku/v2/payload"
)

var log = logging.Logger("filter")
var pubSubTopic = protocol.DefaultPubsubTopic()

const dnsDiscoveryUrl = "enrtree://AOGECG2SPND25EEFMAJ5WF3KSGJNSGV356DSTL2YVLLZWIV6SAYBM@prod.waku.nodes.status.im"
const nameServer = "1.1.1.1" // your local dns provider might be blocking entr

type Config struct {
	Ofname       string
	ContentTopic string
	Iat          time.Duration
	Duration     time.Duration
}

var conf = Config{}

func init() {
	// args
  fmt.Println("Populating CLI params...")
	flag.DurationVar(&conf.Duration, "d", 1000*time.Second,
		"Specify the duration (1s,2m,4h)")
	flag.DurationVar(&conf.Iat, "i", 100*time.Millisecond,
		"Specify the interarrival time in millisecs")
	flag.StringVar(&conf.Ofname, "o", "lightpush.out",
		"Specify the output file")
	flag.StringVar(&conf.ContentTopic, "c", "d608b04e6b6fd7006afdfe916f08b5d",
		"Specify the content topic")
}

func main() {

	flag.Parse()

	// setup the log
	lvl, err := logging.LevelFromString("info")
	if err != nil {
		panic(err)
	}
	logging.SetAllLoggers(lvl)

	// create the waku node
	hostAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:60000")
	ctx := context.Background()
	lightNode, err := node.New(ctx,
		//node.WithWakuRelay(),
		node.WithHostAddress(hostAddr),
		node.WithWakuFilter(false),
	)
	if err != nil {
		panic(err)
	}

	// find the list of full node fleet peers
	fmt.Printf("attempting DNS discovery with %s\n", dnsDiscoveryUrl)
	nodes, err := dnsdisc.RetrieveNodes(ctx, dnsDiscoveryUrl, dnsdisc.WithNameserver(nameServer))
	if err != nil {
		panic(err.Error())
	}

	// connect to the first peer
	var nodeList []multiaddr.Multiaddr
	for _, n := range nodes {
		nodeList = append(nodeList, n.Addresses...)
	}
	fmt.Printf("Discovered and connecting to %v \n", nodeList[0])
	peerID, err := nodeList[0].ValueForProtocol(multiaddr.P_P2P)
	if err != nil {
		fmt.Printf("could not get peerID: %s \n", err)
		panic(err)
	}
	err = lightNode.DialPeerWithMultiAddress(ctx, nodeList[0])
	if err != nil {
		fmt.Printf("could not connect to %s: %s \n", peerID, err)
		panic(err)
	}

	fmt.Println("STARTING THE LIGHTNODE ", conf.ContentTopic)
	// start the light node
	err = lightNode.Start()
	if err != nil {
		panic(err)
	}

	fmt.Println("SUBSCRIBING TO THE TOPIC ", conf.ContentTopic)
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
			panic(err)
		}
		defer f.Close()

		for env := range theFilter.Chan {
			msg := env.Message()
			log.Info("Light node received msg, ", string(msg.Payload))
			str := fmt.Sprintf("Received msg, @", string(msg.ContentTopic), "@", msg.Timestamp, "@", utils.GetUnixEpochFrom(lightNode.Timesource().Now()))
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
