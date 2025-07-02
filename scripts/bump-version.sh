#!/bin/bash

# Script to automatically bump version number for arbor library
# Usage: ./bump-version.sh [major|minor|patch] [optional-custom-version]

set -e

# Default to patch increment
BUMP_TYPE=${1:-patch}
CUSTOM_VERSION=$2

# Function to increment version
increment_version() {
    local version=$1
    local bump_type=$2
    
    # Extract major, minor, patch from version (format: X.Y.Z)
    if [[ $version =~ ^([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
        major="${BASH_REMATCH[1]}"
        minor="${BASH_REMATCH[2]}"
        patch="${BASH_REMATCH[3]}"
    else
        echo "Error: Invalid version format: $version"
        exit 1
    fi
    
    case $bump_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            echo "Error: Invalid bump type: $bump_type"
            echo "Valid types: major, minor, patch"
            exit 1
            ;;
    esac
    
    echo "$major.$minor.$patch"
}

# Get current version from latest git tag
CURRENT_TAG=$(git tag --sort=-version:refname | head -1)
if [[ -z "$CURRENT_TAG" ]]; then
    echo "Error: No tags found. Creating initial version 1.0.0"
    CURRENT_VERSION="1.0.0"
else
    # Remove 'v' prefix if present
    CURRENT_VERSION="${CURRENT_TAG#v}"
fi

echo "Current version: $CURRENT_VERSION"

# Calculate new version
if [[ -n "$CUSTOM_VERSION" ]]; then
    NEW_VERSION="$CUSTOM_VERSION"
    echo "Using custom version: $NEW_VERSION"
else
    NEW_VERSION=$(increment_version "$CURRENT_VERSION" "$BUMP_TYPE")
    echo "Bumping $BUMP_TYPE version: $CURRENT_VERSION -> $NEW_VERSION"
fi

# Update config.yml if it exists
if [[ -f "config.yml" ]]; then
    echo "Updating config.yml..."
    sed -i "s/^  version: \".*\"/  version: \"$NEW_VERSION\"/" config.yml
    echo "✓ Updated config.yml"
fi

# Update CHANGELOG.md if it exists
if [[ -f "CHANGELOG.md" ]]; then
    echo "Updating CHANGELOG.md..."
    # Add new version entry at the top
    sed -i "6i\\## [${NEW_VERSION}] - $(date +%Y-%m-%d)\\n\\n### Added\\n- Version bump to ${NEW_VERSION}\\n" CHANGELOG.md
    echo "✓ Updated CHANGELOG.md"
fi

echo ""
echo "🎉 Version bump complete!"
echo "📝 New version: $NEW_VERSION"
echo ""
echo "Next steps:"
echo "1. Commit the changes: git add . && git commit -m \"Bump version to $NEW_VERSION\""
echo "2. Tag the release: git tag v$NEW_VERSION"
echo "3. Push: git push origin main --tags"
