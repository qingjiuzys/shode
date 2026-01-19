#!/bin/sh

# Test GetQueryResult in function handler

StartHTTPServer 8080

function handleGetPosts() {
    echo "[DEBUG] handleGetPosts: Querying database..."
    QueryDB "SELECT 'post1' as title, 'content1' as body"
    echo "[DEBUG] handleGetPosts: Getting query result..."
    result = GetQueryResult
    echo "[DEBUG] handleGetPosts: Result = $result"
    SetHTTPResponse 200 "$result"
}

RegisterRouteWithResponse "GET" "/api/posts" handleGetPosts
