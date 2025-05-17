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
	"github.com/zeabur/zbpack/pkg/plan"
)

// SupportedIdentifiers returns all supported identifiers
// note that they are in the order of priority
func SupportedIdentifiers(config plan.ImmutableProjectConfiguration) []plan.IdentifierV2 {
	identifiers := []plan.IdentifierV2{
		plan.WrapV2(dart.NewIdentifier()),
		plan.WrapV2(php.NewIdentifier()),
		plan.WrapV2(ruby.NewIdentifier()),
		plan.WrapV2(bun.NewIdentifier()),
		plan.WrapV2(python.NewIdentifier()),
		plan.WrapV2(nodejs.NewIdentifier()),
		plan.WrapV2(golang.NewIdentifier()),
		plan.WrapV2(java.NewIdentifier()),
		plan.WrapV2(deno.NewIdentifier()),
		plan.WrapV2(rust.NewIdentifier()),
		plan.WrapV2(dotnet.NewIdentifier()),
		plan.WrapV2(elixir.NewIdentifier()),
		plan.WrapV2(gleam.NewIdentifier()),
		plan.WrapV2(swift.NewIdentifier()),
		plan.WrapV2(static.NewIdentifier()),
	}

	if !plan.Cast(config.Get("ignore_nix"), plan.ToWeakBoolE).TakeOr(false) {
		identifiers = append([]plan.IdentifierV2{plan.WrapV2(nix.NewIdentifier())}, identifiers...)
	}

	// if ignore_dockerfile in config is true, or ZBPACK_IGNORE_DOCKERFILE is true, ignore dockerfile
	if !plan.Cast(config.Get("ignore_dockerfile"), plan.ToWeakBoolE).TakeOr(false) {
		identifiers = append([]plan.IdentifierV2{dockerfile.NewIdentifier()}, identifiers...)
	}

	return identifiers
}
