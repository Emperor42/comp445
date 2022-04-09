package main

import (
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
	for i := 11; i < len(payload); i++ {
		p.data[i] = byte(payload[i])
	}
	return p
}

//packetType byte, seqNum uint32, host string, port uint16, payload string
func (p Packet) PacketType() byte {
	return p.data[0]
}

func (p Packet) PacketNumber() uint32 {
	return (uint32(p.data[3]) * 256) + uint32(p.data[4])
}

func (p Packet) PeerAddress() string {
	a := strconv.Itoa(int(p.data[5]))
	b := strconv.Itoa(int(p.data[6]))
	c := strconv.Itoa(int(p.data[7]))
	d := strconv.Itoa(int(p.data[8]))
	return a + "." + b + "." + "." + c + "." + d
}

func (p Packet) PeerPort() uint16 {
	return (uint16(p.data[9]) * 256) + uint16(p.data[10])
}

func (p Packet) Payload() string {
	ret := ""
	for i := 11; i < 1024; i++ {
		ret = ret + string(p.data[i])
	}
	return ret
}

type UDP struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func Client(network string) *UDP {
	var u UDP
	u.conn, _ = net.DialUDP(network, nil, nil)
	return &u
}

func Server(network string) *UDP {
	var u UDP
	u.conn, _ = net.ListenUDP(network, nil)
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
