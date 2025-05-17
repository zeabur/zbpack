package zeaburpack

import (
	"github.com/zeabur/zbpack/internal/bun"
	"github.com/zeabur/zbpack/internal/dart"
	"github.com/zeabur/zbpack/internal/deno"
	"github.com/zeabur/zbpack/internal/dockerfile"
	"github.com/zeabur/zbpack/internal/dotnet"
	"github.com/zeabur/zbpack/internal/elixir"
	"github.com/zeabur/zbpack/internal/gleam"
	"github.com/zeabur/zbpack/internal/golang"
	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nix"
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
func SupportedPackers() []packer.V2 {
	return []packer.V2{
		packer.WrapV2(nix.NewPacker()),
		dockerfile.NewPacker(),
		packer.WrapV2(dart.NewPacker()),
		packer.WrapV2(php.NewPacker()),
		packer.WrapV2(bun.NewPacker()),
		packer.WrapV2(nodejs.NewPacker()),
		packer.WrapV2(golang.NewPacker()),
		packer.WrapV2(python.NewPacker()),
		packer.WrapV2(ruby.NewPacker()),
		packer.WrapV2(java.NewPacker()),
		packer.WrapV2(deno.NewPacker()),
		packer.WrapV2(rust.NewPacker()),
		packer.WrapV2(dotnet.NewPacker()),
		packer.WrapV2(elixir.NewPacker()),
		packer.WrapV2(gleam.NewPacker()),
		packer.WrapV2(swift.NewPacker()),
		packer.WrapV2(static.NewPacker()),
	}
}
