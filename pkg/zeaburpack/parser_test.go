package zeaburpack_test

import (
	"testing"

	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/zeaburpack"
)

func TestParseFrom(t *testing.T) {
	t.Parallel()

	testmap := map[string]zeaburpack.FromStatement{
		"FROM alpine": {
			Source: "alpine",
			Stage:  mo.None[string](),
		},
		"FROM alpine:3.12": {
			Source: "alpine:3.12",
			Stage:  mo.None[string](),
		},
		"FROM alpine AS builder": {
			Source: "alpine",
			Stage:  mo.Some("builder"),
		},
		"FROM alpine:3.12 AS builder": {
			Source: "alpine:3.12",
			Stage:  mo.Some("builder"),
		},
		"FROM alpine:3.12 AS builder  # comment": {
			Source: "alpine:3.12",
			Stage:  mo.Some("builder"),
		},
		"FROM alpine:3.12 AS builder # comment": {
			Source: "alpine:3.12",
			Stage:  mo.Some("builder"),
		},
		"FROM    alpine:3.12 AS  builder": {
			Source: "alpine:3.12",
			Stage:  mo.Some("builder"),
		},
		"FROM alpine:3.12    AS  builder": {
			Source: "alpine:3.12",
			Stage:  mo.Some("builder"),
		},
		"FROM alpine # comment": {
			Source: "alpine",
			Stage:  mo.None[string](),
		},
		"FROM alpine:3.12 # comment": {
			Source: "alpine:3.12",
			Stage:  mo.None[string](),
		},
		"FROM alpine AS builder # comment": {
			Source: "alpine",
			Stage:  mo.Some("builder"),
		},
		"FROM --platform=linux/amd64 alpine AS builder": {
			Source: "alpine",
			Stage:  mo.Some("builder"),
		},
		"FROM --platform=$BUILDERPLATFORM alpine:3.12 AS builder": {
			Source: "alpine:3.12",
			Stage:  mo.Some("builder"),
		},
	}

	for k, v := range testmap {
		k := k
		v := v
		t.Run(k, func(t *testing.T) {
			t.Parallel()

			pf, ok := zeaburpack.ParseFrom(k)
			assert.True(t, ok)
			assert.Equal(t, v, pf)
		})
	}
}

func TestParseFrom_String(t *testing.T) {
	t.Parallel()

	testmap := []struct {
		Input  zeaburpack.FromStatement
		Output string
	}{
		{
			Input: zeaburpack.FromStatement{
				Source: "alpine",
				Stage:  mo.None[string](),
			},
			Output: "FROM alpine",
		},
		{
			Input: zeaburpack.FromStatement{
				Source: "alpine:3.12",
				Stage:  mo.None[string](),
			},
			Output: "FROM alpine:3.12",
		},
		{
			Input: zeaburpack.FromStatement{
				Source: "alpine",
				Stage:  mo.Some("builder"),
			},
			Output: "FROM alpine AS builder",
		},
		{
			Input: zeaburpack.FromStatement{
				Source: "alpine:3.12",
				Stage:  mo.Some("builder"),
			},
			Output: "FROM alpine:3.12 AS builder",
		},
	}

	for _, tv := range testmap {
		tv := tv

		t.Run(tv.Output, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tv.Output, tv.Input.String())
		})
	}
}

func TestParseFrom_String_AllowReplacing(t *testing.T) {
	fs := zeaburpack.FromStatement{
		Source: "alpine",
		Stage:  mo.Some("builder"),
	}
	assert.Equal(t, "FROM alpine AS builder", fs.String())

	fs.Stage = mo.None[string]()
	assert.Equal(t, "FROM alpine", fs.String())

	fs.Stage = mo.Some("builder")
	assert.Equal(t, "FROM alpine AS builder", fs.String())

	fs.Source = "alpine:3.12"
	assert.Equal(t, "FROM alpine:3.12 AS builder", fs.String())

	fs.Stage = mo.None[string]()
	assert.Equal(t, "FROM alpine:3.12", fs.String())
}
