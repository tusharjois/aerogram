package tftp

import (
	// "bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// TODO: make this an option
// const optionBlocksize uint16 = 65464
const optionBlocksize uint16 = 1428

func ListenForWriteRequest(addr, outFile string) error {
	// fmt.Printf("% x\n", createWriteRequest("rfc1350.txt"))
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	listenConn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}

	defer listenConn.Close()

	//for {
	// Wait for a connection.
	buf := make([]byte, optionBlocksize+4)
	n, srcAddr, err := listenConn.ReadFromUDP(buf)
	if err != nil {
		return err
	}

	// fmt.Printf("%v bytes from %v: %x\n", n, srcAddr, buf[:n])
	handleFileTransfer(listenConn, srcAddr, buf[:n], outFile)
	// TODO: For go routines, need to track open connections
	// TODO: Might have to use channels for this

	// Handle the connection in a new goroutine.
	// The loop then returns to accepting, so that
	// multiple connections may be served concurrently.
	// go func(c net.UDPConn) {
	// 	fmt.Print()
	// 	// Shut down the connection.
	// 	c.Close()
	// }(conn)
	//}

	return nil
}

func handleFileTransfer(listenConn *net.UDPConn, srcAddr *net.UDPAddr,
	buf []byte, outFile string) {
	opcode, data, _, err := parsePacket(buf)
	if err != nil {
		log.Print(err)
		return
	}
	if opcode != opWrq {
		errPkt := createError(errIllegalOp)
		_, err := listenConn.WriteToUDP(errPkt.Bytes(), srcAddr)
		if err != nil {
			log.Print(fmt.Errorf("error in responding to error packet: %v", err))
		}
		return
	}

	var f *os.File
	log.Printf("[INFO] server: receiving file %s\n", string(data))
	if outFile == "" {
		f = os.Stdout
	} else {
		f, err = os.Create(outFile)
		if err != nil {
			log.Fatal(fmt.Errorf("cannot open file %s: %v", data, err))
		}
	}
	defer f.Close()

	var blockNum uint16 = 0
	initAck := createAck(blockNum)
	_, err = listenConn.WriteToUDP(initAck.Bytes(), srcAddr)
	if err != nil {
		log.Print(fmt.Errorf("error in responding to WRQ: %v", err))
		return
	}

	lastPacket := initAck

	for {
		// TODO: timeouts when not lockstep
		recv := make([]byte, optionBlocksize+4)
		// TODO: SrcData?
		n, srcAddr, err := listenConn.ReadFromUDP(recv)
		// fmt.Printf("%v bytes from %v: %x\n", n, srcAddr, recv[:n])
		if err != nil {
			log.Print(fmt.Errorf("error when receiving data: %v", err))
			return
		}
		opcode, data, block, err := parsePacket(recv[:n])
		if err != nil {
			log.Print(fmt.Errorf("error when parsing packet: %v", err))
			return
		}
		if opcode != opData {
			errPkt := createError(errIllegalOp)
			_, err := listenConn.WriteToUDP(errPkt.Bytes(), srcAddr)
			if err != nil {
				log.Print(fmt.Errorf("error in responding to incorrect %d packet: %v", opcode, err))
			}
			return
		}

		if block == blockNum {
			_, err = listenConn.WriteToUDP(lastPacket.Bytes(), srcAddr)
			if err != nil {
				log.Print(fmt.Errorf("error in sending duplicate ACK %d: %v",
					blockNum, err))
				return
			}
			continue
		} // TODO: greater than? error

		if blockNum == 65535 {
			blockNum = 0
		} else {
			blockNum++
		}
		lastPacket = createAck(blockNum)
		_, err = listenConn.WriteToUDP(lastPacket.Bytes(), srcAddr)
		if err != nil {
			log.Print(fmt.Errorf("error in sending ACK %d: %v",
				blockNum, err))
			return
		}

		_, err = f.Write(data)
		if err != nil {
			log.Fatal(fmt.Errorf("error in writing file: %v", err))
		}
		if len(data) < int(optionBlocksize) {
			listenConn.Close()
			break
		}
	}
}

func WriteFileToServer(fname, addr string) error {
	var f *os.File
	var err error
	if fname == "" || fname == "stdout" {
		f = os.Stdin
		fname = "stdout"
	} else {
		f, err = os.Open(fname)
		if err != nil {
			return fmt.Errorf("error opening file %v: %v", fname, err)
		}
	}
	defer f.Close()

	// reader := bufio.NewReader(f)
	reader := f

	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	serverConn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}
	defer serverConn.Close()

	wrq := createWriteRequest(fname, optionBlocksize)
	_, err = serverConn.Write(wrq.Bytes())
	if err != nil {
		return err
	}

	lastPacket := wrq
	var blockNum uint16 = 0

	for {
		buf := make([]byte, optionBlocksize+4)
		n, srvAddr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		log.Printf("[INFO] client: %v bytes from %v: %v\n", n, srvAddr, buf[:n])

		opcode, _, block, err := parsePacket(buf[:n])
		if err != nil || opcode != opAck {
			errPkt := createError(errIllegalOp)
			_, err := serverConn.Write(errPkt.Bytes())
			if err != nil {
				return fmt.Errorf("error in responding to incorrect %d packet: %v", opcode, err)
			}
			return fmt.Errorf("protocol error")
		}

		// TODO: out-of-order messages
		if block != blockNum {
			_, err = serverConn.Write(lastPacket.Bytes())
			if err != nil {
				log.Print(fmt.Errorf("error in sending duplicate data %d: %v",
					blockNum, err))
			}
			continue
		} // TODO: greater than? error

		if blockNum == 65535 {
			blockNum = 0
		} else {
			blockNum++
		}
		fileData := make([]byte, optionBlocksize)
		// n, err = f.Read(fileData)
		n, rErr := reader.Read(fileData)
		if rErr != nil && rErr != io.EOF {
			log.Fatal(fmt.Errorf("error in reading file %v: %v", fname, err))
		}
		// Perfect multiple of 512, what happens on server side?
		// Nothing. The ACK is sent for an empty DAT packet
		lastPacket = createData(blockNum, fileData[:n])
		n, err = serverConn.Write(lastPacket.Bytes())
		if err != nil {
			log.Print(fmt.Errorf("error in sending data %d: %v",
				blockNum, err))
			return err
			// TODO: Make the error packet a function
		}
		//if n < int(optionBlocksize)+4 {
		if rErr == io.EOF {
			// TODO: get final ack
			return nil
		}
	}
}
