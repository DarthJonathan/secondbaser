package main

import (
	"fmt"
	"github.com/DarthJonathan/secondbaser/config"
	"os"

)

func main() {
	if err := config.RunServer(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
