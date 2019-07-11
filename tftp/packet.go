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

func parsePacket(buf []byte) (tftpOp, []byte, uint16, error) {
	var block uint16

	var opcode tftpOp
	opcode_reader := bytes.NewReader(buf[0:2])
	err := binary.Read(opcode_reader, binary.BigEndian, &opcode)
	if err != nil {
		return opError, nil, block, fmt.Errorf("error reading opcode: %v", err)
	}

	// pkt := bytes.NewReader(buf)
	switch opcode {
	case opWrq:
		var lenFilename = 0
		for 2+lenFilename+1 < len(buf) && buf[2+lenFilename] != 0x00 {
			lenFilename++
		}
		if lenFilename+1 > len(buf) {
			return opError, nil, block, fmt.Errorf("no end of buffer after %v bytes", lenFilename)
		}
		fname := make([]byte, lenFilename)
		fnameReader := bytes.NewReader(buf[2 : 2+lenFilename])
		err := binary.Read(fnameReader, binary.BigEndian, &fname)
		if err != nil {
			return opError, nil, block, fmt.Errorf("error in reading filename for WRQ: %v", err)
		}

		mode := make([]byte, 5)
		startMode := 2 + lenFilename + 1
		if startMode+6 > len(buf) {
			return opError, nil, block, fmt.Errorf("invalid cannot find valid mode for WRQ")
		}
		modeReader := bytes.NewReader(buf[startMode : startMode+5])
		err = binary.Read(modeReader, binary.BigEndian, &mode)
		if err != nil || string(mode) != "octet" {
			return opError, nil, block, fmt.Errorf("invalid mode for WRQ")
		}

		blksize := make([]byte, 7)
		startSize := startMode + 6
		if startSize+8+3 > len(buf) {
			fmt.Printf("%d %v\n", startSize+8, buf)
			return opError, nil, block, fmt.Errorf("missing blksize for WRQ")
		}
		sizeReader := bytes.NewReader(buf[startSize : startSize+7])
		err = binary.Read(sizeReader, binary.BigEndian, &blksize)
		if err != nil || string(blksize) != "blksize" {
			return opError, nil, block, fmt.Errorf("invalid option %s for WRQ", string(blksize))
		}

		return opcode, fname, getPacketBlock(buf[len(buf)-3 : len(buf)-1]), nil

	case opAck:
		if len(buf) == 4 {
			return opcode, buf[2:], getPacketBlock(buf[2:4]), nil
		}
		return opError, nil, block, fmt.Errorf("invalid format for ACK")
	case opData:
		if len(buf) >= 4 {
			return opcode, buf[4:], getPacketBlock(buf[2:4]), nil
		}
		return opError, nil, block, fmt.Errorf("invalid format for DAT")

	default:
		return opError, nil, block, fmt.Errorf("invalid opcode %v", opcode)
	}
}

func getPacketBlock(num []byte) uint16 {
	var blockNum uint16
	if len(num) != 2 {
		return 0 // Error
	}
	binary.Read(bytes.NewReader(num), binary.BigEndian, &blockNum)
	return blockNum
}

func createWriteRequest(filename string, blockSize uint16) *bytes.Buffer {
	if blockSize < 8 {
		blockSize = 8
	} else if blockSize > 65464 {
		blockSize = 65464
	}

	var data = []interface{}{
		opWrq,             // opcode (WRQ = 02)
		[]byte(filename),  // filename
		uint8(0),          // NUL
		[]byte("octet"),   // mode
		uint8(0),          // NUL
		[]byte("blksize"), // Blocksize Option (RFC 2348)
		uint8(0),          // NUL
		uint16(blockSize), // #blocks (8 - 65464)
		uint8(0),          // NUL
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
