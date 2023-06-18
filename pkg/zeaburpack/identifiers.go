package zeaburpack

import (
	"github.com/zeabur/zbpack/internal/deno"
	dockerfilePkg "github.com/zeabur/zbpack/internal/dockerfile"
	"github.com/zeabur/zbpack/internal/dotnet"
	"github.com/zeabur/zbpack/internal/elixir"
	"github.com/zeabur/zbpack/internal/golang"
	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/internal/ruby"
	"github.com/zeabur/zbpack/internal/rust"
	"github.com/zeabur/zbpack/internal/static"
	"github.com/zeabur/zbpack/pkg/plan"
)

// SupportedIdentifiers returns all supported identifiers
// note that they are in the order of priority
func SupportedIdentifiers() []plan.Identifier {
	return []plan.Identifier{
		dockerfilePkg.NewIdentifier(),
		php.NewIdentifier(),
		nodejs.NewIdentifier(),
		golang.NewIdentifier(),
		python.NewIdentifier(),
		ruby.NewIdentifier(),
		java.NewIdentifier(),
		deno.NewIdentifier(),
		rust.NewIdentifier(),
		dotnet.NewIdentifier(),
		elixir.NewIdentifier(),
		static.NewIdentifier(),
	}
}
