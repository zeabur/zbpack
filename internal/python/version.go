package python

import (
	"log"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

const defaultPython3Version = "3.10"

// python3Versions is a list of all the Python 3 versions.
var python3Versions = []*semver.Version{
	semver.MustParse("3.0"),
	semver.MustParse("3.1"),
	semver.MustParse("3.2"),
	semver.MustParse("3.3"),
	semver.MustParse("3.4"),
	semver.MustParse("3.5"),
	semver.MustParse("3.6"),
	semver.MustParse("3.7"),
	semver.MustParse("3.8"),
	semver.MustParse("3.9"),
	semver.MustParse("3.10"),
	semver.MustParse("3.11"),
	semver.MustParse("3.12"),
}
var explicitVersionRegex = regexp.MustCompile(`^v?(\d+(?:\.\d+(?:\.\d+)?)?)$`)

func getPython3Version(versionRange string) string {
	if versionRange == "" {
		return defaultPython3Version
	}

	// If there is an explicit version, we pick it.
	if submatch := explicitVersionRegex.FindStringSubmatch(versionRange); len(submatch) > 1 {
		return submatch[1]
	}

	// create a version constraint from versionRange
	constraint, err := semver.NewConstraint(versionRange)
	if err != nil {
		log.Println("invalid python version constraint", err)
		return defaultPython3Version
	}

	// find the nearest version which satisfies the constraint
	for _, version := range python3Versions {
		if constraint.Check(version) {
			return version.Original()
		}
	}

	// when no version satisfies the constraint, return the default version
	log.Printf("no version satisfies the constraint %s", versionRange)
	return defaultPython3Version
}
