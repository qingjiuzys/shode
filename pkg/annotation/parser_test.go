package annotation

import (
	"testing"
)

func TestParser_ParseAnnotation_Simple(t *testing.T) {
	parser := NewParser()

	annotation, err := parser.ParseAnnotation("@Service")
	if err != nil {
		t.Fatalf("Failed to parse annotation: %v", err)
	}

	if annotation.Name != "Service" {
		t.Errorf("Expected name 'Service', got '%s'", annotation.Name)
	}

	if len(annotation.Arguments) != 0 {
		t.Errorf("Expected no arguments, got %d", len(annotation.Arguments))
	}
}

func TestParser_ParseAnnotation_WithValue(t *testing.T) {
	parser := NewParser()

	annotation, err := parser.ParseAnnotation(`@Controller("/api/users")`)
	if err != nil {
		t.Fatalf("Failed to parse annotation: %v", err)
	}

	if annotation.Name != "Controller" {
		t.Errorf("Expected name 'Controller', got '%s'", annotation.Name)
	}

	value := annotation.GetDefaultValue()
	if value != "/api/users" {
		t.Errorf("Expected value '/api/users', got '%s'", value)
	}
}

func TestParser_ParseAnnotation_WithKeyValue(t *testing.T) {
	parser := NewParser()

	annotation, err := parser.ParseAnnotation(`@RequestMapping(path="/api", method="GET")`)
	if err != nil {
		t.Fatalf("Failed to parse annotation: %v", err)
	}

	if annotation.Name != "RequestMapping" {
		t.Errorf("Expected name 'RequestMapping', got '%s'", annotation.Name)
	}

	path := annotation.GetValue("path")
	if path != "/api" {
		t.Errorf("Expected path '/api', got '%s'", path)
	}

	method := annotation.GetValue("method")
	if method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", method)
	}
}

func TestParser_ParseAnnotations_Multiple(t *testing.T) {
	parser := NewParser()

	code := `@Service
@Repository
function UserService() {
    # ...
}`

	annotations, err := parser.ParseAnnotations(code)
	if err != nil {
		t.Fatalf("Failed to parse annotations: %v", err)
	}

	if len(annotations) != 2 {
		t.Errorf("Expected 2 annotations, got %d", len(annotations))
	}

	if annotations[0].Name != "Service" {
		t.Errorf("Expected first annotation 'Service', got '%s'", annotations[0].Name)
	}

	if annotations[1].Name != "Repository" {
		t.Errorf("Expected second annotation 'Repository', got '%s'", annotations[1].Name)
	}
}
