package semver

import (
	"fmt"
	"strings"
)

// operator represents a comparison operator
type operator int

const (
	opEQ operator = iota // =
	opLT                 // <
	opLE                 // <=
	opGT                 // >
	opGE                 // >=
)

// constraint represents a single version constraint
type constraint struct {
	op      operator
	version *Version
}

// Match checks if version satisfies the constraint
func (c *constraint) Match(v *Version) bool {
	result := v.Compare(c.version)

	switch c.op {
	case opEQ:
		return result == 0
	case opLT:
		return result < 0
	case opLE:
		return result <= 0
	case opGT:
		return result > 0
	case opGE:
		return result >= 0
	default:
		return false
	}
}

// String returns the string representation of the constraint
func (c *constraint) String() string {
	var opStr string
	switch c.op {
	case opEQ:
		opStr = ""
	case opLT:
		opStr = "<"
	case opLE:
		opStr = "<="
	case opGT:
		opStr = ">"
	case opGE:
		opStr = ">="
	}

	if opStr == "" {
		return c.version.String()
	}
	return opStr + " " + c.version.String()
}

// ParseConstraint parses a single constraint string
func ParseConstraint(s string) (*constraint, error) {
	s = strings.TrimSpace(s)

	// Handle empty or wildcard
	if s == "" || s == "*" || s == "x" || s == "X" {
		return &constraint{
			op:      opGE,
			version: &Version{Major: 0, Minor: 0, Patch: 0},
		}, nil
	}

	operators := []struct {
		op      operator
		prefix  string
	}{
		{opGE, ">="},
		{opGT, ">"},
		{opLE, "<="},
		{opLT, "<"},
		{opEQ, "="},
	}

	for _, op := range operators {
		if strings.HasPrefix(s, op.prefix) {
			versionStr := strings.TrimPrefix(s, op.prefix)
			version, err := ParseVersion(versionStr)
			if err != nil {
				return nil, fmt.Errorf("invalid version in constraint %q: %v", s, err)
			}
			return &constraint{
				op:      op.op,
				version: version,
			}, nil
		}
	}

	// No operator means exact match
	version, err := ParseVersion(s)
	if err != nil {
		return nil, fmt.Errorf("invalid constraint: %s", s)
	}

	return &constraint{
		op:      opEQ,
		version: version,
	}, nil
}

// MustParseConstraint parses a constraint and panics on error
func MustParseConstraint(s string) *constraint {
	c, err := ParseConstraint(s)
	if err != nil {
		panic(err)
	}
	return c
}
