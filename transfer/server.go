package transfer

import (
	"bufio"
	"compress/gzip"
	"io"
	"log"
	"net"
	"os"
)

func ReceiveAerogram(conn net.Conn, filename string, useGzip bool) {
	log.Printf("[INFO] server: received conn from %v\n", conn.RemoteAddr())
	defer conn.Close()

	var f *os.File
	var err error

	if filename == "" {
		log.Print("[INFO] server: writing to stdout\n")
		f = os.Stdout
	} else {
		f, err = os.Create(filename)
		if err != nil {
			log.Printf("[ERR] server: %v\n", err)
			log.Fatal("[ERR] server: cannot write to file\n")
		}
		log.Printf("[INFO] server: writing to %v\n", filename)
	}
	writer := bufio.NewWriter(f)

	var reader io.Reader = conn
	if useGzip {
		log.Print("[INFO] server: using decompression\n")
		gReader, err := gzip.NewReader(conn)
		if err != nil {
			log.Printf("[ERR] server: %v\n", err)
			log.Fatal("[ERR] server: cannot use compression\n")
		}
		defer gReader.Close()
		reader = gReader
	}

	n, err := io.Copy(writer, reader)
	if err != nil {
		log.Printf("[ERR] server: %v\n", err)
		log.Print("[ERR] server: cannot receive aerogram\n")
		// TODO: Errors?
	}
	log.Printf("[INFO] server: received %v bytes\n", n)
	if n == 0 {
		log.Fatalf("[ERR] server: no bytes received from %v\n", conn.RemoteAddr())
	}
}
