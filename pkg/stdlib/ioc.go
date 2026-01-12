package stdlib

import (
	"gitee.com/com_818cloud/shode/pkg/ioc"
)

// RegisterBean registers a bean in the IoC container
// Usage: RegisterBean "beanName" "singleton" factoryFunction
func (sl *StdLib) RegisterBean(name, scope string, factory interface{}) error {
	var beanScope ioc.BeanScope
	if scope == "prototype" {
		beanScope = ioc.ScopePrototype
	} else {
		beanScope = ioc.ScopeSingleton
	}

	return sl.iocContainer.RegisterBean(name, beanScope, factory)
}

// GetBean retrieves a bean from the IoC container
// Usage: GetBean "beanName"
func (sl *StdLib) GetBean(name string) (interface{}, error) {
	return sl.iocContainer.GetBean(name)
}

// ContainsBean checks if a bean is registered
// Usage: ContainsBean "beanName"
func (sl *StdLib) ContainsBean(name string) bool {
	return sl.iocContainer.ContainsBean(name)
}

// GetBeanNames returns all registered bean names
// Usage: GetBeanNames
func (sl *StdLib) GetBeanNames() []string {
	return sl.iocContainer.GetBeanNames()
}
