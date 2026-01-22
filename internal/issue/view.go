package issue

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/dannygim/bgl/internal/backlog"
)

// ViewOptions contains options for the view command.
type ViewOptions struct {
	Raw bool
}

// View displays an issue by its key or ID.
func View(issueKeyOrID string, opts ViewOptions) error {
	client, err := backlog.NewClient()
	if err != nil {
		return err
	}

	data, err := client.GetIssue(issueKeyOrID)
	if err != nil {
		return err
	}

	if opts.Raw {
		// Pretty print JSON
		var prettyJSON map[string]any
		if err := json.Unmarshal(data, &prettyJSON); err != nil {
			// If pretty print fails, output raw
			fmt.Println(string(data))
			return nil
		}
		formatted, err := json.MarshalIndent(prettyJSON, "", "  ")
		if err != nil {
			fmt.Println(string(data))
			return nil
		}
		fmt.Println(string(formatted))
		return nil
	}

	issue, err := backlog.ParseIssue(data)
	if err != nil {
		return err
	}

	markdown := backlog.FormatIssueMarkdown(issue)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		// Fallback to plain output if renderer fails
		fmt.Print(markdown)
		return nil
	}

	rendered, err := renderer.Render(markdown)
	if err != nil {
		fmt.Print(markdown)
		return nil
	}

	fmt.Print(rendered)
	return nil
}
