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
	Major uint32
	// Minor version.
	Minor uint32
	// MinorSet indicates if the minor version is set.
	MinorSet bool
	// Patch version.
	Patch uint32
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
		if parts[i] == "*" {
			break // ignorable
		}

		switch i {
		case 0: // major
			parsedVersionMajor, err := strconv.ParseUint(parts[i], 10, 32)
			if err != nil {
				return parsedVersion, err
			}
			parsedVersion.Major = uint32(parsedVersionMajor)
		case 1: // minor
			parsedVersionMinor, err := strconv.ParseUint(parts[i], 10, 32)
			if err != nil {
				return parsedVersion, err
			}
			parsedVersion.Minor = uint32(parsedVersionMinor)
			parsedVersion.MinorSet = true
		case 2: // patch
			patchStr, prereleaseStr, ok := strings.Cut(parts[i], "-")
			if ok {
				parsedVersionPatch, err := strconv.ParseUint(patchStr, 10, 32)
				if err != nil {
					return parsedVersion, err
				}
				parsedVersion.Patch = uint32(parsedVersionPatch)
				parsedVersion.PatchSet = true
				parsedVersion.Prerelease = prereleaseStr
			} else {
				parsedVersionPatch, err := strconv.ParseUint(parts[i], 10, 32)
				if err != nil {
					return parsedVersion, err
				}
				parsedVersion.Patch = uint32(parsedVersionPatch)
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
	minus := false // for less-than operator, we should minus one bit

	for _, r := range constraintList {
		cleanR := r
		majorOnly := false

		if minorSpecifierRegex.MatchString(r) {
			cleanR = minorSpecifierRegex.ReplaceAllString(r, "")
		}
		if majorSpecifierRegex.MatchString(r) {
			cleanR = majorSpecifierRegex.ReplaceAllString(r, "")
			majorOnly = true
			minus = r[0] == '<' && r[1] != '='
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
			minorBit := parsedVersion.Minor & minorMask
			thisVersion |= minorBit
		} else {
			thisVersion |= noMinorFlag
		}

		majorBit := (parsedVersion.Major << 18) & majorMask
		thisVersion |= majorBit | initializedFlag
		if thisVersion > determinedVersion {
			determinedVersion = thisVersion
		}
	}

	if determinedVersion&initializedFlag == 0 {
		return defaultVersion
	}

	major := (determinedVersion & majorMask) >> 18
	if determinedVersion&noMinorFlag != 0 {
		if minus && major > 0 {
			major--
		}

		majorString := strconv.FormatUint(uint64(major), 10)
		return majorString
	}

	minor := determinedVersion & minorMask
	majorString := strconv.FormatUint(uint64(major), 10)
	minorString := strconv.FormatUint(uint64(minor), 10)
	return majorString + "." + minorString
}
