package comment

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

// ViewList displays comments for an issue.
func ViewList(issueKeyOrID string, opts ViewOptions) error {
	client, err := backlog.NewClient()
	if err != nil {
		return err
	}

	data, err := client.GetComments(issueKeyOrID)
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

	comments, err := backlog.ParseComments(data)
	if err != nil {
		return err
	}

	if len(comments) == 0 {
		fmt.Println("No comments found.")
		return nil
	}

	markdown := backlog.FormatCommentsMarkdown(comments)

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

// View displays a single comment.
func View(issueKeyOrID string, commentID string, opts ViewOptions) error {
	client, err := backlog.NewClient()
	if err != nil {
		return err
	}

	data, err := client.GetComment(issueKeyOrID, commentID)
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

	comment, err := backlog.ParseComment(data)
	if err != nil {
		return err
	}

	markdown := backlog.FormatCommentMarkdown(comment)

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
