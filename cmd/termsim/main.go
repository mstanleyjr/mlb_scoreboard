package main

import (
	"fmt"
	"os"

	"github.com/mstanleyjr/mlb_scoreboard/internal/termsim"
)

func main() {
	if err := termsim.RunCLI(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
