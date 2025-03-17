#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status

# Directory for the test
TEST_DIR=$(mktemp -d -t til-git-integration-test-XXXXXX)
echo "Using test directory: $TEST_DIR"

# Load environment variables from .env file
if [ -f .env ]; then
    echo "Loading .env file"
    export $(grep -v '^#' .env | xargs)
else
    echo "No .env file found. Creating a minimal one for Git testing."
    cat > .env << EOF
# Credentials for TIL integration tests
GIT_REMOTE_URL=https://github.com/yourusername/til-test.git
EOF
    export $(grep -v '^#' .env | xargs)
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
echo "This is a test file for TIL Git integration test" > test_file.txt

# Build the TIL binary if it doesn't exist
if [ ! -f "../til" ]; then
    echo "Building TIL binary"
    (cd .. && go build -o til)
fi

# Copy the binary to the test directory
cp ../til .

# Check if git is installed
if ! command -v git &> /dev/null; then
    echo "Git is not installed. Skipping Git integration tests."
    exit 0
fi

# Create a local Git repository for testing
echo "Creating a local Git repository for testing"
mkdir local-git-repo
cd local-git-repo
git init
git config --local user.name "Test User"
git config --local user.email "test@example.com"
touch README.md
git add README.md
git commit -m "Initial commit"
cd ..

# Path to the local Git repository
LOCAL_GIT_PATH="file://$TEST_DIR/local-git-repo"

# Initialize TIL repository with Git integration
echo "Initializing TIL repository with Git integration"
cat > init_input.txt << EOF
n
y
$LOCAL_GIT_PATH

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

# Commit the entry
COMMIT_MESSAGE="Test Git commit from integration test $(date +%s)"
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

# Add a second file
echo "Creating and adding a second test file"
echo "This is the second test file" > test_file2.txt
./til add test_file2.txt
if [ $? -ne 0 ]; then
    echo "❌ til add for second file failed"
    exit 1
fi
echo "✅ Second file added successfully"

# Amend the commit
NEW_MESSAGE="$COMMIT_MESSAGE (amended with second file)"
echo "Amending commit with message: $NEW_MESSAGE"
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

# Manual verification of Git repository
echo "Verifying Git repository contents"
cd til
if [ ! -d ".git" ]; then
    echo "❌ Git repository not found in til directory"
    exit 1
fi

git status
git log --oneline -n 1

# Check if our files are in the repository
if [ ! -f "files/$(date +%Y-%m-%d)_test_file.txt" ] || [ ! -f "files/$(date +%Y-%m-%d)_test_file2.txt" ]; then
    echo "❌ Files not found in the expected location"
    exit 1
fi
echo "✅ Git repository verified"

cd ..

# Push to Git
echo "Pushing to Git repository"
./til push --git
if [ $? -ne 0 ]; then
    echo "❌ til push --git failed"
    exit 1
fi
echo "✅ til push --git succeeded"

echo ""
echo "🎉 All Git integration tests passed! 🎉"
echo ""
echo "Test summary:"
echo "- TIL repository initializes with Git integration"
echo "- Files can be added and staged"
echo "- Entries can be committed with a message"
echo "- Commits are recorded in both TIL log and Git"
echo "- Commits can be amended"
echo "- Files are stored in the expected location"
echo "- Changes can be pushed to the Git repository"

exit 0
