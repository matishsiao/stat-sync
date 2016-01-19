package main

import (
	"net"
	"log"
	"fmt"
	"os"
	"time"
	"encoding/json"
)


type PeerServer struct {
	PeersList []*PeerConnection
	Signal chan *PeerMessage
}

func (p *PeerServer) Listen(ip string,port int) {
	log.Printf("[Server] %v:%v start listen.\n",ip,port)
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d",ip,port))
	if err != nil {
		log.Printf("Listen Error:%v\n",err)
		os.Exit(1)
		return
	}
	ln := l.(*net.TCPListener)
	p.Signal = make(chan *PeerMessage)
	go p.ProcessSignal()
	for {
    	conn, err := ln.AcceptTCP()
    	if err != nil {
    		log.Println("Accept Error:",err)
    	} else {
    		go p.ProcessConn(conn)
    	}
	}
}

func (p *PeerServer) ProcessSignal() {
	for signal := range p.Signal {
		if signal.Type == 100 {
			log.Println("ProcessSignal Signal:",signal)
		}
		//log.Println("ProcessSignal Signal:",signal)
	}
}

func (p *PeerServer) ProcessConn(c *net.TCPConn) {
	var pc PeerConnection
	pc.Init(p.Signal,c)
	p.PeersList = append(p.PeersList,&pc)
}

func (p *PeerServer) HeartBeatCheck() {
	for _,client := range p.PeersList {
		client.HeartBeat(CONFIGS.Name)
	}
}

func (p *PeerServer) SendPeerStatus() {
	for {
		p.sendPeerStatus()
		time.Sleep(10 * time.Second)
	}
}

func (p *PeerServer) sendPeerStatus() {
	if len(PeerStatusList) > 0 {
		var msg PeerMessage
		msg.Sender = CONFIGS.Name
		msg.MessageTime = time.Now().Unix()
		msg.Type = MESSAGE_TYPE_PEER_STATUS
		json,err := json.Marshal(PeerStatusList)
		if err != nil {
			return
		}
		msg.Message = string(json)
		for _,v := range p.PeersList {
			go v.Write(msg)
		}
	}
}

