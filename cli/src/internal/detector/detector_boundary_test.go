//go:build integration

package detector

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFindNodeProjectsRespectsBoundary tests that FindNodeProjects doesn't traverse outside the root directory.
func TestFindNodeProjectsRespectsBoundary(t *testing.T) {
	// Create a temporary directory structure:
	// /tmp/test-root/
	//   ├── parent-project/
	//   │   └── package.json (should NOT be found)
	//   └── workspace/
	//       ├── azure.yaml
	//       └── service/
	//           └── package.json (should be found)

	tmpDir := t.TempDir()

	// Create parent project (outside the workspace)
	parentProjectDir := filepath.Join(tmpDir, "parent-project")
	if err := os.MkdirAll(parentProjectDir, 0o755); err != nil {
		t.Fatalf("Failed to create parent project dir: %v", err)
	}
	parentPackageJSON := filepath.Join(parentProjectDir, "package.json")
	if err := os.WriteFile(parentPackageJSON, []byte(`{"name": "parent-project"}`), 0o644); err != nil {
		t.Fatalf("Failed to create parent package.json: %v", err)
	}

	// Create workspace directory
	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("Failed to create workspace dir: %v", err)
	}

	// Create azure.yaml in workspace
	azureYamlPath := filepath.Join(workspaceDir, "azure.yaml")
	azureYamlContent := `name: test-app
services:
  api:
    project: ./service
    language: node
    host: containerapp
`
	if err := os.WriteFile(azureYamlPath, []byte(azureYamlContent), 0o644); err != nil {
		t.Fatalf("Failed to create azure.yaml: %v", err)
	}

	// Create service directory inside workspace
	serviceDir := filepath.Join(workspaceDir, "service")
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}
	servicePackageJSON := filepath.Join(serviceDir, "package.json")
	if err := os.WriteFile(servicePackageJSON, []byte(`{"name": "service"}`), 0o644); err != nil {
		t.Fatalf("Failed to create service package.json: %v", err)
	}

	// Test: Search from workspace directory (where azure.yaml is located)
	projects, err := FindNodeProjects(workspaceDir)
	if err != nil {
		t.Fatalf("FindNodeProjects failed: %v", err)
	}

	// Verify: Only service project should be found, not parent project
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, found %d", len(projects))
		for i, p := range projects {
			t.Logf("Project %d: %s", i, p.Dir)
		}
	}

	if len(projects) > 0 && projects[0].Dir != serviceDir {
		t.Errorf("Expected service dir %s, got %s", serviceDir, projects[0].Dir)
	}

	// Verify: Parent project should NOT be in the results
	for _, p := range projects {
		if p.Dir == parentProjectDir {
			t.Errorf("Parent project should not be found when searching from workspace directory")
		}
	}
}

// TestFindPythonProjectsRespectsBoundary tests that FindPythonProjects doesn't traverse outside the root directory.
func TestFindPythonProjectsRespectsBoundary(t *testing.T) {
	// Create a temporary directory structure:
	// /tmp/test-root/
	//   ├── parent-project/
	//   │   └── requirements.txt (should NOT be found)
	//   └── workspace/
	//       ├── azure.yaml
	//       └── api/
	//           └── requirements.txt (should be found)

	tmpDir := t.TempDir()

	// Create parent project (outside the workspace)
	parentProjectDir := filepath.Join(tmpDir, "parent-project")
	if err := os.MkdirAll(parentProjectDir, 0o755); err != nil {
		t.Fatalf("Failed to create parent project dir: %v", err)
	}
	parentReqs := filepath.Join(parentProjectDir, "requirements.txt")
	if err := os.WriteFile(parentReqs, []byte("flask==2.0.0\n"), 0o644); err != nil {
		t.Fatalf("Failed to create parent requirements.txt: %v", err)
	}

	// Create workspace directory
	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("Failed to create workspace dir: %v", err)
	}

	// Create azure.yaml in workspace
	azureYamlPath := filepath.Join(workspaceDir, "azure.yaml")
	azureYamlContent := `name: test-app
services:
  api:
    project: ./api
    language: python
    host: containerapp
`
	if err := os.WriteFile(azureYamlPath, []byte(azureYamlContent), 0o644); err != nil {
		t.Fatalf("Failed to create azure.yaml: %v", err)
	}

	// Create api directory inside workspace
	apiDir := filepath.Join(workspaceDir, "api")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		t.Fatalf("Failed to create api dir: %v", err)
	}
	apiReqs := filepath.Join(apiDir, "requirements.txt")
	if err := os.WriteFile(apiReqs, []byte("fastapi==0.100.0\n"), 0o644); err != nil {
		t.Fatalf("Failed to create api requirements.txt: %v", err)
	}

	// Test: Search from workspace directory (where azure.yaml is located)
	projects, err := FindPythonProjects(workspaceDir)
	if err != nil {
		t.Fatalf("FindPythonProjects failed: %v", err)
	}

	// Verify: Only api project should be found, not parent project
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, found %d", len(projects))
		for i, p := range projects {
			t.Logf("Project %d: %s", i, p.Dir)
		}
	}

	if len(projects) > 0 && projects[0].Dir != apiDir {
		t.Errorf("Expected api dir %s, got %s", apiDir, projects[0].Dir)
	}

	// Verify: Parent project should NOT be in the results
	for _, p := range projects {
		if p.Dir == parentProjectDir {
			t.Errorf("Parent project should not be found when searching from workspace directory")
		}
	}
}

// TestFindDotnetProjectsRespectsBoundary tests that FindDotnetProjects doesn't traverse outside the root directory.
func TestFindDotnetProjectsRespectsBoundary(t *testing.T) {
	// Create a temporary directory structure:
	// /tmp/test-root/
	//   ├── parent-project/
	//   │   └── Parent.csproj (should NOT be found)
	//   └── workspace/
	//       ├── azure.yaml
	//       └── api/
	//           └── Api.csproj (should be found)

	tmpDir := t.TempDir()

	// Create parent project (outside the workspace)
	parentProjectDir := filepath.Join(tmpDir, "parent-project")
	if err := os.MkdirAll(parentProjectDir, 0o755); err != nil {
		t.Fatalf("Failed to create parent project dir: %v", err)
	}
	parentCsproj := filepath.Join(parentProjectDir, "Parent.csproj")
	if err := os.WriteFile(parentCsproj, []byte(`<Project Sdk="Microsoft.NET.Sdk"></Project>`), 0o644); err != nil {
		t.Fatalf("Failed to create parent csproj: %v", err)
	}

	// Create workspace directory
	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("Failed to create workspace dir: %v", err)
	}

	// Create azure.yaml in workspace
	azureYamlPath := filepath.Join(workspaceDir, "azure.yaml")
	azureYamlContent := `name: test-app
services:
  api:
    project: ./api
    language: dotnet
    host: containerapp
`
	if err := os.WriteFile(azureYamlPath, []byte(azureYamlContent), 0o644); err != nil {
		t.Fatalf("Failed to create azure.yaml: %v", err)
	}

	// Create api directory inside workspace
	apiDir := filepath.Join(workspaceDir, "api")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		t.Fatalf("Failed to create api dir: %v", err)
	}
	apiCsproj := filepath.Join(apiDir, "Api.csproj")
	if err := os.WriteFile(apiCsproj, []byte(`<Project Sdk="Microsoft.NET.Sdk.Web"></Project>`), 0o644); err != nil {
		t.Fatalf("Failed to create api csproj: %v", err)
	}

	// Test: Search from workspace directory (where azure.yaml is located)
	projects, err := FindDotnetProjects(workspaceDir)
	if err != nil {
		t.Fatalf("FindDotnetProjects failed: %v", err)
	}

	// Verify: Only api project should be found, not parent project
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, found %d", len(projects))
		for i, p := range projects {
			t.Logf("Project %d: %s", i, p.Path)
		}
	}

	if len(projects) > 0 && projects[0].Path != apiCsproj {
		t.Errorf("Expected api csproj %s, got %s", apiCsproj, projects[0].Path)
	}

	// Verify: Parent project should NOT be in the results
	for _, p := range projects {
		if p.Path == parentCsproj {
			t.Errorf("Parent project should not be found when searching from workspace directory")
		}
	}
}

// TestFindAppHostRespectsBoundary tests that FindAppHost doesn't traverse outside the root directory.
func TestFindAppHostRespectsBoundary(t *testing.T) {
	// Create a temporary directory structure:
	// /tmp/test-root/
	//   ├── parent-apphost/
	//   │   ├── AppHost.csproj
	//   │   └── Program.cs (should NOT be found)
	//   └── workspace/
	//       ├── azure.yaml
	//       └── AppHost/
	//           ├── AppHost.csproj
	//           └── Program.cs (should be found)

	tmpDir := t.TempDir()

	// Create parent AppHost project (outside the workspace)
	parentAppHostDir := filepath.Join(tmpDir, "parent-apphost")
	if err := os.MkdirAll(parentAppHostDir, 0o755); err != nil {
		t.Fatalf("Failed to create parent apphost dir: %v", err)
	}
	parentCsproj := filepath.Join(parentAppHostDir, "AppHost.csproj")
	if err := os.WriteFile(parentCsproj, []byte(`<Project Sdk="Microsoft.NET.Sdk"></Project>`), 0o644); err != nil {
		t.Fatalf("Failed to create parent csproj: %v", err)
	}
	parentProgram := filepath.Join(parentAppHostDir, "Program.cs")
	if err := os.WriteFile(parentProgram, []byte(`// Parent Program.cs`), 0o644); err != nil {
		t.Fatalf("Failed to create parent Program.cs: %v", err)
	}

	// Create workspace directory
	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("Failed to create workspace dir: %v", err)
	}

	// Create azure.yaml in workspace
	azureYamlPath := filepath.Join(workspaceDir, "azure.yaml")
	azureYamlContent := `name: test-app
services:
  api:
    project: ./api
    language: dotnet
    host: containerapp
`
	if err := os.WriteFile(azureYamlPath, []byte(azureYamlContent), 0o644); err != nil {
		t.Fatalf("Failed to create azure.yaml: %v", err)
	}

	// Create AppHost directory inside workspace
	appHostDir := filepath.Join(workspaceDir, "AppHost")
	if err := os.MkdirAll(appHostDir, 0o755); err != nil {
		t.Fatalf("Failed to create apphost dir: %v", err)
	}
	appHostCsproj := filepath.Join(appHostDir, "AppHost.csproj")
	if err := os.WriteFile(appHostCsproj, []byte(`<Project Sdk="Microsoft.NET.Sdk"></Project>`), 0o644); err != nil {
		t.Fatalf("Failed to create apphost csproj: %v", err)
	}
	appHostProgram := filepath.Join(appHostDir, "Program.cs")
	if err := os.WriteFile(appHostProgram, []byte(`// AppHost Program.cs`), 0o644); err != nil {
		t.Fatalf("Failed to create apphost Program.cs: %v", err)
	}

	// Test: Search from workspace directory (where azure.yaml is located)
	project, err := FindAppHost(workspaceDir)
	if err != nil {
		t.Fatalf("FindAppHost failed: %v", err)
	}

	// Verify: Only workspace AppHost should be found
	if project == nil {
		t.Fatal("Expected to find AppHost in workspace")
	}

	if project.Dir != appHostDir {
		t.Errorf("Expected AppHost dir %s, got %s", appHostDir, project.Dir)
	}

	if project.ProjectFile != appHostCsproj {
		t.Errorf("Expected AppHost csproj %s, got %s", appHostCsproj, project.ProjectFile)
	}

	// Verify: Parent AppHost should NOT be found
	if project.Dir == parentAppHostDir {
		t.Errorf("Parent AppHost should not be found when searching from workspace directory")
	}
}
