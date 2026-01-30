package semver

import (
	"testing"
)

func TestParseRange_Valid(t *testing.T) {
	tests := []struct {
		expr       string
		shouldFail bool
	}{
		// Valid ranges
		{"^1.2.3", false},
		{"^0.2.3", false},
		{"^0.0.3", false},
		{"~1.2.3", false},
		{"~1.2", false},
		{"~1", false},
		{">=1.0.0", false},
		{">1.0.0", false},
		{"<2.0.0", false},
		{"<=2.0.0", false},
		{"1.2.x", false},
		{"1.x.x", false},
		{"1.x", false},
		{"*", false},
		{"x", false},
		{"X", false},
		{"1.2.3 - 2.3.4", false},
		{"1.0.0", false},
		{"=1.2.3", false},

		// Invalid ranges
		{"invalid", true},
		{">", true},
		{"1.2.3 -", true},
		{"- 1.2.3", true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			_, err := ParseRange(tt.expr)
			if tt.shouldFail && err == nil {
				t.Errorf("ParseRange(%q) should fail but succeeded", tt.expr)
			} else if !tt.shouldFail && err != nil {
				t.Errorf("ParseRange(%q) failed: %v", tt.expr, err)
			}
		})
	}
}

func TestRangeMatching_Caret(t *testing.T) {
	tests := []struct {
		version   string
		rangeExpr string
		matches   bool
	}{
		{"1.2.3", "^1.2.0", true},
		{"1.3.0", "^1.2.0", true},
		{"1.2.4", "^1.2.0", true},
		{"2.0.0", "^1.2.0", false},
		{"0.3.0", "^0.2.0", false}, // Breaking changes in 0.x
		{"0.2.5", "^0.2.0", true},
		{"0.3.0", "^0.3.0", true},
		{"0.0.5", "^0.0.3", false}, // ^0.0.3 => >=0.0.3 <0.0.4 (only patch updates)
		{"0.0.4", "^0.0.3", false}, // 0.0.4 is NOT < 0.0.4
		{"0.0.2", "^0.0.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.version+" in "+tt.rangeExpr, func(t *testing.T) {
			v := MustParseVersion(tt.version)
			r, err := ParseRange(tt.rangeExpr)
			if err != nil {
				t.Fatalf("ParseRange(%q) failed: %v", tt.rangeExpr, err)
			}

			matches := r.Match(v)
			if matches != tt.matches {
				t.Errorf("Match(%s, %s) = %v, want %v", tt.version, tt.rangeExpr, matches, tt.matches)
			}
		})
	}
}

func TestRangeMatching_Tilde(t *testing.T) {
	tests := []struct {
		version   string
		rangeExpr string
		matches   bool
	}{
		{"1.2.3", "~1.2.0", true},
		{"1.2.4", "~1.2.0", true},
		{"1.3.0", "~1.2.0", false},
		{"1.1.9", "~1.2.0", false},
		{"1.2.5", "~1.2", true},
		{"1.3.0", "~1.2", false},
		{"1.4.0", "~1", true}, // ~1 => >=1.0.0 <2.0.0
		{"2.0.0", "~1", false},
		{"1.5.0", "~1", true},
	}

	for _, tt := range tests {
		t.Run(tt.version+" in "+tt.rangeExpr, func(t *testing.T) {
			v := MustParseVersion(tt.version)
			r, err := ParseRange(tt.rangeExpr)
			if err != nil {
				t.Fatalf("ParseRange(%q) failed: %v", tt.rangeExpr, err)
			}

			matches := r.Match(v)
			if matches != tt.matches {
				t.Errorf("Match(%s, %s) = %v, want %v", tt.version, tt.rangeExpr, matches, tt.matches)
			}
		})
	}
}

func TestRangeMatching_Comparison(t *testing.T) {
	tests := []struct {
		version   string
		rangeExpr string
		matches   bool
	}{
		{"1.2.3", ">=1.2.0", true},
		{"1.2.3", ">1.2.0", true},
		{"1.2.3", ">=1.2.3", true},
		{"1.2.3", ">1.2.3", false},
		{"1.2.3", "<=1.2.3", true},
		{"1.2.3", "<1.2.3", false},
		{"1.2.3", "<=1.2.4", true},
		{"1.2.3", "<1.2.4", true},
		{"1.2.3", "=1.2.3", true},
		{"1.2.4", "=1.2.3", false},
		{"1.2.3-alpha", ">=1.2.3", false}, // Pre-release is less than stable
		{"1.2.3", ">=1.2.3-alpha", true},    // Stable is greater than pre-release
	}

	for _, tt := range tests {
		t.Run(tt.version+" in "+tt.rangeExpr, func(t *testing.T) {
			v := MustParseVersion(tt.version)
			r, err := ParseRange(tt.rangeExpr)
			if err != nil {
				t.Fatalf("ParseRange(%q) failed: %v", tt.rangeExpr, err)
			}

			matches := r.Match(v)
			if matches != tt.matches {
				t.Errorf("Match(%s, %s) = %v, want %v", tt.version, tt.rangeExpr, matches, tt.matches)
			}
		})
	}
}

func TestRangeMatching_Wildcard(t *testing.T) {
	tests := []struct {
		version   string
		rangeExpr string
		matches   bool
	}{
		{"1.2.3", "1.2.x", true},
		{"1.2.4", "1.2.x", true},
		{"1.3.0", "1.2.x", false},
		{"1.2.3", "1.x.x", true},
		{"2.0.0", "1.x.x", false},
		{"1.5.5", "1.x.x", true},
		{"1.2.3", "*", true},
		{"2.0.0", "*", true},
		{"0.0.1", "*", true},
		{"1.2.3", "x", true},
		{"1.2.3", "X", true},
	}

	for _, tt := range tests {
		t.Run(tt.version+" in "+tt.rangeExpr, func(t *testing.T) {
			v := MustParseVersion(tt.version)
			r, err := ParseRange(tt.rangeExpr)
			if err != nil {
				t.Fatalf("ParseRange(%q) failed: %v", tt.rangeExpr, err)
			}

			matches := r.Match(v)
			if matches != tt.matches {
				t.Errorf("Match(%s, %s) = %v, want %v", tt.version, tt.rangeExpr, matches, tt.matches)
			}
		})
	}
}

func TestRangeMatching_Hyphen(t *testing.T) {
	tests := []struct {
		version   string
		rangeExpr string
		matches   bool
	}{
		{"1.2.3", "1.2.0 - 1.3.0", true},
		{"1.3.0", "1.2.0 - 1.3.0", true},
		{"1.3.1", "1.2.0 - 1.3.0", false},
		{"1.1.9", "1.2.0 - 1.3.0", false},
		{"1.0.0", "1.0.0 - 2.0.0", true},
		{"2.0.0", "1.0.0 - 2.0.0", true},
		{"2.0.1", "1.0.0 - 2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version+" in "+tt.rangeExpr, func(t *testing.T) {
			v := MustParseVersion(tt.version)
			r, err := ParseRange(tt.rangeExpr)
			if err != nil {
				t.Fatalf("ParseRange(%q) failed: %v", tt.rangeExpr, err)
			}

			matches := r.Match(v)
			if matches != tt.matches {
				t.Errorf("Match(%s, %s) = %v, want %v", tt.version, tt.rangeExpr, matches, tt.matches)
			}
		})
	}
}

func TestRangeMaxSatisfying(t *testing.T) {
	versions := []*Version{
		MustParseVersion("1.0.0"),
		MustParseVersion("1.2.0"),
		MustParseVersion("1.2.3"),
		MustParseVersion("1.3.0"),
		MustParseVersion("2.0.0"),
	}

	tests := []struct {
		rangeExpr   string
		expectVer   string
		expectFound bool
	}{
		{"^1.2.0", "1.3.0", true},
		{"~1.2.0", "1.2.3", true},
		{">=1.2.0", "2.0.0", true},
		{"1.2.x", "1.2.3", true},
		{"^2.0.0", "2.0.0", true}, // ^2.0.0 => >=2.0.0 <3.0.0
		{"^3.0.0", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.rangeExpr, func(t *testing.T) {
			r, err := ParseRange(tt.rangeExpr)
			if err != nil {
				t.Fatalf("ParseRange(%q) failed: %v", tt.rangeExpr, err)
			}

			max := r.MaxSatisfying(versions)
			found := max != nil

			if found != tt.expectFound {
				t.Errorf("MaxSatisfying() found = %v, want %v", found, tt.expectFound)
			}

			if found && max.String() != tt.expectVer {
				t.Errorf("MaxSatisfying() = %s, want %s", max.String(), tt.expectVer)
			}
		})
	}
}

func TestRangeMinSatisfying(t *testing.T) {
	versions := []*Version{
		MustParseVersion("1.0.0"),
		MustParseVersion("1.2.0"),
		MustParseVersion("1.2.3"),
		MustParseVersion("1.3.0"),
		MustParseVersion("2.0.0"),
	}

	tests := []struct {
		rangeExpr   string
		expectVer   string
		expectFound bool
	}{
		{"^1.2.0", "1.2.0", true},
		{"~1.2.0", "1.2.0", true},
		{">=1.2.0", "1.2.0", true},
		{">=1.2.3", "1.2.3", true},
		{">=2.0.0", "2.0.0", true},
		{">=3.0.0", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.rangeExpr, func(t *testing.T) {
			r, err := ParseRange(tt.rangeExpr)
			if err != nil {
				t.Fatalf("ParseRange(%q) failed: %v", tt.rangeExpr, err)
			}

			min := r.MinSatisfying(versions)
			found := min != nil

			if found != tt.expectFound {
				t.Errorf("MinSatisfying() found = %v, want %v", found, tt.expectFound)
			}

			if found && min.String() != tt.expectVer {
				t.Errorf("MinSatisfying() = %s, want %s", min.String(), tt.expectVer)
			}
		})
	}
}

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		expr       string
		shouldFail bool
	}{
		{">=1.0.0", false},
		{">1.0.0", false},
		{"<2.0.0", false},
		{"<=2.0.0", false},
		{"=1.2.3", false},
		{"1.2.3", false},
		{"*", false},
		{"x", false},
		{">", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			_, err := ParseConstraint(tt.expr)
			if tt.shouldFail && err == nil {
				t.Errorf("ParseConstraint(%q) should fail but succeeded", tt.expr)
			} else if !tt.shouldFail && err != nil {
				t.Errorf("ParseConstraint(%q) failed: %v", tt.expr, err)
			}
		})
	}
}

func BenchmarkRangeParsing(b *testing.B) {
	ranges := []string{"^1.2.3", "~1.2.3", ">=1.0.0", "1.2.x", "*"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseRange(ranges[i%len(ranges)])
	}
}

func BenchmarkRangeMatching(b *testing.B) {
	v := MustParseVersion("1.2.3")
	r, _ := ParseRange("^1.2.0")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Match(v)
	}
}
