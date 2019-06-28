package main

import (
	"flag"
	"github.com/tusharjois/aerogram/tftp"
)

func main() {
	// TODO: actual command line flags
	var isServer = flag.Bool("server", false, "runs in server mode if enabled")
	flag.Parse()

	if *isServer {
		tftp.ListenForWriteRequest()
	} else {
		tftp.WriteFileToServer()
	}
}
