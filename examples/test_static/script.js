// Test JavaScript file
console.log('Shode static file server - JavaScript loaded successfully!');

function showAlert() {
    alert('JavaScript is working in Shode!');
}

function getCurrentTime() {
    const now = new Date();
    return now.toLocaleString();
}

// Log to console when page loads
document.addEventListener('DOMContentLoaded', function() {
    console.log('Page loaded at: ' + getCurrentTime());
    console.log('Static file server is fully functional!');
});
