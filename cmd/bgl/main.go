package main

import (
	"fmt"
	"os"

	"github.com/dannygim/bgl/internal/auth"
	"github.com/dannygim/bgl/internal/comment"
	"github.com/dannygim/bgl/internal/issue"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
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
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	case "auth":
		handleAuth()
	case "issue":
		handleIssue()
	case "comment":
		handleComment()
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
	fmt.Println("  comment view [--raw] <issueKey> [commentId]   View comments for an issue")
	fmt.Println("  comment add [--raw] [--yes] <issueKey> [message]   Add a comment to an issue")
	fmt.Println("  help                    Show this help message")
	fmt.Println("  version                 Show version information")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println("  -v, --version   Show version information")
	fmt.Println()
	fmt.Printf("Version: %s (commit: %s, built: %s)\n", version, commit, date)
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

func handleComment() {
	if len(os.Args) < 3 {
		printCommentUsage()
		os.Exit(1)
	}

	switch os.Args[2] {
	case "view":
		handleCommentView()
	case "add":
		handleCommentAdd()
	case "-h", "--help", "help":
		printCommentUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown comment command: %s\n", os.Args[2])
		printCommentUsage()
		os.Exit(1)
	}
}

func handleCommentView() {
	// Parse arguments: bgl comment view [--raw] <issueKey> [commentId]
	args := os.Args[3:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: issue key is required")
		printCommentViewUsage()
		os.Exit(1)
	}

	opts := comment.ViewOptions{}
	var issueKey string
	var commentID string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			opts.Raw = true
		case "-h", "--help":
			printCommentViewUsage()
			return
		default:
			if issueKey == "" {
				issueKey = args[i]
			} else if commentID == "" {
				commentID = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", args[i])
				printCommentViewUsage()
				os.Exit(1)
			}
		}
	}

	if issueKey == "" {
		fmt.Fprintln(os.Stderr, "Error: issue key is required")
		printCommentViewUsage()
		os.Exit(1)
	}

	var err error
	if commentID != "" {
		// View single comment
		err = comment.View(issueKey, commentID, opts)
	} else {
		// View comment list
		err = comment.ViewList(issueKey, opts)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printCommentUsage() {
	fmt.Println("Usage: bgl comment <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  view [--raw] <issueKey> [commentId]   View comments for an issue")
	fmt.Println("  add [--raw] [--yes] <issueKey> [message]   Add a comment to an issue")
}

func handleCommentAdd() {
	// Parse arguments: bgl comment add [--raw] [--yes] <issueKey> [message]
	args := os.Args[3:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: issue key is required")
		printCommentAddUsage()
		os.Exit(1)
	}

	opts := comment.AddOptions{}
	var issueKey string
	var message string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			opts.Raw = true
		case "--yes", "-y":
			opts.Yes = true
		case "-h", "--help":
			printCommentAddUsage()
			return
		default:
			if issueKey == "" {
				issueKey = args[i]
			} else if message == "" {
				message = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", args[i])
				printCommentAddUsage()
				os.Exit(1)
			}
		}
	}

	if issueKey == "" {
		fmt.Fprintln(os.Stderr, "Error: issue key is required")
		printCommentAddUsage()
		os.Exit(1)
	}

	if err := comment.Add(issueKey, message, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printCommentAddUsage() {
	fmt.Println("Usage: bgl comment add [options] <issueKey> [message]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  issueKey    The issue key (e.g., PROJECT-123) or issue ID")
	fmt.Println("  message     The comment message (optional, will prompt if omitted)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --raw       Output raw JSON response")
	fmt.Println("  --yes, -y   Skip confirmation prompt")
	fmt.Println("  -h, --help  Show this help message")
}

func printCommentViewUsage() {
	fmt.Println("Usage: bgl comment view [options] <issueKey> [commentId]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  issueKey    The issue key (e.g., PROJECT-123) or issue ID")
	fmt.Println("  commentId   The comment ID (optional, if omitted shows all comments)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --raw       Output raw JSON response")
	fmt.Println("  -h, --help  Show this help message")
}
