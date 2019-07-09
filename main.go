package main

import (
	"flag"
	"fmt"
	"github.com/tusharjois/aerogram/tftp"
)

func main() {
	// TODO: actual command line flags
	var isServer = flag.Bool("server", false, "runs in server mode if enabled")
	flag.Parse()

	if *isServer {
		err := tftp.ListenForWriteRequest("localhost:26465")
		fmt.Println(err)
	} else {
		err := tftp.WriteFileToServer("txfile", "localhost:26465")
		fmt.Println(err)
	}
}
