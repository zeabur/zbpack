package zeaburpack

import (
	"log"
	"strings"

	"github.com/distribution/reference"
)

// referenceConstructor constructs the standardized image references
// for the image builder. It also provides the ability to prepend
// a proxy registry to the image reference.
type referenceConstructor struct {
	// proxyRegistry indicates the registry to be used for the image.
	// It is designed for Harbor's Proxy Cache (Ref 1).
	//
	// It will be prepended to any image in zeaburpack. If an image
	// contains a domain in the image name, the proxy registry will not
	// be applied to the image (however, if the domain is `docker.io`,
	// it will still be replaced).
	//
	// (Ref 1) https://goharbor.io/docs/2.1.0/administration/configure-proxy-cache/
	proxyRegistry *string

	// stage is a special image reference that will not be extended
	stage map[string]struct{}
}

func newReferenceConstructor(proxyRegistry *string) referenceConstructor {
	// Add `/` suffix to the proxy registry if it is not empty.
	if proxyRegistry != nil {
		cleanedProxyRegistry := strings.TrimSuffix(*proxyRegistry, "/") + "/"
		return referenceConstructor{proxyRegistry: &cleanedProxyRegistry}
	}

	return referenceConstructor{
		proxyRegistry: proxyRegistry,
	}
}

// Construct constructs a new image reference from the given raw
func (rc *referenceConstructor) Construct(rawRefString string) string {
	proxyRegistryPtr := rc.proxyRegistry

	// If the proxy registry is not set, we don't need to do anything.
	if proxyRegistryPtr == nil || len(*proxyRegistryPtr) == 0 {
		return rawRefString
	}

	// If ref is `scratch`, we skip.
	if rawRefString == "scratch" {
		return rawRefString
	}

	// If ref is a stage, we skip.
	if _, ok := rc.stage[rawRefString]; ok {
		return rawRefString
	}

	// Safety: ptr != nil.
	proxyRegistry := *proxyRegistryPtr

	// Parse the user-provided reference.
	ref, err := reference.ParseAnyReference(rawRefString)
	if err != nil {
		log.Println("failed to parse image reference:", err.Error())
		return rawRefString
	}

	// If the image reference contains a domain, we don't need to
	// apply the proxy registry (unless the domain is `docker.io`).
	imageRef, ok := ref.(reference.Named)
	if !ok {
		return rawRefString
	}

	domain := reference.Domain(imageRef)
	// If the domain is not `docker.io`, we leave it as it is.
	if domain != "docker.io" {
		return rawRefString
	}

	// Construct a new reference with the proxy registry.
	path := reference.Path(imageRef)
	switch ref := imageRef.(type) {
	case reference.NamedTagged:
		tag := ref.Tag()

		return proxyRegistry + path + ":" + tag
	case reference.Canonical:
		digest := ref.Digest()

		return proxyRegistry + path + "@" + digest.String()
	default:
		return proxyRegistry + path
	}
}

// AddStage marks the given image reference as a stage, so we won't
// extend such a special stage as a dependency.
//
// It is not thread-safe.
func (rc *referenceConstructor) AddStage(stage string) {
	if rc.stage == nil {
		rc.stage = make(map[string]struct{})
	}

	rc.stage[stage] = struct{}{}
}
