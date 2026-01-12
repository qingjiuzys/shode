#!/usr/bin/env shode

# Book Management Functions

function listBooks() {
    categoryId = GetHTTPQuery "category_id"
    
    if categoryId != "" {
        QueryDB "SELECT b.id, b.title, b.author, b.isbn, c.name as category, b.price, b.stock FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE b.category_id = ? ORDER BY b.title" categoryId
    } else {
        QueryDB "SELECT b.id, b.title, b.author, b.isbn, c.name as category, b.price, b.stock FROM books b LEFT JOIN categories c ON b.category_id = c.id ORDER BY b.title"
    }
    
    result = GetQueryResult
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

function getBook() {
    bookId = GetHTTPQuery "id"
    
    if bookId == "" {
        SetHTTPResponse 400 "Book ID is required"
        return
    }
    
    QueryRowDB "SELECT b.id, b.title, b.author, b.isbn, c.name as category, b.price, b.stock FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE b.id = ?" bookId
    result = GetQueryResult
    
    if result == "" {
        SetHTTPResponse 404 "Book not found"
        return
    }
    
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

function createBook() {
    title = GetHTTPQuery "title"
    author = GetHTTPQuery "author"
    isbn = GetHTTPQuery "isbn"
    categoryId = GetHTTPQuery "category_id"
    price = GetHTTPQuery "price"
    stock = GetHTTPQuery "stock"
    
    if title == "" {
        SetHTTPResponse 400 "Book title is required"
        return
    }
    
    if stock == "" {
        stock = "0"
    }
    if price == "" {
        price = "0"
    }
    
    ExecDB "INSERT INTO books (title, author, isbn, category_id, price, stock) VALUES (?, ?, ?, ?, ?, ?)" title author isbn categoryId price stock
    SetHTTPResponse 201 "Book created"
}

function updateBook() {
    bookId = GetHTTPQuery "id"
    title = GetHTTPQuery "title"
    author = GetHTTPQuery "author"
    isbn = GetHTTPQuery "isbn"
    categoryId = GetHTTPQuery "category_id"
    price = GetHTTPQuery "price"
    stock = GetHTTPQuery "stock"
    
    if bookId == "" {
        SetHTTPResponse 400 "Book ID is required"
        return
    }
    
    if title != "" {
        ExecDB "UPDATE books SET title = ? WHERE id = ?" title bookId
    }
    if author != "" {
        ExecDB "UPDATE books SET author = ? WHERE id = ?" author bookId
    }
    if isbn != "" {
        ExecDB "UPDATE books SET isbn = ? WHERE id = ?" isbn bookId
    }
    if categoryId != "" {
        ExecDB "UPDATE books SET category_id = ? WHERE id = ?" categoryId bookId
    }
    if price != "" {
        ExecDB "UPDATE books SET price = ? WHERE id = ?" price bookId
    }
    if stock != "" {
        ExecDB "UPDATE books SET stock = ? WHERE id = ?" stock bookId
    }
    
    SetHTTPResponse 200 "Book updated"
}

function deleteBook() {
    bookId = GetHTTPQuery "id"
    
    if bookId == "" {
        SetHTTPResponse 400 "Book ID is required"
        return
    }
    
    ExecDB "DELETE FROM books WHERE id = ?" bookId
    SetHTTPResponse 200 "Book deleted"
}
