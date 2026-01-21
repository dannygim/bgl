package main

import (
	"fmt"
	"os"

	"github.com/dannygim/bgl/internal/auth"
)

const (
	version = "0.1.0"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		printUsage()
	case "-v", "--version", "version":
		fmt.Printf("bgl version %s\n", version)
	case "auth":
		handleAuth()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("bgl - A command line tool for Backlog")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  bgl <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  auth login      Login to Backlog using OAuth 2.0")
	fmt.Println("  help            Show this help message")
	fmt.Println("  version         Show version information")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println("  -v, --version   Show version information")
	fmt.Println()
	fmt.Printf("Version: %s\n", version)
}

func handleAuth() {
	if len(os.Args) < 3 {
		printAuthUsage()
		os.Exit(1)
	}

	switch os.Args[2] {
	case "login":
		if err := auth.Login(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		printAuthUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown auth command: %s\n", os.Args[2])
		printAuthUsage()
		os.Exit(1)
	}
}

func printAuthUsage() {
	fmt.Println("Usage: bgl auth <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  login    Login to Backlog using OAuth 2.0")
}
