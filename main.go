package main

import (
	"errors"
	"fmt"
	"os"

	"xpier/internal/xpier"
)

func main() {
	if err := xpier.Run(os.Args[1:]); err != nil {
		if !errors.Is(err, xpier.ErrUsage) {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}
