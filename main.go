package main

import (
	"danmu/config"
	"danmu/platform/bilibili"
	"danmu/platform/douyu"
	"danmu/utils"
	"danmu/utils/log"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
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
		return
	}

	if config.Env != "" {
		confFile = config.Env
	}

	if config.Debug != "" {
		log.SetType(log.FileLog, map[string]string{"fileName": config.Debug})
		log.SetLevel(log.LevelInfo)
	}

	conf, err := ini.LoadFile(confFile)
	config.Conf = conf
	if err != nil {
		log.Fatal(err)
	}

	// platform url
	if len(os.Args) > 1 {
		platformURL := os.Args[len(os.Args)-1]
		u, err := url.ParseRequestURI(platformURL)
		if err != nil {
			log.Fatal(err)
		}
		host := utils.Domain(u.Host)
		log.Info(host)

		switch host {
		case "Douyu":
			douyu.Douyu(platformURL)
		case "Bilibili":
			bilibili.Bilibili(platformURL)
		}

	}

}
