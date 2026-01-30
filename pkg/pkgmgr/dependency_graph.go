package pkg

import (
	"fmt"

	"gitee.com/com_818cloud/shode/pkg/semver"
)

// DependencyGraph represents the dependency graph
type DependencyGraph struct {
	nodes map[string]*DependencyNode
	edges map[string][]string // package -> [dependencies]
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	Name        string
	Version     *semver.Version
	Constraint  string
	Dependencies []*DependencyNode
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*DependencyNode),
		edges: make(map[string][]string),
	}
}

// AddNode adds a dependency node to the graph
func (dg *DependencyGraph) AddNode(name string, version *semver.Version, constraint string) {
	if _, exists := dg.nodes[name]; !exists {
		dg.nodes[name] = &DependencyNode{
			Name:       name,
			Version:    version,
			Constraint: constraint,
		}
	}
}

// AddEdge adds a dependency relationship between two nodes
func (dg *DependencyGraph) AddEdge(from, to string) {
	if dg.edges[from] == nil {
		dg.edges[from] = []string{}
	}
	dg.edges[from] = append(dg.edges[from], to)
}

// GetNode retrieves a node by name
func (dg *DependencyGraph) GetNode(name string) (*DependencyNode, bool) {
	node, exists := dg.nodes[name]
	return node, exists
}

// TopologicalSort performs topological sort on the graph using Kahn's algorithm
// Returns the sorted order of packages (dependees come before dependents)
func (dg *DependencyGraph) TopologicalSort() ([]string, error) {
	// Calculate in-degree for each node
	inDegree := make(map[string]int)
	for name := range dg.nodes {
		inDegree[name] = 0
	}

	for _, toList := range dg.edges {
		for _, to := range toList {
			inDegree[to]++
		}
	}

	// Find all nodes with zero in-degree
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var result []string
	for len(queue) > 0 {
		// Remove a node with zero in-degree
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree for all neighbors
		for _, neighbor := range dg.edges[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(result) != len(dg.nodes) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// GetInstallOrder returns the installation order (dependencies before dependents)
// This is the reverse of topological sort
func (dg *DependencyGraph) GetInstallOrder() ([]string, error) {
	order, err := dg.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Reverse the order to get installation order
	installOrder := make([]string, len(order))
	for i, pkg := range order {
		installOrder[len(order)-1-i] = pkg
	}

	return installOrder, nil
}

// DetectCycles detects circular dependencies using DFS
// Returns a list of detected cycles
func (dg *DependencyGraph) DetectCycles() [][]string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycles [][]string

	var path []string

	var dfs func(string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range dg.edges[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			} else if recStack[neighbor] {
				// Found cycle
				cycleStart := -1
				for i, p := range path {
					if p == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
			}
		}

		recStack[node] = false
		path = path[:len(path)-1]
	}

	for node := range dg.nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// ResolveConflicts resolves version conflicts in the dependency graph
// Returns a map of package names to resolved versions
func (dg *DependencyGraph) ResolveConflicts() (map[string]*semver.Version, error) {
	// Collect all version requirements for each package
	versions := make(map[string][]*semver.Version)
	constraints := make(map[string][]string)

	for name, node := range dg.nodes {
		versions[name] = append(versions[name], node.Version)
		constraints[name] = append(constraints[name], node.Constraint)
	}

	resolved := make(map[string]*semver.Version)

	for name, versionList := range versions {
		constraintList := constraints[name]

		// Build the intersection of all constraints
		var commonRange *semver.Range
		for _, constraint := range constraintList {
			r, err := semver.ParseRange(constraint)
			if err != nil {
				return nil, fmt.Errorf("invalid constraint %s for %s: %v", constraint, name, err)
			}

			if commonRange == nil {
				commonRange = r
			} else {
				// Intersect ranges
				commonRange = commonRange.Intersect(r)
			}
		}

		if commonRange == nil {
			return nil, fmt.Errorf("no common range for %s", name)
		}

		// Find the highest version from the list that satisfies the common range
		var bestVersion *semver.Version
		for _, version := range versionList {
			if commonRange.Match(version) {
				if bestVersion == nil || version.Compare(bestVersion) > 0 {
					bestVersion = version
				}
			}
		}

		if bestVersion == nil {
			return nil, fmt.Errorf("no version satisfies constraints for %s: %v", name, constraintList)
		}

		resolved[name] = bestVersion
	}

	return resolved, nil
}

// GetDependencies returns all dependencies of a package
func (dg *DependencyGraph) GetDependencies(name string) []string {
	if deps, exists := dg.edges[name]; exists {
		return deps
	}
	return []string{}
}

// HasCycle checks if the graph has any cycles
func (dg *DependencyGraph) HasCycle() bool {
	cycles := dg.DetectCycles()
	return len(cycles) > 0
}

// String returns a string representation of the graph
func (dg *DependencyGraph) String() string {
	var result string
	for from, toList := range dg.edges {
		result += fmt.Sprintf("%s -> %v\n", from, toList)
	}
	return result
}
