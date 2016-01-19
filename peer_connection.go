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
	if p.Status {
		 p.Outgoing <- data
	}
}

func (p *PeerConnection) Close() {
	p.Status = false
	p.Conn.Close()
	log.Println("Close Connection:",p.Addr)
}

func (p *PeerConnection) ProcessWrite() {
	for v := range p.Outgoing {
		
		data,err := json.Marshal(v)
		if err != nil {
			log.Println("HeartBeat JSON Error:",err)
			continue
		}
		
		err = p.write(data)
		if err != nil {
			break
		}
	}
	p.Close()
}

func (p *PeerConnection) write(data []byte) error {
	_, err := p.Conn.Write(data)
	if err != nil {
		log.Printf("Srv Client Send Error:%v RemoteAddr:%s data:%s\n", err, p.Addr,string(data))
		
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
		if p.Status {
			msg,err := p.read()
			if msg != nil && err == nil {
				switch msg.Type {
					case MESSAGE_TYPE_HEARTBEAT:
						var hb HeartBeatStatus
						err := json.Unmarshal([]byte(msg.Message),&hb)
						if err != nil {
							//nothing need to do
							continue
						}
						if v,ok := PeerStatusList[hb.PeerId];!ok {
							NewPeer(hb.PeerId)
							//PeerStatusList[hb.PeerId] = &PeerStatus{PeerId:hb.PeerId,Status:hb.Status,ReportTime:msg.MessageTime}
						} else {
							v.Status = hb.Status
							v.ReportTime = msg.MessageTime
						}
					default:
						continue	
				}
				go p.SendSignal(msg)
			} else if err != nil {
				p.Close()
			}
		} else {
			break
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