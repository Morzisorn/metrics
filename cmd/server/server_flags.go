package main

import (
	"github.com/spf13/pflag"
)

var flagServerAddr string

func parseServerFlags() {
	pflag.StringVar(&flagServerAddr, "a", ":8080", "address and port to run server")
	pflag.Parse()
}
