package issue

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dannygim/bgl/internal/backlog"
)

// AddOptions contains options for the add command.
type AddOptions struct {
	Raw            bool
	Yes            bool
	ProjectIDOrKey string
	Summary        string
	IssueTypeID    string
	PriorityID     string
	ParentIssueID  string
	Description    string
	AssigneeID     string
	StartDate      string
	DueDate        string
	CategoryIDs    string
	MilestoneIDs   string
	VersionIDs     string
}

// Add creates a new issue. Required fields not given as options are
// prompted interactively.
func Add(opts AddOptions) error {
	client, err := backlog.NewClient()
	if err != nil {
		return err
	}

	// Resolve the project key to its numeric ID (the Add Issue API only
	// accepts a numeric projectId).
	projectData, err := client.GetProject(opts.ProjectIDOrKey)
	if err != nil {
		return err
	}
	project, err := backlog.ParseProject(projectData)
	if err != nil {
		return err
	}

	summary := opts.Summary
	if summary == "" {
		if err := huh.NewInput().
			Title("Summary").
			Description("Enter the issue summary").
			Value(&summary).
			Run(); err != nil {
			return fmt.Errorf("failed to get summary input: %w", err)
		}

		if strings.TrimSpace(summary) == "" {
			return fmt.Errorf("summary cannot be empty")
		}
	}

	issueTypeID := opts.IssueTypeID
	if issueTypeID == "" {
		data, err := client.GetIssueTypes(opts.ProjectIDOrKey)
		if err != nil {
			return err
		}
		issueTypes, err := backlog.ParseIssueTypes(data)
		if err != nil {
			return err
		}
		if len(issueTypes) == 0 {
			return fmt.Errorf("no issue types found in project %s", opts.ProjectIDOrKey)
		}

		options := make([]huh.Option[string], len(issueTypes))
		for i, issueType := range issueTypes {
			options[i] = huh.NewOption(issueType.Name, strconv.Itoa(issueType.ID))
		}
		if err := huh.NewSelect[string]().
			Title("Issue Type").
			Options(options...).
			Value(&issueTypeID).
			Run(); err != nil {
			return fmt.Errorf("failed to select issue type: %w", err)
		}
	}

	priorityID := opts.PriorityID
	if priorityID == "" {
		data, err := client.GetPriorities()
		if err != nil {
			return err
		}
		priorities, err := backlog.ParsePriorities(data)
		if err != nil {
			return err
		}
		if len(priorities) == 0 {
			return fmt.Errorf("no priorities found")
		}

		options := make([]huh.Option[string], len(priorities))
		for i, priority := range priorities {
			options[i] = huh.NewOption(priority.Name, strconv.Itoa(priority.ID))
		}
		if err := huh.NewSelect[string]().
			Title("Priority").
			Options(options...).
			Value(&priorityID).
			Run(); err != nil {
			return fmt.Errorf("failed to select priority: %w", err)
		}
	}

	// Show confirmation unless --yes is specified
	if !opts.Yes {
		var confirm bool
		if err := huh.NewConfirm().
			Title("Create Issue?").
			Description(fmt.Sprintf("Project: %s\nSummary: %s", project.ProjectKey, summary)).
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

	data := url.Values{}
	data.Set("projectId", strconv.Itoa(project.ID))
	data.Set("summary", summary)
	data.Set("issueTypeId", issueTypeID)
	data.Set("priorityId", priorityID)
	if opts.ParentIssueID != "" {
		data.Set("parentIssueId", opts.ParentIssueID)
	}
	if opts.Description != "" {
		data.Set("description", opts.Description)
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

	result, err := client.AddIssue(data)
	if err != nil {
		return err
	}

	if opts.Raw {
		// Pretty print JSON
		var prettyJSON map[string]any
		if err := json.Unmarshal(result, &prettyJSON); err != nil {
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

	created, err := backlog.ParseIssue(result)
	if err != nil {
		return err
	}

	issueURL := fmt.Sprintf("https://%s/view/%s", client.GetSpace(), created.IssueKey)

	fmt.Println("Issue created successfully!")
	fmt.Printf("Key: %s\n", created.IssueKey)
	fmt.Printf("URL: %s\n", issueURL)

	return nil
}

// addMultiValues splits a comma-separated ID list and adds each value under key.
func addMultiValues(data url.Values, key string, ids string) {
	if ids == "" {
		return
	}
	for id := range strings.SplitSeq(ids, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			data.Add(key, id)
		}
	}
}
