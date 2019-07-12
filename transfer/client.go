package transfer

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
)

func SendAerogram(conn net.Conn, filename string) error {
	log.Printf("[INFO] server: received conn from %v\n", conn.RemoteAddr())

	var f *os.File
	if filename == "" {
		log.Print("[INFO] client: reading from stdin\n")
		f = os.Stdin
	} else {
		log.Printf("[INFO] client: reading from %v\n", filename)
	}
	reader := bufio.NewReader(f)

	n, err := io.Copy(conn, reader)
	if err != nil {
		log.Printf("[ERR] client: %v\n", err)
		log.Print("[ERR] client: cannot send aerogram\n")
		return err
	}

	log.Printf("[INFO] client: sent %v bytes\n", n)
	return nil
}
