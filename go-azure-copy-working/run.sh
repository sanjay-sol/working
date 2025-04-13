
# set -e

# # --- CONFIG ---
# REPO_URL="https://github.com/your-username/your-repo.git"  # 👈 Replace with actual repo
# REPO_DIR="your-repo"  # 👈 Replace with your repo folder name after clone
# DOCKER_IMAGE_NAME="uploader-app"
# AZURE_CONTAINER=${1:-default-container}  # First arg, fallback to 'default-container'
# # Add more envs if needed
# AZURE_ACCOUNT_NAME=${AZURE_ACCOUNT_NAME:-"harvestedstorage2"}
# AZURE_ACCOUNT_KEY=${AZURE_ACCOUNT_KEY:-"bihW5fxPa/VdaATbn5iBgj+yd6XBmn6LQaXEjgHiThbJ3RcW+M6TtQc5Ml3cfihXruNRQRjzYGpU+AStU/OnHA=="}
# IMAGE_SOURCE=${IMAGE_SOURCE:-"/Users/sanjaysirangi/Desktop/go-azure-copy-working/files"}
# RETRY_PATH=${RETRY_PATH:-"/tmp/retry"}

# echo "🔍 Checking prerequisites..."

# # --- Docker Check ---
# if ! command -v docker &> /dev/null; then
#     echo "🐳 Docker not found. Installing Docker..."
#     sudo apt-get update
#     sudo apt-get install -y docker.io
#     sudo systemctl enable docker
#     sudo systemctl start docker
#     sudo usermod -aG docker $USER
#     echo "✅ Docker installed."
# else
#     echo "✅ Docker already installed."
# fi

# # --- Go Check ---
# if ! command -v go &> /dev/null; then
#     echo "🔧 Go not found. Installing Go..."
#     sudo apt-get update
#     sudo apt-get install -y golang
#     echo "✅ Go installed."
# else
#     echo "✅ Go already installed."
# fi

# # --- Git Check ---
# if ! command -v git &> /dev/null; then
#     echo "🔧 Git not found. Installing Git..."
#     sudo apt-get update
#     sudo apt-get install -y git
#     echo "✅ Git installed."
# else
#     echo "✅ Git already installed."
# fi

# # --- Clone Repo ---
# if [ ! -d "$REPO_DIR" ]; then
#     echo "📦 Cloning repo..."
#     git clone "$REPO_URL"
# else
#     echo "📦 Repo already exists. Pulling latest..."
#     cd "$REPO_DIR" && git pull && cd ..
# fi

# # --- Docker Build ---
# cd "$REPO_DIR"
# echo "🏗️  Building Docker image..."
# docker build -t "$DOCKER_IMAGE_NAME" .

# # --- Docker Run ---
# echo "🚀 Starting uploader container..."
# docker run --rm \
#     --env AZURE_ACCOUNT_NAME="$AZURE_ACCOUNT_NAME" \
#     --env AZURE_ACCOUNT_KEY="$AZURE_ACCOUNT_KEY" \
#     --env AZURE_CONTAINER="$AZURE_CONTAINER" \
#     --env IMAGE_SOURCE="$IMAGE_SOURCE" \
#     --env RETRY_PATH="$RETRY_PATH" \
#     "$DOCKER_IMAGE_NAME"

#!/bin/bash

set -e

# --- CONFIG ---
REPO_URL="https://github.com/sanjay-sol/harvested_working.git"  # Replace this
REPO_DIR="harvested_working"  # Replace with actual repo dir name
DOCKER_IMAGE_NAME="uploader-app"
AZURE_CONTAINER=${1:-default-container}  # 👈 First argument = container name
AZURE_ACCOUNT_NAME=${AZURE_ACCOUNT_NAME:-"harvestedstorage2"}
AZURE_ACCOUNT_KEY=${AZURE_ACCOUNT_KEY:-"bihW5fxPa/VdaATbn5iBgj+yd6XBmn6LQaXEjgHiThbJ3RcW+M6TtQc5Ml3cfihXruNRQRjzYGpU+AStU/OnHA=="}  # Update
IMAGE_SOURCE=${IMAGE_SOURCE:-"/Users/sanjaysirangi/Desktop/mimic/images"}
RETRY_PATH=${RETRY_PATH:-"/tmp/retry"}

echo "🧪 Checking prerequisites..."

command -v docker &> /dev/null || {
    echo "🐳 Installing Docker..."
    sudo apt-get update
    sudo apt-get install -y docker.io
    sudo systemctl enable docker
    sudo systemctl start docker
    sudo usermod -aG docker $USER
}

command -v go &> /dev/null || {
    echo "🔧 Installing Go..."
    sudo apt-get update
    sudo apt-get install -y golang
}

command -v git &> /dev/null || {
    echo "🔧 Installing Git..."
    sudo apt-get update
    sudo apt-get install -y git
}

if [ ! -d "$REPO_DIR" ]; then
    echo "📥 Cloning repository..."
    git clone "$REPO_URL"
else
    echo "📥 Updating repository..."
    cd "$REPO_DIR" && git pull && cd ..
fi

cd "$REPO_DIR"

echo "🐳 Building Docker image..."
docker build -t "$DOCKER_IMAGE_NAME" .

echo "🚀 Running Docker container..."
docker run --rm \
    -e AZURE_ACCOUNT_NAME="$AZURE_ACCOUNT_NAME" \
    -e AZURE_ACCOUNT_KEY="$AZURE_ACCOUNT_KEY" \
    -e AZURE_CONTAINER="$AZURE_CONTAINER" \
    -e IMAGE_SOURCE="$IMAGE_SOURCE" \
    -e RETRY_PATH="$RETRY_PATH" \
    -e RETRY_INTERVAL=10 \
    -e QUEUE_SIZE=500 \
    -e MAX_WORKERS=100 \
    -e TOTAL_RETRY_WORKERS=10 \
    -e LOG_LEVEL=info \
    -e RETRY_LIMIT=3 \
    "$DOCKER_IMAGE_NAME"
