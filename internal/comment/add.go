package comment

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dannygim/bgl/internal/backlog"
)

// AddOptions contains options for the add command.
type AddOptions struct {
	Raw bool
	Yes bool
}

// Add adds a comment to an issue.
func Add(issueKeyOrID string, content string, opts AddOptions) error {
	// If content is empty, prompt for input
	if content == "" {
		if err := huh.NewText().
			Title("Comment").
			Description("Enter your comment").
			Value(&content).
			Run(); err != nil {
			return fmt.Errorf("failed to get comment input: %w", err)
		}

		if strings.TrimSpace(content) == "" {
			return fmt.Errorf("comment content cannot be empty")
		}
	}

	// Show confirmation unless --yes is specified
	if !opts.Yes {
		var confirm bool
		if err := huh.NewConfirm().
			Title("Add Comment?").
			Description(fmt.Sprintf("Issue: %s\nContent:\n%s", issueKeyOrID, content)).
			Affirmative("Confirm").
			Negative("Cancel").
			Value(&confirm).
			Run(); err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}

		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	client, err := backlog.NewClient()
	if err != nil {
		return err
	}

	data, err := client.AddComment(issueKeyOrID, content)
	if err != nil {
		return err
	}

	if opts.Raw {
		// Pretty print JSON
		var prettyJSON map[string]any
		if err := json.Unmarshal(data, &prettyJSON); err != nil {
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

	// Parse the response to get the comment ID
	comment, err := backlog.ParseComment(data)
	if err != nil {
		return err
	}

	// Build and display the comment URL
	space := client.GetSpace()
	commentURL := fmt.Sprintf("https://%s/view/%s#comment-%d", space, issueKeyOrID, comment.ID)

	fmt.Println("Comment added successfully!")
	fmt.Printf("URL: %s\n", commentURL)

	return nil
}
