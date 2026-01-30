package semver

import (
	"testing"
)

func TestParseVersion_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected *Version
	}{
		{
			"1.2.3",
			&Version{Major: 1, Minor: 2, Patch: 3, Pre: []string{}, Build: ""},
		},
		{
			"0.0.0",
			&Version{Major: 0, Minor: 0, Patch: 0, Pre: []string{}, Build: ""},
		},
		{
			"10.20.30",
			&Version{Major: 10, Minor: 20, Patch: 30, Pre: []string{}, Build: ""},
		},
		{
			"1.0.0-alpha",
			&Version{Major: 1, Minor: 0, Patch: 0, Pre: []string{"alpha"}, Build: ""},
		},
		{
			"1.0.0-alpha.1",
			&Version{Major: 1, Minor: 0, Patch: 0, Pre: []string{"alpha", "1"}, Build: ""},
		},
		{
			"1.0.0-0.3.7",
			&Version{Major: 1, Minor: 0, Patch: 0, Pre: []string{"0", "3", "7"}, Build: ""},
		},
		{
			"1.0.0-x.7.z.92",
			&Version{Major: 1, Minor: 0, Patch: 0, Pre: []string{"x", "7", "z", "92"}, Build: ""},
		},
		{
			"1.0.0-alpha+001",
			&Version{Major: 1, Minor: 0, Patch: 0, Pre: []string{"alpha"}, Build: "001"},
		},
		{
			"1.0.0+20130313144700",
			&Version{Major: 1, Minor: 0, Patch: 0, Pre: []string{}, Build: "20130313144700"},
		},
		{
			"1.0.0-beta+exp.sha.5114f85",
			&Version{Major: 1, Minor: 0, Patch: 0, Pre: []string{"beta"}, Build: "exp.sha.5114f85"},
		},
		{
			"1.0.0-alpha.1+build.1",
			&Version{Major: 1, Minor: 0, Patch: 0, Pre: []string{"alpha", "1"}, Build: "build.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseVersion(tt.input)
			if err != nil {
				t.Errorf("ParseVersion(%q) failed: %v", tt.input, err)
				return
			}

			if result.Major != tt.expected.Major ||
				result.Minor != tt.expected.Minor ||
				result.Patch != tt.expected.Patch ||
				len(result.Pre) != len(tt.expected.Pre) ||
				result.Build != tt.expected.Build {
				t.Errorf("ParseVersion(%q) = %+v, want %+v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseVersion_Invalid(t *testing.T) {
	tests := []string{
		"",
		"1",
		"1.2",
		"v1.2.3",
		// Note: "1.2.3-beta" is a valid version (pre-release)
		"invalid",
		"1.2.3.4",
		"-1.2.3",
		"1.-2.3",
		"1.2.-3",
		"01.02.03", // Leading zeros are not allowed
		"1. 2.3",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseVersion(input)
			if err == nil {
				t.Errorf("ParseVersion(%q) should fail but succeeded", input)
			}
		})
	}
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		v1     string
		v2     string
		expect int
	}{
		// Basic comparison
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -1},
		{"1.2.4", "1.2.3", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},

		// Pre-release comparison
		{"1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.beta", -1},
		{"1.0.0-alpha.beta", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-beta.2", -1},
		{"1.0.0-beta.2", "1.0.0-beta.11", -1},
		{"1.0.0-beta.11", "1.0.0-rc.1", -1},
		{"1.0.0-rc.1", "1.0.0", -1},

		// Pre-release vs stable
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0", -1},

		// Build metadata should be ignored
		{"1.2.3+build", "1.2.3", 0},
		{"1.2.3+build.1", "1.2.3+build.2", 0},
		{"1.2.3-alpha+build", "1.2.3-alpha", 0},
	}

	for _, tt := range tests {
		t.Run(tt.v1+" vs "+tt.v2, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			if err != nil {
				t.Fatalf("ParseVersion(%q) failed: %v", tt.v1, err)
			}

			v2, err := ParseVersion(tt.v2)
			if err != nil {
				t.Fatalf("ParseVersion(%q) failed: %v", tt.v2, err)
			}

			result := v1.Compare(v2)
			if result != tt.expect {
				t.Errorf("Compare(%q, %q) = %d, expect %d", tt.v1, tt.v2, result, tt.expect)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		version string
		expect  string
	}{
		{"1.2.3", "1.2.3"},
		{"1.0.0-alpha", "1.0.0-alpha"},
		{"1.0.0-alpha.1", "1.0.0-alpha.1"},
		{"1.0.0+build", "1.0.0+build"},
		{"1.0.0-alpha+build", "1.0.0-alpha+build"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("ParseVersion(%q) failed: %v", tt.version, err)
			}

			result := v.String()
			if result != tt.expect {
				t.Errorf("String() = %q, want %q", result, tt.expect)
			}
		})
	}
}

func TestVersionIncrement(t *testing.T) {
	tests := []struct {
		version     string
		incMajor    string
		incMinor    string
		incPatch    string
	}{
		{"1.2.3", "2.0.0", "1.3.0", "1.2.4"},
		{"0.0.0", "1.0.0", "0.1.0", "0.0.1"},
		{"1.0.0-alpha", "2.0.0", "1.1.0", "1.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("ParseVersion(%q) failed: %v", tt.version, err)
			}

			if v.IncrementMajor().String() != tt.incMajor {
				t.Errorf("IncrementMajor() = %q, want %q", v.IncrementMajor().String(), tt.incMajor)
			}

			if v.IncrementMinor().String() != tt.incMinor {
				t.Errorf("IncrementMinor() = %q, want %q", v.IncrementMinor().String(), tt.incMinor)
			}

			if v.IncrementPatch().String() != tt.incPatch {
				t.Errorf("IncrementPatch() = %q, want %q", v.IncrementPatch().String(), tt.incPatch)
			}
		})
	}
}

func TestVersionComparisonHelpers(t *testing.T) {
	v1 := MustParseVersion("1.2.3")
	v2 := MustParseVersion("1.2.4")
	v3 := MustParseVersion("1.2.3")

	if !v1.LessThan(v2) {
		t.Error("LessThan failed")
	}
	if !v1.LessThanOrEqual(v2) {
		t.Error("LessThanOrEqual failed")
	}
	if !v1.LessThanOrEqual(v3) {
		t.Error("LessThanOrEqual (equal) failed")
	}
	if !v2.GreaterThan(v1) {
		t.Error("GreaterThan failed")
	}
	if !v2.GreaterThanOrEqual(v1) {
		t.Error("GreaterThanOrEqual failed")
	}
	if !v3.Equal(v1) {
		t.Error("Equal failed")
	}
}

func TestVersionIsPreRelease(t *testing.T) {
	tests := []struct {
		version   string
		preRelease bool
	}{
		{"1.0.0", false},
		{"1.0.0-alpha", true},
		{"1.0.0-alpha.1", true},
		{"1.0.0-beta", true},
		{"1.0.0-rc.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v := MustParseVersion(tt.version)
			if result := v.IsPreRelease(); result != tt.preRelease {
				t.Errorf("IsPreRelease() = %v, want %v", result, tt.preRelease)
			}
		})
	}
}

func BenchmarkVersionParsing(b *testing.B) {
	versions := []string{"1.2.3", "1.0.0-alpha", "1.0.0-beta+build"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseVersion(versions[i%len(versions)])
	}
}

func BenchmarkVersionComparison(b *testing.B) {
	v1 := MustParseVersion("1.2.3")
	v2 := MustParseVersion("1.2.4")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v1.Compare(v2)
	}
}
