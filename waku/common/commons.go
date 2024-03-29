package common

import (
  "time"
  "math/rand"
  "flag"
)

const StartPort = 60000
const PortRange = 1000

const GraceWait = 10 // percentage
const InterPubSubDelay = 25 // seconds

const DnsDiscoveryUrl = "enrtree://ANEDLO25QVUGJOUTQFRYKWX6P4Z4GKVESBMHML7DZ6YK4LGS5FC5O@prod.wakuv2.nodes.status.im"
const NameServer = "1.1.1.1"
const LocalHost = "0.0.0.0"

const ExpBackOffInit time.Duration = 10*time.Second // 10s, 20s, 40s, 80s, 160s
const ExpBackOffRetries int = 5

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
	flag.StringVar(&(*conf).Ofname, "o", "output.out",
		"Specify the output file")
	flag.StringVar(&(*conf).ContentTopic, "c", "6b6fd7006afdfe916f08b5d",
		"Specify the content topic")
}
