package main

import (
	"os"
	"path/filepath"
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

