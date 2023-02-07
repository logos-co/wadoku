package main

import (
	"context"
	"fmt"
  "flag"
 // "net"
	//"os"
	"time"
  "strconv"


	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	logging "github.com/ipfs/go-log/v2"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

  "github.com/logos-co/wadoku/waku/common"
)


var log = logging.Logger("publish")
var seqNumber int32 = 0
var conf = common.Config{}
var nodeType = "publish"

func init() {
	// args
  fmt.Println("Populating CLI params...")
  common.ArgInit(&conf)
}

const defaultPubSubTopic = "default"

func main() {
	flag.Parse()

	// setup the log  
	lvl, err := logging.LevelFromString(conf.LogLevel)
	if err != nil {
		panic(err)
	}
	logging.SetAllLoggers(lvl)

  tcpEndPoint :=   "/ip4/" + common.LocalHost +
                   "/tcp/" +
                   strconv.Itoa(common.StartPort + common.RandInt(0, common.PortRange))
//	hostAddr, _ := net.ResolveTCPAddr("tcp", tcpEndPoint)

	// create a new libp2p Host that listens on a random TCP port
	// we can specify port like /ip4/0.0.0.0/tcp/3326
	host, err := libp2p.New(libp2p.ListenAddrStrings(tcpEndPoint))
	if err != nil {
		panic(err)
	}

	// view host details and addresses
	log.Info("host ID ", host.ID().Pretty())
	log.Info("following are the assigned addresses")
	for _, addr := range host.Addrs() {
		log.Info(addr.String())
	}

	// create a new PubNode 
	ctx := context.Background()
	pubNode, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}

	// setup local mDNS discovery
	if err := setupDiscovery(host); err != nil {
		panic(err)
	}

	// join the pubsub topic called librum
	topic, err := pubNode.Join(conf.ContentTopic)
	if err != nil {
		panic(err)
	}

	// create publisher
	go publish(ctx, topic)

  <-time.After(conf.Duration)
  log.Error(conf.Duration, " elapsed, stopping the " + nodeType + " node!");

	// shut the nodes down
	host.Close()
}


// start publisher to topic
func publish(ctx context.Context, topic *pubsub.Topic) {
	for {
		  time.Sleep(conf.Iat)
      seqNumber++
      tstamp := time.Time.UnixNano(time.Now())
			msg := fmt.Sprintf("%s, %s, %s", fmt.Sprint(seqNumber), fmt.Sprint(tstamp), conf.ContentTopic)
			bytes := []byte(msg)
			topic.Publish(ctx, bytes)
      log.Info(msg)
	}
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, defaultPubSubTopic, &discoveryNotifee{h: h})
	return s.Start()
}
