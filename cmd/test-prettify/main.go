package main

import (
	"fmt"
	"os"

	"github.com/dougdomingos/test-prettify/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
