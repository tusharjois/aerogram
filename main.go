package main

import (
	"flag"
	"fmt"
	"github.com/hashicorp/mdns"
	"github.com/tusharjois/aerogram/tftp"
	"io/ioutil"
	"log"
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
		// Setup our service export
		host, _ := os.Hostname()
		info := []string{"simple 1:1 filesharing over a local network"}
		service, _ := mdns.NewMDNSService(host, "_aerogram._udp", "", "", 26465, nil, info)
		server, _ := mdns.NewServer(&mdns.Config{Zone: service})
		defer server.Shutdown()
		err := tftp.ListenForWriteRequest("0.0.0.0:26465", *outFile)
		if err != nil {
			log.Fatalf("[ERR] server: %s\n", err)
		}
	} else {
		// Make a channel for results and start listening
		// TODO: allow for more than 1
		entriesCh := make(chan *mdns.ServiceEntry, 1)

		go func() {
			// Start the lookup
			qParams := mdns.DefaultParams("_aerogram._udp")
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
		log.Printf("[INFO] client: connecting to server %s\n", connString)
		err := tftp.WriteFileToServer(*inFile, connString)
		if err != nil {
			log.Fatalf("[ERR] client: %s\n", err)
		}
		os.Exit(0)
	}
}
