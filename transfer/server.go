package transfer

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
)

func ReceiveAerogram(conn net.Conn, filename string) {
	log.Printf("[INFO] server: received conn from %v\n", conn.RemoteAddr())

	var f *os.File
	if filename == "" {
		log.Print("[INFO] server: writing to stdout\n")
		f = os.Stdout
	} else {
		log.Printf("[INFO] server: writing to %v\n", filename)
	}
	writer := bufio.NewWriter(f)

	n, err := io.Copy(writer, conn)
	if err != nil {
		log.Printf("[ERR] server: %v\n", err)
		log.Print("[ERR] server: cannot receive aerogram\n")
		// TODO: Errors?
	}
	log.Printf("[INFO] server: received %v bytes\n", n)
}
