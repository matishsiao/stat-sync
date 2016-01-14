package main
import (
	"net"
	"bytes"
	"encoding/json"
	"time"
	"log"
)

type PeerConnection struct {
	Addr string
	Conn *net.TCPConn
	Status bool
	RecvBuf	bytes.Buffer
	Signal chan *PeerMessage
	Outgoing chan PeerMessage
	HB HeartBeatStatus
	PS PeersStatus
}

func (p *PeerConnection) Init(signal chan *PeerMessage,conn *net.TCPConn) {
	p.Conn = conn
	p.Addr = conn.RemoteAddr().String()
	p.Status = true
	p.Signal = signal
	p.Outgoing = make(chan PeerMessage)
	go p.Read()
	go p.ProcessWrite()
}

func (p *PeerConnection) Write(data PeerMessage) {
	 p.Outgoing <- data
}

func (p *PeerConnection) ProcessWrite() {
	for v := range p.Outgoing {
		data,err := json.Marshal(v)
		if err != nil {
			log.Println("HeartBeat JSON Error:",err)
			continue
		}
		p.write(data)
	}
}

func (p *PeerConnection) write(data []byte) error {
	_, err := p.Conn.Write(data)
	if err != nil {
		log.Printf("Srv Client Send Error:%v RemoteAddr:%s\n", err, p.Addr)
		p.Conn.Close()
		p.Status = false
	}
	return err
}

func (p *PeerConnection) HeartBeat(id string) {
	var msg PeerMessage
	msg.MessageTime = time.Now().Unix()
	msg.Type = MESSAGE_TYPE_HEARTBEAT
	json,err := json.Marshal(HeartBeatStatus{PeerId:id,Status:true})
	if err != nil {
		return
	}
	msg.Message = string(json)
	p.Write(msg)
}


func (p *PeerConnection) Read() {
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
			go p.SendSignal(msg)
		}
	}
}

func  (p *PeerConnection) SendSignal(msg *PeerMessage) {
	p.Signal <- msg
}


func (p *PeerConnection) read() (*PeerMessage, error) {
	tmp := make([]byte,102400)
	for {
		resp := p.Parse()
		if resp != nil {
			p.RecvBuf.Reset()
			return resp, nil
		}
		n, err := p.Conn.Read(tmp)
		if err != nil {
			return nil, err
		}
		p.RecvBuf.Write(tmp[0:n])
	}
}

func (p *PeerConnection) Parse() *PeerMessage {
	buf := p.RecvBuf.Bytes()
	var msg PeerMessage
	err := json.Unmarshal(buf,&msg)
	if err != nil {
		//nothing need to do
		return nil
	}
	return &msg
}