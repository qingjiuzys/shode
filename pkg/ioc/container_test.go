package ioc

import (
	"strings"
	"testing"
)

// Test types
type UserRepository interface {
	Find(id int) string
}

type UserRepositoryImpl struct {
	db string
}

func (r *UserRepositoryImpl) Find(id int) string {
	return "user-" + string(rune(id))
}

type UserService struct {
	repo UserRepository
}

func NewUserRepository(db string) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func TestContainer_RegisterBean(t *testing.T) {
	container := NewContainer()

	// Register singleton bean
	err := container.RegisterBean("userRepo", ScopeSingleton, func() *UserRepositoryImpl {
		return NewUserRepository("test-db")
	})
	if err != nil {
		t.Fatalf("Failed to register bean: %v", err)
	}

	// Try to register duplicate
	err = container.RegisterBean("userRepo", ScopeSingleton, func() *UserRepositoryImpl {
		return NewUserRepository("test-db")
	})
	if err == nil {
		t.Error("Expected error when registering duplicate bean")
	}
}

func TestContainer_GetBean_Singleton(t *testing.T) {
	container := NewContainer()

	// Register singleton
	err := container.RegisterBean("userRepo", ScopeSingleton, func() *UserRepositoryImpl {
		return NewUserRepository("test-db")
	})
	if err != nil {
		t.Fatalf("Failed to register bean: %v", err)
	}

	// Get bean twice - should return same instance
	bean1, err := container.GetBean("userRepo")
	if err != nil {
		t.Fatalf("Failed to get bean: %v", err)
	}

	bean2, err := container.GetBean("userRepo")
	if err != nil {
		t.Fatalf("Failed to get bean: %v", err)
	}

	if bean1 != bean2 {
		t.Error("Singleton beans should return same instance")
	}
}

func TestContainer_GetBean_Prototype(t *testing.T) {
	container := NewContainer()

	// Register prototype
	err := container.RegisterBean("userRepo", ScopePrototype, func() *UserRepositoryImpl {
		return NewUserRepository("test-db")
	})
	if err != nil {
		t.Fatalf("Failed to register bean: %v", err)
	}

	// Get bean twice - should return different instances
	bean1, err := container.GetBean("userRepo")
	if err != nil {
		t.Fatalf("Failed to get bean: %v", err)
	}

	bean2, err := container.GetBean("userRepo")
	if err != nil {
		t.Fatalf("Failed to get bean: %v", err)
	}

	if bean1 == bean2 {
		t.Error("Prototype beans should return different instances")
	}
}

func TestContainer_DependencyInjection(t *testing.T) {
	container := NewContainer()

	// Register repository
	err := container.RegisterBean("userRepo", ScopeSingleton, func() *UserRepositoryImpl {
		return NewUserRepository("test-db")
	})
	if err != nil {
		t.Fatalf("Failed to register repository: %v", err)
	}

	// Register service with dependency
	err = container.RegisterBean("userService", ScopeSingleton, func(repo *UserRepositoryImpl) *UserService {
		return NewUserService(repo)
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Get service - should have repository injected
	service, err := container.GetBean("userService")
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}

	userService, ok := service.(*UserService)
	if !ok {
		t.Fatal("Service is not of correct type")
	}

	if userService.repo == nil {
		t.Error("Repository should be injected into service")
	}
}

func TestContainer_CircularDependency(t *testing.T) {
	container := NewContainer()

	// Register beans that depend on each other (circular dependency)
	err := container.RegisterBean("serviceA", ScopeSingleton, func(serviceB *UserService) *UserRepositoryImpl {
		// This creates a dependency on serviceB
		return NewUserRepository("test-db")
	})
	if err != nil {
		t.Fatalf("Failed to register serviceA: %v", err)
	}

	// Register serviceB that depends on serviceA (circular)
	err = container.RegisterBean("serviceB", ScopeSingleton, func(serviceA *UserRepositoryImpl) *UserService {
		// This creates circular dependency
		return NewUserService(serviceA)
	})
	if err != nil {
		t.Fatalf("Failed to register serviceB: %v", err)
	}

	// Try to get bean - should detect circular dependency
	_, err = container.GetBean("serviceA")
	if err == nil {
		t.Error("Expected error for circular dependency")
	} else if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("Expected circular dependency error, got: %v", err)
	}
}

func TestContainer_ContainsBean(t *testing.T) {
	container := NewContainer()

	if container.ContainsBean("test") {
		t.Error("Bean should not exist")
	}

	err := container.RegisterBean("test", ScopeSingleton, func() string {
		return "test"
	})
	if err != nil {
		t.Fatalf("Failed to register bean: %v", err)
	}

	if !container.ContainsBean("test") {
		t.Error("Bean should exist")
	}
}

func TestContainer_GetBeanNames(t *testing.T) {
	container := NewContainer()

	container.RegisterBean("bean1", ScopeSingleton, func() string { return "1" })
	container.RegisterBean("bean2", ScopeSingleton, func() string { return "2" })
	container.RegisterBean("bean3", ScopeSingleton, func() string { return "3" })

	names := container.GetBeanNames()
	if len(names) != 3 {
		t.Errorf("Expected 3 bean names, got %d", len(names))
	}
}
