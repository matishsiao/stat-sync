package main

import (
	"net"
	"bytes"
	"encoding/json"
	"time"
	"log"
	"fmt"
)

type PeerClient struct {
	Addr string
	Conn *net.TCPConn
	Status bool
	RecvBuf	bytes.Buffer
	Signal chan *PeerMessage
	Outgoing chan PeerMessage
	HB HeartBeatStatus
	PS PeersStatus
}

func (p *PeerClient) Init(signal chan *PeerMessage,addr string) {
	p.Conn = p.Connect(addr)
	p.Addr = addr
	p.Status = true
	p.Signal = signal
	p.Outgoing = make(chan PeerMessage)
	go p.Read()
	go p.ProcessWrite()
}

func (p *PeerClient) RetryConnect(addr string) *net.TCPConn {
	for {
		log.Println("Retry Connect to ",addr)
		ClientPeerStatusList[addr].Status = false
		ClientPeerStatusList[addr].ReportTime = time.Now().Unix()
		conn := p.connect(addr)
		if conn != nil {
			p.Conn = conn
			ClientPeerStatusList[addr].Status = true
			return p.Conn
		}
		time.Sleep(10 * time.Second)
	}
}

func (p *PeerClient) Connect(addr string) *net.TCPConn {
	conn := p.connect(addr)
	if conn == nil {
		conn = p.RetryConnect(addr)
	}
	return conn
}

func (p *PeerClient) connect(addr string) *net.TCPConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
    if err != nil {
        log.Println("ResolveTCPAddr failed:", err.Error())
        return nil
    }
	conn, err := net.DialTCP("tcp",nil, tcpAddr)
	if err != nil {
		log.Println("PeerClient Connection failed.",err)
		return nil
	}
	p.Status = true
	return conn
}

func (p *PeerClient) Write(data PeerMessage) {
	 p.Outgoing <- data
}

func (p *PeerClient) ProcessWrite() {
	for v := range p.Outgoing {
		data,err := json.Marshal(v)
		if err != nil {
			log.Println("HeartBeat JSON Error:",err)
			continue
		}
		err = p.write(data)
		if err != nil {
			log.Printf("Peer Client Write Error:%v\n", err)
		}
	}
}

func (p *PeerClient) write(data []byte) error {
	if p.Status {
		_, err := p.Conn.Write(data)
		if err != nil {
			log.Printf("Peer Client Send Error:%v RemoteAddr:%s\n", err, p.Addr)
			p.Conn.Close()
			go p.RetryConnect(p.Addr)
			p.Status = false
		}
		return err
	} else {
		return fmt.Errorf("Lost connection.")
	}
	return nil
}

func (p *PeerClient) HeartBeat(id string) {
	var msg PeerMessage
	msg.MessageTime = time.Now().Unix()
	msg.Type = MESSAGE_TYPE_HEARTBEAT
	json,err := json.Marshal(HeartBeatStatus{PeerId:id,Status:true})
	if err != nil {
		return
	}
	msg.Message = string(json)
	log.Println("HeartBeat:",msg)
	p.Write(msg)
}


func (p *PeerClient) Read() {
	for {
		msg,err := p.read()
		if msg != nil && err == nil {
			switch msg.Type {
				case MESSAGE_TYPE_PEER_STATUS:
					var ps PeersStatus
					err := json.Unmarshal([]byte(msg.Message),&ps)
					if err != nil {
						//nothing need to do
						continue
					}
					p.PS = ps
				case MESSAGE_TYPE_HEARTBEAT:
					var hb HeartBeatStatus
					err := json.Unmarshal([]byte(msg.Message),&hb)
					if err != nil {
						//nothing need to do
						continue
					}
					p.HB = hb
				default:
					continue	
			}
			go p.sendSignal(msg)
		}
	}
}

func  (p *PeerClient) sendSignal(msg *PeerMessage) {
	p.Signal <- msg
}

func (p *PeerClient) read() (*PeerMessage, error) {
	tmp := make([]byte,102400)
	for {
		resp := p.Parse()
		if resp != nil {
			return resp, nil
		}
		n, err := p.Conn.Read(tmp)
		if err != nil {
			return nil, err
		}
		p.RecvBuf.Write(tmp[0:n])
	}
}

func (p *PeerClient) Parse() *PeerMessage {
	buf := p.RecvBuf.Bytes()
	var msg PeerMessage
	err := json.Unmarshal(buf,&msg)
	if err != nil {
		return nil
	}
	return &msg
}
