#!/bin/bash

# Function to get the size of a branch in a Git repository without cloning it
get_branch_size() {
    local repo_url=$1
    local branch_name=$2
    
    # Fetch branch information without cloning the repository
    local commit_hash=$(git ls-remote --heads "$repo_url" "$branch_name" | awk '{ print $1 }')
    if [ -z "$commit_hash" ]; then
        echo "Failed to fetch commit hash of branch '$branch_name'"
        exit 1
    fi
    
    # Calculate the size of the branch
    local branch_size=$(git archive --format=tar --remote "$repo_url" "$commit_hash" | wc -c)
    if [ -z "$branch_size" ]; then
        echo "Failed to calculate size of branch '$branch_name'"
        exit 1
    fi
    
    echo "$branch_size"
}

# Replace these variables with your actual repository URL and branch name
REPO_URL="http://ec2-18-194-139-24.eu-central-1.compute.amazonaws.com:7990/scm/ten/tensorflow.git"
BRANCH_NAME="main"

# Get the size of the branch
branch_size=$(get_branch_size "$REPO_URL" "$BRANCH_NAME")
echo "Size of branch '$BRANCH_NAME': $branch_size bytes"

