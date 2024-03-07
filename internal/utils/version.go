package utils

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semver version.
type Version struct {
	// Major version.
	Major int
	// Minor version.
	Minor int
	// MinorSet indicates if the minor version is set.
	MinorSet bool
	// Patch version.
	Patch int
	// PatchSet indicates if the patch version is set.
	PatchSet bool
	// Prerelease version. Empty = not set.
	Prerelease string
}

// SplitVersion splits the version string into major, minor, patch, and prerelease.
//
// If the version string is "8.0.0-beta1", the result will be 8, 0, 0, "beta1".
// If the version is a wildcard, the field will be -1.
func SplitVersion(version string) (Version, error) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	parsedVersion := Version{}

	if version == "" {
		return parsedVersion, fmt.Errorf("empty version")
	}

	parts := strings.SplitN(version, ".", 3)
	for i := 0; i < len(parts); i++ {
		var err error

		if parts[i] == "*" {
			break // ignorable
		}

		switch i {
		case 0: // major
			parsedVersion.Major, err = strconv.Atoi(parts[i])
			if err != nil {
				return parsedVersion, err
			}
		case 1: // minor
			parsedVersion.Minor, err = strconv.Atoi(parts[i])
			if err != nil {
				return parsedVersion, err
			}
			parsedVersion.MinorSet = true
		case 2: // patch
			patchStr, prereleaseStr, ok := strings.Cut(parts[i], "-")
			if ok {
				parsedVersion.Patch, err = strconv.Atoi(patchStr)
				if err != nil {
					return parsedVersion, err
				}
				parsedVersion.PatchSet = true
				parsedVersion.Prerelease = prereleaseStr
			} else {
				parsedVersion.Patch, err = strconv.Atoi(parts[i])
				if err != nil {
					return parsedVersion, err
				}
				parsedVersion.PatchSet = true
			}
		}
	}

	return parsedVersion, nil
}

// ConstraintToVersion converts the Semver version constraint to a minor-only version.
func ConstraintToVersion(constraints string, defaultVersion string) string {
	majorSpecifierRegex := regexp.MustCompile(`^[><^]=?`)
	minorSpecifierRegex := regexp.MustCompile(`^[~=]`)

	// for example, ^8 ~8.3
	constraintList := strings.Split(strings.Replace(constraints, "||", " ", -1), " ")

	// From the lower bit to the upper bit:
	//     15: minor
	//      1: no-minor (no-minor has higher priority than minor)
	//     15: major
	//      1: initialized
	determinedVersion := uint32(0)
	minorMask := uint32((1 << 16) - 1)
	majorMask := uint32((1 << 31) - 1 ^ (1 << 17) - 1)
	noMinorFlag := uint32(1 << 16)
	initializedFlag := uint32(1 << 31)

	for _, r := range constraintList {
		cleanR := r
		majorOnly := false

		if minorSpecifierRegex.MatchString(r) {
			cleanR = minorSpecifierRegex.ReplaceAllString(r, "")
		}
		if majorSpecifierRegex.MatchString(r) {
			cleanR = majorSpecifierRegex.ReplaceAllString(r, "")
			majorOnly = true
		}

		parsedVersion, err := SplitVersion(cleanR)
		if err != nil {
			log.Println("invalid version", err)
			continue
		}

		if !parsedVersion.MinorSet {
			majorOnly = true
		}

		thisVersion := uint32(0)
		if !majorOnly {
			minorBit := uint32(parsedVersion.Minor) & minorMask
			thisVersion |= minorBit
		} else {
			thisVersion |= noMinorFlag
		}

		majorBit := (uint32(parsedVersion.Major) << 18) & majorMask
		thisVersion |= majorBit | initializedFlag
		if thisVersion > determinedVersion {
			determinedVersion = thisVersion
		}
	}

	if determinedVersion&initializedFlag == 0 {
		return defaultVersion
	}

	major := (determinedVersion & majorMask) >> 18
	majorString := strconv.FormatUint(uint64(major), 10)
	if determinedVersion&noMinorFlag != 0 {
		return majorString
	}

	minor := determinedVersion & minorMask
	minorString := strconv.FormatUint(uint64(minor), 10)
	return majorString + "." + minorString
}
