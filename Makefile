.PHONY: build clean

build:
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | xargs) && \
		go build -ldflags "-X github.com/dannygim/bgl/internal/config.ClientID=$$BACKLOG_CLIENT_ID -X github.com/dannygim/bgl/internal/config.ClientSecret=$$BACKLOG_CLIENT_SECRET" -o bgl ./cmd/bgl; \
	else \
		echo "Error: .env file not found"; \
		exit 1; \
	fi

clean:
	rm -f bgl
	rm -rf dist
