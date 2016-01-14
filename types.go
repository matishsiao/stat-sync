package main 

import (
)
const MESSAGE_TYPE_PEER_STATUS = 1
const MESSAGE_TYPE_HEARTBEAT = 2

type PeerMessage struct {
	MessageTime int64 `json:"time"`
	Type int `json:"type"`
	Message string `json:"message"`
}

type HeartBeatStatus struct {
	PeerId string `json:"peerid"`
	Status bool `json:"status"`
}

type PeerStatus struct {
	PeerId string `json:"peerid"`
	Status bool `json:"status"`
	ReportTime int64 `json:"report"`
}

type PeersStatus struct {
	Sender string `json:"sender"`
	SendTime int64 `json:"sendtime"`
	Group []PeerStatus `json:"group"`
}

type Configs struct {
	Host string `json:"host"`
	Port int `json:"port"`
	Name string	`json:"name"`
	PeerList []string `json:"peerlist"`
}


