package attachment

import (
	"fmt"
	"os"

	"github.com/dannygim/bgl/internal/backlog"
)

// DownloadOptions contains options for the download command.
type DownloadOptions struct {
	Output string
}

// Download downloads an issue's attachment file and saves it to disk.
func Download(issueKeyOrID string, attachmentID string, opts DownloadOptions) error {
	client, err := backlog.NewClient()
	if err != nil {
		return err
	}

	data, filename, err := client.DownloadIssueAttachment(issueKeyOrID, attachmentID)
	if err != nil {
		return err
	}

	path := opts.Output
	if path == "" {
		path = filename
	}
	if path == "" {
		path = "attachment-" + attachmentID
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Downloaded: %s (%d bytes)\n", path, len(data))
	return nil
}
