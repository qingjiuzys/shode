package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version (major.minor.patch)
// Format: MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
// Example: 1.2.3-alpha.1+build123
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   []string // Pre-release identifiers (e.g., ["alpha", "1"])
	Build string   // Build metadata (ignored in comparisons)
}

// ParseVersion parses a version string into a Version struct
func ParseVersion(v string) (*Version, error) {
	if v == "" {
		return nil, fmt.Errorf("empty version string")
	}

	// Split build metadata
	parts := strings.SplitN(v, "+", 2)
	versionPart := parts[0]
	build := ""
	if len(parts) == 2 {
		build = parts[1]
	}

	// Split pre-release
	parts = strings.SplitN(versionPart, "-", 2)
	releasePart := parts[0]
	pre := []string{}
	if len(parts) == 2 {
		pre = strings.Split(parts[1], ".")
	}

	// Parse release version (major.minor.patch)
	releaseParts := strings.Split(releasePart, ".")
	if len(releaseParts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s (expected MAJOR.MINOR.PATCH)", v)
	}

	major, err := strconv.Atoi(releaseParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", releaseParts[0])
	}

	minor, err := strconv.Atoi(releaseParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", releaseParts[1])
	}

	patch, err := strconv.Atoi(releaseParts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", releaseParts[2])
	}

	// Validate non-negative
	if major < 0 || minor < 0 || patch < 0 {
		return nil, fmt.Errorf("version numbers must be non-negative")
	}

	// Reject leading zeros (e.g., "01.02.03")
	if releaseParts[0] != strconv.Itoa(major) ||
		releaseParts[1] != strconv.Itoa(minor) ||
		releaseParts[2] != strconv.Itoa(patch) {
		return nil, fmt.Errorf("version numbers must not have leading zeros: %s", v)
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   pre,
		Build: build,
	}, nil
}

// MustParseVersion parses a version string and panics on error
func MustParseVersion(v string) *Version {
	version, err := ParseVersion(v)
	if err != nil {
		panic(err)
	}
	return version
}

// Compare compares two versions
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
func (v *Version) Compare(other *Version) int {
	// Compare major
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	// Compare minor
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	// Compare patch
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Compare pre-release
	// A version with pre-release is less than without
	if len(v.Pre) == 0 && len(other.Pre) > 0 {
		return 1
	}
	if len(v.Pre) > 0 && len(other.Pre) == 0 {
		return -1
	}
	if len(v.Pre) == 0 && len(other.Pre) == 0 {
		return 0
	}

	// Compare pre-release identifiers
	maxLen := len(v.Pre)
	if len(other.Pre) > maxLen {
		maxLen = len(other.Pre)
	}

	for i := 0; i < maxLen; i++ {
		// Missing identifier is less than present
		if i >= len(v.Pre) {
			return -1
		}
		if i >= len(other.Pre) {
			return 1
		}

		vId := v.Pre[i]
		otherId := other.Pre[i]

		// Numeric identifiers have lower precedence than non-numeric
		vNum, vErr := strconv.Atoi(vId)
		otherNum, otherErr := strconv.Atoi(otherId)

		if vErr == nil && otherErr == nil {
			// Both numeric
			if vNum != otherNum {
				if vNum < otherNum {
					return -1
				}
				return 1
			}
		} else if vErr == nil {
			// vId is numeric, otherId is not - numeric has lower precedence
			return -1
		} else if otherErr == nil {
			// otherId is numeric, vId is not
			return 1
		} else {
			// Both non-numeric - compare lexicographically
			if vId != otherId {
				if vId < otherId {
					return -1
				}
				return 1
			}
		}
	}

	return 0
}

// String returns the string representation of the version
func (v *Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Pre) > 0 {
		result += "-" + strings.Join(v.Pre, ".")
	}
	if v.Build != "" {
		result += "+" + v.Build
	}
	return result
}

// IsPreRelease returns true if this is a pre-release version
func (v *Version) IsPreRelease() bool {
	return len(v.Pre) > 0
}

// Matches checks if version satisfies a constraint string
func (v *Version) Matches(constraint string) bool {
	r, err := ParseRange(constraint)
	if err != nil {
		return false
	}
	return r.Match(v)
}

// LessThan checks if v < other
func (v *Version) LessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// LessThanOrEqual checks if v <= other
func (v *Version) LessThanOrEqual(other *Version) bool {
	return v.Compare(other) <= 0
}

// GreaterThan checks if v > other
func (v *Version) GreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// GreaterThanOrEqual checks if v >= other
func (v *Version) GreaterThanOrEqual(other *Version) bool {
	return v.Compare(other) >= 0
}

// Equal checks if v == other
func (v *Version) Equal(other *Version) bool {
	return v.Compare(other) == 0
}

// IncrementMajor increments the major version and resets minor and patch
func (v *Version) IncrementMajor() *Version {
	return &Version{
		Major: v.Major + 1,
		Minor: 0,
		Patch: 0,
		Pre:   []string{},
	}
}

// IncrementMinor increments the minor version and resets patch
func (v *Version) IncrementMinor() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor + 1,
		Patch: 0,
		Pre:   []string{},
	}
}

// IncrementPatch increments the patch version
func (v *Version) IncrementPatch() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
		Pre:   []string{},
	}
}
