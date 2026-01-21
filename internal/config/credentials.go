package config

// ClientID and ClientSecret are set at build time using ldflags.
// Build with: go build -ldflags "-X github.com/dannygim/bgl/internal/config.ClientID=xxx -X github.com/dannygim/bgl/internal/config.ClientSecret=yyy"
var (
	ClientID     string
	ClientSecret string
)
