package zeaburpack

import (
	"github.com/zeabur/zbpack/internal/bun"
	"github.com/zeabur/zbpack/internal/deno"
	"github.com/zeabur/zbpack/internal/dockerfile"
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
	"github.com/zeabur/zbpack/internal/swift"
	"github.com/zeabur/zbpack/pkg/packer"
)

// SupportedPackers returns all supported identifiers
func SupportedPackers() []packer.Packer {
	return []packer.Packer{
		dockerfile.NewPacker(),
		php.NewPacker(),
		bun.NewPacker(),
		nodejs.NewPacker(),
		golang.NewPacker(),
		python.NewPacker(),
		ruby.NewPacker(),
		java.NewPacker(),
		deno.NewPacker(),
		rust.NewPacker(),
		dotnet.NewPacker(),
		elixir.NewPacker(),
		swift.NewPacker(),
		static.NewPacker(),
	}
}
