#!/bin/sh
# Shode Advanced Features Demo
# This script demonstrates the new execution engine capabilities

echo "=== Shode Advanced Features Demo ==="
echo ""

# 1. Variable Assignment
echo "1. Variable Assignment:"
NAME="Shode"
VERSION="0.2.0"
echo "Project: $NAME v$VERSION"
echo ""

# 2. Pipelines
echo "2. Pipeline Support:"
echo "Creating test file..."
echo "apple
banana
cherry
date
elderberry" > /tmp/fruits.txt

echo "Counting lines with pipeline:"
cat /tmp/fruits.txt | wc -l
echo ""

# 3. Redirection
echo "3. Input/Output Redirection:"
echo "Writing to file..."
echo "Hello from Shode!" > /tmp/shode_output.txt
echo "Appending to file..."
echo "This is line 2" >> /tmp/shode_output.txt
echo "Reading file back:"
cat /tmp/shode_output.txt
echo ""

# 4. Control Flow - If Statement
echo "4. If-Then-Else Statement:"
if test -f /tmp/shode_output.txt; then
    echo "✓ File exists"
else
    echo "✗ File not found"
fi
echo ""

# 5. For Loop
echo "5. For Loop:"
echo "Iterating over fruits:"
for fruit in apple banana cherry; do
    echo "  - $fruit"
done
echo ""

# 6. While Loop (limited demo)
echo "6. While Loop:"
counter=0
while test $counter -lt 3; do
    echo "  Counter: $counter"
    counter=$((counter + 1))
done
echo ""

# 7. Standard Library Functions
echo "7. Standard Library Functions:"
echo "Using Shode built-in functions..."

# File operations
echo "FileExists check:"
if FileExists "/tmp/shode_output.txt"; then
    echo "✓ File found using FileExists"
fi

# String operations
echo "String operations:"
text="hello world"
echo "Original: $text"
echo "Uppercase: $(ToUpper "$text")"
echo "Contains 'world': $(Contains "$text" "world")"
echo ""

# 8. Complex Pipeline
echo "8. Complex Pipeline Example:"
echo "Creating log file..."
echo "ERROR: Something went wrong
INFO: Starting process
ERROR: Another error
INFO: Process complete
ERROR: Final error" > /tmp/app.log

echo "Filtering errors and counting:"
cat /tmp/app.log | grep "ERROR" | wc -l
echo ""

# 9. Environment Variables
echo "9. Environment Variables:"
export SHODE_ENV="production"
echo "SHODE_ENV: $SHODE_ENV"
echo "Current directory: $(pwd)"
echo ""

# Cleanup
echo "Cleaning up temporary files..."
rm -f /tmp/fruits.txt /tmp/shode_output.txt /tmp/app.log
echo ""

echo "=== Demo Complete ==="
echo "Shode v0.2.0 - Modern Shell Scripting Platform"
