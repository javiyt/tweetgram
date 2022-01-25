//go:build tools
// +build tools

package main

import (
	_ "github.com/c-sto/encembed"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/google/wire/cmd/wire"
	_ "github.com/mailru/easyjson/easyjson"
	_ "github.com/vektra/mockery/v2"
	_ "github.com/zimmski/go-mutesting/cmd/go-mutesting"
	_ "golang.org/x/tools/cmd/stringer"
	_ "mvdan.cc/gofumpt"
)
