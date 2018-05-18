package main

import (
	"danmu/config"
	"danmu/util/log"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	ini "github.com/vaughan0/go-ini"
)

var wg sync.WaitGroup

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.BoolVar(&config.Version, "v", false, "Show version")
	flag.StringVar(&config.Debug, "d", "", "Debug mode")
	flag.StringVar(&config.Env, "c", "", "Env file")
}

func main() {
	flag.Parse()
	var confFile = "./env.ini"

	if config.Version {
		fmt.Println("Current Version ", config.VERSION)
		os.Exit(1)
	}

	if config.Env != "" {
		confFile = config.Env
	}

	if config.Debug != "" {
		log.SetType(log.FileLog, map[string]string{"fileName": config.Debug})
		log.SetLevel(log.LevelError)
	}

	conf, err := ini.LoadFile(confFile)
	config.Conf = conf
	if err != nil {
		log.Fatal(err)
	}

	// platform url
	if len(os.Args) > 2 {
		url := os.Args[len(os.Args)-1]
		fmt.Printf("%v", url)
	}

}
