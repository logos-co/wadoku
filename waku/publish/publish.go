package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"bytes"
  //"math/rand"
  "strconv"
	"encoding/binary"
	"os"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/payload"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"github.com/logos-co/wadoku/waku/common"

	//"crypto/rand"
	//"encoding/hex"
	//"github.com/ethereum/go-ethereum/crypto"
	//"github.com/waku-org/go-waku/waku/v2/protocol/filter"
	//"github.com/waku-org/go-waku/waku/v2/protocol"

  //"github.com/wadoku/wadoku/utils"
)

var log = logging.Logger("lightpush")
var seqNumber int32 = 0
var conf = common.Config{}

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
		node.WithHostAddress(hostAddr),
		//node.WithNTP(),  // don't use NTP, fails at msec granularity
		node.WithWakuRelay(),
		node.WithWakuFilter(false),
	)
	if err != nil {
		panic(err)
	}

  log.Info("config: ", conf)
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

	log.Info("STARTING THE LIGHTPUSH NODE ", conf.ContentTopic)
	// start the light node
	err = wakuNode.Start()
	if err != nil {
		log.Error("COULD NOT START THE LIGHTPUSH ", peerID, err)
		panic(err)
	}

	go writeLoop(ctx, &conf, wakuNode)

	<-time.After(conf.Duration)
  log.Error(conf.Duration, " elapsed, closing the node!");

	// shut the nodes down
	wakuNode.Stop()
}

func writeLoop(ctx context.Context, conf *common.Config, wakuNode *node.WakuNode) {
	log.Info("IN THE WRITELOOP ", conf.ContentTopic)

	f, err := os.OpenFile(conf.Ofname, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
    log.Error("Could not open file: ", err)
		panic(err)
	}
	defer f.Close()

  prevTStamp := wakuNode.Timesource().Now()
	for {
		time.Sleep(conf.Iat)
    seqNumber++

		// build the message & seq number
		p := new(payload.Payload)
    wbuf := new(bytes.Buffer)
    err := binary.Write(wbuf, binary.LittleEndian, seqNumber)
    if err != nil {
        log.Error("binary.Write failed:", err)
        panic(err)
    }
    p.Data = wbuf.Bytes()
		var version uint32 = 0
		payload, err := p.Encode(version)
		if err != nil {
			log.Error("Could not Encode: ", err)
      panic(err)
		}
		msg := &pb.WakuMessage{
			Payload:      payload,
			Version:      version,
			ContentTopic: conf.ContentTopic,
			Timestamp:    utils.GetUnixEpochFrom(wakuNode.Timesource().Now()),
		}

		// publish the message
		_, err = wakuNode.Relay().Publish(ctx, msg)
		if err != nil {
			log.Error("Could not publish: ", err)
			return
		}

    iat := time.Since(prevTStamp)
		str := fmt.Sprintf("SENT: %d %s %d\n", seqNumber, msg, iat.Milliseconds())
		if _, err = f.WriteString(str); err != nil {
			panic(err)
		}
		log.Info(str)
    prevTStamp = time.Unix(0, msg.Timestamp)
	}
  log.Error("Out of the Write loop: Message channel closed (timeout?)!")
}