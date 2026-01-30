package pkg

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/semver"
)

func TestDependencyGraph_AddNode(t *testing.T) {
	dg := NewDependencyGraph()

	v := semver.MustParseVersion("1.2.3")
	dg.AddNode("pkg-a", v, "^1.0.0")

	node, exists := dg.GetNode("pkg-a")
	if !exists {
		t.Fatal("Node not found")
	}

	if node.Name != "pkg-a" {
		t.Errorf("Expected name 'pkg-a', got '%s'", node.Name)
	}

	if node.Version.String() != "1.2.3" {
		t.Errorf("Expected version '1.2.3', got '%s'", node.Version.String())
	}
}

func TestDependencyGraph_AddEdge(t *testing.T) {
	dg := NewDependencyGraph()

	dg.AddEdge("pkg-a", "pkg-b")
	dg.AddEdge("pkg-a", "pkg-c")

	deps := dg.GetDependencies("pkg-a")
	if len(deps) != 2 {
		t.Fatalf("Expected 2 dependencies, got %d", len(deps))
	}

	// Sort for consistent comparison
	sort.Strings(deps)
	expected := []string{"pkg-b", "pkg-c"}
	if !reflect.DeepEqual(deps, expected) {
		t.Errorf("Expected dependencies %v, got %v", expected, deps)
	}
}

func TestTopologicalSort(t *testing.T) {
	dg := NewDependencyGraph()

	// Add nodes
	dg.AddNode("a", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("b", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("c", semver.MustParseVersion("1.0.0"), "")

	// Add edges: a -> b -> c (a depends on b, b depends on c)
	dg.AddEdge("a", "b")
	dg.AddEdge("b", "c")

	// For installation, we need the reverse order
	installOrder, err := dg.GetInstallOrder()
	if err != nil {
		t.Fatalf("GetInstallOrder failed: %v", err)
	}

	// Install order should be [c b a] (dependencies first)
	cIdx := indexOf(installOrder, "c")
	bIdx := indexOf(installOrder, "b")
	aIdx := indexOf(installOrder, "a")

	if cIdx != 0 || bIdx != 1 || aIdx != 2 {
		t.Errorf("Wrong install order: %v (c=%d, b=%d, a=%d). Expected [c b a]", installOrder, cIdx, bIdx, aIdx)
	}
}

func TestTopologicalSort_Complex(t *testing.T) {
	dg := NewDependencyGraph()

	// Add nodes
	dg.AddNode("a", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("b", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("c", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("d", semver.MustParseVersion("1.0.0"), "")

	// Add edges:
	// a -> b, c
	// b -> d
	// c -> d
	dg.AddEdge("a", "b")
	dg.AddEdge("a", "c")
	dg.AddEdge("b", "d")
	dg.AddEdge("c", "d")

	result, err := dg.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	// d should be last (has no dependents)
	dIdx := indexOf(result, "d")
	if dIdx != len(result)-1 {
		t.Errorf("Expected d to be last, got position %d", dIdx)
	}

	// a should be first (no dependencies)
	aIdx := indexOf(result, "a")
	if aIdx != 0 {
		t.Errorf("Expected a to be first, got position %d", aIdx)
	}
}

func TestTopologicalSort_Cycle(t *testing.T) {
	dg := NewDependencyGraph()

	// Add nodes
	dg.AddNode("a", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("b", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("c", semver.MustParseVersion("1.0.0"), "")

	// Create cycle: a -> b -> c -> a
	dg.AddEdge("a", "b")
	dg.AddEdge("b", "c")
	dg.AddEdge("c", "a")

	_, err := dg.TopologicalSort()
	if err == nil {
		t.Error("Expected error for cyclic graph")
	}

	if err.Error() != "circular dependency detected" {
		t.Errorf("Expected 'circular dependency detected' error, got: %v", err)
	}
}

func TestDetectCycles_NoCycle(t *testing.T) {
	dg := NewDependencyGraph()

	dg.AddNode("a", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("b", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("c", semver.MustParseVersion("1.0.0"), "")

	dg.AddEdge("a", "b")
	dg.AddEdge("b", "c")

	cycles := dg.DetectCycles()
	if len(cycles) != 0 {
		t.Errorf("Expected no cycles, got %d", len(cycles))
	}
}

func TestDetectCycles_SimpleCycle(t *testing.T) {
	dg := NewDependencyGraph()

	dg.AddNode("a", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("b", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("c", semver.MustParseVersion("1.0.0"), "")

	// Create cycle: a -> b -> c -> a
	dg.AddEdge("a", "b")
	dg.AddEdge("b", "c")
	dg.AddEdge("c", "a")

	cycles := dg.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("Expected to detect cycle")
	}

	// Verify cycle contains all three nodes
	if len(cycles[0]) != 3 {
		t.Errorf("Expected cycle of length 3, got %d", len(cycles[0]))
	}
}

func TestDetectCycles_SelfLoop(t *testing.T) {
	dg := NewDependencyGraph()

	dg.AddNode("a", semver.MustParseVersion("1.0.0"), "")

	// Self loop: a -> a
	dg.AddEdge("a", "a")

	cycles := dg.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("Expected to detect self-loop")
	}
}

func TestHasCycle(t *testing.T) {
	dg := NewDependencyGraph()

	dg.AddNode("a", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("b", semver.MustParseVersion("1.0.0"), "")

	// No cycle
	dg.AddEdge("a", "b")
	if dg.HasCycle() {
		t.Error("Expected no cycle")
	}

	// Add cycle
	dg.AddEdge("b", "a")
	if !dg.HasCycle() {
		t.Error("Expected to detect cycle")
	}
}

func TestResolveConflicts_NoConflict(t *testing.T) {
	dg := NewDependencyGraph()

	dg.AddNode("pkg-a", semver.MustParseVersion("1.2.0"), "^1.0.0")
	dg.AddNode("pkg-b", semver.MustParseVersion("2.0.0"), "^2.0.0")

	resolved, err := dg.ResolveConflicts()
	if err != nil {
		t.Fatalf("ResolveConflicts failed: %v", err)
	}

	if resolved["pkg-a"].String() != "1.2.0" {
		t.Errorf("Expected pkg-a version 1.2.0, got %s", resolved["pkg-a"].String())
	}

	if resolved["pkg-b"].String() != "2.0.0" {
		t.Errorf("Expected pkg-b version 2.0.0, got %s", resolved["pkg-b"].String())
	}
}

func TestResolveConflicts_CompatibleConstraints(t *testing.T) {
	dg := NewDependencyGraph()

	// Same package with a version that satisfies both constraints
	// pkg-a needs ^1.0.0, pkg-b needs ^1.2.0
	// Available version is 1.3.0 which satisfies both
	dg.AddNode("pkg", semver.MustParseVersion("1.3.0"), "^1.0.0")

	// Simulate adding another constraint by updating the node
	dg.nodes["pkg"].Constraint = "^1.2.0"

	resolved, err := dg.ResolveConflicts()
	if err != nil {
		t.Fatalf("ResolveConflicts failed: %v", err)
	}

	// Should resolve to 1.3.0 (satisfies both ^1.0.0 and ^1.2.0)
	if resolved["pkg"].String() != "1.3.0" {
		t.Errorf("Expected pkg version 1.3.0, got %s", resolved["pkg"].String())
	}
}

func TestResolveConflicts_IncompatibleConstraints(t *testing.T) {
	dg := NewDependencyGraph()

	// pkg with version 1.0.0, but constraints ^1.0.0 and ^2.0.0
	// which have no intersection
	dg.AddNode("pkg", semver.MustParseVersion("1.0.0"), "^1.0.0")

	// Manually set multiple constraints by modifying the graph structure
	// In real usage, this would come from different packages requiring different versions
	dg.nodes["pkg"].Constraint = "^2.0.0"

	_, err := dg.ResolveConflicts()
	if err == nil {
		t.Error("Expected error for incompatible constraints")
	}
}

func TestResolveConflicts_WildcardConstraint(t *testing.T) {
	dg := NewDependencyGraph()

	// pkg with version 1.3.0 and wildcard constraint
	dg.AddNode("pkg", semver.MustParseVersion("1.3.0"), "1.x")

	resolved, err := dg.ResolveConflicts()
	if err != nil {
		t.Fatalf("ResolveConflicts failed: %v", err)
	}

	// Should resolve to 1.3.0
	if resolved["pkg"].String() != "1.3.0" {
		t.Errorf("Expected pkg version 1.3.0, got %s", resolved["pkg"].String())
	}
}

func TestDependencyGraph_String(t *testing.T) {
	dg := NewDependencyGraph()

	dg.AddNode("a", semver.MustParseVersion("1.0.0"), "")
	dg.AddNode("b", semver.MustParseVersion("1.0.0"), "")

	dg.AddEdge("a", "b")

	str := dg.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}

func BenchmarkTopologicalSort(b *testing.B) {
	dg := NewDependencyGraph()

	// Create a large DAG
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("pkg-%d", i)
		dg.AddNode(name, semver.MustParseVersion("1.0.0"), "")

		// Add dependency to previous package
		if i > 0 {
			prev := fmt.Sprintf("pkg-%d", i-1)
			dg.AddEdge(name, prev)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dg.TopologicalSort()
	}
}

func BenchmarkDetectCycles(b *testing.B) {
	dg := NewDependencyGraph()

	// Create a large graph without cycles
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("pkg-%d", i)
		dg.AddNode(name, semver.MustParseVersion("1.0.0"), "")

		if i > 0 {
			prev := fmt.Sprintf("pkg-%d", i-1)
			dg.AddEdge(prev, name)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dg.DetectCycles()
	}
}

// Helper functions

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
