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
	var inFile = flag.String("sendfile", "", "filename to send, leave blank for stdin")
	var outFile = flag.String("recvfile", "", "filename to save as, leave blank for stdout")
	var useGzip = flag.Bool("gzip", false, "compress/decompress with gzip")
	var isDebug = flag.Bool("debug", false, "show debug log")
	flag.Parse()

	if !*isDebug {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	st, _ := os.Stdin.Stat()
	isStdin := st.Size() > 0

	if !isStdin || *outFile != "" {
		// Run in server mode.
		ln, err := net.Listen("tcp", ":26465")
		if err != nil {
			log.Printf("[ERR] server: %v\n", err)
			fmt.Fprintf(os.Stderr, "cannot listen for aerogram on %v\n", "26465")
			os.Exit(1)
		}

		// Setup our service export.
		host, _ := os.Hostname()
		info := []string{"simple 1:1 filesharing over a local network"}

		ifaces, err := net.Interfaces()
		if err != nil {
			log.Printf("[ERR] server: %v\n", err)
			fmt.Fprintln(os.Stderr, "error: cannot load network interfaces")
			os.Exit(1)
		}

		// Get the list of addresses on which to advertise.
		var ips []net.IP
		for _, i := range ifaces {
			addrs, _ := i.Addrs()
			if err != nil {
				log.Printf("[ERR] server: %v\n", err)
				fmt.Fprintf(os.Stderr, "error: cannot load addresses for interface %v\n", i.Name)
				os.Exit(1)
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				ips = append(ips, ip)
			}
		}

		// Stand up the mDNS receiver.
		service, _ := mdns.NewMDNSService(host, "_aerogram._tcp.", "", "", 26465, ips, info)
		server, _ := mdns.NewServer(&mdns.Config{Zone: service})
		// TODO: Handle Shutdown in the os.Exit(1) case.
		defer server.Shutdown()

		// Accept a send request.
		// TODO: How would this look with more than one connection?
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[ERR] server: %v\n", err)
			fmt.Fprintln(os.Stderr, "error: cannot accept connection")
			os.Exit(1)
		}

		// Receive the aerogram.
		err = transfer.ReceiveAerogram(conn, *outFile, *useGzip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot receive aerogram from %v\n", conn.RemoteAddr().String())
			os.Exit(1)
		}
	} else {
		// Run in client mode.
		// Make a channel for results and start listening.
		// TODO: Allow for more than 1 entry.
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
			fmt.Fprintln(os.Stderr, "request timed out")
			os.Exit(1)
		}
		log.Printf("[INFO] client: found entry %v\n", entry)
		connString := fmt.Sprintf("%s:%d", entry.AddrV4.String(), entry.Port)

		// Connect to the aerogram instance.
		conn, err := net.Dial("tcp", connString)
		if err != nil {
			log.Printf("[ERR] client: %v\n", err)
			fmt.Fprintf(os.Stderr, "error: cannot connect to %v\n", connString)
			os.Exit(1)
		}

		// Send the aerogram.
		err = transfer.SendAerogram(conn, *inFile, *useGzip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot send aerogram to %v\n", connString)
			os.Exit(1)
		}
	}
}
