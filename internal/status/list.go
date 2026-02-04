package status

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/dannygim/bgl/internal/backlog"
)

// ListOptions contains options for the list command.
type ListOptions struct {
	Raw bool
}

// List displays the status list for a project.
func List(projectIDOrKey string, opts ListOptions) error {
	client, err := backlog.NewClient()
	if err != nil {
		return err
	}

	data, err := client.GetProjectStatuses(projectIDOrKey)
	if err != nil {
		return err
	}

	if opts.Raw {
		// Pretty print JSON
		var prettyJSON []any
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

	statuses, err := backlog.ParseProjectStatuses(data)
	if err != nil {
		return err
	}

	markdown := backlog.FormatProjectStatusesMarkdown(statuses)

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
