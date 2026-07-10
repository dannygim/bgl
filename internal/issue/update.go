package issue

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/charmbracelet/glamour"
	"github.com/dannygim/bgl/internal/backlog"
)

// UpdateOptions contains options for the update command.
type UpdateOptions struct {
	Raw          bool
	StatusID     string
	Summary      string
	Description  string
	IssueTypeID  string
	PriorityID   string
	AssigneeID   string
	StartDate    string
	DueDate      string
	CategoryIDs  string
	MilestoneIDs string
	VersionIDs   string
	Comment      string
}

// Update updates an issue and displays the result.
func Update(issueKeyOrID string, opts UpdateOptions) error {
	client, err := backlog.NewClient()
	if err != nil {
		return err
	}

	data := url.Values{}
	if opts.StatusID != "" {
		data.Set("statusId", opts.StatusID)
	}
	if opts.Summary != "" {
		data.Set("summary", opts.Summary)
	}
	if opts.Description != "" {
		data.Set("description", opts.Description)
	}
	if opts.IssueTypeID != "" {
		data.Set("issueTypeId", opts.IssueTypeID)
	}
	if opts.PriorityID != "" {
		data.Set("priorityId", opts.PriorityID)
	}
	if opts.AssigneeID != "" {
		data.Set("assigneeId", opts.AssigneeID)
	}
	if opts.StartDate != "" {
		data.Set("startDate", opts.StartDate)
	}
	if opts.DueDate != "" {
		data.Set("dueDate", opts.DueDate)
	}
	addMultiValues(data, "categoryId[]", opts.CategoryIDs)
	addMultiValues(data, "milestoneId[]", opts.MilestoneIDs)
	addMultiValues(data, "versionId[]", opts.VersionIDs)
	if opts.Comment != "" {
		data.Set("comment", opts.Comment)
	}

	if len(data) == 0 {
		return fmt.Errorf("no update options specified")
	}

	result, err := client.UpdateIssue(issueKeyOrID, data)
	if err != nil {
		return err
	}

	if opts.Raw {
		// Pretty print JSON
		var prettyJSON map[string]any
		if err := json.Unmarshal(result, &prettyJSON); err != nil {
			// If pretty print fails, output raw
			fmt.Println(string(result))
			return nil
		}
		formatted, err := json.MarshalIndent(prettyJSON, "", "  ")
		if err != nil {
			fmt.Println(string(result))
			return nil
		}
		fmt.Println(string(formatted))
		return nil
	}

	issue, err := backlog.ParseIssue(result)
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
