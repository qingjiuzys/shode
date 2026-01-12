#!/usr/bin/env shode

# Category Management Functions

function listCategories() {
    QueryDB "SELECT id, name, description FROM categories ORDER BY name"
    result = GetQueryResult
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

function createCategory() {
    name = GetHTTPQuery "name"
    description = GetHTTPQuery "description"
    
    if name == "" {
        SetHTTPResponse 400 "Category name is required"
        return
    }
    
    ExecDB "INSERT INTO categories (name, description) VALUES (?, ?)" name description
    SetHTTPResponse 201 "Category created"
}

function updateCategory() {
    categoryId = GetHTTPQuery "id"
    name = GetHTTPQuery "name"
    description = GetHTTPQuery "description"
    
    if categoryId == "" {
        SetHTTPResponse 400 "Category ID is required"
        return
    }
    
    if name != "" {
        ExecDB "UPDATE categories SET name = ? WHERE id = ?" name categoryId
    }
    if description != "" {
        ExecDB "UPDATE categories SET description = ? WHERE id = ?" description categoryId
    }
    
    SetHTTPResponse 200 "Category updated"
}

function deleteCategory() {
    categoryId = GetHTTPQuery "id"
    
    if categoryId == "" {
        SetHTTPResponse 400 "Category ID is required"
        return
    }
    
    # Check if category has books
    QueryRowDB "SELECT COUNT(*) FROM books WHERE category_id = ?" categoryId
    bookCount = GetQueryResult
    
    if Contains bookCount "0" {
        ExecDB "DELETE FROM categories WHERE id = ?" categoryId
        SetHTTPResponse 200 "Category deleted"
    } else {
        SetHTTPResponse 400 "Cannot delete category with books"
    }
}
