# Test Projects

This directory contains test projects used to validate the App Extension commands.

## Structure

```
test-projects/
├── node/               # Node.js test projects
│   ├── test-node-project/    (npm with dependencies)
│   └── test-npm-project/     (simple npm project)
├── python/             # Python test projects
│   ├── test-poetry-project/  (poetry)
│   ├── test-python-project/  (pip)
│   └── test-uv-project/      (uv)
├── boundary-test/      # Tests boundary checking (no parent traversal)
│   ├── package.json          (parent - should NOT be found)
│   └── workspace/
│       ├── azure.yaml        (workspace root)
│       ├── web/              (should be found)
│       └── api/              (should be found)
└── azure/              # Azure configuration test files
    ├── azure.yaml
    ├── azure-backup.yaml
    └── azure-fail.yaml
```

## Usage

These projects are used to test:
- `azd app deps` - Installing dependencies across different package managers
- `azd app run` - Running development environments
- Detection logic for package managers (npm, pnpm, pip, poetry, uv)
- **Boundary checking** - Ensuring projects outside `azure.yaml` workspace are not detected

### Boundary Test Project

The `boundary-test/` project specifically tests that the detector functions:
- ✅ Only search within the workspace defined by `azure.yaml` location
- ✅ Do NOT traverse outside the workspace to parent directories
- ❌ Do NOT detect projects in sibling/parent directories

See `boundary-test/README.md` for detailed test instructions.

## Running Tests

From the root directory:
```bash
# Test deps command
azd app deps

# Test run command
azd app run
```
