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
	var isServer = flag.Bool("recv", false, "receive a file")
	var isClient = flag.Bool("send", false, "send a file")
	var inFile = flag.String("infile", "", "filename to send, leave blank for stdin")
	var outFile = flag.String("outfile", "", "filename to save as, leave blank for stdout")
	var isDebug = flag.Bool("debug", false, "show debug log")
	var useGzip = flag.Bool("gzip", false, "compress/decompress with gzip")
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

		ifaces, _ := net.Interfaces()
		var ips []net.IP
		// TODO: handle err
		for _, i := range ifaces {
			addrs, _ := i.Addrs()
			// TODO: handle err
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				// process IP address
				ips = append(ips, ip)
			}
		}

		service, _ := mdns.NewMDNSService(host, "_aerogram._tcp.", "", "", 26465, ips, info)
		server, _ := mdns.NewServer(&mdns.Config{Zone: service})
		defer server.Shutdown()

		for {
			conn, err := ln.Accept()
			if err != nil {
				// TODO handle error
			}
			//go func() {
			err = transfer.ReceiveAerogram(conn, *outFile, *useGzip)
			if err != nil {
				log.Fatalf("[ERR] server: %v\n", err)
			}
			//}()
			return
		}
	} else if *isClient {
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
		fmt.Println(entry)
		connString := fmt.Sprintf("%s:%d", entry.AddrV4.String(), entry.Port)
		conn, err := net.Dial("tcp", connString)
		if err != nil {
			// handle error
			log.Printf("[ERR] client: ", err)
			log.Fatalf("[ERR] client: cannot connect to %v\n", connString)
		}
		transfer.SendAerogram(conn, *inFile, *useGzip)
		// err := tftp.WriteFileToServer(*inFile, connString)
		// if err != nil {
		// 	log.Fatalf("[ERR] client: %s\n", err)
		// }
	} else {
		fmt.Fprintln(os.Stderr, "please specify --send or --recv")
		os.Exit(1)
	}
}
