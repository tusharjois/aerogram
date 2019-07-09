package tftp

import (
	"fmt"
	"log"
	"net"
	"os"
)

func ListenForWriteRequest(addr string) error {
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

	for {
		// Wait for a connection.
		buf := make([]byte, 512)
		n, srcAddr, err := listenConn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		fmt.Printf("%v bytes from %v: %x\n", n, srcAddr, buf[:n])
		go handleFileTransfer(listenConn, srcAddr, buf[:n])

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		// go func(c net.UDPConn) {
		// 	fmt.Print()
		// 	// Shut down the connection.
		// 	c.Close()
		// }(conn)
	}
}

func handleFileTransfer(listenConn *net.UDPConn, srcAddr *net.UDPAddr,
	buf []byte) {
	opcode, data, err := parsePacket(buf)
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

	f, err := os.Create(string(data))
	if err != nil {
		log.Fatal(fmt.Errorf("cannot open file %s: %v", data, err))
	}
	defer f.Close()

	var blockNum uint16 = 0
	initAck := createAck(blockNum)
	_, err = listenConn.WriteToUDP(initAck.Bytes(), srcAddr)
	if err != nil {
		log.Print(fmt.Errorf("error in responding to WRQ: %v", err))
	}
	// TODO: timeouts when not lockstep
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
	_, err = serverConn.Write(wrq.Bytes())
	if err != nil {
		return err
	}

	for {
		buf := make([]byte, 32)
		n, srvAddr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		fmt.Printf("%v bytes from %v: %v\n", n, srvAddr, buf[:n])
		break
	}

	return nil
}
