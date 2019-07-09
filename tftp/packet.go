package tftp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type tftpError uint8

const (
	errUndef tftpError = iota
	errNotFound
	errAccess
	errNoSpace
	errIllegalOp
)

type tftpOp uint16

const (
	opWrq   tftpOp = 2
	opData  tftpOp = 3
	opAck   tftpOp = 4
	opError tftpOp = 5
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

func parsePacket(buf []byte) (tftpOp, []byte, error) {
	var opcode tftpOp
	opcode_reader := bytes.NewReader(buf[0:2])
	err := binary.Read(opcode_reader, binary.BigEndian, &opcode)
	if err != nil {
		return opError, nil, fmt.Errorf("error reading opcode: %v", err)
	}

	// pkt := bytes.NewReader(buf)

	switch opcode {
	case opWrq:
		var lenFilename = 0
		for lenFilename+1 < len(buf) && buf[2+lenFilename] != 0x00 {
			lenFilename++
		}
		if lenFilename+1 > len(buf) {
			return opError, nil, fmt.Errorf("no end of buffer after %v bytes", lenFilename)
		}
		fname := make([]byte, lenFilename)
		fnameReader := bytes.NewReader(buf[2 : 2+lenFilename])
		err := binary.Read(fnameReader, binary.BigEndian, &fname)
		fmt.Printf("%s\n", fname)
		if err != nil {
			return opError, nil, fmt.Errorf("error in reading filename for WRQ: %v", err)
		}

		mode := make([]byte, 5)
		modeReader := bytes.NewReader(buf[2+lenFilename+1 : len(buf)-1])
		err = binary.Read(modeReader, binary.BigEndian, &mode)
		if err != nil || string(mode) != "octet" {
			return opError, nil, fmt.Errorf("invalid mode for WRQ")
		}

		return opcode, fname, nil

	default:
		return opError, nil, fmt.Errorf("invalid opcode %v", opcode)
	}
}

func createWriteRequest(filename string) *bytes.Buffer {
	var data = []interface{}{
		opWrq,            // opcode (WRQ = 02)
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
		opData,           // opcode (DAT = 03)
		uint16(blockNum), // block number
		dataBlock,        // data
	}
	pkt := createPacket(data)
	return pkt
}

func createAck(blockNum uint16) *bytes.Buffer {
	var data = []interface{}{
		opAck,            // opcode (ACK = 04)
		uint16(blockNum), // block number
	}
	pkt := createPacket(data)
	return pkt
}

func createError(e tftpError) *bytes.Buffer {
	var data = []interface{}{
		opError,  // opcode (ERR = 05)
		uint8(e), // error number
		uint8(0), // NUL
	}
	pkt := createPacket(data)
	return pkt
}
