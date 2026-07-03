package main

import (
	"fmt"
	"os"

	"github.com/bougou/go-ceph/cmd/goceph/root"
)

func main() {
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "goceph: %v\n", err)
		os.Exit(1)
	}
}
