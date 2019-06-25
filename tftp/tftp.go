package tftp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func createPacket(data []interface{}) *bytes.Buffer {
	pkt := new(bytes.Buffer)
	for _, v := range data {
		err := binary.Write(pkt, binary.LittleEndian, v)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
	}
	return pkt
}

func CreateWritePacket(filename string) *bytes.Buffer {
	var data = []interface{}{
		uint8(0),
		uint8(2),         // opcode (WRQ)
		[]byte(filename), // filename
		uint8(0),         // NUL
		[]byte("octet"),  // mode
		uint8(0),         // NUL
	}
	pkt := createPacket(data)
	fmt.Printf("% x\n", pkt.Bytes())
	return pkt
}
