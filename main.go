package main

import (
	"github.com/tusharjois/aerogram/tftp"
)

func main() {
	tftp.CreateWritePacket("hello.txt")
}
