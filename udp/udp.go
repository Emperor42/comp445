package udp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Packet struct {
	data []byte
}

func generatePacket(packetType byte, seqNum uint32, host string, port uint16, payload string) Packet {
	if len(payload) > 1013 {
		panic("Packet payload is too large!")
	}
	var p Packet
	p.data = make([]byte, 1024)
	p.data[0] = packetType
	p.data[1] = 0
	p.data[2] = 0
	p.data[3] = byte(seqNum / 256)
	p.data[4] = byte(seqNum % 256)
	addr := strings.Split(host, ".")
	a, _ := strconv.Atoi(addr[0])
	b, _ := strconv.Atoi(addr[1])
	c, _ := strconv.Atoi(addr[2])
	d, _ := strconv.Atoi(addr[3])
	p.data[5] = byte(a)
	p.data[6] = byte(b)
	p.data[7] = byte(c)
	p.data[8] = byte(d)
	p.data[9] = byte(port / 256)
	p.data[10] = byte(port % 256)
	if payload != "" {
		for i := 11; i < len(payload); i++ {
			p.data[i] = byte(payload[i])
		}
	}
	return p
}

func (u *UDP) generatePacket(packetType byte, seqNum uint32, payload string) Packet {
	if len(payload) > 1013 {
		panic("Packet payload is too large!")
	}
	var p Packet
	p.data = make([]byte, 1024)
	p.data[0] = packetType
	p.data[1] = 0
	p.data[2] = 0
	p.data[3] = byte(seqNum / 256)
	p.data[4] = byte(seqNum % 256)
	a := u.addr.IP.To4()
	port := (uint16(u.addr.Port))
	if a != nil {
		p.data[5] = a[0]
		p.data[6] = a[1]
		p.data[7] = a[2]
		p.data[8] = a[3]
	}
	p.data[9] = byte(port / 256)
	p.data[10] = byte(port % 256)
	if payload != "" {
		for i := 11; i < len(payload); i++ {
			p.data[i] = byte(payload[i])
		}
	}
	return p
}

//packetType byte, seqNum uint32, host string, port uint16, payload string
func (p *Packet) PacketType() byte {
	return p.data[0]
}

func (p *Packet) PacketNumber() uint32 {
	return (uint32(p.data[3]) * 256) + uint32(p.data[4])
}

func (p *Packet) PeerAddress() string {
	a := strconv.Itoa(int(p.data[5]))
	b := strconv.Itoa(int(p.data[6]))
	c := strconv.Itoa(int(p.data[7]))
	d := strconv.Itoa(int(p.data[8]))
	return a + "." + b + "." + "." + c + "." + d
}

func (p *Packet) PeerPort() uint16 {
	return (uint16(p.data[9]) * 256) + uint16(p.data[10])
}

func (p *Packet) Payload() string {
	ret := ""
	for i := 11; i < 1024; i++ {
		ret = ret + string(p.data[i])
	}
	return ret
}

func (p *Packet) Data() bool {
	if p.data != nil && len(p.data) > 0 {
		return p.data[0] == 0
	}
	return false
}

func (p *Packet) SYN() bool {
	if p.data != nil && len(p.data) > 0 {
		return p.data[0] == 1
	}
	return false
}

func (p *Packet) SYNACK() bool {
	if p.data != nil && len(p.data) > 0 {
		return p.data[0] == 2
	}
	return false
}

func (p *Packet) ACK() bool {
	if p.data != nil && len(p.data) > 0 {
		return p.data[0] == 3
	}
	return false
}

func (p *Packet) NAK() bool {
	if p.data != nil && len(p.data) > 0 {
		return p.data[0] == 4
	}
	return false
}

type UDP struct {
	conn      *net.UDPConn
	lastRead  int
	status    error
	addr      *net.UDPAddr
	client    bool
	buffer    [1024](*Packet)
	index     int
	connected bool
	timeout   int
	sent      int
	network   string
	port      int
}

func (u *UDP) Buffer() [1024](*Packet) {
	u.process()
	return u.buffer
}

func (u *UDP) process() {
	offset := -1
	//find minimum sequence number, that is our zero
	for i := 0; i < 1024; i++ {
		p := u.buffer[i]
		if p != nil {
			num := int(p.PacketNumber())
			//find min seqeuence number
			if offset == -1 {
				offset = num
			} else {
				if offset < num {
					offset = num
				}
			}
		}
	}
	//assign packets to correct locations
	var tempBuffer [1024](*Packet)
	for i := 0; i < 1024; i++ {
		p := u.buffer[i]
		if p != nil {
			num := int(p.PacketNumber()) - offset
			//find min seqeuence number
			tempBuffer[num] = p
		}
	}
	u.buffer = tempBuffer
}

func (u *UDP) includePacket(p *Packet) {
	if p != nil {
		u.buffer[u.index%1024] = p
		u.index++
	}
}

func (u *UDP) Connection() *net.UDPConn {
	return u.conn
}

func (u *UDP) Handshake() (*net.UDPConn, error) {
	synPacket := u.generatePacket(1, uint32(u.sent), "")
	synackPacket := u.generatePacket(2, uint32(u.sent), "")
	ackPacket := u.generatePacket(3, uint32(u.sent), "")
	if u.client {
		u.Write(&synPacket)
		for i := 0; i < u.timeout; i++ {
			var temp Packet
			p := &temp
			u.Read(p)
			u.includePacket(p)
			var ack bool
			for i := 0; i < 1024; i++ {
				p := u.buffer[i]
				if p != nil {
					if !ack {
						ack = p.SYNACK()
					}
				}
			}
			if ack {
				u.connected = true
				u.Write(&ackPacket)
				break
			}
		}
	} else {
		for i := 0; i < u.timeout; i++ {
			var temp Packet
			p := &temp
			u.Read(p)
			u.includePacket(p)
			var syn, ack bool
			for i := 0; i < 1024; i++ {
				p := u.buffer[i]
				if p != nil {
					if !syn {
						syn = p.SYN()
					}
					if !ack {
						ack = p.ACK()
					}
				}
			}
			if syn && ack {
				u.connected = true
				break
			} else if syn && !ack {
				u.Write(&synackPacket)
			}
		}
	}
	return u.conn, nil
}

func (u *UDP) Read(p *Packet) *net.UDPAddr {
	u.lastRead, u.addr, u.status = u.conn.ReadFromUDP(p.data)
	if u.status == nil {
		u.process()
		return u.addr
	}
	return nil
}

func (u *UDP) Write(p *Packet) *net.UDPAddr {
	u.lastRead, u.status = u.conn.WriteToUDP(p.data, u.addr)
	if u.status == nil {
		u.process()
		return u.addr
	}
	return nil
}

func Client(network string, port, timeout int) *UDP {
	var u UDP
	u.client = true
	u.network = network
	u.port = port
	u.connected = false
	u.timeout = timeout
	u.addr, u.status = net.ResolveUDPAddr(network, fmt.Sprintf(":%d", port))
	u.conn, u.status = net.ListenUDP(network, u.addr)
	return &u
}

func Server(network string, port, timeout int) *UDP {
	var u UDP
	u.client = true
	u.network = network
	u.port = port
	u.connected = false
	u.timeout = timeout
	u.addr, u.status = net.ResolveUDPAddr(network, fmt.Sprintf(":%d", port))
	u.conn, u.status = net.ListenUDP(network, u.addr)
	return &u
}

func (u *UDP) recvFrom() (Packet, string) {
	var p Packet
	s := ""
	_, u.addr, _ = u.conn.ReadFromUDP(p.data)
	return p, s
}

func (u *UDP) sendTo(p Packet, hostport string) {
	u.conn.WriteToUDP(p.data, u.addr)
}

func main() {}
