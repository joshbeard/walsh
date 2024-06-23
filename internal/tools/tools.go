//go:build tools

package tools

import (
	_ "github.com/boumenot/gocover-cobertura"               // Test coverage reporting (make coverage).
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // Linting (make lint).
	_ "github.com/segmentio/golines"                        // Long line fixer.
	_ "golang.org/x/vuln/cmd/govulncheck"                   // Code vulnerability checks (make check-vuln).
	_ "mvdan.cc/gofumpt"                                    // Formatting and linting (make gofumpt).
)
