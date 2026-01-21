package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	version = "0.1.0"
)

func main() {
	// Define flags
	var help bool
	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&help, "help", false, "show help")
	
	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "bgl - A command line tool for Backlog\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  bgl [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -h, --help     Show this help message\n")
		fmt.Fprintf(os.Stderr, "\nVersion: %s\n", version)
	}
	
	flag.Parse()
	
	// Show help if requested
	if help {
		flag.Usage()
		os.Exit(0)
	}
	
	// Default behavior when no flags are provided
	if flag.NFlag() == 0 {
		fmt.Println("bgl - A command line tool for Backlog")
		fmt.Println("Run 'bgl -h' or 'bgl --help' for usage information")
	}
}
