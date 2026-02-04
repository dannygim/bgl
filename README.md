# bgl
A command line tool for Backlog

## Installation

### Homebrew

```bash
brew tap dannygim/tap                                                                         
brew install bgl                                                                              
bgl --help
```

## Usage

### Authentication

#### Login

Login to Backlog using OAuth 2.0:

```bash
bgl auth login
```

This will:
1. Prompt you to enter your Backlog space (e.g., `myspace.backlog.com` or `myspace.backlog.jp`)
2. Open your browser for authentication
3. After successful login, save the access token and refresh token to `~/.config/bgl/config.json`

#### Logout

Logout and remove stored tokens:

```bash
bgl auth logout
```

This will remove the access token and refresh token from `~/.config/bgl/config.json`.

### Issue

#### View Issue

View an issue by its key or ID:

```bash
bgl issue view PROJECT-123
```

This displays the issue in Markdown format with the following information:
- Summary
- Assignee
- Status
- Description

To output the raw JSON response:

```bash
bgl issue view --raw PROJECT-123
```

#### Update Issue

Update an issue's status:

```bash
bgl issue update --status=2 PROJECT-123
```

This updates the issue status and displays the updated issue in Markdown format (same as `issue view`).

To get the available status IDs for a project, use `bgl status list <projectId>`.

To output the raw JSON response:

```bash
bgl issue update --raw --status=2 PROJECT-123
```

### Comment

#### View Comments

View all comments for an issue:

```bash
bgl comment view PROJECT-123
```

This displays comments in Markdown format with the following information:
- Comment Id
- User (name and email)
- Datetime
- Content

Comments are separated by `---`.

To view a specific comment by ID:

```bash
bgl comment view PROJECT-123 12345
```

To output the raw JSON response:

```bash
bgl comment view --raw PROJECT-123
bgl comment view --raw PROJECT-123 12345
```

#### Add Comment

Add a comment to an issue interactively (prompts for message input):

```bash
bgl comment add PROJECT-123
```

Add a comment with a message directly:

```bash
bgl comment add PROJECT-123 "This is my comment"
```

When providing a message directly, you will be prompted to confirm before adding the comment. To skip the confirmation prompt, use `--yes` or `-y`:

```bash
bgl comment add --yes PROJECT-123 "This is my comment"
bgl comment add -y PROJECT-123 "This is my comment"
```

After successfully adding a comment, the URL to the comment will be displayed.

To output the raw JSON response:

```bash
bgl comment add --raw PROJECT-123 "This is my comment"
```

### Status

#### List Statuses

List all statuses for a project:

```bash
bgl status list PROJECT
```

This displays the project statuses in Markdown format:

```
## Status
- Open (id: 1)
- Close (id: 2)
```

To output the raw JSON response:

```bash
bgl status list --raw PROJECT
```

### Other Commands

```bash
bgl --help      # Show help message
bgl --version   # Show version information
```

## Configuration

Tokens are stored in `~/.config/bgl/config.json`:

```json
{
  "space": "myspace.backlog.com",
  "access_token": "...",
  "refresh_token": "..."
}
```

## Development

### Building

To build with OAuth credentials embedded at build time:

```bash
go build -ldflags "-X github.com/dannygim/bgl/internal/config.ClientID=YOUR_CLIENT_ID -X github.com/dannygim/bgl/internal/config.ClientSecret=YOUR_CLIENT_SECRET" -o bgl ./cmd/bgl
```

### Setting up OAuth Application

1. Go to your Backlog space settings
2. Navigate to Developer Applications
3. Register a new application
4. Set the redirect URI to `http://localhost:18765`
5. Note your Client ID and Client Secret
