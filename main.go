package main

import (
	"flag"
	"fmt"
	"github.com/micro/mdns"
	"github.com/tusharjois/aerogram/transfer"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	// TODO: actual command line flags
	var isServer = flag.Bool("server", false, "runs in server mode if enabled")
	var inFile = flag.String("infile", "", "filename to send, leave blank for stdin")
	var outFile = flag.String("outfile", "", "filename to save as, leave blank for stdout")
	var isDebug = flag.Bool("debug", false, "show debug log")
	// var timeoutFlag = flag.Duration("timeout", time.Minute, "amount of time to wait for a connection")
	flag.Parse()

	if !*isDebug {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	if *isServer {
		ln, err := net.Listen("tcp", ":26465")
		if err != nil {
			// TODO handle error
		}

		// Setup our service export
		host, _ := os.Hostname()
		info := []string{"simple 1:1 filesharing over a local network"}
		service, _ := mdns.NewMDNSService(host, "_aerogram._tcp.", "", "", 26465, nil, info)
		server, _ := mdns.NewServer(&mdns.Config{Zone: service})
		defer server.Shutdown()

		for {
			conn, err := ln.Accept()
			if err != nil {
				// handle error
			}
			go transfer.ReceiveAerogram(conn, *outFile)
		}
	} else {
		// Make a channel for results and start listening
		// TODO: allow for more than 1
		entriesCh := make(chan *mdns.ServiceEntry, 1)

		go func() {
			// Start the lookup
			qParams := mdns.DefaultParams("_aerogram._tcp.")
			qParams.Timeout = time.Minute
			qParams.Entries = entriesCh
			mdns.Query(qParams)
			close(entriesCh)
		}()

		entry, present := <-entriesCh
		if !present {
			log.Fatalln("[WARN] aerogram: request timed out")
		}
		connString := fmt.Sprintf("%s:%d", entry.AddrV4.String(), entry.Port)
		conn, err := net.Dial("tcp", connString)
		if err != nil {
			// handle error
			log.Printf("[ERR] client: ", err)
			log.Fatalf("[ERR] client: cannot connect to %v\n", connString)
		}
		transfer.SendAerogram(conn, *inFile)
		// err := tftp.WriteFileToServer(*inFile, connString)
		// if err != nil {
		// 	log.Fatalf("[ERR] client: %s\n", err)
		// }
		os.Exit(0)
	}
}
