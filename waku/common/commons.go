package common

import (
  "time"
  "math/rand"
  "flag"
)

const StartPort = 60000
const PortRange = 1000

const DnsDiscoveryUrl = "enrtree://AOGECG2SPND25EEFMAJ5WF3KSGJNSGV356DSTL2YVLLZWIV6SAYBM@prod.waku.nodes.status.im"
const NameServer = "1.1.1.1"
const LocalHost = "0.0.0.0"

type Config struct {
	LogLevel     string
	Ofname       string
	ContentTopic string
	Iat          time.Duration
	Duration     time.Duration
}

func RandInt(min, max int) int {
    return min + rand.Intn(max - min)
}

func ArgInit(conf *Config){
	flag.DurationVar(&(*conf).Duration, "d", 1000*time.Second,
		"Specify the duration (1s,2m,4h)")
	flag.DurationVar(&(*conf).Iat, "i", 300*time.Millisecond,
		"Specify the interarrival time in millisecs")
	flag.StringVar(&(*conf).LogLevel, "l", "info",
		"Specify the log level")
	flag.StringVar(&(*conf).Ofname, "o", "lightpush.out",
		"Specify the output file")
	flag.StringVar(&(*conf).ContentTopic, "c", "6b6fd7006afdfe916f08b5d",
		"Specify the content topic")
}
