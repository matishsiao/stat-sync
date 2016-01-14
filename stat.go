package main 

import (
	"flag"
	"log"
	"fmt"
	"time"
)

var (
	CONFIGS Configs
	configPath string
	Server PeerServer
	version string = "0.0.1"
)

func main() {
	
	flag.StringVar(&configPath,"c","stat.json","config file path")
	flag.Parse()
	log.Printf("Version:%s Config File:%s\n",version,configPath)
	config,err := LoadConfigs(configPath)
	if err != nil {
		log.Println("Config Error:",err)
	}
	CONFIGS = config
	CONFIGS.Name = fmt.Sprintf("%s:%d",CONFIGS.Host,CONFIGS.Port)
	StartService()
	for {
		DumpClientStatus()
		time.Sleep(10 * time.Second)
	}
}

func StartService() {
	go HeartBeatInit()
	go Server.Listen(CONFIGS.Host,CONFIGS.Port)
}

