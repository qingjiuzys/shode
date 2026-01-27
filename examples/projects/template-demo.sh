#!/usr/bin/env shode
# Template Engine Demo with HTTP Server

StartHTTPServer "8090"

# API: Render template from string
function renderTemplate() {
    SetHTTPResponseTemplateString 200 "Hello {{name}}! Welcome to {{app}}." "name=User" "app=Shode"
}
RegisterHTTPRoute "GET" "/api/render" "function" "renderTemplate"

# API: Render complex template
function renderAdvanced() {
    SetHTTPResponseTemplateString 200 "{{#if show}}<h1>Welcome {{name}}</h1>{{/if}}" "show=true" "name=Alice"
}
RegisterHTTPRoute "GET" "/api/advanced" "function" "renderAdvanced"

# API: Test conditional rendering (false)
function renderHidden() {
    SetHTTPResponseTemplateString 200 "{{#if show}}This should be hidden{{/if}}Otherwise: This is shown" "show=false"
}
RegisterHTTPRoute "GET" "/api/hidden" "function" "renderHidden"

# Serve static files
RegisterStaticRoute "/" "./examples/test_static"

Println "Template Engine Demo Server"
Println "http://localhost:8090"
Println ""
Println "Endpoints:"
Println "  /api/render    - Simple template"
Println "  /api/advanced  - Complex template with conditionals"
Println "  /api/hidden   - Conditional rendering (false)"

for i in $(seq 1 100000); do sleep 1; done
