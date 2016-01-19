package main

import (
	"log"
	"time"
	"encoding/json"	
)

var (
	peerClientList []*PeerClient
	PeerStatusList map[string]*PeerStatus
	PeerSignal chan *PeerMessage
)

func HeartBeatInit() {
	PeerSignal = make(chan *PeerMessage)
	PeerStatusList = make(map[string]*PeerStatus)
	for _,v := range CONFIGS.PeerList {
		go NewPeer(v)
	}
	go ListenPeerSignal()
	SendHeartBeat()
}

func NewPeer(addr string) {
	if addr != CONFIGS.Name {
		var ps PeerStatus 
		ps.PeerId = addr
		ps.Status = true
		ps.ReportTime = time.Now().Unix()
		PeerStatusList[addr] = &ps
		var pc PeerClient
		pc.Init(PeerSignal,addr)
		peerClientList = append(peerClientList,&pc)
		log.Println("NewPeer:",addr)
	}
}

func DumpPeerStatus() {
	for k,v := range PeerStatusList {
		log.Printf("Peer Status[%s] Status:%v Report:%d\n",k,v.Status,v.ReportTime)
	}
}

func DiffPeerStatus(msg *PeerMessage) {
	if msg.Type == MESSAGE_TYPE_PEER_STATUS {
		var ps map[string]PeerStatus
		err := json.Unmarshal([]byte(msg.Message),&ps)
		if err != nil {
			return
		}
		for k,peer := range ps {
			if k != CONFIGS.Name {
				if v,ok := PeerStatusList[k]; ok {
					if v.ReportTime > peer.ReportTime {
						continue
					} else {
						v.ReportTime = peer.ReportTime
						v.Status = peer.Status
					}
				} else {
					go NewPeer(k)
				}
			}
		}
	}
}

func ListenPeerSignal() {
	for msg := range PeerSignal {
		if msg.Type == 1 {
			DiffPeerStatus(msg)
		}
	}
}

func SendHeartBeat() {
	for {
		for _,peer := range peerClientList {
			go peer.HeartBeat(CONFIGS.Name)
		}
		time.Sleep(time.Duration(CONFIGS.SyncTime) * time.Second)
	}
}
