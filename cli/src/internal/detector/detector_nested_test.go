//go:build integration

package detector

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNestedWorkspaceScenario tests the exact scenario from the bug report:
// User is in a subdirectory, azure.yaml is found in parent, but projects
// from OUTSIDE the azure.yaml directory should not be detected.
//
// Directory structure:
// /tmp/
//
//	├── outside-project/
//	│   └── package.json (SHOULD NOT BE FOUND)
//	└── my-workspace/
//	    ├── azure.yaml (workspace root)
//	    ├── frontend/
//	    │   └── package.json (SHOULD BE FOUND)
//	    └── backend/
//	        └── package.json (SHOULD BE FOUND)
//
// When running from my-workspace/frontend, the system should:
// 1. Find azure.yaml in parent (my-workspace)
// 2. Search from my-workspace directory (NOT from /)
// 3. Find frontend and backend projects
// 4. NOT find outside-project
func TestNestedWorkspaceScenario(t *testing.T) {
	// Create temporary directory structure
	tmpRoot := t.TempDir()

	// 1. Create an outside project (sibling to workspace)
	outsideProjectDir := filepath.Join(tmpRoot, "outside-project")
	if err := os.MkdirAll(outsideProjectDir, 0o755); err != nil {
		t.Fatalf("Failed to create outside project dir: %v", err)
	}
	outsidePackageJSON := filepath.Join(outsideProjectDir, "package.json")
	if err := os.WriteFile(outsidePackageJSON, []byte(`{
		"name": "outside-project",
		"scripts": {"dev": "vite"}
	}`), 0o644); err != nil {
		t.Fatalf("Failed to create outside package.json: %v", err)
	}

	// 2. Create workspace directory
	workspaceDir := filepath.Join(tmpRoot, "my-workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("Failed to create workspace dir: %v", err)
	}

	// 3. Create azure.yaml in workspace
	azureYamlPath := filepath.Join(workspaceDir, "azure.yaml")
	azureYamlContent := `name: my-app
services:
  frontend:
    project: ./frontend
    language: node
  backend:
    project: ./backend
    language: node
`
	if err := os.WriteFile(azureYamlPath, []byte(azureYamlContent), 0o644); err != nil {
		t.Fatalf("Failed to create azure.yaml: %v", err)
	}

	// 4. Create frontend service
	frontendDir := filepath.Join(workspaceDir, "frontend")
	if err := os.MkdirAll(frontendDir, 0o755); err != nil {
		t.Fatalf("Failed to create frontend dir: %v", err)
	}
	frontendPackageJSON := filepath.Join(frontendDir, "package.json")
	if err := os.WriteFile(frontendPackageJSON, []byte(`{
		"name": "frontend",
		"scripts": {"dev": "vite"}
	}`), 0o644); err != nil {
		t.Fatalf("Failed to create frontend package.json: %v", err)
	}

	// 5. Create backend service
	backendDir := filepath.Join(workspaceDir, "backend")
	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		t.Fatalf("Failed to create backend dir: %v", err)
	}
	backendPackageJSON := filepath.Join(backendDir, "package.json")
	if err := os.WriteFile(backendPackageJSON, []byte(`{
		"name": "backend",
		"scripts": {"dev": "node server.js"}
	}`), 0o644); err != nil {
		t.Fatalf("Failed to create backend package.json: %v", err)
	}

	// TEST 1: Search from workspace directory
	t.Run("search_from_workspace_root", func(t *testing.T) {
		projects, err := FindNodeProjects(workspaceDir)
		if err != nil {
			t.Fatalf("FindNodeProjects failed: %v", err)
		}

		// Should find exactly 2 projects
		if len(projects) != 2 {
			t.Errorf("Expected 2 projects, found %d", len(projects))
			for i, p := range projects {
				t.Logf("  Project %d: %s", i, p.Dir)
			}
		}

		// Verify outside project is NOT in results
		for _, p := range projects {
			if p.Dir == outsideProjectDir {
				t.Errorf("Outside project should NOT be found! Found at: %s", p.Dir)
			}
		}

		// Verify expected projects ARE in results
		foundFrontend := false
		foundBackend := false
		for _, p := range projects {
			if p.Dir == frontendDir {
				foundFrontend = true
			}
			if p.Dir == backendDir {
				foundBackend = true
			}
		}

		if !foundFrontend {
			t.Error("Frontend project should be found")
		}
		if !foundBackend {
			t.Error("Backend project should be found")
		}
	})

	// TEST 2: Simulate the bug - FindAzureYaml from subdirectory, then search
	// This is what the actual app does
	t.Run("simulate_real_workflow", func(t *testing.T) {
		// First, find azure.yaml from a subdirectory (like user running from frontend/)
		foundAzureYamlPath, err := FindAzureYaml(frontendDir)
		if err != nil {
			t.Fatalf("FindAzureYaml failed: %v", err)
		}

		if foundAzureYamlPath != azureYamlPath {
			t.Errorf("Expected to find azure.yaml at %s, got %s", azureYamlPath, foundAzureYamlPath)
		}

		// Now search from the azure.yaml directory (this is what the fix does)
		searchRoot := filepath.Dir(foundAzureYamlPath)
		projects, err := FindNodeProjects(searchRoot)
		if err != nil {
			t.Fatalf("FindNodeProjects failed: %v", err)
		}

		// Should find exactly 2 projects (frontend and backend)
		if len(projects) != 2 {
			t.Errorf("Expected 2 projects, found %d", len(projects))
			for i, p := range projects {
				t.Logf("  Project %d: %s", i, p.Dir)
			}
		}

		// CRITICAL: Verify outside project is NOT found
		for _, p := range projects {
			if p.Dir == outsideProjectDir {
				t.Errorf("BUG REPRODUCED! Outside project was found at: %s", p.Dir)
				t.Error("This means the boundary checking is not working!")
			}
		}
	})

	// TEST 3: Verify the OLD buggy behavior would have failed
	// (searching from tmpRoot would find ALL 3 projects)
	t.Run("demonstrate_bug_if_no_boundary", func(t *testing.T) {
		// If we search from the root, we'd find all 3 projects (the bug)
		allProjects, err := FindNodeProjects(tmpRoot)
		if err != nil {
			t.Fatalf("FindNodeProjects failed: %v", err)
		}

		// This SHOULD find 3 projects (demonstrating the bug we fixed)
		if len(allProjects) != 3 {
			t.Logf("Note: Found %d projects when searching from root", len(allProjects))
		} else {
			t.Logf("Without boundary: Would have found %d projects (including outside-project)", len(allProjects))
		}
	})
}
