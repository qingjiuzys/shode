package semver

import (
	"fmt"
	"strings"
)

// Range represents a version range constraint
type Range struct {
	constraints []*constraint
}

// ParseRange parses a version range string
// Supports: ^1.2.3, ~1.2.3, >=1.0.0, 1.x.*, *, 1.2.3 - 2.3.4
func ParseRange(r string) (*Range, error) {
	r = strings.TrimSpace(r)

	if r == "" || r == "*" || r == "x" || r == "X" {
		return &Range{constraints: []*constraint{{op: opGE, version: &Version{Major: 0, Minor: 0, Patch: 0}}}}, nil
	}

	// Handle hyphen range: 1.2.3 - 2.3.4
	if strings.Contains(r, " - ") {
		return parseHyphenRange(r)
	}

	// Handle caret range: ^1.2.3
	if strings.HasPrefix(r, "^") {
		return parseCaretRange(r)
	}

	// Handle tilde range: ~1.2.3
	if strings.HasPrefix(r, "~") {
		return parseTildeRange(r)
	}

	// Handle wildcard: 1.x.x, 1.2.x
	if strings.Contains(r, "x") || strings.Contains(r, "X") {
		return parseWildcardRange(r)
	}

	// Handle comparison operators: >=, >, <=, <, =
	return parseComparisonRange(r)
}

// Match checks if version satisfies the range
func (r *Range) Match(v *Version) bool {
	for _, c := range r.constraints {
		if !c.Match(v) {
			return false
		}
	}
	return true
}

// MaxSatisfying returns the maximum version satisfying the range
func (r *Range) MaxSatisfying(versions []*Version) *Version {
	var max *Version
	for _, v := range versions {
		if r.Match(v) {
			if max == nil || v.Compare(max) > 0 {
				max = v
			}
		}
	}
	return max
}

// MinSatisfying returns the minimum version satisfying the range
func (r *Range) MinSatisfying(versions []*Version) *Version {
	var min *Version
	for _, v := range versions {
		if r.Match(v) {
			if min == nil || v.Compare(min) < 0 {
				min = v
			}
		}
	}
	return min
}

// Intersect returns the intersection of two ranges
func (r *Range) Intersect(other *Range) *Range {
	return &Range{
		constraints: append(r.constraints, other.constraints...),
	}
}

// String returns the string representation of the range
func (r *Range) String() string {
	var parts []string
	for _, c := range r.constraints {
		parts = append(parts, c.String())
	}
	return strings.Join(parts, " ")
}

// parseHyphenRange parses hyphen range: "1.2.3 - 2.3.4"
func parseHyphenRange(r string) (*Range, error) {
	parts := strings.SplitN(r, " - ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid hyphen range: %s", r)
	}

	minVersion, err := ParseVersion(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid min version: %v", err)
	}

	maxVersion, err := ParseVersion(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid max version: %v", err)
	}

	return &Range{
		constraints: []*constraint{
			{op: opGE, version: minVersion},
			{op: opLE, version: maxVersion},
		},
	}, nil
}

// parseCaretRange parses caret range: "^1.2.3"
// ^1.2.3 => >=1.2.3 <2.0.0
// ^0.2.3 => >=0.2.3 <0.3.0
// ^0.0.3 => >=0.0.3 <0.0.4
func parseCaretRange(r string) (*Range, error) {
	version, err := ParseVersion(strings.TrimPrefix(r, "^"))
	if err != nil {
		return nil, err
	}

	var maxVersion *Version
	if version.Major > 0 {
		// ^1.2.3 => >=1.2.3 <2.0.0
		maxVersion = &Version{Major: version.Major + 1, Minor: 0, Patch: 0, Pre: []string{}}
	} else if version.Minor > 0 {
		// ^0.2.3 => >=0.2.3 <0.3.0
		maxVersion = &Version{Major: 0, Minor: version.Minor + 1, Patch: 0, Pre: []string{}}
	} else {
		// ^0.0.3 => >=0.0.3 <0.0.4
		maxVersion = &Version{Major: 0, Minor: 0, Patch: version.Patch + 1, Pre: []string{}}
	}

	return &Range{
		constraints: []*constraint{
			{op: opGE, version: version},
			{op: opLT, version: maxVersion},
		},
	}, nil
}

// parseTildeRange parses tilde range: "~1.2.3"
// ~1.2.3 => >=1.2.3 <1.3.0
// ~1.2 => >=1.2.0 <1.3.0
// ~1 => >=1.0.0 <2.0.0
func parseTildeRange(r string) (*Range, error) {
	versionStr := strings.TrimPrefix(r, "~")
	versionStr = strings.TrimSpace(versionStr)

	// Handle ~1, ~1.2 formats
	parts := strings.Split(versionStr, ".")
	if len(parts) == 1 {
		// ~1 => >=1.0.0 <2.0.0
		versionStr = versionStr + ".0.0"
	} else if len(parts) == 2 {
		// ~1.2 => >=1.2.0 <1.3.0
		versionStr = versionStr + ".0"
	}

	version, err := ParseVersion(versionStr)
	if err != nil {
		return nil, err
	}

	var maxVersion *Version
	if len(parts) == 1 {
		// ~1 => >=1.0.0 <2.0.0
		maxVersion = &Version{
			Major: version.Major + 1,
			Minor: 0,
			Patch: 0,
		}
	} else {
		// ~1.2 or ~1.2.3 => minor increment
		maxVersion = &Version{
			Major: version.Major,
			Minor: version.Minor + 1,
			Patch: 0,
		}
	}

	return &Range{
		constraints: []*constraint{
			{op: opGE, version: version},
			{op: opLT, version: maxVersion},
		},
	}, nil
}

// parseWildcardRange parses wildcard range: "1.x.x", "1.2.x", "1.x", "x"
func parseWildcardRange(r string) (*Range, error) {
	parts := strings.Split(r, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return nil, fmt.Errorf("invalid wildcard range: %s", r)
	}

	// Normalize to 3 parts
	for len(parts) < 3 {
		parts = append(parts, "x")
	}

	// Convert wildcards to 0
	for i, part := range parts {
		if part == "x" || part == "X" || part == "*" {
			parts[i] = "0"
		}
	}

	minVersion, err := ParseVersion(strings.Join(parts, "."))
	if err != nil {
		return nil, err
	}

	var maxVersion *Version
	if parts[0] == "0" && parts[1] == "0" && parts[2] == "0" {
		// "x.x.x" or "*" => >=0.0.0 (no upper bound)
		return &Range{
			constraints: []*constraint{
				{op: opGE, version: minVersion},
			},
		}, nil
	}

	if parts[1] == "0" && parts[2] == "0" {
		// "1.x.x" => >=1.0.0 <2.0.0
		maxVersion = &Version{Major: minVersion.Major + 1, Minor: 0, Patch: 0}
	} else if parts[2] == "0" {
		// "1.2.x" => >=1.2.0 <1.3.0
		maxVersion = &Version{Major: minVersion.Major, Minor: minVersion.Minor + 1, Patch: 0}
	} else {
		// "1.2.3" => exact match
		return &Range{
			constraints: []*constraint{
				{op: opEQ, version: minVersion},
			},
		}, nil
	}

	return &Range{
		constraints: []*constraint{
			{op: opGE, version: minVersion},
			{op: opLT, version: maxVersion},
		},
	}, nil
}

// parseComparisonRange parses comparison range: ">=1.0.0", ">1.0.0", "<=2.0.0", "<2.0.0", "=1.2.3"
func parseComparisonRange(r string) (*Range, error) {
	operators := []struct {
		op      operator
		prefix  string
	}{
		{opGE, ">="},
		{opGT, ">"},
		{opLE, "<="},
		{opLT, "<"},
		{opEQ, "="},
		{opEQ, ""}, // Empty prefix means exact match
	}

	for _, op := range operators {
		if strings.HasPrefix(r, op.prefix) {
			versionStr := strings.TrimPrefix(r, op.prefix)
			version, err := ParseVersion(versionStr)
			if err != nil {
				return nil, err
			}
			return &Range{
				constraints: []*constraint{
					{op: op.op, version: version},
				},
			}, nil
		}
	}

	// Try parsing as exact version
	version, err := ParseVersion(r)
	if err != nil {
		return nil, fmt.Errorf("invalid range: %s", r)
	}

	return &Range{
		constraints: []*constraint{
			{op: opEQ, version: version},
		},
	}, nil
}
