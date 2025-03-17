#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status

# Directory for the test
TEST_DIR=$(mktemp -d -t til-integration-test-XXXXXX)
echo "Using test directory: $TEST_DIR"

# Load environment variables from .env file
if [ -f .env ]; then
    echo "Loading .env file"
    export $(grep -v '^#' .env | xargs)
else
    echo "No .env file found. Please create one with NOTION_API_KEY and NOTION_DB_ID."
    exit 1
fi

# Check required environment variables
if [ -z "$NOTION_API_KEY" ] || [ -z "$NOTION_DB_ID" ]; then
    echo "NOTION_API_KEY and NOTION_DB_ID must be set in .env file."
    exit 1
fi

# Function to clean up the test directory
cleanup() {
    echo "Cleaning up test directory: $TEST_DIR"
    rm -rf "$TEST_DIR"
}

# Register the cleanup function to be called on script exit
trap cleanup EXIT

# Change to the test directory
cd "$TEST_DIR"

# Create a test file
echo "Creating test file"
echo "This is a test file for TIL integration test" > test_file.txt

# Build the TIL binary if it doesn't exist
if [ ! -f "../til" ]; then
    echo "Building TIL binary"
    (cd .. && go build -o til)
fi

# Copy the binary to the test directory
cp ../til .

# Test til version
echo "Testing 'til version'"
./til version
if [ $? -ne 0 ]; then
    echo "❌ til version failed"
    exit 1
fi
echo "✅ til version succeeded"

# Initialize TIL repository with Notion integration
echo "Initializing TIL repository with Notion integration"
cat > init_input.txt << EOF
y
$NOTION_API_KEY
$NOTION_DB_ID
n

EOF

cat init_input.txt | ./til init
if [ $? -ne 0 ]; then
    echo "❌ til init failed"
    exit 1
fi
echo "✅ til init succeeded"

# Verify the repository was initialized
if [ ! -d "til" ] || [ ! -f "til/til.md" ]; then
    echo "❌ Repository initialization failed, til directory or til.md not found"
    exit 1
fi
echo "✅ Repository structure verified"

# Add the test file
echo "Adding test file"
./til add test_file.txt
if [ $? -ne 0 ]; then
    echo "❌ til add failed"
    exit 1
fi
echo "✅ til add succeeded"

# Verify the file was staged
if [ ! -f ".til/staging/test_file.txt" ]; then
    echo "❌ File was not staged"
    exit 1
fi
echo "✅ File staging verified"

# Commit the entry
COMMIT_MESSAGE="Test commit from integration test $(date +%s)"
echo "Committing entry with message: $COMMIT_MESSAGE"
./til commit -m "$COMMIT_MESSAGE"
if [ $? -ne 0 ]; then
    echo "❌ til commit failed"
    exit 1
fi
echo "✅ til commit succeeded"

# Verify the commit
./til log -n 1 | grep "$COMMIT_MESSAGE"
if [ $? -ne 0 ]; then
    echo "❌ Commit verification failed, entry not found in log"
    exit 1
fi
echo "✅ Commit verified in log"

# Test log output
echo "Testing 'til log'"
./til log -n 5
if [ $? -ne 0 ]; then
    echo "❌ til log failed"
    exit 1
fi
echo "✅ til log succeeded"

# Push to Notion
echo "Pushing to Notion"
./til push --notion
if [ $? -ne 0 ]; then
    echo "❌ til push --notion failed"
    exit 1
fi
echo "✅ til push --notion succeeded"

# Amend the commit
NEW_MESSAGE="$COMMIT_MESSAGE (amended)"
echo "Amending commit with message: $NEW_MESSAGE"
echo "This is an additional line for the amended test file" >> test_file.txt
./til add test_file.txt
./til commit --amend -m "$NEW_MESSAGE"
if [ $? -ne 0 ]; then
    echo "❌ til commit --amend failed"
    exit 1
fi
echo "✅ til commit --amend succeeded"

# Verify the amended commit
./til log -n 1 | grep "$NEW_MESSAGE"
if [ $? -ne 0 ]; then
    echo "❌ Amend verification failed, amended entry not found in log"
    exit 1
fi
echo "✅ Amended commit verified in log"

# Push to Notion again
echo "Pushing amended entry to Notion"
./til push --notion
if [ $? -ne 0 ]; then
    echo "❌ Second til push --notion failed"
    exit 1
fi
echo "✅ Second til push --notion succeeded"

echo ""
echo "🎉 All tests passed! 🎉"
echo ""
echo "Test summary:"
echo "- TIL version command works"
echo "- TIL repository initializes with Notion integration"
echo "- Files can be added and staged"
echo "- Entries can be committed with a message"
echo "- Commits are recorded in the log"
echo "- Entries can be pushed to Notion"
echo "- Commits can be amended"
echo "- Amended entries can be pushed to Notion"

exit 0
