package main

import (
	"flag"

	"github.com/lmika/gopkgs/cli"
)

func main() {
	flagQueue := flag.String("q", "", "queue to drain")
	flag.Parse()

	if *flagQueue == "" {
		cli.Fatalf("-q flag needs to be specified")
	}

	
}
