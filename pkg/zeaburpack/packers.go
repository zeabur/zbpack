package zeaburpack

import (
	"github.com/salamer/zbpack/internal/bun"
	"github.com/salamer/zbpack/internal/dart"
	"github.com/salamer/zbpack/internal/deno"
	"github.com/salamer/zbpack/internal/dockerfile"
	"github.com/salamer/zbpack/internal/dotnet"
	"github.com/salamer/zbpack/internal/elixir"
	"github.com/salamer/zbpack/internal/gleam"
	"github.com/salamer/zbpack/internal/golang"
	"github.com/salamer/zbpack/internal/java"
	"github.com/salamer/zbpack/internal/nix"
	"github.com/salamer/zbpack/internal/nodejs"
	"github.com/salamer/zbpack/internal/php"
	"github.com/salamer/zbpack/internal/python"
	"github.com/salamer/zbpack/internal/ruby"
	"github.com/salamer/zbpack/internal/rust"
	"github.com/salamer/zbpack/internal/static"
	"github.com/salamer/zbpack/internal/swift"
	"github.com/salamer/zbpack/pkg/packer"
)

// SupportedPackers returns all supported identifiers
func SupportedPackers() []packer.Packer {
	return []packer.Packer{
		nix.NewPacker(),
		dockerfile.NewPacker(),
		dart.NewPacker(),
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
		gleam.NewPacker(),
		swift.NewPacker(),
		static.NewPacker(),
	}
}
