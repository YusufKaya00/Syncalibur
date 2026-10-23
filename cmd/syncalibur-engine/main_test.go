package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCalculateShannonEntropy(t *testing.T) {
	// Normal low-entropy text
	normalText := "hello world this is a normal test string"
	entNormal := calculateShannonEntropy(normalText)
	if entNormal > 4.2 {
		t.Errorf("expected normal text entropy to be < 4.2, got %f", entNormal)
	}

	// High entropy random hash / secret
	randomSecret := "aZ9#kL2$vM8@xP4%qR7^wT1&yU5*"
	entSecret := calculateShannonEntropy(randomSecret)
	if entSecret < 4.4 {
		t.Errorf("expected random secret entropy to be >= 4.4, got %f", entSecret)
	}
}

func TestDetectSecretsInFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Clean file
	cleanFile := filepath.Join(tmpDir, "clean.go")
	_ = os.WriteFile(cleanFile, []byte("package main\n\nfunc main() {\n\tprintln(\"clean\")\n}\n"), 0644)
	if hasSecret, reason := detectSecretsInFile(cleanFile); hasSecret {
		t.Errorf("clean file flagged as secret: %s", reason)
	}

	// 2. File with AWS Key
	awsFile := filepath.Join(tmpDir, "aws.txt")
	_ = os.WriteFile(awsFile, []byte("AWS_KEY=AKIA1234567890ABCDEF\n"), 0644)
	if hasSecret, _ := detectSecretsInFile(awsFile); !hasSecret {
		t.Errorf("expected AWS key to be detected")
	}

	// 3. File with GitHub PAT
	ghFile := filepath.Join(tmpDir, "gh.txt")
	_ = os.WriteFile(ghFile, []byte("TOKEN=ghp_123456789012345678901234567890123456\n"), 0644)
	if hasSecret, _ := detectSecretsInFile(ghFile); !hasSecret {
		t.Errorf("expected GitHub PAT to be detected")
	}

	// 4. File with Private Key Header
	keyFile := filepath.Join(tmpDir, "priv.key")
	_ = os.WriteFile(keyFile, []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----\n"), 0644)
	if hasSecret, _ := detectSecretsInFile(keyFile); !hasSecret {
		t.Errorf("expected Private Key header to be detected")
	}

	// 5. Benign File with UUID (should NOT trigger secret quarantine)
	uuidFile := filepath.Join(tmpDir, "uuid.txt")
	_ = os.WriteFile(uuidFile, []byte("SESSION_ID=c82c1970-4ff7-461a-80d0-e958b320a7ec\n"), 0644)
	if hasSecret, reason := detectSecretsInFile(uuidFile); hasSecret {
		t.Errorf("expected UUID to be ignored, got false alarm: %s", reason)
	}

	// 6. Benign File with base64 data URI (should NOT trigger secret quarantine)
	dataUriFile := filepath.Join(tmpDir, "asset.js")
	_ = os.WriteFile(dataUriFile, []byte("const icon = \"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==\";\n"), 0644)
	if hasSecret, reason := detectSecretsInFile(dataUriFile); hasSecret {
		t.Errorf("expected data URI to be ignored, got false alarm: %s", reason)
	}

	// 7. File with 'test' in name that contains a secret must NOT be bypassed
	testCredFile := filepath.Join(tmpDir, "latest_creds.py")
	_ = os.WriteFile(testCredFile, []byte("AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\n"), 0644)
	if hasSecret, _ := detectSecretsInFile(testCredFile); !hasSecret {
		t.Errorf("expected latest_creds.py with secret to be detected, but it bypassed scanning!")
	}

	// 8. Test .env file must be quarantined even if named test
	testEnvFile := filepath.Join(tmpDir, ".env.test")
	_ = os.WriteFile(testEnvFile, []byte("DB_PASSWORD=supersecret\n"), 0644)
	if hasSecret, _ := detectSecretsInFile(testEnvFile); !hasSecret {
		t.Errorf("expected .env.test to match sensitive blacklist")
	}
}

func TestValidateFileSyntax(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Valid Go file
	validGo := filepath.Join(tmpDir, "valid.go")
	_ = os.WriteFile(validGo, []byte("package test\nfunc Add(a, b int) int { return a + b }\n"), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "valid.go"); !valid {
		t.Errorf("expected valid Go file to pass syntax check, got: %s", errReason)
	}

	// 2. Broken Go file (unclosed brace)
	brokenGo := filepath.Join(tmpDir, "broken.go")
	_ = os.WriteFile(brokenGo, []byte("package test\nfunc Broken() {\n"), 0644)
	if valid, _ := validateFileSyntax(tmpDir, "broken.go"); valid {
		t.Errorf("expected broken Go file (missing brace) to fail syntax check")
	}

	// 3. Valid JSON
	validJSON := filepath.Join(tmpDir, "valid.json")
	_ = os.WriteFile(validJSON, []byte(`{"name": "syncalibur", "active": true}`), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "valid.json"); !valid {
		t.Errorf("expected valid JSON to pass, got: %s", errReason)
	}

	// 4. Broken JSON
	brokenJSON := filepath.Join(tmpDir, "broken.json")
	_ = os.WriteFile(brokenJSON, []byte(`{"name": "syncalibur", broken`), 0644)
	if valid, _ := validateFileSyntax(tmpDir, "broken.json"); valid {
		t.Errorf("expected broken JSON to fail syntax check")
	}

	// 5. Valid Bash script with case statement and hash comments
	validSh := filepath.Join(tmpDir, "valid.sh")
	_ = os.WriteFile(validSh, []byte("#!/bin/bash\n# comment with (unclosed paren\ncase \"$1\" in\n  start) echo ok ;;\n  stop)  echo bye ;;\nesac\n"), 0755)
	if valid, errReason := validateFileSyntax(tmpDir, "valid.sh"); !valid {
		t.Errorf("expected valid Bash script with case and comment to pass, got: %s", errReason)
	}

	// 6. Broken Bash script (unclosed if statement)
	brokenSh := filepath.Join(tmpDir, "broken.sh")
	_ = os.WriteFile(brokenSh, []byte("#!/bin/bash\nif [ 1 -eq 1 ]; then echo ok\n"), 0755)
	if valid, _ := validateFileSyntax(tmpDir, "broken.sh"); valid {
		t.Errorf("expected broken Bash script to fail syntax check")
	}

	// 7. Valid Java file
	validJava := filepath.Join(tmpDir, "App.java")
	_ = os.WriteFile(validJava, []byte("public class App { public static void main(String[] args) { System.out.println(\"ok\"); } }"), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "App.java"); !valid {
		t.Errorf("expected valid Java file to pass, got: %s", errReason)
	}

	// 8. Broken Java file (missing closing brace)
	brokenJava := filepath.Join(tmpDir, "BrokenApp.java")
	_ = os.WriteFile(brokenJava, []byte("public class BrokenApp { public static void main(String[] args) {"), 0644)
	if valid, _ := validateFileSyntax(tmpDir, "BrokenApp.java"); valid {
		t.Errorf("expected broken Java file to fail syntax check")
	}

	// 9. Valid React TSX
	validTSX := filepath.Join(tmpDir, "Component.tsx")
	_ = os.WriteFile(validTSX, []byte("export const Header = () => { return <div><h1>Title</h1></div>; };"), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "Component.tsx"); !valid {
		t.Errorf("expected valid React TSX to pass, got: %s", errReason)
	}

	// 9b. Valid TypeScript with async/await and missing ambient types
	validTS := filepath.Join(tmpDir, "async_service.ts")
	_ = os.WriteFile(validTS, []byte("export async function loadData(): Promise<any> { return { ok: true }; }"), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "async_service.ts"); !valid {
		t.Errorf("expected valid async TypeScript to pass, got: %s", errReason)
	}

	// 10. Broken React TSX (unclosed JSX expression brace)
	brokenTSX := filepath.Join(tmpDir, "BrokenComponent.tsx")
	_ = os.WriteFile(brokenTSX, []byte("export const Header = () => { return <div>{items.map(i => <span>{i}</span>); };"), 0644)
	if valid, _ := validateFileSyntax(tmpDir, "BrokenComponent.tsx"); valid {
		t.Errorf("expected broken React TSX to fail syntax check")
	}

	// 11. Valid Python with hash comment and multiline docstring
	validPy := filepath.Join(tmpDir, "script.py")
	pyContent := `# Python script with (unmatched parenthesis in comment
"""
Docstring with [unmatched brackets] and (parentheses)
"""
def hello():
    return "world"
`
	_ = os.WriteFile(validPy, []byte(pyContent), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "script.py"); !valid {
		t.Errorf("expected valid Python with docstring and comments to pass, got: %s", errReason)
	}

	// 12. Broken Python (unclosed bracket in code)
	brokenPy := filepath.Join(tmpDir, "broken.py")
	_ = os.WriteFile(brokenPy, []byte("def broken():\n    return [1, 2, 3\n"), 0644)
	if valid, _ := validateFileSyntax(tmpDir, "broken.py"); valid {
		t.Errorf("expected broken Python to fail syntax check")
	}

	// 13. Valid Helm ConfigMap with embedded HTML and CSS containing hex colors (#0f172a)
	helmConfigMap := filepath.Join(tmpDir, "configmap.yaml")
	helmContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-html
data:
  index.html: |
    <!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <title>{{ .Values.appContent.title }}</title>
      <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #0f172a; color: #f8fafc; }
        .card { background: #1e293b; padding: 2.5rem; border-radius: 1rem; box-shadow: 0 10px 25px rgba(0,0,0,0.5); text-align: center; max-width: 600px; border: 1px solid #334155; }
        h1 { color: #38bdf8; margin-bottom: 1rem; }
        p { color: #94a3b8; font-size: 1.1rem; line-height: 1.6; }
        .badge { display: inline-block; padding: 0.25rem 0.75rem; background: #0284c7; color: white; border-radius: 9999px; font-size: 0.875rem; font-weight: 600; margin-top: 1rem; }
      </style>
    </head>
    <body>
      <div class="card">
        <h1>{{ .Values.appContent.heading }}</h1>
        <p>{{ .Values.appContent.message }}</p>
        <div class="badge">Running on Multi-Node K3s (Azure)</div>
      </div>
    </body>
    </html>
`
	_ = os.WriteFile(helmConfigMap, []byte(helmContent), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "configmap.yaml"); !valid {
		t.Errorf("expected Helm ConfigMap with embedded CSS to pass syntax check, got: %s", errReason)
	}

	// 14. Broken Helm YAML (unclosed template tag)
	brokenHelm := filepath.Join(tmpDir, "broken_helm.yaml")
	_ = os.WriteFile(brokenHelm, []byte("metadata:\n  name: {{ .Release.Name\n"), 0644)
	if valid, _ := validateFileSyntax(tmpDir, "broken_helm.yaml"); valid {
		t.Errorf("expected broken Helm template (unclosed '{{') to fail syntax check")
	}

	// 15. Broken YAML (forbidden tab indentation)
	tabYAML := filepath.Join(tmpDir, "tab.yaml")
	_ = os.WriteFile(tabYAML, []byte("key:\n\tvalue: 1\n"), 0644)
	if valid, _ := validateFileSyntax(tmpDir, "tab.yaml"); valid {
		t.Errorf("expected YAML with tabs to fail syntax check")
	}

	// 16. Valid Terraform HCL (.tf)
	validTF := filepath.Join(tmpDir, "main.tf")
	tfContent := `resource "azurerm_resource_group" "rg" {
  name     = "rg-test"
  location = "West Europe"
  tags = {
    environment = "dev"
  }
}
`
	_ = os.WriteFile(validTF, []byte(tfContent), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "main.tf"); !valid {
		t.Errorf("expected valid Terraform HCL to pass, got: %s", errReason)
	}

	// 17. Broken Terraform HCL (unclosed brace)
	brokenTF := filepath.Join(tmpDir, "broken.tf")
	_ = os.WriteFile(brokenTF, []byte("resource \"azurerm_resource_group\" \"rg\" {\n  name = \"rg-test\"\n"), 0644)
	if valid, _ := validateFileSyntax(tmpDir, "broken.tf"); valid {
		t.Errorf("expected broken Terraform HCL to fail syntax check")
	}

	// 18. Valid CSS with hex colors and ID selectors
	validCSS := filepath.Join(tmpDir, "style.css")
	cssContent := `
#header {
  background-color: #0f172a;
  color: #fff;
  border: 1px solid #334155;
}
.card {
  padding: 10px;
}
`
	_ = os.WriteFile(validCSS, []byte(cssContent), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "style.css"); !valid {
		t.Errorf("expected valid CSS with hex colors to pass, got: %s", errReason)
	}

	// 19. Broken CSS (missing closing brace)
	brokenCSS := filepath.Join(tmpDir, "broken.css")
	_ = os.WriteFile(brokenCSS, []byte(".card { color: #fff;\n"), 0644)
	if valid, _ := validateFileSyntax(tmpDir, "broken.css"); valid {
		t.Errorf("expected broken CSS to fail syntax check")
	}

	// 20. Valid JSONC (JSON with comments)
	validJSONC := filepath.Join(tmpDir, "tsconfig.json")
	jsoncContent := `{
  // TypeScript configuration
  "compilerOptions": {
    "target": "ES2022", /* Modern target */
    "module": "commonjs"
  }
}`
	_ = os.WriteFile(validJSONC, []byte(jsoncContent), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "tsconfig.json"); !valid {
		t.Errorf("expected valid JSONC to pass, got: %s", errReason)
	}

	// 21. Valid Dockerfile
	validDocker := filepath.Join(tmpDir, "Dockerfile")
	dockerContent := "FROM alpine:3.19\nRUN apk add --no-cache curl\nCMD [\"sh\"]\n"
	_ = os.WriteFile(validDocker, []byte(dockerContent), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "Dockerfile"); !valid {
		t.Errorf("expected valid Dockerfile to pass, got: %s", errReason)
	}

	// 22. Valid SQL (.sql)
	validSQL := filepath.Join(tmpDir, "query.sql")
	sqlContent := "-- Select users with permissions\nSELECT id, name FROM users WHERE role IN ('admin', 'editor');\n"
	_ = os.WriteFile(validSQL, []byte(sqlContent), 0644)
	if valid, errReason := validateFileSyntax(tmpDir, "query.sql"); !valid {
		t.Errorf("expected valid SQL to pass, got: %s", errReason)
	}
}

func TestInspectGoAST(t *testing.T) {
	tmpDir := t.TempDir()
	sampleGo := filepath.Join(tmpDir, "sample.go")
	src := `package sample

type AuthEngine struct{}

func (a *AuthEngine) ValidateToken(t string) bool {
	return t != ""
}

func HelperFunc() int {
	return 42
}
`
	_ = os.WriteFile(sampleGo, []byte(src), 0644)

	// Check line 6 (inside ValidateToken method)
	res := inspectGoAST(sampleGo, 6, 6)
	if res == nil {
		t.Fatalf("expected AST inspection to find symbol at line 6")
	}
	if res.Kind != "method" || res.Symbol != "(*AuthEngine).ValidateToken" {
		t.Errorf("expected (*AuthEngine).ValidateToken, got kind=%s, symbol=%s", res.Kind, res.Symbol)
	}

	// Check line 10 (inside HelperFunc)
	resFunc := inspectGoAST(sampleGo, 10, 10)
	if resFunc == nil {
		t.Fatalf("expected AST inspection to find symbol at line 10")
	}
	if resFunc.Kind != "func" || resFunc.Symbol != "HelperFunc()" {
		t.Errorf("expected HelperFunc(), got kind=%s, symbol=%s", resFunc.Kind, resFunc.Symbol)
	}
}

func TestParseConventionalCommit(t *testing.T) {
	cases := []struct {
		input       string
		wantType    string
		wantScope   string
		wantBreak   bool
		wantDesc    string
	}{
		{
			input:     "feat(engine): add multi-language pre-flight AST gates",
			wantType:  "feat",
			wantScope: "engine",
			wantBreak: false,
			wantDesc:  "add multi-language pre-flight AST gates",
		},
		{
			input:     "fix(yaml): resolve CSS hex color inside ConfigMap block scalar",
			wantType:  "fix",
			wantScope: "yaml",
			wantBreak: false,
			wantDesc:  "resolve CSS hex color inside ConfigMap block scalar",
		},
		{
			input:     "feat(protocol)!: redesign synchronization wire protocol",
			wantType:  "feat",
			wantScope: "protocol",
			wantBreak: true,
			wantDesc:  "redesign synchronization wire protocol",
		},
		{
			input:     "perf(tui): optimize terminal winsize ioctl",
			wantType:  "perf",
			wantScope: "tui",
			wantBreak: false,
			wantDesc:  "optimize terminal winsize ioctl",
		},
		{
			input:     "docs: update README with release workflow",
			wantType:  "docs",
			wantScope: "",
			wantBreak: false,
			wantDesc:  "update README with release workflow",
		},
	}

	for _, c := range cases {
		cType, cScope, isBreak, cDesc := parseConventionalCommit(c.input)
		if cType != c.wantType || cScope != c.wantScope || isBreak != c.wantBreak || cDesc != c.wantDesc {
			t.Errorf("parseConventionalCommit(%q) = (%q, %q, %v, %q); want (%q, %q, %v, %q)",
				c.input, cType, cScope, isBreak, cDesc, c.wantType, c.wantScope, c.wantBreak, c.wantDesc)
		}
	}
}

func TestDetermineSemVerBump(t *testing.T) {
	// Case 1: Major bump due to breaking change
	majorCommits := []CommitInfo{
		{Type: "fix", Subject: "fix small bug"},
		{Type: "feat", Subject: "feat(wire)!: break backward compatibility", IsBreaking: true},
	}
	if bump := determineSemVerBump(majorCommits); bump != BumpMajor {
		t.Errorf("expected BumpMajor, got %v", bump)
	}

	// Case 2: Minor bump due to feat
	minorCommits := []CommitInfo{
		{Type: "fix", Subject: "fix small bug"},
		{Type: "feat", Subject: "feat(ui): add emblem cockpit"},
	}
	if bump := determineSemVerBump(minorCommits); bump != BumpMinor {
		t.Errorf("expected BumpMinor, got %v", bump)
	}

	// Case 3: Patch bump due to fixes only
	patchCommits := []CommitInfo{
		{Type: "fix", Subject: "fix small bug"},
		{Type: "perf", Subject: "perf: improve speed"},
	}
	if bump := determineSemVerBump(patchCommits); bump != BumpPatch {
		t.Errorf("expected BumpPatch, got %v", bump)
	}
}

func TestCalculateNextVersion(t *testing.T) {
	cases := []struct {
		currentTag string
		bump       SemVerBump
		want       string
	}{
		{"", BumpMajor, "v2.0.0"},
		{"v1.0.0", BumpMajor, "v2.0.0"},
		{"v2.0.0", BumpMinor, "v2.1.0"},
		{"v2.1.0", BumpPatch, "v2.1.1"},
		{"v2.1.1", BumpMajor, "v3.0.0"},
	}

	for _, c := range cases {
		got := calculateNextVersion(c.currentTag, c.bump)
		if got != c.want {
			t.Errorf("calculateNextVersion(%q, %v) = %q; want %q", c.currentTag, c.bump, got, c.want)
		}
	}
}

func TestGenerateChangelog(t *testing.T) {
	commits := []CommitInfo{
		{Hash: "d65f710", Type: "feat", Scope: "ui", Subject: "feat(ui): add emblem cockpit", Description: "add emblem cockpit"},
		{Hash: "ea4cb4e", Type: "fix", Scope: "yaml", Subject: "fix(yaml): resolve CSS hex color inside block scalar", Description: "resolve CSS hex color inside block scalar"},
		{Hash: "846e690", Type: "perf", Scope: "tui", Subject: "perf(tui): direct TIOCGWINSZ ioctl", Description: "direct TIOCGWINSZ ioctl"},
	}

	changelog := generateChangelog("v2.0.0", "2026-10-20", commits)
	if !strings.Contains(changelog, "## [v2.0.0] - 2026-10-20") {
		t.Errorf("expected version header in changelog, got:\n%s", changelog)
	}
	if !strings.Contains(changelog, "### 🚀 Features") {
		t.Errorf("expected Features section in changelog")
	}
	if !strings.Contains(changelog, "### 🐛 Bug Fixes") {
		t.Errorf("expected Bug Fixes section in changelog")
	}
	if !strings.Contains(changelog, "d65f710") {
		t.Errorf("expected commit hash d65f710 in changelog")
	}
}

func TestLoadLogoLines(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("failed to get repo root: %v", err)
	}
	logoPath := filepath.Join(repoRoot, "assets", "syncalibur-logo.ansi")
	lines := loadLogoLines(logoPath)
	if len(lines) != 21 {
		t.Errorf("expected 21 logo lines from file, got %d", len(lines))
	}

	// Test fallback when running in external repository where assets does not exist
	fallbackLines := loadLogoLines("/nonexistent/path/syncalibur-logo.ansi")
	if len(fallbackLines) != 21 {
		t.Errorf("expected 21 fallback logo lines, got %d", len(fallbackLines))
	}
	hasTrueColor := false
	for _, l := range fallbackLines {
		if strings.Contains(l, "\033[38;2;") {
			hasTrueColor = true
			break
		}
	}
	if !hasTrueColor {
		t.Errorf("expected fallback logo to contain 24-bit TrueColor ANSI escape sequences")
	}
}

func TestDetectSecretsShellExpansion(t *testing.T) {
	tmpDir := t.TempDir()

	// File with shell parameter expansion and model names (should NOT be flagged)
	scriptFile := filepath.Join(tmpDir, "script.sh")
	content := `#!/usr/bin/env bash
OLLAMA_MODEL="${SYNCALIBUR_OLLAMA_MODEL:-qwen2.5-coder:1.5b}"
# Comments with long strings should not trigger entropy
`
	_ = os.WriteFile(scriptFile, []byte(content), 0644)
	if hasSecret, reason := detectSecretsInFile(scriptFile); hasSecret {
		t.Errorf("expected shell parameter expansion to NOT be flagged as secret, got: %s", reason)
	}
}

func TestPlanSmartStaging(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	_ = exec.Command("git", "-C", tmpDir, "init").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Initial commit
	initFile := filepath.Join(tmpDir, "README.md")
	_ = os.WriteFile(initFile, []byte("# Initial"), 0644)
	_ = exec.Command("git", "-C", tmpDir, "add", ".").Run()
	_ = exec.Command("git", "-C", tmpDir, "commit", "-m", "chore: init").Run()

	// Create multiple changes across scopes: docs, ci, code, and test
	_ = os.WriteFile(initFile, []byte("# Updated README"), 0644) // docs
	_ = os.MkdirAll(filepath.Join(tmpDir, ".github", "workflows"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, ".github", "workflows", "ci.yml"), []byte("name: CI"), 0644) // ci
	_ = os.MkdirAll(filepath.Join(tmpDir, "cmd", "app"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "cmd", "app", "main.go"), []byte("package main"), 0644) // code:app
	_ = os.WriteFile(filepath.Join(tmpDir, "cmd", "app", "main_test.go"), []byte("package main"), 0644) // test

	groups := planSmartStaging(tmpDir)
	if len(groups) < 3 {
		t.Errorf("expected at least 3 partition groups (docs, ci, code/test), got %d", len(groups))
	}

	foundDocs := false
	foundCI := false
	foundCode := false
	for _, g := range groups {
		if g.Scope == "docs" {
			foundDocs = true
		}
		if g.Scope == "ci" {
			foundCI = true
		}
		if g.Scope == "app" {
			foundCode = true
		}
	}

	if !foundDocs || !foundCI || !foundCode {
		t.Errorf("expected groups to contain docs, ci, and app, got: %+v", groups)
	}
}

func TestParseFleetEntry(t *testing.T) {
	// 1. Simple path
	e1 := parseFleetEntry("/home/user/project")
	if e1 == nil || e1.Path != "/home/user/project" || e1.Mode != "" {
		t.Errorf("expected path /home/user/project, got %+v", e1)
	}

	// 2. Bracketed attributes with canonicalized mode
	e2 := parseFleetEntry("/home/user/repo [mode=commit-generator] [remote=upstream]")
	if e2 == nil || e2.Mode != "autopilot" || e2.Remote != "upstream" {
		t.Errorf("expected mode=autopilot, remote=upstream, got %+v", e2)
	}

	e2b := parseFleetEntry("/home/user/repo [mode=autopilot] [remote=upstream]")
	if e2b == nil || e2b.Mode != "autopilot" || e2b.Remote != "upstream" {
		t.Errorf("expected mode=autopilot, remote=upstream, got %+v", e2b)
	}

	// 3. Inline key=value
	e3 := parseFleetEntry("/home/user/app mode=manual remote=origin")
	if e3 == nil || e3.Mode != "manual" || e3.Remote != "origin" {
		t.Errorf("expected mode=manual, got %+v", e3)
	}

	// 4. Positional mode
	e4 := parseFleetEntry("/home/user/test autopilot")
	if e4 == nil || e4.Mode != "autopilot" {
		t.Errorf("expected positional mode autopilot, got %+v", e4)
	}

	e4b := parseFleetEntry("/home/user/test manual")
	if e4b == nil || e4b.Mode != "manual" {
		t.Errorf("expected positional mode manual, got %+v", e4b)
	}

	// 5. Comment line
	e5 := parseFleetEntry("# comment line")
	if e5 != nil {
		t.Errorf("expected nil for comment line, got %+v", e5)
	}
}

func TestExtractHostPort(t *testing.T) {
	cases := []struct {
		url          string
		expectedHost string
		expectedPort string
	}{
		{"git@github.com:org/repo.git", "github.com", "22"},
		{"git@gitlab.corp.internal:group/project.git", "gitlab.corp.internal", "22"},
		{"https://github.com/org/repo.git", "github.com", "443"},
		{"https://mygit.local:8443/repo.git", "mygit.local", "8443"},
		{"ssh://git@custom.host:2222/repo.git", "custom.host", "2222"},
	}

	for _, c := range cases {
		host, port := extractHostPort(c.url)
		if host != c.expectedHost || port != c.expectedPort {
			t.Errorf("extractHostPort(%s) = (%s, %s); want (%s, %s)", c.url, host, port, c.expectedHost, c.expectedPort)
		}
	}
}

func TestResolveEffectiveRepoMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Case 1: Default
	m1 := resolveEffectiveRepoMode(tmpDir, "")
	if m1 != "autopilot" {
		t.Errorf("expected default autopilot, got %s", m1)
	}

	// Case 2: Fleet line mode takes precedence and canonicalizes aliases
	m2 := resolveEffectiveRepoMode(tmpDir, "commit-generator")
	if m2 != "autopilot" {
		t.Errorf("expected fleet mode commit-generator to resolve to autopilot, got %s", m2)
	}

	m2b := resolveEffectiveRepoMode(tmpDir, "manual")
	if m2b != "manual" {
		t.Errorf("expected fleet mode manual, got %s", m2b)
	}

	// Case 3: Repo-local .syncalibur file
	cfgPath := filepath.Join(tmpDir, ".syncalibur")
	_ = os.WriteFile(cfgPath, []byte("MODE=\"manual\"\n"), 0644)
	m3 := resolveEffectiveRepoMode(tmpDir, "")
	if m3 != "manual" {
		t.Errorf("expected repo .syncalibur mode manual, got %s", m3)
	}
}

func TestGetCommitVelocity(t *testing.T) {
	tmpDir := t.TempDir()
	_ = exec.Command("git", "-C", tmpDir, "init").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	f := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(f, []byte("hello"), 0644)
	_ = exec.Command("git", "-C", tmpDir, "add", ".").Run()
	_ = exec.Command("git", "-C", tmpDir, "commit", "-m", "feat: first").Run()

	velo := getCommitVelocity(tmpDir)
	if velo.Today != 1 || velo.Total7d != 1 {
		t.Errorf("expected Today=1, Total7d=1, got %+v", velo)
	}
	if !strings.Contains(velo.Sparkline, "1 commits/7d") {
		t.Errorf("expected Sparkline to contain '1 commits/7d', got %s", velo.Sparkline)
	}
}

func TestDevOpsKubernetesPortMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Valid matching Service & Deployment
	validK8s := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  labels:
    app: web
spec:
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:1.25.3-alpine
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: web-service
spec:
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 8080
`
	_ = os.WriteFile(filepath.Join(tmpDir, "valid.yaml"), []byte(validK8s), 0644)
	issuesValid := validateDevOpsGate(tmpDir, []string{"valid.yaml"})
	var portErrorsValid []InfraIssue
	for _, iss := range issuesValid {
		if iss.Code == "K8S_PORT_MISMATCH" {
			portErrorsValid = append(portErrorsValid, iss)
		}
	}
	if len(portErrorsValid) != 0 {
		t.Errorf("expected 0 port mismatch errors for matching Service/Deployment, got: %+v", portErrorsValid)
	}

	// 2. Mismatched targetPort (Service targetPort 9090 vs Container port 8080)
	mismatchedK8s := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  labels:
    app: web
spec:
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:1.25.3-alpine
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: web-service
spec:
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 9090
`
	_ = os.WriteFile(filepath.Join(tmpDir, "bad.yaml"), []byte(mismatchedK8s), 0644)
	issuesBad := validateDevOpsGate(tmpDir, []string{"bad.yaml"})
	var portErrorsBad []InfraIssue
	for _, iss := range issuesBad {
		if iss.Code == "K8S_PORT_MISMATCH" {
			portErrorsBad = append(portErrorsBad, iss)
		}
	}
	if len(portErrorsBad) == 0 {
		t.Errorf("expected port mismatch error for targetPort 9090 vs containerPort 8080, got 0 errors")
	}
}

func TestDevOpsConfigMapSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	badConfigMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  AWS_SECRET_ACCESS_KEY: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  DATABASE_URL: "postgres://user:secret123@db:5432/prod"
`
	_ = os.WriteFile(filepath.Join(tmpDir, "cm.yaml"), []byte(badConfigMap), 0644)
	issues := validateDevOpsGate(tmpDir, []string{"cm.yaml"})
	var cmErrors []InfraIssue
	for _, iss := range issues {
		if iss.Code == "CONFIGMAP_PLAINTEXT_SECRET" {
			cmErrors = append(cmErrors, iss)
		}
	}
	if len(cmErrors) == 0 {
		t.Errorf("expected secret detection error in ConfigMap data, got 0")
	}
}

func TestDevOpsContainerProbes(t *testing.T) {
	tmpDir := t.TempDir()
	badProbe := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-service
spec:
  template:
    spec:
      containers:
      - name: api
        image: myapp:v1.0.0
        ports:
        - containerPort: 8080
        readinessProbe:
          httpGet:
            path: /healthz
            port: 3000
`
	_ = os.WriteFile(filepath.Join(tmpDir, "probe.yaml"), []byte(badProbe), 0644)
	issues := validateDevOpsGate(tmpDir, []string{"probe.yaml"})
	var probeErrors []InfraIssue
	for _, iss := range issues {
		if iss.Code == "PROBE_PORT_MISMATCH" {
			probeErrors = append(probeErrors, iss)
		}
	}
	if len(probeErrors) == 0 {
		t.Errorf("expected readinessProbe port mismatch error for port 3000, got 0")
	}
}

func TestDevOpsDockerfileLint(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Dockerfile with :latest and no USER directive
	badDocker := filepath.Join(tmpDir, "Dockerfile")
	badData := []byte("FROM golang:latest\nCOPY . .\nCMD [\"./app\"]\n")
	_ = os.WriteFile(badDocker, badData, 0644)
	errs := lintDockerfile(badData, "Dockerfile", tmpDir)
	if len(errs) == 0 {
		t.Errorf("expected lintDockerfile to flag ':latest' and missing USER non-root, got 0 errors")
	}

	// 2. Multi-stage Dockerfile with non-root USER and pinned tag
	goodDocker := filepath.Join(tmpDir, "Dockerfile.good")
	goodContent := []byte(`FROM golang:1.22.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/server .

FROM alpine:3.19.1
RUN adduser -D -u 10001 appuser
USER 10001
WORKDIR /app
COPY --from=builder /app/server /app/server
ENTRYPOINT ["/app/server"]
`)
	_ = os.WriteFile(goodDocker, goodContent, 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".dockerignore"), []byte(".git\n*.env\n"), 0644)
	goodErrs := lintDockerfile(goodContent, "Dockerfile.good", tmpDir)
	if len(goodErrs) != 0 {
		t.Errorf("expected clean multi-stage Dockerfile to pass, got: %+v", goodErrs)
	}
}

func TestDevOpsTerraformHygiene(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Missing .gitignore with .tf file present
	tfFile := filepath.Join(tmpDir, "main.tf")
	_ = os.WriteFile(tfFile, []byte("resource \"aws_s3_bucket\" \"b\" {}\n"), 0644)
	errs := lintTerraformHygiene(tmpDir, []string{"main.tf"})
	if len(errs) == 0 {
		t.Errorf("expected missing .gitignore error when .tf files exist, got 0")
	}

	// 2. Plaintext credentials in terraform.tfvars
	_ = os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(".terraform/\n*.tfstate*\n"), 0644)
	tfvarsFile := filepath.Join(tmpDir, "terraform.tfvars")
	_ = os.WriteFile(tfvarsFile, []byte("db_password = \"supersecret123456\"\n"), 0644)
	varErrs := lintTerraformHygiene(tmpDir, []string{"main.tf", "terraform.tfvars"})
	if len(varErrs) == 0 {
		t.Errorf("expected plaintext credential error in terraform.tfvars, got 0")
	}
}

func TestDetectConflictMarkers(t *testing.T) {
	tmpDir := t.TempDir()

	// File with Git conflict markers
	conflictFile := filepath.Join(tmpDir, "merge.go")
	conflictContent := `package main

<<<<<<< HEAD
func Greet() string { return "Hello from main" }
=======
func Greet() string { return "Hello from feature" }
>>>>>>> feature-branch
`
	_ = os.WriteFile(conflictFile, []byte(conflictContent), 0644)
	found, msg := detectConflictMarkers(conflictFile)
	if !found {
		t.Errorf("expected detectConflictMarkers to find conflict, got none")
	}
	if !strings.Contains(msg, "<<<<<<<") || !strings.Contains(msg, ":3:") {
		t.Errorf("expected conflict message to mention line 3 and <<<<<<<, got %s", msg)
	}
}

func TestIsBinaryFile(t *testing.T) {
	tmpDir := t.TempDir()

	// UTF-8 plain text file
	textFile := filepath.Join(tmpDir, "readme.txt")
	_ = os.WriteFile(textFile, []byte("Hello World, this is normal text.\n"), 0644)
	isBin, err := isBinaryFile(textFile)
	if err != nil || isBin {
		t.Errorf("expected plain text file NOT to be flagged as binary (isBin=%v, err=%v)", isBin, err)
	}

	// Binary file with null bytes
	binFile := filepath.Join(tmpDir, "data.bin")
	_ = os.WriteFile(binFile, []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x00, 0x00}, 0644)
	isBin2, err := isBinaryFile(binFile)
	if err != nil || !isBin2 {
		t.Errorf("expected file with null bytes to be flagged as binary (isBin=%v, err=%v)", isBin2, err)
	}
}

func TestModeCycleAndBadge(t *testing.T) {
	// Mode badge checks
	if !strings.Contains(getModeBadge("autopilot"), "Autopilot") {
		t.Errorf("expected getModeBadge('autopilot') to contain 'Autopilot', got %s", getModeBadge("autopilot"))
	}
	if !strings.Contains(getModeBadge("instant-push"), "Autopilot") {
		t.Errorf("expected legacy instant-push to resolve to Autopilot badge, got %s", getModeBadge("instant-push"))
	}
	if !strings.Contains(getModeBadge("manual"), "Manual") {
		t.Errorf("expected getModeBadge('manual') to contain 'Manual', got %s", getModeBadge("manual"))
	}

	// 2-state toggle check (autopilot <-> manual)
	tmpDir := t.TempDir()
	_ = exec.Command("git", "-C", tmpDir, "init").Run()
	_ = os.WriteFile(filepath.Join(tmpDir, ".syncalibur"), []byte("MODE=\"autopilot\"\n"), 0644)

	// autopilot -> manual
	m1 := cycleActiveMode("syncalibur.sh", tmpDir)
	if m1 != "manual" {
		t.Errorf("expected manual after autopilot, got %s", m1)
	}

	// manual -> autopilot
	m2 := cycleActiveMode("syncalibur.sh", tmpDir)
	if m2 != "autopilot" {
		t.Errorf("expected autopilot after manual, got %s", m2)
	}
}

func TestCockpitAnimations(t *testing.T) {
	// 1. EKG Wave
	w1 := getEkgWave(0, 10)
	if len([]rune(w1)) != 10 {
		t.Errorf("expected ekg wave rune length 10, got %d (%s)", len([]rune(w1)), w1)
	}
	w2 := getEkgWave(5, 12)
	if len([]rune(w2)) != 12 {
		t.Errorf("expected ekg wave rune length 12, got %d (%s)", len([]rune(w2)), w2)
	}
	if getEkgWave(1, 0) != "" {
		t.Errorf("expected empty string for 0 width, got %s", getEkgWave(1, 0))
	}

	// 2. Animated Mode Badge
	modes := []string{"instant-push", "devops", "pr", "developer"}
	for _, m := range modes {
		b0 := getAnimatedModeBadge(m, 0)
		b1 := getAnimatedModeBadge(m, 4)
		if b0 == "" || b1 == "" {
			t.Errorf("mode badge for %s must not be empty", m)
		}
	}

	// 3. boostAnsiColors
	testLine := "\033[38;2;100;100;100mX\033[0m"
	boosted := boostAnsiColors(testLine, 1.5, 10, 20, 30)
	if !strings.Contains(boosted, "\033[38;2;160;170;180m") {
		t.Errorf("boostAnsiColors failed, got %s", boosted)
	}

	// 4. Logo Animation Rune Width Invariance
	logo := defaultEmbeddedLogoLines
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	for frame := 0; frame < 30; frame++ {
		anim := applyAnimatedLogo(logo, frame)
		if len(anim) != 21 {
			t.Fatalf("frame %d: expected 21 logo lines, got %d", frame, len(anim))
		}
		for r, line := range anim {
			clean := re.ReplaceAllString(line, "")
			runeCount := len([]rune(clean))
			if runeCount != 44 {
				t.Fatalf("frame %d row %d: expected exactly 44 visible runes, got %d (%s)", frame, r, runeCount, clean)
			}
		}
	}
}

func TestIsTypeScriptSyntaxError(t *testing.T) {
	syntaxErrors := []string{
		"Component.tsx(1,40): error TS1005: ';' expected.",
		"Component.tsx(1,15): error TS1109: Expression expected.",
		"Component.tsx(5,1): error TS1128: Declaration or statement expected.",
	}
	for _, se := range syntaxErrors {
		if !isTypeScriptSyntaxError(se) {
			t.Errorf("expected '%s' to be recognized as TS syntax error", se)
		}
	}

	semanticErrors := []string{
		"Component.tsx(1,38): error TS7026: JSX element implicitly has type 'any' because no interface 'JSX.IntrinsicElements' exists.",
		"Component.tsx(1,20): error TS2304: Cannot find name 'document'.",
		"Component.tsx(1,20): error TS2307: Cannot find module 'react' or its corresponding type declarations.",
		"Component.tsx(1,20): error TS2584: Cannot find name 'console'. Do you need to change your target library?",
		"Component.tsx(1,20): error TS2318: Cannot find global type 'Boolean'.",
	}
	for _, se := range semanticErrors {
		if isTypeScriptSyntaxError(se) {
			t.Errorf("expected semantic error '%s' to NOT be flagged as syntax error", se)
		}
	}
}

func TestGetLatestSemVerTag(t *testing.T) {
	tmpDir := t.TempDir()
	_ = exec.Command("git", "-C", tmpDir, "init").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
	_ = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644)
	_ = exec.Command("git", "-C", tmpDir, "add", ".").Run()
	_ = exec.Command("git", "-C", tmpDir, "commit", "-m", "chore: init").Run()
	_ = exec.Command("git", "-C", tmpDir, "tag", "v2.0.0").Run()
	_ = exec.Command("git", "-C", tmpDir, "tag", "v2.3.0").Run()
	_ = exec.Command("git", "-C", tmpDir, "tag", "v2.1.0").Run()

	latest := getLatestSemVerTag(tmpDir)
	if latest != "v2.3.0" {
		t.Errorf("expected latest tag 'v2.3.0', got '%s'", latest)
	}
}

func TestShieldToggleAndBypass(t *testing.T) {
	tmpDir := t.TempDir()
	_ = exec.Command("git", "-C", tmpDir, "init").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	gitDir := filepath.Join(tmpDir, ".git")

	// 1. By default, shields should be enabled
	if areShieldsDisabled(tmpDir) {
		t.Errorf("expected shields to be enabled by default")
	}

	// 2. Test file with dummy secret
	secretFile := filepath.Join(tmpDir, "credentials.py")
	_ = os.WriteFile(secretFile, []byte("api_key = \"sk-live-dummysecretpassword1234567890abcdef\"\n"), 0644)

	// Preflight should detect secret and exit with code 3
	exitCode, _ := runPreflight(tmpDir, false, false)
	if exitCode != 3 {
		t.Errorf("expected preflight to detect secret and return 3, got %d", exitCode)
	}

	// 3. Disable shields via syncalibur.noshield marker file
	noShieldFile := filepath.Join(gitDir, "syncalibur.noshield")
	_ = os.WriteFile(noShieldFile, []byte("1\n"), 0644)
	if !areShieldsDisabled(tmpDir) {
		t.Errorf("expected shields to be disabled when syncalibur.noshield exists")
	}

	// Preflight should now pass (exit code 0) because shields are disabled
	exitCodeBypass, _ := runPreflight(tmpDir, false, false)
	if exitCodeBypass != 0 {
		t.Errorf("expected preflight to pass (0) when shields are disabled, got %d", exitCodeBypass)
	}

	// 4. Re-enable shields by removing marker file
	_ = os.Remove(noShieldFile)
	if areShieldsDisabled(tmpDir) {
		t.Errorf("expected shields to be re-enabled after removing syncalibur.noshield")
	}

	// 5. Test explicit --no-shield parameter to runPreflight
	exitCodeExplicit, _ := runPreflight(tmpDir, false, false, true)
	if exitCodeExplicit != 0 {
		t.Errorf("expected preflight to pass (0) when explicit noShield is true, got %d", exitCodeExplicit)
	}

	// 6. Test environment variable SYNCALIBUR_SECRET_SHIELD="false"
	os.Setenv("SYNCALIBUR_SECRET_SHIELD", "false")
	defer os.Unsetenv("SYNCALIBUR_SECRET_SHIELD")
	if !areShieldsDisabled(tmpDir) {
		t.Errorf("expected shields to be disabled when SYNCALIBUR_SECRET_SHIELD=false")
	}
}

func TestIsSafePathDeletedFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "syncalibur-safepath-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. Deleted file should be reported safe (not an error)
	safe, reason := isSafePath(tmpDir, "deleted-file.txt")
	if !safe {
		t.Errorf("expected deleted file to be safe, got: %s", reason)
	}

	// 2. Existing regular file
	realFile := filepath.Join(tmpDir, "real.txt")
	_ = os.WriteFile(realFile, []byte("hello"), 0644)
	safeReal, reasonReal := isSafePath(tmpDir, "real.txt")
	if !safeReal {
		t.Errorf("expected existing file to be safe, got: %s", reasonReal)
	}

	// 3. Path traversal outside repo
	safeEscape, _ := isSafePath(tmpDir, "../outside.txt")
	if safeEscape {
		t.Errorf("expected path traversal outside repo to be rejected")
	}
}

func TestIsSafePathSymlinkDirectoryTraversal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "syncalibur-symlink-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outsideDir, err := os.MkdirTemp("", "syncalibur-outside-dir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outsideDir)

	secretFile := filepath.Join(outsideDir, "secret.txt")
	_ = os.WriteFile(secretFile, []byte("super-secret"), 0644)

	// Create directory symlink inside repo pointing outside
	symlinkPath := filepath.Join(tmpDir, "leak_dir")
	if err := os.Symlink(outsideDir, symlinkPath); err != nil {
		t.Skip("symlinks not supported on this environment")
	}

	// isSafePath on leak_dir/secret.txt must be rejected!
	safe, reason := isSafePath(tmpDir, "leak_dir/secret.txt")
	if safe {
		t.Errorf("expected directory symlink traversal to be rejected, but was marked safe")
	} else if !strings.Contains(reason, "traverses outside repository via symlink") {
		t.Errorf("expected traversal error message, got: %s", reason)
	}
}

func TestDetectConflictMarkersMarkdownH1(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "syncalibur-doc-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// CommonMark Setext H1 header format
	mdContent := "# Project Documentation\nSome description here.\n\nSection Title\n=======\nMore text.\n"
	_, _ = tmpFile.WriteString(mdContent)
	_ = tmpFile.Close()

	hasConflict, reason := detectConflictMarkers(tmpFile.Name())
	if hasConflict {
		t.Errorf("CommonMark Setext H1 header was falsely flagged as merge conflict: %s", reason)
	}

	// Now add an actual Git merge conflict block
	conflictContent := "<<<<<<< HEAD\nlocal change\n=======\nremote change\n>>>>>>> feature\n"
	_ = os.WriteFile(tmpFile.Name(), []byte(conflictContent), 0644)
	hasRealConflict, _ := detectConflictMarkers(tmpFile.Name())
	if !hasRealConflict {
		t.Errorf("expected genuine Git conflict to be flagged, but was missed")
	}
}

func TestDetectSecretsUnquotedAndUTF16(t *testing.T) {
	// 1. Unquoted YAML API key
	tmpYaml, err := os.CreateTemp("", "syncalibur-test-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpYaml.Name())

	_, _ = tmpYaml.WriteString("service:\n  api_key: AIzaSyD123456789012345678901234\n")
	_ = tmpYaml.Close()

	hasSecret, reason := detectSecretsInFile(tmpYaml.Name())
	if !hasSecret {
		t.Errorf("expected unquoted YAML secret to be detected, but was missed")
	} else if !strings.Contains(reason, "matches known credential signature") {
		t.Errorf("unexpected reason: %s", reason)
	}

	// 2. UTF-16 LE encoded secret file
	tmpUtf16, err := os.CreateTemp("", "syncalibur-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpUtf16.Name())

	secretPlain := "aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\n"
	u16Runes := utf16.Encode([]rune(secretPlain))
	buf := []byte{0xFF, 0xFE} // UTF-16 LE BOM
	for _, r := range u16Runes {
		buf = append(buf, byte(r), byte(r>>8))
	}
	_ = os.WriteFile(tmpUtf16.Name(), buf, 0644)

	hasU16Secret, u16Reason := detectSecretsInFile(tmpUtf16.Name())
	if !hasU16Secret {
		t.Errorf("expected UTF-16 encoded AWS secret to be detected, but was bypassed")
	} else if !strings.Contains(u16Reason, "matches known credential signature") {
		t.Errorf("unexpected UTF-16 reason: %s", u16Reason)
	}
}

func TestDetectSecretsLargeFileNoBypass(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "syncalibur-large-*.sql")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write 500 lines of SQL and inject a secret at line 250
	for i := 1; i <= 500; i++ {
		if i == 250 {
			_, _ = tmpFile.WriteString("INSERT INTO configs (k, v) VALUES ('github_token', 'ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789');\n")
		} else {
			_, _ = tmpFile.WriteString(fmt.Sprintf("INSERT INTO users (id, name) VALUES (%d, 'user_%d');\n", i, i))
		}
	}
	_ = tmpFile.Close()

	hasSecret, reason := detectSecretsInFile(tmpFile.Name())
	if !hasSecret {
		t.Errorf("expected secret in SQL dump to be detected, but was missed")
	} else if !strings.Contains(reason, "matches known credential signature") {
		t.Errorf("unexpected reason: %s", reason)
	}
}




