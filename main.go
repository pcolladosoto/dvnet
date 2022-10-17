package main

import (
	"fmt"

	"github.com/pcolladosoto/dvnet/dvnet"
)

func main() {
	fmt.Printf("booting up the dvnet network driver...\n")
	h := dvnet.GetHandler()

	if err := h.ServeUnix("dvnet", 0); err != nil {
		fmt.Printf("unable to listen over a Unix socket: %v\n", err)
	}

	// if err := h.ServeTCP("dvnet", ":7777", "", nil); err != nil {
	// 	fmt.Printf("unable to listen over TCP: %v\n", err)
	// }
}
