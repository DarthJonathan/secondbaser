package main

import (
	"fmt"
	"os"

	"github.com/darthjonathan/secondbaser/config"
)

func main() {
	if err := config.RunServer(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
