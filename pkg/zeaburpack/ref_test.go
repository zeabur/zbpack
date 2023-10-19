package zeaburpack

import (
	"testing"

	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
)

func TestReferenceConstructor_Construct(t *testing.T) {
	t.Parallel()

	proxy := "zeabur.tld/proxyowo/"
	testMap := map[string]string{
		"docker.io/library/alpine":        proxy + "library/alpine",
		"docker.io/library/alpine:latest": proxy + "library/alpine:latest",
		"docker.io/library/alpine:3.12":   proxy + "library/alpine:3.12",
		"docker.io/library/alpine@sha256:28a392d143b7d67dea564499865c2136371022d0098fde486e9872f0427cdada": proxy + "library/alpine@sha256:28a392d143b7d67dea564499865c2136371022d0098fde486e9872f0427cdada",
		"alpine":                            proxy + "library/alpine",
		"library/alpine":                    proxy + "library/alpine",
		"alpine:3.12":                       proxy + "library/alpine:3.12",
		"docker.io:1234/library/alpine":     "docker.io:1234/library/alpine",
		"other.io/library/alpine":           "other.io/library/alpine",
		"other.io/library/alpine:latest":    "other.io/library/alpine:latest",
		"other.io/library/alpine:3.12":      "other.io/library/alpine:3.12",
		"other.io:1234/library/alpine:3.12": "other.io:1234/library/alpine:3.12",
		"scratch":                           "scratch",
	}

	ref := newReferenceConstructor(&proxy)

	for k, v := range testMap {
		k := k
		v := v
		t.Run(k, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, v, ref.Construct(k))
		})
	}
}

func TestReferenceConstructor_Construct_Stage(t *testing.T) {
	t.Parallel()

	proxy := "zeabur.tld/proxyowo/"
	st := []FromStatement{
		{
			Source: "alpine",
			Stage:  mo.Some("builder"),
		},
		{
			Source: "builder",
			Stage:  mo.None[string](),
		},
	}

	expectedRef := []string{
		"zeabur.tld/proxyowo/library/alpine",
		"builder",
	}

	ref := newReferenceConstructor(&proxy)
	for i, v := range st {
		i := i
		v := v

		assert.Equal(t, expectedRef[i], ref.Construct(v.Source))
		if stage, ok := v.Stage.Get(); ok {
			ref.AddStage(stage)
		}
	}
}
