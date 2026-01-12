#!/usr/bin/env shode

# HTTP Route Handlers

function handleLogin() {
    username = GetHTTPQuery "username"
    password = GetHTTPQuery "password"
    
    if username == "" || password == "" {
        SetHTTPResponse 400 "Username and password are required"
        return
    }
    
    login username password
    token = GetEnv "login_token"
    success = GetEnv "login_success"
    
    if success == "false" || token == "" {
        SetHTTPResponse 401 "Invalid username or password"
        return
    }
    
    SetHTTPHeader "Content-Type" "application/json"
    response = "{\"token\":\"" + token + "\",\"username\":\"" + username + "\"}"
    SetHTTPResponse 200 response
}

function handleListCategories() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    listCategories
}

function handleCreateCategory() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    createCategory
}

function handleUpdateCategory() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    updateCategory
}

function handleDeleteCategory() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    deleteCategory
}

function handleListBooks() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    listBooks
}

function handleGetBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    getBook
}

function handleCreateBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    createBook
}

function handleUpdateBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    updateBook
}

function handleDeleteBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    deleteBook
}
