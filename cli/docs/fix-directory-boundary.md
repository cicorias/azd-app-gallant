# Fix: Directory Boundary Issue in Project Detection

## Problem

The `app deps` and `app run` commands were sometimes detecting projects outside the intended workspace. This happened because:

1. `FindAzureYaml()` searches **up** the directory tree to find `azure.yaml`
2. When `azure.yaml` was found in a parent directory, service paths were resolved relative to that directory
3. However, `FindNodeProjects()`, `FindPythonProjects()`, and other detector functions did not have boundary checking
4. They could traverse outside the workspace and find unrelated projects in parent directories

## Solution

### 1. Added Boundary Checking to Detector Functions

Modified all detector functions to prevent traversal outside the search root:
- `FindNodeProjects()` 
- `FindPythonProjects()`
- `FindDotnetProjects()`
- `FindAppHost()`

Each function now:
- Computes the absolute path of each visited file/directory
- Checks if the path is within the root directory using `filepath.Rel()`
- Skips any directory that would go outside the boundary (when `relPath` starts with `..`)

### 2. Updated `executeDeps()` to Use Azure.yaml Directory

Modified `cli/src/cmd/app/commands/core.go` to:
- Find the `azure.yaml` location first
- Use the `azure.yaml` directory as the search root for project detection
- Fall back to current working directory if no `azure.yaml` is found

This ensures that all project detection starts from the workspace root, not from arbitrary subdirectories.

## Code Changes

### detector.go
- Added boundary check in all `Find*Projects` functions
- Uses `filepath.Rel()` to detect if path is outside root
- Returns `filepath.SkipDir` for paths starting with `..`

### core.go
- Calls `detector.FindAzureYaml(cwd)` to locate workspace root
- Uses `filepath.Dir(azureYamlPath)` as `searchRoot`
- Passes `searchRoot` to all `Find*Projects` calls

## Testing

### Integration Tests
Created `detector_boundary_test.go` with 4 integration tests:
- `TestFindNodeProjectsRespectsBoundary` ✅
- `TestFindPythonProjectsRespectsBoundary` ✅
- `TestFindDotnetProjectsRespectsBoundary` ✅
- `TestFindAppHostRespectsBoundary` ✅

All tests verify that projects outside the workspace are NOT detected.

### Manual Test Project
Created `cli/tests/projects/boundary-test/` with:
- Parent `package.json` (should NOT be found)
- Workspace with `azure.yaml`
- Two services inside workspace (should be found)

Test confirms only workspace projects are detected:
```
📦 Found 1 Node.js project(s)  ✅
🐍 Found 1 Python project(s)   ✅
```

Parent project is correctly ignored! ✅

## Files Changed

1. `cli/src/internal/detector/detector.go` - Added boundary checking
2. `cli/src/cmd/app/commands/core.go` - Use azure.yaml directory as search root
3. `cli/src/internal/detector/detector_boundary_test.go` - New integration tests
4. `cli/tests/projects/boundary-test/` - New test project

## Impact

- Prevents detection of unintended projects
- Services only run from within the workspace
- Dependencies only installed for workspace projects
- Clearer scope for multi-project repositories
