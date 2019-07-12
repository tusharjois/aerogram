package transfer

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
)

func SendAerogram(conn net.Conn, filename string) error {
	log.Printf("[INFO] client: connected to server %v\n", conn.RemoteAddr())

	var f *os.File
	var err error

	if filename == "" {
		log.Print("[INFO] client: reading from stdin\n")
		f = os.Stdin
	} else {
		f, err = os.Open(filename)
		if err != nil {
			log.Printf("[ERR] client: %v\n", err)
			log.Fatal("[ERR] client: cannot write to file")
		}
		log.Printf("[INFO] client: reading from %v\n", filename)
	}
	reader := bufio.NewReader(f)

	n, err := io.Copy(conn, reader)
	if err != nil {
		log.Printf("[ERR] client: %v\n", err)
		log.Print("[ERR] client: cannot send aerogram\n")
		return err
	}
	conn.Close()

	log.Printf("[INFO] client: sent %v bytes\n", n)
	return nil
}
