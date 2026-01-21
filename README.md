# bgl
A command line tool for Backlog

## Installation

### Build from source

```bash
# Build with OAuth credentials
go build -ldflags "-X github.com/dannygim/bgl/internal/config.ClientID=YOUR_CLIENT_ID -X github.com/dannygim/bgl/internal/config.ClientSecret=YOUR_CLIENT_SECRET" -o bgl ./cmd/bgl
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
