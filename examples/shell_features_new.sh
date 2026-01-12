#!/bin/sh
# Shode New Shell Features Demo
# Demonstrates: Background Jobs, Command Substitution, Arrays

echo "=== Shode New Shell Features Demo ==="
echo ""

# 1. Command Substitution
echo "1. Command Substitution:"
echo "Current date: $(date)"
echo "Current user: $(whoami)"
echo "Working directory: $(pwd)"
echo ""

# Using backticks
echo "Using backticks:"
echo "Hostname: `hostname`"
echo ""

# 2. Arrays
echo "2. Array Support:"
fruits=(apple banana cherry date elderberry)
echo "Fruits array defined: fruits=(apple banana cherry date elderberry)"
echo "Array as string: $fruits"
echo ""

# 3. Background Jobs
echo "3. Background Jobs:"
echo "Starting background task..."
echo "This is a background task" &
echo "Main script continues immediately"
echo ""

# 4. Combining Features
echo "4. Combining Features:"
# Use command substitution to get list of files
files=$(ls -1 2>/dev/null | head -5)
echo "Files from command substitution: $files"
echo ""

# Array with command substitution
echo "Creating array from command substitution:"
file_array=($(ls -1 2>/dev/null | head -3))
echo "File array: $file_array"
echo ""

echo "=== Demo Complete ==="
