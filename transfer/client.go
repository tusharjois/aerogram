package transfer

import (
	"bufio"
	"compress/gzip"
	"io"
	"log"
	"net"
	"os"
)

func SendAerogram(conn net.Conn, filename string, useGzip bool) error {
	log.Printf("[INFO] client: connected to server %v\n", conn.RemoteAddr())
	defer conn.Close()

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
	defer f.Close()
	reader := bufio.NewReader(f)

	var writer io.Writer = conn
	if useGzip {
		log.Print("[INFO] client: using compression\n")
		gWriter := gzip.NewWriter(conn)
		defer gWriter.Close()
		writer = gWriter
	}

	n, err := io.Copy(writer, reader)
	if err != nil {
		log.Printf("[ERR] client: %v\n", err)
		log.Print("[ERR] client: cannot send aerogram\n")
		return err
	}

	log.Printf("[INFO] client: sent %v bytes\n", n)
	return nil
}
