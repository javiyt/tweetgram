//go:build tools
// +build tools

package main

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/google/wire/cmd/wire"
	_ "github.com/vektra/mockery/v2"
	_ "golang.org/x/tools/cmd/stringer"
	_ "github.com/mailru/easyjson/easyjson"
)
