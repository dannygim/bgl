package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dannygim/bgl/internal/auth"
	"github.com/dannygim/bgl/internal/category"
	"github.com/dannygim/bgl/internal/comment"
	"github.com/dannygim/bgl/internal/issue"
	"github.com/dannygim/bgl/internal/issuetype"
	"github.com/dannygim/bgl/internal/milestone"
	"github.com/dannygim/bgl/internal/status"
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
	case "status":
		handleStatus()
	case "category":
		handleCategory()
	case "milestone":
		handleMilestone()
	case "issuetype":
		handleIssueType()
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
	fmt.Println("  issue add [--raw] [--yes] --project=<projectIdOrKey> [options]   Create a new issue")
	fmt.Println("  issue update [--raw] [options] <issueKey>   Update an issue")
	fmt.Println("  comment view [--raw] <issueKey> [commentId]   View comments for an issue")
	fmt.Println("  comment add [--raw] [--yes] <issueKey> [message]   Add a comment to an issue")
	fmt.Println("  status list [--raw] <projectId>   List statuses for a project")
	fmt.Println("  category list [--raw] <projectId>   List categories for a project")
	fmt.Println("  milestone list [--raw] <projectId>   List versions/milestones for a project")
	fmt.Println("  issuetype list [--raw] <projectId>   List issue types for a project")
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
	case "add":
		handleIssueAdd()
	case "update":
		handleIssueUpdate()
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
	fmt.Println("  add [--raw] [--yes] --project=<projectIdOrKey> [options]   Create a new issue")
	fmt.Println("  update [--raw] [options] <issueKey>   Update an issue")
}

func handleIssueAdd() {
	// Parse arguments: bgl issue add [--raw] [--yes] --project=<projectIdOrKey> [options]
	args := os.Args[3:]

	opts := issue.AddOptions{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--raw":
			opts.Raw = true
		case arg == "--yes" || arg == "-y":
			opts.Yes = true
		case arg == "-h" || arg == "--help":
			printIssueAddUsage()
			return
		case strings.HasPrefix(arg, "--project="):
			opts.ProjectIDOrKey = strings.TrimPrefix(arg, "--project=")
		case strings.HasPrefix(arg, "--summary="):
			opts.Summary = strings.TrimPrefix(arg, "--summary=")
		case strings.HasPrefix(arg, "--type="):
			opts.IssueTypeID = strings.TrimPrefix(arg, "--type=")
		case strings.HasPrefix(arg, "--priority="):
			opts.PriorityID = strings.TrimPrefix(arg, "--priority=")
		case strings.HasPrefix(arg, "--parent="):
			opts.ParentIssueID = strings.TrimPrefix(arg, "--parent=")
		case strings.HasPrefix(arg, "--description="):
			opts.Description = strings.TrimPrefix(arg, "--description=")
		case strings.HasPrefix(arg, "--assignee="):
			opts.AssigneeID = strings.TrimPrefix(arg, "--assignee=")
		case strings.HasPrefix(arg, "--start-date="):
			opts.StartDate = strings.TrimPrefix(arg, "--start-date=")
		case strings.HasPrefix(arg, "--due-date="):
			opts.DueDate = strings.TrimPrefix(arg, "--due-date=")
		case strings.HasPrefix(arg, "--category="):
			opts.CategoryIDs = strings.TrimPrefix(arg, "--category=")
		case strings.HasPrefix(arg, "--milestone="):
			opts.MilestoneIDs = strings.TrimPrefix(arg, "--milestone=")
		case strings.HasPrefix(arg, "--version="):
			opts.VersionIDs = strings.TrimPrefix(arg, "--version=")
		default:
			fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", arg)
			printIssueAddUsage()
			os.Exit(1)
		}
	}

	if opts.ProjectIDOrKey == "" {
		fmt.Fprintln(os.Stderr, "Error: --project is required")
		printIssueAddUsage()
		os.Exit(1)
	}

	if err := issue.Add(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printIssueAddUsage() {
	fmt.Println("Usage: bgl issue add [options] --project=<projectIdOrKey>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --project=<idOrKey>     Project ID or key (required)")
	fmt.Println("  --summary=<text>        Issue summary (prompted if omitted)")
	fmt.Println("  --type=<id>             Issue type ID (prompted if omitted)")
	fmt.Println("  --priority=<id>         Priority ID (prompted if omitted)")
	fmt.Println("  --parent=<issueId>      Parent issue ID (numeric ID, not issue key)")
	fmt.Println("  --description=<text>    Issue description")
	fmt.Println("  --assignee=<id>         Assignee user ID")
	fmt.Println("  --start-date=<date>     Start date (yyyy-MM-dd)")
	fmt.Println("  --due-date=<date>       Due date (yyyy-MM-dd)")
	fmt.Println("  --category=<id,...>     Category IDs (comma-separated)")
	fmt.Println("  --milestone=<id,...>    Milestone IDs (comma-separated)")
	fmt.Println("  --version=<id,...>      Version IDs (comma-separated)")
	fmt.Println("  --raw                   Output raw JSON response")
	fmt.Println("  --yes, -y               Skip confirmation prompt")
	fmt.Println("  -h, --help              Show this help message")
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

func handleIssueUpdate() {
	// Parse arguments: bgl issue update [--raw] [options] <issueKey>
	args := os.Args[3:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: issue key is required")
		printIssueUpdateUsage()
		os.Exit(1)
	}

	opts := issue.UpdateOptions{}
	var issueKey string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--raw":
			opts.Raw = true
		case arg == "-h" || arg == "--help":
			printIssueUpdateUsage()
			return
		case strings.HasPrefix(arg, "--status="):
			opts.StatusID = strings.TrimPrefix(arg, "--status=")
		case strings.HasPrefix(arg, "--summary="):
			opts.Summary = strings.TrimPrefix(arg, "--summary=")
		case strings.HasPrefix(arg, "--description="):
			opts.Description = strings.TrimPrefix(arg, "--description=")
		case strings.HasPrefix(arg, "--type="):
			opts.IssueTypeID = strings.TrimPrefix(arg, "--type=")
		case strings.HasPrefix(arg, "--priority="):
			opts.PriorityID = strings.TrimPrefix(arg, "--priority=")
		case strings.HasPrefix(arg, "--assignee="):
			opts.AssigneeID = strings.TrimPrefix(arg, "--assignee=")
		case strings.HasPrefix(arg, "--start-date="):
			opts.StartDate = strings.TrimPrefix(arg, "--start-date=")
		case strings.HasPrefix(arg, "--due-date="):
			opts.DueDate = strings.TrimPrefix(arg, "--due-date=")
		case strings.HasPrefix(arg, "--category="):
			opts.CategoryIDs = strings.TrimPrefix(arg, "--category=")
		case strings.HasPrefix(arg, "--milestone="):
			opts.MilestoneIDs = strings.TrimPrefix(arg, "--milestone=")
		case strings.HasPrefix(arg, "--version="):
			opts.VersionIDs = strings.TrimPrefix(arg, "--version=")
		case strings.HasPrefix(arg, "--comment="):
			opts.Comment = strings.TrimPrefix(arg, "--comment=")
		default:
			if issueKey == "" {
				issueKey = arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", arg)
				printIssueUpdateUsage()
				os.Exit(1)
			}
		}
	}

	if issueKey == "" {
		fmt.Fprintln(os.Stderr, "Error: issue key is required")
		printIssueUpdateUsage()
		os.Exit(1)
	}

	if err := issue.Update(issueKey, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printIssueUpdateUsage() {
	fmt.Println("Usage: bgl issue update [options] <issueKey>")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  issueKey                The issue key (e.g., PROJECT-123) or issue ID")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --status=<id>           Status ID to set")
	fmt.Println("  --summary=<text>        Issue summary")
	fmt.Println("  --description=<text>    Issue description")
	fmt.Println("  --type=<id>             Issue type ID")
	fmt.Println("  --priority=<id>         Priority ID")
	fmt.Println("  --assignee=<id>         Assignee user ID")
	fmt.Println("  --start-date=<date>     Start date (yyyy-MM-dd)")
	fmt.Println("  --due-date=<date>       Due date (yyyy-MM-dd)")
	fmt.Println("  --category=<id,...>     Category IDs (comma-separated)")
	fmt.Println("  --milestone=<id,...>    Milestone IDs (comma-separated)")
	fmt.Println("  --version=<id,...>      Version IDs (comma-separated)")
	fmt.Println("  --comment=<text>        Comment to add with the update")
	fmt.Println("  --raw                   Output raw JSON response")
	fmt.Println("  -h, --help              Show this help message")
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

func handleStatus() {
	if len(os.Args) < 3 {
		printStatusUsage()
		os.Exit(1)
	}

	switch os.Args[2] {
	case "list":
		handleStatusList()
	case "-h", "--help", "help":
		printStatusUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown status command: %s\n", os.Args[2])
		printStatusUsage()
		os.Exit(1)
	}
}

func handleStatusList() {
	// Parse arguments: bgl status list [--raw] <projectId>
	args := os.Args[3:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: project ID is required")
		printStatusListUsage()
		os.Exit(1)
	}

	opts := status.ListOptions{}
	var projectID string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			opts.Raw = true
		case "-h", "--help":
			printStatusListUsage()
			return
		default:
			if projectID == "" {
				projectID = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", args[i])
				printStatusListUsage()
				os.Exit(1)
			}
		}
	}

	if projectID == "" {
		fmt.Fprintln(os.Stderr, "Error: project ID is required")
		printStatusListUsage()
		os.Exit(1)
	}

	if err := status.List(projectID, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printStatusUsage() {
	fmt.Println("Usage: bgl status <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list [--raw] <projectId>   List statuses for a project")
}

func printStatusListUsage() {
	fmt.Println("Usage: bgl status list [options] <projectId>")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  projectId   The project ID or project key")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --raw       Output raw JSON response")
	fmt.Println("  -h, --help  Show this help message")
}

func handleCategory() {
	if len(os.Args) < 3 {
		printCategoryUsage()
		os.Exit(1)
	}

	switch os.Args[2] {
	case "list":
		handleCategoryList()
	case "-h", "--help", "help":
		printCategoryUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown category command: %s\n", os.Args[2])
		printCategoryUsage()
		os.Exit(1)
	}
}

func handleCategoryList() {
	// Parse arguments: bgl category list [--raw] <projectId>
	args := os.Args[3:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: project ID is required")
		printCategoryListUsage()
		os.Exit(1)
	}

	opts := category.ListOptions{}
	var projectID string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			opts.Raw = true
		case "-h", "--help":
			printCategoryListUsage()
			return
		default:
			if projectID == "" {
				projectID = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", args[i])
				printCategoryListUsage()
				os.Exit(1)
			}
		}
	}

	if projectID == "" {
		fmt.Fprintln(os.Stderr, "Error: project ID is required")
		printCategoryListUsage()
		os.Exit(1)
	}

	if err := category.List(projectID, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printCategoryUsage() {
	fmt.Println("Usage: bgl category <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list [--raw] <projectId>   List categories for a project")
}

func printCategoryListUsage() {
	fmt.Println("Usage: bgl category list [options] <projectId>")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  projectId   The project ID or project key")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --raw       Output raw JSON response")
	fmt.Println("  -h, --help  Show this help message")
}

func handleMilestone() {
	if len(os.Args) < 3 {
		printMilestoneUsage()
		os.Exit(1)
	}

	switch os.Args[2] {
	case "list":
		handleMilestoneList()
	case "-h", "--help", "help":
		printMilestoneUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown milestone command: %s\n", os.Args[2])
		printMilestoneUsage()
		os.Exit(1)
	}
}

func handleMilestoneList() {
	// Parse arguments: bgl milestone list [--raw] <projectId>
	args := os.Args[3:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: project ID is required")
		printMilestoneListUsage()
		os.Exit(1)
	}

	opts := milestone.ListOptions{}
	var projectID string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			opts.Raw = true
		case "-h", "--help":
			printMilestoneListUsage()
			return
		default:
			if projectID == "" {
				projectID = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", args[i])
				printMilestoneListUsage()
				os.Exit(1)
			}
		}
	}

	if projectID == "" {
		fmt.Fprintln(os.Stderr, "Error: project ID is required")
		printMilestoneListUsage()
		os.Exit(1)
	}

	if err := milestone.List(projectID, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printMilestoneUsage() {
	fmt.Println("Usage: bgl milestone <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list [--raw] <projectId>   List versions/milestones for a project")
}

func printMilestoneListUsage() {
	fmt.Println("Usage: bgl milestone list [options] <projectId>")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  projectId   The project ID or project key")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --raw       Output raw JSON response")
	fmt.Println("  -h, --help  Show this help message")
}

func handleIssueType() {
	if len(os.Args) < 3 {
		printIssueTypeUsage()
		os.Exit(1)
	}

	switch os.Args[2] {
	case "list":
		handleIssueTypeList()
	case "-h", "--help", "help":
		printIssueTypeUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown issuetype command: %s\n", os.Args[2])
		printIssueTypeUsage()
		os.Exit(1)
	}
}

func handleIssueTypeList() {
	// Parse arguments: bgl issuetype list [--raw] <projectId>
	args := os.Args[3:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: project ID is required")
		printIssueTypeListUsage()
		os.Exit(1)
	}

	opts := issuetype.ListOptions{}
	var projectID string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			opts.Raw = true
		case "-h", "--help":
			printIssueTypeListUsage()
			return
		default:
			if projectID == "" {
				projectID = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", args[i])
				printIssueTypeListUsage()
				os.Exit(1)
			}
		}
	}

	if projectID == "" {
		fmt.Fprintln(os.Stderr, "Error: project ID is required")
		printIssueTypeListUsage()
		os.Exit(1)
	}

	if err := issuetype.List(projectID, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printIssueTypeUsage() {
	fmt.Println("Usage: bgl issuetype <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list [--raw] <projectId>   List issue types for a project")
}

func printIssueTypeListUsage() {
	fmt.Println("Usage: bgl issuetype list [options] <projectId>")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  projectId   The project ID or project key")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --raw       Output raw JSON response")
	fmt.Println("  -h, --help  Show this help message")
}
