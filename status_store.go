package main

import (
	"log"
	"time"
)

var (
	peerClientList []*PeerClient
	ServerPeerStatusList map[string]*PeerStatus
	ClientPeerStatusList map[string]*PeerStatus
	PeerSignal chan *PeerMessage
)

func HeartBeatInit() {
	PeerSignal = make(chan *PeerMessage)
	ServerPeerStatusList = make(map[string]*PeerStatus)
	ClientPeerStatusList = make(map[string]*PeerStatus)
	for _,v := range CONFIGS.PeerList {
		var ps PeerStatus 
		ps.PeerId = v
		ps.Status = true
		ps.ReportTime = time.Now().Unix()
		ClientPeerStatusList[v] = &ps
		var pc PeerClient
		go pc.Init(PeerSignal,v)
		peerClientList = append(peerClientList,&pc)
	}
	go ListenPeerSignal()
	go SendHeartBeat()
}

func DumpClientStatus() {
	for k,v := range ClientPeerStatusList {
		log.Printf("Client Peer[%s] Status:%v Report:%d\n",k,v.Status,v.ReportTime)
	}
}

func ListenPeerSignal() {
	for msg := range PeerSignal {
		log.Println("ListenPeerSignal:",msg)
	}
}


func SendHeartBeat() {
	for {
		for _,peer := range peerClientList {
			go peer.HeartBeat(CONFIGS.Name)
		}
		time.Sleep(5 * time.Second)
	}
}
