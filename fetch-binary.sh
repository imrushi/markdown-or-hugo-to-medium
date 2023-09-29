#!/bin/bash

# Get the release JSON from GitHub API
release_json=$(curl -s https://api.github.com/repos/imrushi/markdown-or-hugo-to-medium/releases/latest)

# Extract the binary URL from the release assets
binary_url=$(echo "$release_json" | jq -r '.assets[] | select(.name == "HugoToMedium") | .browser_download_url')

# Fetch the binary and make it executable
curl -o HugoToMedium -L "$binary_url"
chmod +x HugoToMedium

# Print a message to indicate the binary has been fetched
echo "HugoToMedium binary has been fetched and is ready for use."
