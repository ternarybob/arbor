# Release Process Improvements

## Overview

The release process has been improved to follow a proper workflow that runs tests before committing and prevents duplicate CI/CD pipeline triggers.

## Changes Made

### 1. PowerShell Release Script Updates (`../helpers/create-release.ps1`)

The script now follows this improved process:

1. **Step 1: Run Tests and Validation**
   - Runs `go test -v ./...` 
   - Runs `go vet ./...` for static analysis
   - Runs `go build ./...` for build validation
   - **If any step fails, the release is aborted**

2. **Step 2: Commit Changes**
   - Checks for uncommitted changes
   - Creates a commit with auto-generated message if needed
   - Uses format: `Release - YYYY-MM-DD HH:MM by username`

3. **Step 3: Create Tag**
   - Creates annotated git tag with release version
   - Tag message format: `Release vX.Y.Z`

4. **Step 4: Push to Remote**
   - Pushes both commit and tag to GitHub
   - Only the tag push triggers the release workflow (no duplicates)

### 2. GitHub Actions Workflow Separation

#### Main Branch CI (`ci-main.yml`)
- Triggers on: Push to `main` branch, Pull Requests to `main`
- Purpose: Continuous integration for regular development
- Jobs: Test & Validate only (no releases)

#### Release Pipeline (`ci.yml`)
- Triggers on: Push of tags matching `v*` pattern
- Purpose: Create GitHub releases for tagged versions
- Jobs: 
  - Test & Validate
  - Create Release (only runs if tests pass)

### 3. Key Improvements

#### Prevents Duplicate Workflows
- **Before**: Both commit to `main` AND tag push triggered workflows → 2 pipelines
- **After**: Only tag push triggers release workflow → 1 pipeline

#### Enforces Test-First Approach
- **Before**: Script would commit/push without running tests locally
- **After**: Tests run locally before any commits, failing fast if issues exist

#### Proper Dependency Chain
- **Before**: Release job could run even if tests failed
- **After**: Release job requires explicit test success with `if: success()` condition

#### Better Error Handling
- Script exits immediately on any test/build failure
- GitHub Actions properly chain job dependencies
- Clear error messages for each failure point

## Usage

### Create a Release
```powershell
# From the helpers directory
./create-release.ps1 -ProjectName arbor

# With specific version bump
./create-release.ps1 -ProjectName arbor -Minor
./create-release.ps1 -ProjectName arbor -Major

# With explicit version
./create-release.ps1 -ProjectName arbor -Version v1.5.0
```

### Expected Flow
1. Developer runs release script
2. Script validates all tests pass locally
3. Script commits changes and creates tag
4. Script pushes to GitHub
5. **Only one** GitHub Actions workflow triggers (release pipeline)
6. GitHub Actions runs tests again in CI environment
7. If CI tests pass → GitHub release is created
8. If CI tests fail → No release is created

## Benefits

- **Reliability**: Tests must pass before any release artifacts are created
- **Efficiency**: No duplicate workflow runs wasting CI resources
- **Consistency**: Same validation steps locally and in CI
- **Safety**: Multiple validation checkpoints prevent broken releases
- **Clarity**: Clear separation between development CI and release processes

## Workflow Files

- `.github/workflows/ci-main.yml` - Main branch continuous integration
- `.github/workflows/ci.yml` - Release pipeline (tag-triggered only)
- `../helpers/create-release.ps1` - Enhanced release script with validation

## Testing the Process

The release script now includes comprehensive validation:
- Unit tests (`go test -v ./...`)
- Static analysis (`go vet ./...`) 
- Build validation (`go build ./...`)

All validation steps must pass before any commits or tags are created, ensuring only quality code reaches production.
