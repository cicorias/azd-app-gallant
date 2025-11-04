# Development Scripts

This folder contains helper scripts for local development and testing. These scripts are **not** required for end users - they're only for contributors working on the azd-app CLI extension.

> **⚠️ IMPORTANT:** Use `mage` as the primary interface for all development tasks. Build and install operations are implemented in pure Go via magefile for cross-platform support.

## Quick Start with Mage

```powershell
# Install mage
go install github.com/magefile/mage@latest

# See all available commands
mage -l

# Common tasks
mage build          # Build for current platform
mage buildall       # Build for all platforms (Windows/Linux/macOS)
mage install        # Build and install locally
mage installfast    # Quick install without dashboard build
mage test           # Run tests
mage testintegration # Run integration tests
mage run            # Run directly in test project
mage release        # Create a release
```

## Environment Variables for Mage

Some mage targets support configuration via environment variables:

### `mage testintegration`
```powershell
# Run specific package
$env:TEST_PACKAGE="installer"; mage testintegration

# Run specific test
$env:TEST_NAME="TestInstallNodeDependencies"; mage testintegration

# Set custom timeout (default: 10m)
$env:TEST_TIMEOUT="20m"; mage testintegration
```

### `mage run`
```powershell
# Run in specific project
$env:PROJECT_DIR="tests/projects/node/test-npm-project"; mage run

# Run specific command (default: run)
$env:COMMAND="deps"; mage run
```

## Scripts Overview

### 🚀 Build & Release

| Script | Purpose | Mage Command |
|--------|---------|--------------|
| `release.ps1` | Release automation (triggers GitHub workflow) | `mage release` |

### 🔧 Quick Development

| Script | Purpose | Mage Command |
|--------|---------|--------------|
| `install.ps1` | Install extension locally | `mage install` (or `mage installfast` for quick iteration) |
| `watch.ps1` | Auto-rebuild on file changes | N/A (file watching best done in shell) |
| `run-direct.ps1` | Run without installing | `mage run` |

### 🧪 Testing

| Script | Purpose | Mage Command |
|--------|---------|--------------|
| `test-integration.ps1` | Run integration tests | `mage testintegration` (supports TEST_PACKAGE, TEST_NAME env vars) |

---

## Detailed Usage

### install.ps1

**Install extension locally** - uses `azd extension build` to build and install the extension.

```powershell
# Install locally
.\scripts\install.ps1

# Uninstall
.\scripts\install.ps1 -Uninstall
```

**Mage equivalents:**
```powershell
# Full install with dashboard build
mage install

# Quick install without dashboard (faster iteration)
mage installfast

# Uninstall
mage uninstall
```

**When to use:** When you want to test the extension as end users would use it (`azd app <command>`).

### watch.ps1

**Auto-rebuild on file changes** - watches for `.go` file changes and automatically rebuilds.

```powershell
# Watch src directory (default)
.\scripts\watch.ps1

# Watch custom directory
.\scripts\watch.ps1 -Path "src/internal"
```

**Note:** File watching is best done via shell script. No direct mage equivalent.

**When to use:** When you want continuous rebuilds while editing code. Press Ctrl+C to stop.

### run-direct.ps1

**Run without installing** - executes the binary directly without installing as an azd extension.

```powershell
# Run default command (run) in default test project
.\scripts\run-direct.ps1

# Run specific command
.\scripts\run-direct.ps1 reqs

# Specify project directory
.\scripts\run-direct.ps1 run -ProjectDir "tests\projects\node\test-npm-project"
```

**Mage equivalent:**
```powershell
# Run in default project
mage run

# Run in specific project with specific command
$env:PROJECT_DIR="tests/projects/node/test-npm-project"; $env:COMMAND="deps"; mage run
```

**When to use:** Quick testing without needing to install the extension.

### test-integration.ps1

**Run integration tests** - executes integration tests that require external tools.

```powershell
# Run all integration tests
.\scripts\test-integration.ps1

# Run specific package
.\scripts\test-integration.ps1 -Package installer

# Run specific test
.\scripts\test-integration.ps1 -Test TestInstallNodeDependenciesIntegration

# Verbose output
.\scripts\test-integration.ps1 -Verbose
```

**Mage equivalent:**
```powershell
# Run all integration tests
mage testintegration

# Run specific package
$env:TEST_PACKAGE="installer"; mage testintegration

# Run specific test  
$env:TEST_NAME="TestInstallNodeDependencies"; mage testintegration

# Custom timeout (default: 10m)
$env:TEST_TIMEOUT="20m"; mage testintegration
```

**When to use:** Before submitting a PR or when testing functionality that requires external dependencies.

### release.ps1

**Create a release** - triggers the GitHub Actions workflow to create a draft release.

```powershell
# Interactive release (prompts for version bump type)
.\scripts\release.ps1

# Specify version directly
.\scripts\release.ps1 -Version 1.2.3

# Automatic version bump
.\scripts\release.ps1 -BumpType Patch
.\scripts\release.ps1 -BumpType Minor
.\scripts\release.ps1 -BumpType Major

# Dry run (shows what would happen)
.\scripts\release.ps1 -DryRun
```

**When to use:** To create a new release. This triggers a GitHub workflow that builds binaries, creates checksums, updates registry.json, and creates a draft release on GitHub.

---

## Recommended Development Workflow

### Option 1: Using Mage (Recommended)

```powershell
# Install mage if not already installed
go install github.com/magefile/mage@latest

# Build and install
mage install

# Run tests
mage test

# Build for all platforms
mage buildall

# See all available commands
mage -l
```

### Option 2: Quick Iteration with Scripts

```powershell
# Terminal 1: Watch for changes
.\scripts\watch.ps1

# Terminal 2: Make code changes and test
# The watch script will auto-rebuild on save
azd app run
```

### Option 3: Manual Testing

```powershell
# Build and install
.\scripts\install.ps1

# Test
cd tests\projects\fullstack-test
azd app run
```

---

## See Also

- [Main README](../README.md) - Project overview and user documentation
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Contribution guidelines
- [magefile.go](../magefile.go) - Build automation targets (primary interface)
- [Release Process](../docs/release.md) - Detailed release documentation
