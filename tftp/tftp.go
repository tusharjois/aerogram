package tftp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type tftpError uint8

const (
	errUndef tftpError = iota
	errNotFound
	errAccess
	errNoSpace
	errIllegalOp
)

func createPacket(data []interface{}) *bytes.Buffer {
	pkt := new(bytes.Buffer)
	for _, v := range data {
		err := binary.Write(pkt, binary.BigEndian, v)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
	}
	return pkt
}

func createWriteRequest(filename string) *bytes.Buffer {
	var data = []interface{}{
		uint16(2),        // opcode (WRQ)
		[]byte(filename), // filename
		uint8(0),         // NUL
		[]byte("octet"),  // mode
		uint8(0),         // NUL
	}
	pkt := createPacket(data)
	return pkt
}

func createData(blockNum uint16, dataBlock []byte) *bytes.Buffer {
	var data = []interface{}{
		uint16(3),        // opcode (ERR)
		uint16(blockNum), // block number
		dataBlock,        // data
	}
	pkt := createPacket(data)
	return pkt
}

func createAck(blockNum uint16) *bytes.Buffer {
	var data = []interface{}{
		uint16(4),        // opcode (ACK)
		uint16(blockNum), // block number
	}
	pkt := createPacket(data)
	return pkt
}

func createError(e tftpError) *bytes.Buffer {
	var data = []interface{}{
		uint16(5), // opcode (ERR)
		uint8(e),  // error number
		uint8(0),  // NUL
	}
	pkt := createPacket(data)
	return pkt
}

func parsePacket(pkt *bytes.Buffer) {

}

func ListenForWriteRequest() {
	// fmt.Printf("% x\n", createWriteRequest("rfc1350.txt"))
}

func WriteFileToServer(fname, addr string) error {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	serverConn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}

	wrq := createWriteRequest(fname)
	_, err := serverConn.Write(wrq.Bytes())
	if err != nil {
		return err
	}

	return nil
}
