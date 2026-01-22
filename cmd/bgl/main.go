package main

import (
	"fmt"
	"os"

	"github.com/dannygim/bgl/internal/auth"
	"github.com/dannygim/bgl/internal/issue"
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
	case "issue":
		handleIssue()
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
	fmt.Println("  auth login              Login to Backlog using OAuth 2.0")
	fmt.Println("  auth logout             Logout and remove stored tokens")
	fmt.Println("  issue view [--raw] <issueKey>   View an issue by key or ID")
	fmt.Println("  help                    Show this help message")
	fmt.Println("  version                 Show version information")
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
	case "logout":
		if err := auth.Logout(); err != nil {
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
	fmt.Println("  login     Login to Backlog using OAuth 2.0")
	fmt.Println("  logout    Logout and remove stored tokens")
}

func handleIssue() {
	if len(os.Args) < 3 {
		printIssueUsage()
		os.Exit(1)
	}

	switch os.Args[2] {
	case "view":
		handleIssueView()
	case "-h", "--help", "help":
		printIssueUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown issue command: %s\n", os.Args[2])
		printIssueUsage()
		os.Exit(1)
	}
}

func handleIssueView() {
	// Parse arguments: bgl issue view [--raw] <issueKey>
	args := os.Args[3:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: issue key is required")
		printIssueViewUsage()
		os.Exit(1)
	}

	opts := issue.ViewOptions{}
	var issueKey string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			opts.Raw = true
		case "-h", "--help":
			printIssueViewUsage()
			return
		default:
			if issueKey == "" {
				issueKey = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", args[i])
				printIssueViewUsage()
				os.Exit(1)
			}
		}
	}

	if issueKey == "" {
		fmt.Fprintln(os.Stderr, "Error: issue key is required")
		printIssueViewUsage()
		os.Exit(1)
	}

	if err := issue.View(issueKey, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printIssueUsage() {
	fmt.Println("Usage: bgl issue <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  view [--raw] <issueKey>   View an issue by key or ID")
}

func printIssueViewUsage() {
	fmt.Println("Usage: bgl issue view [options] <issueKey>")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  issueKey    The issue key (e.g., PROJECT-123) or issue ID")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --raw       Output raw JSON response")
	fmt.Println("  -h, --help  Show this help message")
}
