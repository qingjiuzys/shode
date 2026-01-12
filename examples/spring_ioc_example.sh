#!/usr/bin/env shode

# Spring-like IoC Container Example
# Demonstrates dependency injection and bean management

Println "=== IoC Container Example ==="

# Note: Full IoC implementation requires function references
# This example shows the concept

Println "IoC Container allows:"
Println "1. Bean registration with singleton/prototype scope"
Println "2. Automatic dependency injection"
Println "3. Lifecycle management"

Println ""
Println "Example usage (conceptual):"
Println "# Register a repository bean"
Println "RegisterBean 'userRepository' 'singleton' createUserRepository"
Println ""
Println "# Register a service bean with dependency"
Println "RegisterBean 'userService' 'singleton' createUserService"
Println ""
Println "# Get bean from container"
Println "userService = GetBean 'userService'"
Println ""
Println "# Check if bean exists"
Println "exists = ContainsBean 'userService'"

Println ""
Println "=== IoC Example Complete ==="
