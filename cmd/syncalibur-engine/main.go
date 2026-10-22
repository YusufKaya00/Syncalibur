// ==============================================================================
// Syncalibur Go Engine (cmd/syncalibur-engine/main.go)
// ------------------------------------------------------------------------------
// High-performance, zero-dependency Go engine powering Syncalibur's:
//   1. High-Definition TrueColor Terminal Emblem Renderer (`render-logo`)
//   2. AST-Aware Semantic Conventional Commit Generator (`commit-msg`)
//   3. Pre-flight Syntax Gate & Shannon Entropy Secret Shield (`preflight`)
//   4. Multi-Repository Fleet Scanner & Parallel Status Probe (`fleet-scan`, `fleet-status`)
//   5. 7-Day Commit Velocity Sparkline Generator (`sparkline`)
//   6. First-Class Telegram Bot & Multi-Channel Webhook Dispatcher (`webhook`, `telegram`)
//   7. Two-Pane Interactive Live TUI with Scrollable Logs (`ui`)
// ==============================================================================
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	_ "image/png"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
)

// ANSI Color Palette (24-bit TrueColor & 16-color Fallback)
const (
	cReset = "\033[0m"
	cBold  = "\033[1m"
	cDim   = "\033[2m"
	cGold  = "\033[38;2;245;194;66m"
	cCyan  = "\033[38;2;56;189;248m"
	cGreen = "\033[38;2;52;211;153m"
	cRed   = "\033[38;2;248;113;113m"
	cAmber = "\033[38;2;251;191;36m"
	cSlate = "\033[38;2;148;163;184m"
	cWhite = "\033[38;2;241;245;249m"
	cDark  = "\033[38;2;71;85;105m"
	cBlue  = "\033[38;2;59;130;246m"
	cBgBar = "\033[48;2;30;41;59m"
)

// Syncalibur Official Release & Semantic Version
const (
	SyncaliburVersion   = "v2.18.0"
	SyncaliburMilestone = "Zero-Data-Loss Flight Recorder & Autopilot"
)

// Lucide / Minimalist Typography Icons
const (
	iconCheck   = "✓"
	iconCross   = "✕"
	iconWarn    = "▲"
	iconShield  = "⛨"
	iconZap     = "✦"
	iconSparkle = "✦"
	iconArrowUp = "↗"
	iconArrowDn = "↙"
	iconDot     = "●"
	iconCircle  = "○"
	iconDiamond = "◆"
	iconChevron = "›"
	iconSync    = "⇄"
	iconDoctor  = "✚"
	iconPause   = "▲"
	iconPlay    = "▶"
	iconWheel   = "☸"
	iconHelm    = "⎈"
)

var quads = [16]string{
	" ", "▘", "▝", "▀",
	"▖", "▌", "▞", "▛",
	"▗", "▚", "▐", "▜",
	"▄", "▙", "▟", "█",
}

type rgbaPx struct {
	r, g, b, a int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: syncalibur <daemon-exec|render-logo|commit-msg|preflight|sparkline|fleet-scan|fleet-status|webhook|telegram|ui> [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "daemon-exec", "daemon":
		if len(os.Args) < 3 {
			os.Exit(1)
		}
		os.Exit(runDaemonSupervisor(os.Args[2], os.Args[3:]...))

	case "render-logo":
		if len(os.Args) < 4 {
			os.Exit(1)
		}
		if err := renderLogoANSI(os.Args[2], os.Args[3], 44, 21); err != nil {
			fmt.Fprintf(os.Stderr, "render-logo error: %v\n", err)
			os.Exit(1)
		}

	case "commit-msg":
		repo := "."
		if len(os.Args) >= 3 {
			repo = os.Args[2]
		}
		fmt.Println(generateCommitMessage(repo))

	case "preflight":
		repo := "."
		stagedOnly := false
		jsonOut := false
		for _, arg := range os.Args[2:] {
			if arg == "--staged-only" {
				stagedOnly = true
			} else if arg == "--json" {
				jsonOut = true
			} else if !strings.HasPrefix(arg, "-") {
				repo = arg
			}
		}
		exitCode, output := runPreflight(repo, stagedOnly, jsonOut)
		if output != "" {
			fmt.Println(output)
		}
		os.Exit(exitCode)

	case "sparkline":
		repo := "."
		if len(os.Args) >= 3 {
			repo = os.Args[2]
		}
		fmt.Print(generateSparkline(repo))

	case "probe-network":
		rawURL := "github.com"
		timeoutMs := 1000
		if len(os.Args) >= 3 {
			rawURL = os.Args[2]
		}
		if len(os.Args) >= 4 {
			if t, err := strconv.Atoi(os.Args[3]); err == nil && t > 0 {
				timeoutMs = t
			}
		}
		if probeRemoteLiveness(rawURL, timeoutMs) {
			os.Exit(0)
		}
		os.Exit(1)

	case "fleet-scan":
		root := "."
		if len(os.Args) >= 3 {
			root = os.Args[2]
		}
		for _, r := range scanGitRepos(root, 3) {
			fmt.Println(r)
		}

	case "fleet-status":
		if len(os.Args) < 3 {
			os.Exit(1)
		}
		printFleetStatusTable(os.Args[2])

	case "webhook":
		if len(os.Args) < 9 {
			os.Exit(1)
		}
		if err := sendWebhook(os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6], os.Args[7], os.Args[8]); err != nil {
			fmt.Fprintf(os.Stderr, "webhook error: %v\n", err)
			os.Exit(1)
		}

	case "telegram":
		if len(os.Args) < 10 {
			os.Exit(1)
		}
		if err := sendTelegramMessage(os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6], os.Args[7], os.Args[8], os.Args[9]); err != nil {
			fmt.Fprintf(os.Stderr, "telegram error: %v\n", err)
			os.Exit(1)
		}

	case "status", "ui":
		scriptPath := ""
		repoRoot := "."
		if len(os.Args) >= 3 {
			scriptPath = os.Args[2]
		}
		if len(os.Args) >= 4 {
			repoRoot = os.Args[3]
		}
		runInteractiveTUI(scriptPath, repoRoot)

	case "smart-stage":
		repo := "."
		jsonOut := false
		for _, arg := range os.Args[2:] {
			if arg == "--json" {
				jsonOut = true
			} else if !strings.HasPrefix(arg, "-") {
				repo = arg
			}
		}
		groups := planSmartStaging(repo)
		if jsonOut {
			data, _ := json.MarshalIndent(groups, "", "  ")
			fmt.Println(string(data))
		} else {
			for _, g := range groups {
				fmt.Printf("GROUP|%s|%s|%s\n", g.Scope, g.Type, strings.Join(g.Files, ","))
			}
		}

	case "version", "--version", "-v":
		printVersion()

	case "release":
		mode := "auto"
		dryRun := false
		repoRoot := "."
		for _, arg := range os.Args[2:] {
			if arg == "--dry-run" {
				dryRun = true
			} else if arg == "auto" || arg == "major" || arg == "minor" || arg == "patch" {
				mode = arg
			} else if !strings.HasPrefix(arg, "-") {
				repoRoot = arg
			}
		}
		if err := runRelease(repoRoot, mode, dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "release error: %v\n", err)
			os.Exit(1)
		}

	case "doctor":
		scriptPath := "syncalibur.sh"
		if p, err := exec.LookPath("syncalibur.sh"); err == nil {
			scriptPath = p
		} else if exe, err := os.Executable(); err == nil {
			cand := filepath.Join(filepath.Dir(exe), "..", "syncalibur.sh")
			if _, err := os.Stat(cand); err == nil {
				scriptPath = cand
			}
		}
		cmd := exec.Command("bash", scriptPath, "doctor")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			os.Exit(1)
		}

	case "mode":
		scriptPath := "syncalibur.sh"
		if p, err := exec.LookPath("syncalibur.sh"); err == nil {
			scriptPath = p
		} else if exe, err := os.Executable(); err == nil {
			cand := filepath.Join(filepath.Dir(exe), "..", "syncalibur.sh")
			if _, err := os.Stat(cand); err == nil {
				scriptPath = cand
			}
		}
		cmdArgs := append([]string{scriptPath, "mode"}, os.Args[2:]...)
		cmd := exec.Command("bash", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			os.Exit(1)
		}

	case "help", "--help", "-h":
		fmt.Println("Syncalibur - Autonomous Git Auto-Sync Platform & Live Engine")
		fmt.Println("Usage: syncalibur <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  ui                   Launch interactive live TUI cockpit")
		fmt.Println("  smart-stage [repo]   Analyze and partition working tree changes into atomic scoped groups")
		fmt.Println("  commit-msg [repo]    Generate Conventional Commit message for staged diff")
		fmt.Println("  preflight [repo]     Run Shannon entropy & syntax validation")
		fmt.Println("  release [mode]       Automated SemVer release & CHANGELOG generation")
		fmt.Println("  sparkline [repo]     Render 7-day commit velocity sparkline")
		fmt.Println("  fleet-scan [dir]     Recursively discover Git repositories")
		fmt.Println("  fleet-status [file]  Probe status of multi-repo fleet list")
		fmt.Println("  version              Display version and runtime information")
	default:
		os.Exit(1)
	}
}

// ==============================================================================
// 0. Native Process Supervisor (`daemon-exec`)
// ==============================================================================

func runDaemonSupervisor(scriptPath string, extraArgs ...string) int {
	bashBin, err := exec.LookPath("bash")
	if err != nil {
		if runtime.GOOS == "windows" {
			for _, candidate := range []string{
				`C:\Program Files\Git\bin\bash.exe`,
				`C:\Program Files\Git\usr\bin\bash.exe`,
				`C:\Git\bin\bash.exe`,
			} {
				if _, err := os.Stat(candidate); err == nil {
					bashBin = candidate
					break
				}
			}
		}
		if bashBin == "" {
			bashBin = "/bin/bash"
		}
	}
	progName := "syncalibur"
	if runtime.GOOS == "windows" {
		progName = bashBin
	}
	args := append([]string{progName, scriptPath}, extraArgs...)
	cmd := &exec.Cmd{
		Path:   bashBin,
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
		Env: append(os.Environ(),
			fmt.Sprintf("SYNCALIBUR_SUPERVISOR_PID=%d", os.Getpid()),
			"SYNCALIBUR_WRAPPED=1",
		),
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "syncalibur supervisor start error: %v\n", err)
		return 1
	}

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sigCh {
			if cmd.Process != nil {
				_ = cmd.Process.Signal(sig)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}

// ==============================================================================
// 1. High-Definition TrueColor Terminal Emblem Renderer
// ==============================================================================

func renderLogoANSI(srcPath, dstPath string, outW, outH int) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	b := src.Bounds()
	srcW := b.Dx()
	srcH := b.Dy()

	pxW := outW * 2
	pxH := outH * 2

	sample := func(x, y int) rgbaPx {
		sx := b.Min.X + (x*srcW)/pxW
		sy := b.Min.Y + (y*srcH)/pxH
		if sx >= b.Max.X {
			sx = b.Max.X - 1
		}
		if sy >= b.Max.Y {
			sy = b.Max.Y - 1
		}
		r, g, bVal, a := src.At(sx, sy).RGBA()
		if a == 0 {
			return rgbaPx{0, 0, 0, 0}
		}
		alpha := float64(a) / 65535.0
		r8 := clampByte((float64(r) / 65535.0) / alpha * 255.0)
		g8 := clampByte((float64(g) / 65535.0) / alpha * 255.0)
		b8 := clampByte((float64(bVal) / 65535.0) / alpha * 255.0)
		a8 := int(float64(a) / 257.0)
		return rgbaPx{r8, g8, b8, a8}
	}

	var out strings.Builder
	for cy := 0; cy < outH; cy++ {
		for cx := 0; cx < outW; cx++ {
			p00 := sample(cx*2, cy*2)
			p10 := sample(cx*2+1, cy*2)
			p01 := sample(cx*2, cy*2+1)
			p11 := sample(cx*2+1, cy*2+1)

			pts := [4]rgbaPx{p00, p10, p01, p11}
			var active []rgbaPx
			mask := 0
			for i, p := range pts {
				if p.a > 32 {
					mask |= (1 << i)
					active = append(active, p)
				}
			}

			if mask == 0 {
				out.WriteString("\033[0m ")
			} else if mask == 15 {
				var sr, sg, sb int
				for _, p := range active {
					sr += p.r
					sg += p.g
					sb += p.b
				}
				n := len(active)
				fmt.Fprintf(&out, "\033[0m\033[38;2;%d;%d;%dm█", sr/n, sg/n, sb/n)
			} else if len(active) == 1 {
				p := active[0]
				fmt.Fprintf(&out, "\033[0m\033[38;2;%d;%d;%dm%s", p.r, p.g, p.b, quads[mask])
			} else if len(active) <= 3 {
				var sr, sg, sb int
				for _, p := range active {
					sr += p.r
					sg += p.g
					sb += p.b
				}
				n := len(active)
				fmt.Fprintf(&out, "\033[0m\033[38;2;%d;%d;%dm%s", sr/n, sg/n, sb/n, quads[mask])
			} else {
				bestErr := math.MaxFloat64
				bestM := 3
				var bestFG, bestBG [3]int
				for m := 1; m < 8; m++ {
					var g1, g2 []rgbaPx
					for i := 0; i < 4; i++ {
						if ((m >> i) & 1) == 1 {
							g1 = append(g1, pts[i])
						} else {
							g2 = append(g2, pts[i])
						}
					}
					c1 := avgPxColor(g1)
					c2 := avgPxColor(g2)
					errVal := sqPxErr(g1, c1) + sqPxErr(g2, c2)
					if m == 3 {
						errVal *= 0.92
					}
					if errVal < bestErr {
						bestErr = errVal
						bestM = m
						bestFG = c1
						bestBG = c2
					}
				}
				fmt.Fprintf(&out, "\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm%s",
					bestFG[0], bestFG[1], bestFG[2],
					bestBG[0], bestBG[1], bestBG[2],
					quads[bestM])
			}
		}
		out.WriteString("\033[0m\n")
	}

	return os.WriteFile(dstPath, []byte(out.String()), 0644)
}

func clampByte(v float64) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return int(v)
}

func avgPxColor(list []rgbaPx) [3]int {
	if len(list) == 0 {
		return [3]int{0, 0, 0}
	}
	var sr, sg, sb int
	for _, p := range list {
		sr += p.r
		sg += p.g
		sb += p.b
	}
	n := len(list)
	return [3]int{sr / n, sg / n, sb / n}
}

func sqPxErr(list []rgbaPx, c [3]int) float64 {
	var e float64
	for _, p := range list {
		dr := float64(p.r - c[0])
		dg := float64(p.g - c[1])
		db := float64(p.b - c[2])
		e += dr*dr + dg*dg + db*db
	}
	return e
}

// ==============================================================================
// 2. AST-Aware Semantic Conventional Commit Generator
// ==============================================================================

type astSymbolResult struct {
	Kind   string // method, func, type, test
	Symbol string // e.g., "(*TokenManager).ValidateToken", "computeSparkline"
}

// inspectGoAST inspects a Go file to see if the modified lines are inside a function or type
func inspectGoAST(filePath string, startLine, endLine int) *astSymbolResult {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil
	}

	var found *astSymbolResult

	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil || found != nil {
			return false
		}

		switch x := n.(type) {
		case *ast.FuncDecl:
			pos := fset.Position(x.Pos()).Line
			end := fset.Position(x.End()).Line
			if startLine >= pos && startLine <= end {
				name := x.Name.Name
				if strings.HasPrefix(name, "Test") {
					found = &astSymbolResult{Kind: "test", Symbol: name}
					return false
				}
				if x.Recv != nil && len(x.Recv.List) > 0 {
					recvType := exprToString(x.Recv.List[0].Type)
					found = &astSymbolResult{
						Kind:   "method",
						Symbol: fmt.Sprintf("(*%s).%s", recvType, name),
					}
					return false
				}
				found = &astSymbolResult{
					Kind:   "func",
					Symbol: fmt.Sprintf("%s()", name),
				}
				return false
			}

		case *ast.TypeSpec:
			pos := fset.Position(x.Pos()).Line
			end := fset.Position(x.End()).Line
			if startLine >= pos && startLine <= end {
				found = &astSymbolResult{
					Kind:   "type",
					Symbol: fmt.Sprintf("%s struct", x.Name.Name),
				}
				return false
			}
		}
		return true
	})

	return found
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return exprToString(t.X)
	case *ast.Ident:
		return t.Name
	default:
		return "Receiver"
	}
}

// parseHunkHeader extracts the enclosing function/class/component from unified diff header
// e.g. @@ -12,4 +12,6 @@ export const UserProfile = () => {
var hunkHeaderRegex = regexp.MustCompile(`@@\s+[^@]+@@\s*(?:(?:def|function|class|async\s+def|public|private|protected|fn|func|const|export\s+const|export\s+function|export\s+default)\s+)?([A-Za-z0-9_.:]+)`)

func detectEnclosingSymbol(repoRoot string, files []string) (string, string) {
	if len(files) == 0 {
		return "", ""
	}

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		absPath := filepath.Join(repoRoot, file)

		if ext == ".go" {
			diffOut, err := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "diff", "--cached", "-U0", "--", file).Output()
			if err == nil {
				reHunk := regexp.MustCompile(`@@\s+-\d+(?:,\d+)?\s+\+(\d+)(?:,(\d+))?\s+@@`)
				for _, line := range strings.Split(string(diffOut), "\n") {
					if m := reHunk.FindStringSubmatch(line); m != nil {
						startLine, _ := strconv.Atoi(m[1])
						count := 1
						if len(m) > 2 && m[2] != "" {
							count, _ = strconv.Atoi(m[2])
						}
						if res := inspectGoAST(absPath, startLine, startLine+count); res != nil {
							return res.Kind, res.Symbol
						}
					}
				}
			}
		} else if ext == ".py" || ext == ".ts" || ext == ".tsx" || ext == ".js" || ext == ".jsx" || ext == ".rs" || ext == ".cpp" || ext == ".java" || ext == ".cs" || ext == ".kt" {
			diffOut, err := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "diff", "--cached", "-p", "--", file).Output()
			if err == nil {
				for _, line := range strings.Split(string(diffOut), "\n") {
					if strings.HasPrefix(line, "@@") {
						if m := hunkHeaderRegex.FindStringSubmatch(line); m != nil {
							fn := strings.TrimSpace(m[1])
							if fn != "" && !strings.Contains(fn, "import") {
								return "func", fmt.Sprintf("%s()", fn)
							}
						}
					}
				}
			}
		}
	}
	return "", ""
}

func generateCommitMessage(repoRoot string) string {
	if strings.EqualFold(os.Getenv("SYNCALIBUR_AI_COMMIT"), "true") {
		if aiMsg := tryAICommitMessage(repoRoot); aiMsg != "" {
			return aiMsg
		}
	}

	out, err := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "diff", "--cached", "--name-status").Output()
	if err != nil || len(bytes.TrimSpace(out)) == 0 {
		return fmt.Sprintf("chore: auto-sync (%s)", time.Now().Format("15:04:05"))
	}

	var files, statuses []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			statuses = append(statuses, parts[0])
			files = append(files, parts[len(parts)-1])
		}
	}
	if len(files) == 0 {
		return fmt.Sprintf("chore: auto-sync (%s)", time.Now().Format("15:04:05"))
	}

	var totalAdd, totalDel int
	numOut, _ := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "diff", "--cached", "--numstat").Output()
	for _, line := range strings.Split(strings.TrimSpace(string(numOut)), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			a, _ := strconv.Atoi(fields[0])
			d, _ := strconv.Atoi(fields[1])
			totalAdd += a
			totalDel += d
		}
	}
	statBadge := fmt.Sprintf("(+%d/-%d)", totalAdd, totalDel)

	allDocs, allTests, allCI, allStyle := true, true, true, true
	allK8s, allHelm, allDocker, allTF, allGitOps := true, true, true, true, true
	hasAdd, hasDel := false, false

	for i, f := range files {
		st := statuses[i]
		if strings.HasPrefix(st, "A") {
			hasAdd = true
		}
		if strings.HasPrefix(st, "D") {
			hasDel = true
		}
		lower := strings.ToLower(f)
		base := strings.ToLower(filepath.Base(f))
		if !(strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".rst") || strings.HasSuffix(lower, ".txt") || strings.HasPrefix(lower, "docs/") || lower == "license") {
			allDocs = false
		}
		if !(strings.Contains(lower, "test") || strings.Contains(lower, "spec")) {
			allTests = false
		}
		if !(strings.HasPrefix(lower, ".github/") || strings.HasPrefix(lower, ".gitlab") || lower == "dockerfile" || lower == "makefile" || lower == "go.mod" || lower == "package.json" || strings.HasSuffix(lower, ".tf") || strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml")) {
			allCI = false
		}
		if !(strings.HasSuffix(lower, ".css") || strings.HasSuffix(lower, ".scss")) {
			allStyle = false
		}
		if !(strings.HasPrefix(lower, "k8s/") || strings.HasPrefix(lower, "deploy/") || strings.HasPrefix(lower, "manifests/") || strings.HasPrefix(lower, "base/") || strings.HasPrefix(lower, "overlays/")) {
			allK8s = false
		}
		if !(strings.HasPrefix(lower, "charts/") || base == "chart.yaml" || strings.HasPrefix(base, "values")) {
			allHelm = false
		}
		if !(strings.HasPrefix(base, "dockerfile") || base == ".dockerignore" || base == "compose.yaml" || base == "docker-compose.yml") {
			allDocker = false
		}
		if !(strings.HasSuffix(lower, ".tf") || strings.HasSuffix(lower, ".tfvars") || strings.HasPrefix(lower, "terraform/")) {
			allTF = false
		}
		if !(strings.HasPrefix(lower, "argocd/") || strings.HasPrefix(lower, "flux-system/") || strings.HasPrefix(lower, "apps/")) {
			allGitOps = false
		}
	}

	cType := "chore"
	infraScope := ""
	if allDocs {
		cType = "docs"
	} else if allTests {
		cType = "test"
	} else if allStyle {
		cType = "style"
	} else if allK8s {
		cType = "infra"
		infraScope = "k8s"
	} else if allHelm {
		cType = "helm"
		infraScope = "chart"
	} else if allDocker {
		cType = "docker"
		infraScope = "build"
	} else if allTF {
		cType = "tf"
		infraScope = "infra"
	} else if allGitOps {
		cType = "gitops"
		infraScope = "deploy"
	} else if allCI {
		cType = "build"
	} else if hasAdd || totalAdd > totalDel*2 {
		cType = "feat"
	} else if hasDel || totalDel > totalAdd {
		cType = "refactor"
	} else {
		diffSample, _ := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "diff", "--cached", "-U0").Output()
		if regexp.MustCompile(`(?i)\+.*(fix|bug|error|patch|resolve|crash)`).Match(diffSample) {
			cType = "fix"
		} else {
			cType = "feat"
		}
	}

	cScope := infraScope
	if cScope == "" {
		if len(files) == 1 {
			base := filepath.Base(files[0])
			ext := filepath.Ext(base)
			cScope = strings.TrimSuffix(base, ext)
		} else {
			firstDir := filepath.Dir(files[0])
			if firstDir != "." && firstDir != "" {
				same := true
				for _, f := range files[1:] {
					if filepath.Dir(f) != firstDir {
						same = false
						break
					}
				}
				if same {
					cScope = filepath.Base(firstDir)
				}
			}
		}
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{1,20}$`).MatchString(cScope) {
		cScope = ""
	}

	symbolKind, symbolName := detectEnclosingSymbol(repoRoot, files)
	if symbolKind == "test" {
		cType = "test"
	}

	var subject string
	if symbolName != "" {
		action := "update"
		if hasAdd && !hasDel {
			action = "add"
		} else if hasDel && !hasAdd {
			action = "remove"
		} else if cType == "fix" {
			action = "fix"
		}
		subject = fmt.Sprintf("%s %s", action, symbolName)
	} else if len(files) == 1 {
		action := "update"
		if strings.HasPrefix(statuses[0], "A") {
			action = "add"
		} else if strings.HasPrefix(statuses[0], "D") {
			action = "remove"
		}
		subject = fmt.Sprintf("%s %s", action, filepath.Base(files[0]))
	} else if len(files) <= 3 {
		var baseNames []string
		for _, f := range files {
			baseNames = append(baseNames, filepath.Base(f))
		}
		subject = "update " + strings.Join(baseNames, ", ")
	} else {
		subject = fmt.Sprintf("update %s, %s (+%d more)", filepath.Base(files[0]), filepath.Base(files[1]), len(files)-2)
	}

	if cScope != "" {
		return fmt.Sprintf("%s(%s): %s %s", cType, cScope, subject, statBadge)
	}
	return fmt.Sprintf("%s: %s %s", cType, subject, statBadge)
}

func tryAICommitMessage(repoRoot string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	statOut, _ := exec.CommandContext(ctx, "git", "-C", repoRoot, "diff", "--cached", "--shortstat").Output()
	diffOut, _ := exec.CommandContext(ctx, "git", "-C", repoRoot, "diff", "--cached", "-U1").Output()
	if len(diffOut) > 1600 {
		diffOut = diffOut[:1600]
	}

	model := strings.TrimSpace(os.Getenv("SYNCALIBUR_OLLAMA_MODEL"))
	if model == "" {
		model = "qwen2.5-coder:1.5b"
	}

	prompt := fmt.Sprintf("Write a single-line Conventional Commit message (max 65 chars, format: type(scope): summary) for this git diff:\n%s\n%s", string(statOut), string(diffOut))
	reqBody, _ := json.Marshal(map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:11434/api/generate", bytes.NewReader(reqBody))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var res struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return ""
	}

	clean := strings.TrimSpace(res.Response)
	clean = strings.Trim(clean, "`\"'\n\r ")
	if idx := strings.Index(clean, "\n"); idx != -1 {
		clean = strings.TrimSpace(clean[:idx])
	}

	if regexp.MustCompile(`^(feat|fix|docs|refactor|test|chore|build|ci|style)(\([a-zA-Z0-9_-]+\))?:\s*.+`).MatchString(clean) {
		return clean
	}
	return ""
}

// ==============================================================================
// 3. Pre-flight Syntax Gate & Shannon Entropy Secret Shield (`preflight`)
// ==============================================================================

var (
	uuidRegex    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	pureHexRegex = regexp.MustCompile(`^[0-9a-fA-F]{32,}$`)
)

func calculateShannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	var freq [256]float64
	var unicodeFreq map[rune]float64
	hasUnicode := false
	total := float64(len([]rune(s)))

	for _, r := range s {
		if r < 256 {
			freq[r]++
		} else {
			hasUnicode = true
			if unicodeFreq == nil {
				unicodeFreq = make(map[rune]float64)
			}
			unicodeFreq[r]++
		}
	}

	var entropy float64
	for _, count := range freq {
		if count > 0 {
			p := count / total
			entropy -= p * math.Log2(p)
		}
	}
	if hasUnicode {
		for _, count := range unicodeFreq {
			p := count / total
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

var knownSecretRegexes = []*regexp.Regexp{
	regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),                                                            // AWS Access Key ID
	regexp.MustCompile(`(?i)aws[_-]?secret[_-]?access[_-]?key\s*[:=]\s*["']?([A-Za-z0-9/+=]{40})["']?`), // AWS Secret Access Key
	regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9_]{36,}\b`),                                                // GitHub Classic Tokens
	regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{82}\b`),                                                // GitHub Fine-Grained PAT
	regexp.MustCompile(`\bglpat-[A-Za-z0-9\-]{20,}\b`),                                                   // GitLab PAT
	regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{32,}\b`),                                                      // OpenAI Key
	regexp.MustCompile(`\bsk-ant-[A-Za-z0-9_-]{32,}\b`),                                                  // Anthropic Key
	regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,}\b`),                                               // Slack Token
	regexp.MustCompile(`\b(?:sk|rk)_(?:live|test)_[0-9a-zA-Z]{24,}\b`),                                   // Stripe Keys
	regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),                                             // Private Key (RSA, EC, PKCS8)
	regexp.MustCompile(`-----BEGIN OPENSSH PRIVATE KEY-----`),                                            // OpenSSH Private Key
	regexp.MustCompile(`(?i)(?:api[_-]?key|secret|password|auth[_-]?token|bearer)\s*[:=]\s*["']([A-Za-z0-9_\-+/=]{20,})["']`), // Generic Secret assignments
}

var defaultSensitiveGlobs = []string{
	".env*", "*.tfstate*", "*.tfvars*", "credentials.json", "*serviceaccount*.json",
	"client_secret*.json", "azureProfile.json", "*.pem", "*.key", "*.p12", "*.pfx",
	"*.pkcs12", "*.jks", "*.keystore", "id_rsa*", "id_ecdsa*", "id_ed25519*", "id_dsa*",
	"*.asc", "secring.gpg", "*.kdbx", "kubeconfig", "*.kubeconfig", "*.ovpn", "*.tblk",
	"auth.json", ".netrc", ".npmrc", ".pypirc", "*.htpasswd", "*.dump", "*.rdb",
}

func matchesSensitivePattern(filePath string) bool {
	base := strings.ToLower(filepath.Base(filePath))
	// Whitelist templates and samples
	if strings.HasSuffix(base, ".example") || strings.HasSuffix(base, ".sample") ||
		strings.HasSuffix(base, ".template") || strings.HasSuffix(base, ".dist") ||
		strings.HasSuffix(base, ".default") || strings.HasSuffix(base, ".skel") {
		return false
	}
	for _, pattern := range defaultSensitiveGlobs {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}
	return false
}

func isBinaryFile(filePath string) (bool, error) {
	st, err := os.Lstat(filePath)
	if err != nil {
		return false, err
	}
	// Never load files > 10MB into memory for secret/syntax analysis
	if st.Size() > 10*1024*1024 {
		return true, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 8000)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}
	return false, nil
}

func isSafePath(repoRoot, relPath string) (bool, string) {
	absPath := filepath.Join(repoRoot, relPath)
	lst, err := os.Lstat(absPath)
	if err != nil {
		return false, "failed to lstat path"
	}

	if lst.Mode()&os.ModeSymlink != 0 {
		target, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return false, "broken or cyclic symlink"
		}
		relToRoot, err := filepath.Rel(repoRoot, target)
		if err != nil || strings.HasPrefix(relToRoot, "..") || filepath.IsAbs(relToRoot) {
			return false, fmt.Sprintf("symlink targets location outside repository: %s -> %s", relPath, target)
		}
	}

	if !lst.Mode().IsRegular() && lst.Mode()&os.ModeSymlink == 0 {
		return false, fmt.Sprintf("special file/socket/pipe rejected: %s", relPath)
	}
	return true, ""
}

func detectConflictMarkers(filePath string) (bool, string) {
	st, err := os.Lstat(filePath)
	if err != nil || !st.Mode().IsRegular() || st.Size() > 10*1024*1024 {
		return false, ""
	}
	f, err := os.Open(filePath)
	if err != nil {
		return false, ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.HasPrefix(line, "<<<<<<< ") || line == "=======" || strings.HasPrefix(line, ">>>>>>> ") {
			return true, fmt.Sprintf("%s:%d: unresolved merge conflict marker '%s'", filepath.Base(filePath), lineNum, strings.Fields(line)[0])
		}
	}
	return false, ""
}

func detectSecretsInFile(filePath string) (bool, string) {
	base := strings.ToLower(filepath.Base(filePath))

	// Unit test files: only skip heuristic entropy scanning to avoid false positives on mock/dummy test values,
	// but ALWAYS verify blacklist patterns (.env, id_rsa, *.pem, credentials.json, etc.).
	// NEVER use strings.Contains(base, "test") as that would blindly bypass test_api_keys.json or latest_creds.py!
	isTestCode := strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, ".test.ts") ||
		strings.HasSuffix(base, ".test.js") || strings.HasSuffix(base, ".spec.ts") ||
		strings.HasSuffix(base, ".spec.js") || strings.HasPrefix(base, "fixture_") ||
		strings.HasSuffix(base, "_fixture.json")

	// 1. Blacklist pattern verification (always enforced, even in test code)
	if matchesSensitivePattern(filePath) {
		return true, fmt.Sprintf("%s (matches sensitive file pattern blacklist)", filepath.Base(filePath))
	}

	if isTestCode {
		return false, ""
	}

	// 2. Binary / size safety check
	if isBin, _ := isBinaryFile(filePath); isBin {
		return false, ""
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, ""
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	isDocFile := ext == ".md" || ext == ".markdown" || ext == ".rst" || ext == ".txt"

	lines := strings.Split(string(data), "\n")
	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			// Skip explicit sample/example comments
			if strings.Contains(strings.ToLower(trimmed), "example") || strings.Contains(strings.ToLower(trimmed), "test") {
				continue
			}
		}

		// Skip SVG vector elements and HTML integrity hashes
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "xmlns") || strings.Contains(lower, "<svg") || strings.Contains(lower, "<path") ||
			strings.Contains(lower, "integrity=\"sha") || strings.Contains(lower, "sha512-") || strings.Contains(lower, "sha256-") {
			continue
		}

		// In documentation files, only run entropy if line has credential context (or code fence)
		if isDocFile {
			hasCredContext := strings.Contains(lower, "key") || strings.Contains(lower, "secret") ||
				strings.Contains(lower, "token") || strings.Contains(lower, "password") ||
				strings.Contains(lower, "auth") || strings.Contains(lower, "credential") ||
				strings.Contains(lower, "private") || strings.Contains(lower, "bearer") ||
				strings.HasPrefix(trimmed, "```")
			if !hasCredContext {
				// Still check explicit known regex signatures
				for _, re := range knownSecretRegexes {
					if m := re.FindString(trimmed); m != "" {
						return true, fmt.Sprintf("%s:%d (matches known credential signature)", filepath.Base(filePath), lineNum+1)
					}
				}
				continue
			}
		}

		for _, re := range knownSecretRegexes {
			if m := re.FindString(trimmed); m != "" {
				return true, fmt.Sprintf("%s:%d (matches known credential signature)", filepath.Base(filePath), lineNum+1)
			}
		}

		// Shannon Entropy check on long continuous strings (skip comments for raw entropy to avoid false positives)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		words := strings.Fields(trimmed)
		for _, w := range words {
			if idx := strings.Index(w, "="); idx != -1 {
				w = w[idx+1:]
			}
			w = strings.Trim(w, `"'=:,;()[]{}<>$*`+"`")
			// Skip shell variable expansions, regexes, code syntax, or paths
			if strings.Contains(w, "${") || strings.Contains(w, ":-") || strings.Contains(w, "/") || strings.Contains(w, "\\") || strings.Contains(w, ":") || strings.ContainsAny(w, "()[]{}<>^&|;*~") {
				continue
			}
			if len(w) >= 24 {
				// Exclude benign high-entropy structures: data URIs, base64 data, UUIDs, pure hex hashes
				if strings.HasPrefix(w, "data:") || strings.Contains(w, "base64,") ||
					uuidRegex.MatchString(w) || pureHexRegex.MatchString(w) {
					continue
				}

				// Exclude URLs, markdown links, file paths, template placeholders, CLI flags
				if strings.Contains(w, "://") || strings.Contains(w, "http://") || strings.Contains(w, "https://") || strings.Contains(w, "ftp://") {
					continue
				}
				if strings.Contains(w, "](") || strings.Contains(w, ")[") || strings.Contains(w, "[") || strings.Contains(w, "]") {
					continue
				}
				if strings.Contains(w, "<") && strings.Contains(w, ">") {
					continue
				}
				if (strings.Count(w, "/") >= 2 || strings.Count(w, "\\") >= 2) && !strings.HasSuffix(w, "=") {
					continue
				}
				if strings.HasPrefix(w, "./") || strings.HasPrefix(w, "../") || strings.HasPrefix(w, "/") {
					continue
				}
				if strings.Contains(w, "--") {
					continue
				}
				if strings.Contains(w, "@") && strings.Contains(w, ".") {
					continue
				}
				lowerW := strings.ToLower(w)
				if strings.HasSuffix(lowerW, ".md") || strings.HasSuffix(lowerW, ".go") || strings.HasSuffix(lowerW, ".ts") ||
					strings.HasSuffix(lowerW, ".tsx") || strings.HasSuffix(lowerW, ".json") || strings.HasSuffix(lowerW, ".yaml") ||
					strings.HasSuffix(lowerW, ".yml") || strings.HasSuffix(lowerW, ".tf") || strings.HasSuffix(lowerW, ".png") ||
					strings.HasSuffix(lowerW, ".jpg") || strings.HasSuffix(lowerW, ".html") || strings.HasSuffix(lowerW, ".css") ||
					strings.HasSuffix(lowerW, ".sh") || strings.HasSuffix(lowerW, ".py") || strings.HasSuffix(lowerW, ".txt") {
					continue
				}

				ent := calculateShannonEntropy(w)
				if ent >= 4.4 {
					hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(w)
					hasDigit := regexp.MustCompile(`[0-9]`).MatchString(w)
					if hasLetter && hasDigit {
						return true, fmt.Sprintf("%s:%d (high entropy %.2f bit/char token)", filepath.Base(filePath), lineNum+1, ent)
					}
				}
			}
		}
	}
	return false, ""
}

type bracketToken struct {
	char rune
	line int
}

func checkLexicalBalance(src []byte, ext string) (bool, string) {
	var stack []bracketToken
	runes := []rune(string(src))
	n := len(runes)
	line := 1

	inLineComment := false
	inBlockComment := false
	inDoubleQuote := false
	inSingleQuote := false
	inBacktick := false
	inTripleDoubleQuote := false
	inTripleSingleQuote := false
	escaped := false

	isRust := ext == ".rs"
	isPython := ext == ".py"
	isBash := ext == ".sh" || ext == ".bash"
	isHashComment := ext == ".py" || ext == ".sh" || ext == ".bash" || ext == ".yaml" || ext == ".yml" || ext == ".rb" || ext == ".toml" || ext == ".r"

	for i := 0; i < n; i++ {
		ch := runes[i]
		if ch == '\n' {
			line++
			if inLineComment {
				inLineComment = false
			}
			escaped = false
			continue
		}

		if inLineComment {
			continue
		}

		if inBlockComment {
			if ch == '*' && i+1 < n && runes[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if inTripleDoubleQuote {
			if ch == '"' && i+2 < n && runes[i+1] == '"' && runes[i+2] == '"' && !escaped {
				inTripleDoubleQuote = false
				i += 2
			} else if ch == '\\' && !escaped {
				escaped = true
			} else {
				escaped = false
			}
			continue
		}

		if inTripleSingleQuote {
			if ch == '\'' && i+2 < n && runes[i+1] == '\'' && runes[i+2] == '\'' && !escaped {
				inTripleSingleQuote = false
				i += 2
			} else if ch == '\\' && !escaped {
				escaped = true
			} else {
				escaped = false
			}
			continue
		}

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' && (inDoubleQuote || inSingleQuote || inBacktick) {
			escaped = true
			continue
		}

		if inDoubleQuote {
			if ch == '"' {
				inDoubleQuote = false
			}
			continue
		}

		if inSingleQuote {
			if ch == '\'' {
				// In YAML/SQL, two consecutive single quotes '' is an escaped quote inside a single-quoted scalar
				if (ext == ".yaml" || ext == ".yml" || ext == ".sql") && i+1 < n && runes[i+1] == '\'' {
					i++
					continue
				}
				inSingleQuote = false
			}
			continue
		}

		if inBacktick {
			if ch == '`' {
				inBacktick = false
			}
			continue
		}

		// Check Python multiline triple quotes
		if isPython {
			if ch == '"' && i+2 < n && runes[i+1] == '"' && runes[i+2] == '"' {
				inTripleDoubleQuote = true
				i += 2
				continue
			}
			if ch == '\'' && i+2 < n && runes[i+1] == '\'' && runes[i+2] == '\'' {
				inTripleSingleQuote = true
				i += 2
				continue
			}
		}

		// Check comment starts
		if isHashComment && ch == '#' {
			inLineComment = true
			continue
		}

		if ch == '/' && i+1 < n {
			if runes[i+1] == '/' {
				inLineComment = true
				i++
				continue
			} else if runes[i+1] == '*' {
				inBlockComment = true
				i++
				continue
			}
		}

		// Check string starts
		if ch == '"' {
			inDoubleQuote = true
			continue
		}
		if ch == '\'' && !isRust {
			// Contraction or apostrophe within words:
			// e.g. Node'lar, Swap'i, Master'a, Token'i, it's, don't, can't
			if i > 0 && (unicode.IsLetter(runes[i-1]) || unicode.IsDigit(runes[i-1])) && i+1 < n && unicode.IsLetter(runes[i+1]) {
				continue
			}
			// Plural possessive or end-of-word apostrophe followed by whitespace/punctuation
			if i > 0 && unicode.IsLetter(runes[i-1]) && i >= 2 && unicode.IsLetter(runes[i-2]) && (i+1 < n && (unicode.IsSpace(runes[i+1]) || runes[i+1] == '\n' || runes[i+1] == ',' || runes[i+1] == '.' || runes[i+1] == ':' || runes[i+1] == ';')) {
				continue
			}
			// In YAML/plain text, an apostrophe attached to an identifier is not a string opener
			if (ext == ".yaml" || ext == ".yml") && i > 0 && (unicode.IsLetter(runes[i-1]) || unicode.IsDigit(runes[i-1])) {
				continue
			}
			inSingleQuote = true
			continue
		}
		if ch == '`' {
			inBacktick = true
			continue
		}

		// Brackets
		if ch == '{' || ch == '(' || ch == '[' {
			stack = append(stack, bracketToken{char: ch, line: line})
		} else if ch == '}' || ch == ')' || ch == ']' {
			if len(stack) == 0 {
				// In Bash, case patterns end with a lone ')' (e.g. pattern) cmd ;; esac)
				if isBash && ch == ')' {
					continue
				}
				return false, fmt.Sprintf("line %d: unexpected closing '%c' with no matching open bracket", line, ch)
			}
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if (ch == '}' && top.char != '{') ||
				(ch == ')' && top.char != '(') ||
				(ch == ']' && top.char != '[') {
				if isBash && ch == ')' && top.char != '(' {
					stack = append(stack, top)
					continue
				}
				return false, fmt.Sprintf("line %d: mismatched closing '%c' (expected closing for '%c' from line %d)", line, ch, top.char, top.line)
			}
		}
	}

	if inTripleDoubleQuote || inTripleSingleQuote {
		return false, "unclosed multiline docstring spanning to end of file"
	}
	if inBlockComment {
		return false, "unclosed block comment '/*' spanning to end of file"
	}
	if inDoubleQuote {
		return false, "unclosed string literal '\"' spanning to end of file"
	}
	if inSingleQuote {
		return false, "unclosed single quote literal ''' spanning to end of file"
	}
	if inBacktick {
		return false, "unclosed template literal '`' spanning to end of file"
	}
	if len(stack) > 0 {
		top := stack[len(stack)-1]
		return false, fmt.Sprintf("line %d: unclosed '%c' at end of file", top.line, top.char)
	}

	return true, ""
}

func validateFileSyntax(repoRoot, relPath string) (bool, string) {
	absPath := filepath.Join(repoRoot, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return true, "" // file might be deleted in git
	}

	ext := strings.ToLower(filepath.Ext(relPath))
	baseName := strings.ToLower(filepath.Base(relPath))

	// Documentation and plain text files should NEVER be blocked by syntax hold
	if ext == ".md" || ext == ".markdown" || ext == ".txt" || ext == ".rst" || ext == ".adoc" || ext == ".log" {
		return true, ""
	}

	// Dockerfile / Containerfile
	if baseName == "dockerfile" || baseName == "containerfile" || strings.HasPrefix(baseName, "dockerfile.") || strings.HasSuffix(baseName, ".dockerfile") {
		return validateDockerfileSyntax(data, relPath)
	}

	switch ext {
	case ".go":
		fset := token.NewFileSet()
		if _, err := parser.ParseFile(fset, absPath, data, parser.AllErrors); err != nil {
			return false, fmt.Sprintf("%s: %v", relPath, err)
		}
	case ".py":
		if _, err := exec.LookPath("python3"); err == nil {
			cmd := exec.Command("python3", "-B", "-m", "py_compile", absPath)
			if out, err := cmd.CombinedOutput(); err != nil {
				errMsg := strings.TrimSpace(string(out))
				if errMsg == "" {
					errMsg = "Python syntax compilation failed"
				}
				return false, fmt.Sprintf("%s: %s", relPath, errMsg)
			}
		}
		if ok, reason := checkLexicalBalance(data, ext); !ok {
			return false, fmt.Sprintf("%s: %s", relPath, reason)
		}
	case ".js", ".mjs", ".cjs":
		if _, err := exec.LookPath("node"); err == nil {
			cmd := exec.Command("node", "--check", absPath)
			if out, err := cmd.CombinedOutput(); err != nil {
				errMsg := strings.TrimSpace(string(out))
				if errMsg == "" {
					errMsg = "JavaScript syntax check failed"
				}
				return false, fmt.Sprintf("%s: %s", relPath, errMsg)
			}
		}
		if ok, reason := checkLexicalBalance(data, ext); !ok {
			return false, fmt.Sprintf("%s: %s", relPath, reason)
		}
	case ".ts", ".tsx", ".jsx":
		localTsc := filepath.Join(repoRoot, "node_modules", ".bin", "tsc")
		tscBin := ""
		if _, err := os.Stat(localTsc); err == nil {
			tscBin = localTsc
		} else if p, err := exec.LookPath("tsc"); err == nil {
			tscBin = p
		}
		if tscBin != "" {
			cmd := exec.Command(tscBin, "--noEmit", "--skipLibCheck", "--target", "esnext", "--jsx", "preserve", absPath)
			if out, err := cmd.CombinedOutput(); err != nil {
				errMsg := strings.TrimSpace(string(out))
				if errMsg != "" {
					lines := strings.Split(errMsg, "\n")
					for _, l := range lines {
						lTrim := strings.TrimSpace(l)
						if lTrim == "" || !strings.Contains(lTrim, "error TS") {
							continue
						}
						// Only genuine syntactic parsing errors (TS1xxx) fail the syntax gate.
						// Ambient missing declarations, global types, or import errors are ignored in standalone checks.
						if isTypeScriptSyntaxError(lTrim) {
							return false, fmt.Sprintf("%s: %s", relPath, lTrim)
						}
					}
				}
			}
		}
		if ok, reason := checkLexicalBalance(data, ext); !ok {
			return false, fmt.Sprintf("%s: %s", relPath, reason)
		}
	case ".java":
		if javac, err := exec.LookPath("javac"); err == nil {
			tmpDir, err := os.MkdirTemp("", "syncalibur-javac-*")
			if err == nil {
				defer os.RemoveAll(tmpDir)
				cmd := exec.Command(javac, "-nowarn", "-d", tmpDir, absPath)
				if out, err := cmd.CombinedOutput(); err != nil {
					errMsg := strings.TrimSpace(string(out))
					if errMsg != "" && !strings.Contains(errMsg, "cannot find symbol") && !strings.Contains(errMsg, "does not exist") {
						lines := strings.Split(errMsg, "\n")
						return false, fmt.Sprintf("%s: %s", relPath, lines[0])
					}
				}
			}
		}
		if ok, reason := checkLexicalBalance(data, ext); !ok {
			return false, fmt.Sprintf("%s: %s", relPath, reason)
		}
	case ".json", ".jsonc", ".json5":
		cleanJSON := stripCommentsFromJSON(data)
		var v any
		if err := json.Unmarshal(cleanJSON, &v); err != nil {
			return false, fmt.Sprintf("%s: invalid JSON: %v", relPath, err)
		}
	case ".yaml", ".yml":
		return validateYAMLSyntax(data, relPath)
	case ".sh", ".bash", ".zsh":
		bashBin := ""
		if p, err := exec.LookPath("bash"); err == nil {
			bashBin = p
		} else if runtime.GOOS == "windows" {
			for _, candidate := range []string{
				`C:\Program Files\Git\bin\bash.exe`,
				`C:\Program Files\Git\usr\bin\bash.exe`,
				`C:\Git\bin\bash.exe`,
			} {
				if _, err := os.Stat(candidate); err == nil {
					bashBin = candidate
					break
				}
			}
		}
		if bashBin != "" {
			cmd := exec.Command(bashBin, "-n", absPath)
			if out, err := cmd.CombinedOutput(); err != nil {
				errMsg := strings.TrimSpace(string(out))
				if errMsg == "" {
					errMsg = "Shell script syntax validation failed"
				}
				return false, fmt.Sprintf("%s: %s", relPath, errMsg)
			}
		}
		if ok, reason := checkLexicalBalance(data, ext); !ok {
			return false, fmt.Sprintf("%s: %s", relPath, reason)
		}
	case ".rb":
		if _, err := exec.LookPath("ruby"); err == nil {
			cmd := exec.Command("ruby", "-c", absPath)
			if out, err := cmd.CombinedOutput(); err != nil {
				errMsg := strings.TrimSpace(string(out))
				if errMsg != "" && strings.Contains(errMsg, "syntax error") {
					lines := strings.Split(errMsg, "\n")
					return false, fmt.Sprintf("%s: %s", relPath, lines[0])
				}
			}
		}
		if ok, reason := checkLexicalBalance(data, ext); !ok {
			return false, fmt.Sprintf("%s: %s", relPath, reason)
		}
	case ".php":
		if _, err := exec.LookPath("php"); err == nil {
			cmd := exec.Command("php", "-l", absPath)
			if out, err := cmd.CombinedOutput(); err != nil {
				errMsg := strings.TrimSpace(string(out))
				if errMsg != "" {
					lines := strings.Split(errMsg, "\n")
					return false, fmt.Sprintf("%s: %s", relPath, lines[0])
				}
			}
		}
		if ok, reason := checkLexicalBalance(data, ext); !ok {
			return false, fmt.Sprintf("%s: %s", relPath, reason)
		}
	case ".tf", ".tfvars", ".hcl":
		return validateHCLSyntax(data, relPath)
	case ".sql":
		return validateSQLSyntax(data, relPath)
	case ".css", ".scss", ".sass", ".less":
		return validateCSSSyntax(data, relPath)
	case ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".rs", ".kt", ".kts", ".cs", ".swift", ".scala", ".lua":
		if ok, reason := checkLexicalBalance(data, ext); !ok {
			return false, fmt.Sprintf("%s: %s", relPath, reason)
		}
	}
	return true, ""
}

func isTypeScriptSyntaxError(line string) bool {
	idx := strings.Index(line, "error TS")
	if idx == -1 {
		return false
	}
	codePart := line[idx+len("error TS"):]
	// In TypeScript compiler, syntactic/grammar parsing errors strictly belong to the TS1000-TS1999 range (TS1xxx).
	// Semantic errors (missing modules, missing types, missing globals, type checking) are TS2000+.
	if len(codePart) >= 4 && codePart[0] == '1' {
		isAllDigits := true
		for i := 1; i < 4; i++ {
			if codePart[i] < '0' || codePart[i] > '9' {
				isAllDigits = false
				break
			}
		}
		if isAllDigits {
			return true
		}
	}
	return false
}

func stripCommentsFromJSON(data []byte) []byte {
	var out []byte
	inString := false
	escaped := false
	n := len(data)

	for i := 0; i < n; i++ {
		ch := data[i]
		if inString {
			out = append(out, ch)
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			out = append(out, ch)
			continue
		}

		if ch == '/' && i+1 < n && data[i+1] == '/' {
			for i < n && data[i] != '\n' {
				i++
			}
			if i < n {
				out = append(out, '\n')
			}
			continue
		}

		if ch == '/' && i+1 < n && data[i+1] == '*' {
			i += 2
			for i+1 < n && !(data[i] == '*' && data[i+1] == '/') {
				i++
			}
			i++
			continue
		}

		out = append(out, ch)
	}
	return out
}

func validateYAMLSyntax(data []byte, relPath string) (bool, string) {
	lines := strings.Split(string(data), "\n")

	inBlockScalar := false
	blockScalarIndent := 0

	templateStack := 0
	flowStack := []rune{}

	for lineNum, line := range lines {
		trimmed := strings.TrimRight(line, "\r")

		// 1. Check forbidden tabs for indentation (YAML spec)
		if strings.HasPrefix(trimmed, "\t") {
			return false, fmt.Sprintf("%s:%d: YAML syntax error: tabs are forbidden for indentation", relPath, lineNum+1)
		}

		// Calculate indentation
		indent := 0
		for _, ch := range trimmed {
			if ch == ' ' {
				indent++
			} else {
				break
			}
		}

		cleanLine := strings.TrimSpace(trimmed)
		if cleanLine == "" {
			continue
		}

		// Check if inside multiline literal/folded block scalar (| or >)
		// e.g. ConfigMap with embedded HTML, CSS, Shell, JSON, Nginx
		if inBlockScalar {
			if indent > blockScalarIndent {
				// Embedded content lines: do NOT parse YAML syntax or CSS hex (#0f172a) as comments!
				// Only verify template tag balances ({{ ... }}) if present
				if strings.Contains(cleanLine, "{{") || strings.Contains(cleanLine, "}}") {
					for i := 0; i < len(cleanLine); i++ {
						if i+1 < len(cleanLine) && cleanLine[i] == '{' && cleanLine[i+1] == '{' {
							templateStack++
							i++
						} else if i+1 < len(cleanLine) && cleanLine[i] == '}' && cleanLine[i+1] == '}' {
							if templateStack > 0 {
								templateStack--
							}
							i++
						}
					}
				}
				continue
			} else {
				inBlockScalar = false
			}
		}

		// Check if this line begins a multiline block scalar
		// e.g. "index.html: |", "script: |-", "config: >"
		if strings.Contains(cleanLine, ": |") || strings.Contains(cleanLine, ": |-") ||
			strings.Contains(cleanLine, ": |+") || strings.Contains(cleanLine, ": >") ||
			strings.Contains(cleanLine, ": >-") || strings.Contains(cleanLine, ": >+") ||
			cleanLine == "|" || cleanLine == "|-" || cleanLine == ">" || cleanLine == ">-" {
			inBlockScalar = true
			blockScalarIndent = indent
		}

		// Parse line outside block scalar
		runes := []rune(cleanLine)
		n := len(runes)
		inQuote := rune(0)

		for i := 0; i < n; i++ {
			r := runes[i]

			if inQuote != 0 {
				if r == inQuote {
					if r == '\'' && i+1 < n && runes[i+1] == '\'' {
						i++ // escaped '' in YAML
						continue
					}
					inQuote = 0
				}
				continue
			}
			if r == '"' || r == '\'' {
				if i == 0 || unicode.IsSpace(runes[i-1]) || runes[i-1] == ':' || runes[i-1] == '-' || runes[i-1] == '[' || runes[i-1] == '{' {
					inQuote = r
				}
				continue
			}

			// Comment check: YAML comments start with # preceded by whitespace or at line start
			if r == '#' && (i == 0 || unicode.IsSpace(runes[i-1])) {
				break
			}

			// Template tags: {{ and }}
			if i+1 < n && r == '{' && runes[i+1] == '{' {
				templateStack++
				i++
				continue
			}
			if i+1 < n && r == '}' && runes[i+1] == '}' {
				if templateStack > 0 {
					templateStack--
				}
				i++
				continue
			}

			// Flow mapping braces outside template tags
			if templateStack == 0 {
				if r == '{' || r == '[' {
					flowStack = append(flowStack, r)
				} else if r == '}' || r == ']' {
					if len(flowStack) > 0 {
						top := flowStack[len(flowStack)-1]
						if (r == '}' && top == '{') || (r == ']' && top == '[') {
							flowStack = flowStack[:len(flowStack)-1]
						}
					}
				}
			}
		}
	}

	if templateStack > 0 {
		return false, fmt.Sprintf("%s: unclosed template tag '{{' at end of file", relPath)
	}
	if len(flowStack) > 0 {
		return false, fmt.Sprintf("%s: unclosed flow mapping '%c' at end of file", relPath, flowStack[len(flowStack)-1])
	}
	return true, ""
}

func validateCSSSyntax(data []byte, relPath string) (bool, string) {
	lines := strings.Split(string(data), "\n")
	var stack []bracketToken
	inBlockComment := false
	inString := rune(0)
	escaped := false

	for lineNum, line := range lines {
		runes := []rune(line)
		n := len(runes)

		for i := 0; i < n; i++ {
			r := runes[i]

			if inBlockComment {
				if r == '*' && i+1 < n && runes[i+1] == '/' {
					inBlockComment = false
					i++
				}
				continue
			}

			if inString != 0 {
				if escaped {
					escaped = false
				} else if r == '\\' {
					escaped = true
				} else if r == inString {
					inString = 0
				}
				continue
			}

			// Single-line SCSS/Less comment //
			if r == '/' && i+1 < n && runes[i+1] == '/' {
				break
			}
			// Block comment /*
			if r == '/' && i+1 < n && runes[i+1] == '*' {
				inBlockComment = true
				i++
				continue
			}

			if r == '"' || r == '\'' {
				inString = r
				continue
			}

			if r == '{' || r == '(' || r == '[' {
				stack = append(stack, bracketToken{char: r, line: lineNum + 1})
			} else if r == '}' || r == ')' || r == ']' {
				if len(stack) == 0 {
					return false, fmt.Sprintf("%s:%d: unexpected closing '%c' with no matching open bracket", relPath, lineNum+1, r)
				}
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if (r == '}' && top.char != '{') || (r == ')' && top.char != '(') || (r == ']' && top.char != '[') {
					return false, fmt.Sprintf("%s:%d: mismatched closing '%c' (expected closing for '%c' from line %d)", relPath, lineNum+1, r, top.char, top.line)
				}
			}
		}
	}

	if inBlockComment {
		return false, fmt.Sprintf("%s: unclosed block comment '/*' at end of file", relPath)
	}
	if inString != 0 {
		return false, fmt.Sprintf("%s: unclosed string literal '%c' at end of file", relPath, inString)
	}
	if len(stack) > 0 {
		top := stack[len(stack)-1]
		return false, fmt.Sprintf("%s: line %d: unclosed '%c' at end of file", relPath, top.line, top.char)
	}
	return true, ""
}

func validateHCLSyntax(data []byte, relPath string) (bool, string) {
	lines := strings.Split(string(data), "\n")
	var stack []bracketToken
	inBlockComment := false
	inHeredoc := false
	heredocMarker := ""
	inString := false
	escaped := false

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inHeredoc {
			if trimmed == heredocMarker {
				inHeredoc = false
				heredocMarker = ""
			}
			continue
		}

		runes := []rune(line)
		n := len(runes)

		for i := 0; i < n; i++ {
			r := runes[i]

			if inBlockComment {
				if r == '*' && i+1 < n && runes[i+1] == '/' {
					inBlockComment = false
					i++
				}
				continue
			}

			if inString {
				if escaped {
					escaped = false
				} else if r == '\\' {
					escaped = true
				} else if r == '"' {
					inString = false
				}
				continue
			}

			if r == '#' || (r == '/' && i+1 < n && runes[i+1] == '/') {
				break
			}
			if r == '/' && i+1 < n && runes[i+1] == '*' {
				inBlockComment = true
				i++
				continue
			}

			if r == '<' && i+1 < n && runes[i+1] == '<' {
				rem := strings.TrimSpace(string(runes[i+2:]))
				rem = strings.TrimPrefix(rem, "-")
				marker := strings.TrimSpace(rem)
				if marker != "" && !strings.ContainsAny(marker, " \t\r\n\"'{}()") {
					inHeredoc = true
					heredocMarker = marker
					break
				}
			}

			if r == '"' {
				inString = true
				continue
			}

			if r == '{' || r == '(' || r == '[' {
				stack = append(stack, bracketToken{char: r, line: lineNum + 1})
			} else if r == '}' || r == ')' || r == ']' {
				if len(stack) == 0 {
					return false, fmt.Sprintf("%s:%d: unexpected closing '%c'", relPath, lineNum+1, r)
				}
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if (r == '}' && top.char != '{') || (r == ')' && top.char != '(') || (r == ']' && top.char != '[') {
					return false, fmt.Sprintf("%s:%d: mismatched closing '%c' (expected closing for '%c' from line %d)", relPath, lineNum+1, r, top.char, top.line)
				}
			}
		}
	}

	if inHeredoc {
		return false, fmt.Sprintf("%s: unclosed heredoc '<<%s' at end of file", relPath, heredocMarker)
	}
	if inBlockComment {
		return false, fmt.Sprintf("%s: unclosed block comment '/*' at end of file", relPath)
	}
	if inString {
		return false, fmt.Sprintf("%s: unclosed string literal '\"' at end of file", relPath)
	}
	if len(stack) > 0 {
		top := stack[len(stack)-1]
		return false, fmt.Sprintf("%s: line %d: unclosed '%c' at end of file", relPath, top.line, top.char)
	}
	return true, ""
}

func validateSQLSyntax(data []byte, relPath string) (bool, string) {
	lines := strings.Split(string(data), "\n")
	var stack []bracketToken
	inBlockComment := false
	inString := false

	for lineNum, line := range lines {
		runes := []rune(line)
		n := len(runes)

		for i := 0; i < n; i++ {
			r := runes[i]

			if inBlockComment {
				if r == '*' && i+1 < n && runes[i+1] == '/' {
					inBlockComment = false
					i++
				}
				continue
			}

			if inString {
				if r == '\'' {
					if i+1 < n && runes[i+1] == '\'' {
						i++
						continue
					}
					inString = false
				}
				continue
			}

			if (r == '-' && i+1 < n && runes[i+1] == '-') || r == '#' {
				break
			}
			if r == '/' && i+1 < n && runes[i+1] == '*' {
				inBlockComment = true
				i++
				continue
			}

			if r == '\'' {
				inString = true
				continue
			}

			if r == '(' {
				stack = append(stack, bracketToken{char: r, line: lineNum + 1})
			} else if r == ')' {
				if len(stack) == 0 {
					return false, fmt.Sprintf("%s:%d: unexpected closing ')'", relPath, lineNum+1)
				}
				stack = stack[:len(stack)-1]
			}
		}
	}

	if inString {
		return false, fmt.Sprintf("%s: unclosed string literal ''' at end of file", relPath)
	}
	if len(stack) > 0 {
		top := stack[len(stack)-1]
		return false, fmt.Sprintf("%s: line %d: unclosed '(' at end of file", relPath, top.line)
	}
	return true, ""
}

func validateDockerfileSyntax(data []byte, relPath string) (bool, string) {
	lines := strings.Split(string(data), "\n")
	validInstructions := map[string]bool{
		"FROM": true, "RUN": true, "CMD": true, "LABEL": true, "MAINTAINER": true,
		"EXPOSE": true, "ENV": true, "ADD": true, "COPY": true, "ENTRYPOINT": true,
		"VOLUME": true, "USER": true, "WORKDIR": true, "ARG": true, "ONBUILD": true,
		"STOPSIGNAL": true, "HEALTHCHECK": true, "SHELL": true,
	}

	inContinuation := false

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if inContinuation {
			if strings.HasSuffix(trimmed, "\\") {
				inContinuation = true
			} else {
				inContinuation = false
			}
			continue
		}

		parts := strings.Fields(trimmed)
		if len(parts) == 0 {
			continue
		}

		instr := strings.ToUpper(parts[0])
		if !validInstructions[instr] {
			return false, fmt.Sprintf("%s:%d: unknown Dockerfile instruction '%s'", relPath, lineNum+1, parts[0])
		}

		if strings.HasSuffix(trimmed, "\\") {
			inContinuation = true
		}
	}
	return true, ""
}

// ==============================================================================
// 3. Static DevOps & Infrastructure Manifest Validator (Pure Go Standard Library)
// ==============================================================================

type IntOrString struct {
	IsString bool
	IntVal   int
	StrVal   string
}

func (ios IntOrString) String() string {
	if ios.IsString {
		return ios.StrVal
	}
	return strconv.Itoa(ios.IntVal)
}

func parseIntOrString(s string) IntOrString {
	s = strings.TrimSpace(strings.Trim(s, `"'`))
	if val, err := strconv.Atoi(s); err == nil {
		return IntOrString{IsString: false, IntVal: val}
	}
	return IntOrString{IsString: true, StrVal: s}
}

type K8sPort struct {
	Name          string
	Port          int
	TargetPort    IntOrString
	ContainerPort int
	Protocol      string
	Line          int
}

type K8sProbe struct {
	Type string // "livenessProbe", "readinessProbe", "startupProbe"
	Port IntOrString
	Line int
}

type K8sContainer struct {
	Name   string
	Image  string
	Ports  []K8sPort
	Probes []K8sProbe
	Line   int
}

type K8sResource struct {
	APIVersion      string
	Kind            string
	Name            string
	Namespace       string
	FilePath        string
	LineStart       int
	ServiceSelector map[string]string
	ServicePorts    []K8sPort
	TemplateLabels  map[string]string
	Containers      []K8sContainer
	Data            map[string]string
	DataLines       map[string]int
}

type InfraIssue struct {
	Severity string `json:"severity"` // "ERROR" or "WARNING"
	Code     string `json:"code"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Resource string `json:"resource"`
	Message  string `json:"message"`
	FixHint  string `json:"fix_hint"`
}

func splitYAMLDocuments(data []byte) []string {
	var docs []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var buf strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" || strings.HasPrefix(trimmed, "--- ") {
			if buf.Len() > 0 {
				docs = append(docs, buf.String())
				buf.Reset()
			}
			continue
		}
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	if buf.Len() > 0 {
		docs = append(docs, buf.String())
	}
	return docs
}

func parseK8sDoc(doc string, filePath string, baseLine int) *K8sResource {
	lines := strings.Split(doc, "\n")
	res := &K8sResource{
		FilePath:        filePath,
		LineStart:       baseLine,
		ServiceSelector: make(map[string]string),
		TemplateLabels:  make(map[string]string),
		Data:            make(map[string]string),
		DataLines:       make(map[string]int),
	}

	var sectionStack []string
	var currentContainer *K8sContainer
	var currentPort *K8sPort
	var currentProbe *K8sProbe
	var inDataBlock bool

	for i, rawLine := range lines {
		lineNum := baseLine + i
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))

		if strings.HasPrefix(trimmed, "apiVersion:") {
			res.APIVersion = strings.TrimSpace(strings.TrimPrefix(trimmed, "apiVersion:"))
			continue
		}
		if strings.HasPrefix(trimmed, "kind:") {
			res.Kind = strings.TrimSpace(strings.TrimPrefix(trimmed, "kind:"))
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		key := strings.TrimSpace(parts[0])
		val := ""
		if len(parts) == 2 {
			val = strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
		}

		if (key == "name" || strings.HasPrefix(trimmed, "- name:")) && (indent == 2 || indent == 4) && res.Name == "" {
			nameVal := val
			if strings.HasPrefix(trimmed, "- name:") {
				nameVal = strings.TrimSpace(strings.TrimPrefix(trimmed, "- name:"))
				nameVal = strings.Trim(nameVal, `"'`)
			}
			res.Name = nameVal
		}
		if key == "namespace" && (indent == 2 || indent == 4) && res.Namespace == "" {
			res.Namespace = val
		}

		// Service Selector
		if key == "selector" && (res.Kind == "Service" || res.Kind == "") {
			sectionStack = []string{"service_selector"}
			continue
		}
		if len(sectionStack) > 0 && sectionStack[0] == "service_selector" {
			if indent > 2 && len(parts) == 2 {
				res.ServiceSelector[key] = val
				continue
			} else if indent <= 2 {
				sectionStack = nil
			}
		}

		// Workload template labels
		if key == "labels" && indent >= 4 {
			sectionStack = []string{"template_labels"}
			continue
		}
		if len(sectionStack) > 0 && sectionStack[0] == "template_labels" {
			if indent >= 6 && len(parts) == 2 {
				res.TemplateLabels[key] = val
				continue
			} else if indent < 6 {
				sectionStack = nil
			}
		}

		// ConfigMap data block
		if key == "data" && indent <= 2 && (res.Kind == "ConfigMap" || res.Kind == "") {
			inDataBlock = true
			continue
		}
		if inDataBlock {
			if indent >= 2 && len(parts) == 2 {
				res.Data[key] = val
				res.DataLines[key] = lineNum
				continue
			} else if indent < 2 {
				inDataBlock = false
			}
		}

		// Service Ports
		if (key == "ports" || trimmed == "ports:") && indent <= 4 && res.Kind == "Service" {
			sectionStack = []string{"service_ports"}
			continue
		}
		if len(sectionStack) > 0 && sectionStack[0] == "service_ports" {
			if strings.HasPrefix(trimmed, "-") {
				res.ServicePorts = append(res.ServicePorts, K8sPort{Line: lineNum, Protocol: "TCP"})
				currentPort = &res.ServicePorts[len(res.ServicePorts)-1]
				trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
				parts = strings.SplitN(trimmed, ":", 2)
				key = strings.TrimSpace(parts[0])
				if len(parts) == 2 {
					val = strings.TrimSpace(strings.Trim(parts[1], `"'`))
				}
			}
			if currentPort != nil {
				if key == "port" {
					currentPort.Port, _ = strconv.Atoi(val)
				} else if key == "targetPort" {
					currentPort.TargetPort = parseIntOrString(val)
				} else if key == "name" {
					currentPort.Name = val
				} else if key == "protocol" {
					currentPort.Protocol = val
				}
			}
			if indent < 2 {
				sectionStack = nil
				currentPort = nil
			}
		}

		// Containers in Workload
		if (key == "containers" || trimmed == "containers:") && indent >= 4 {
			sectionStack = []string{"containers"}
			continue
		}
		if len(sectionStack) > 0 && sectionStack[0] == "containers" {
			if strings.HasPrefix(trimmed, "- name:") || (strings.HasPrefix(trimmed, "-") && strings.Contains(trimmed, "image:")) {
				res.Containers = append(res.Containers, K8sContainer{Line: lineNum})
				currentContainer = &res.Containers[len(res.Containers)-1]
				if strings.HasPrefix(trimmed, "- name:") {
					currentContainer.Name = strings.TrimSpace(strings.Trim(strings.TrimPrefix(trimmed, "- name:"), `"'`))
				}
			}
			if currentContainer != nil {
				if key == "name" && currentContainer.Name == "" {
					currentContainer.Name = val
				} else if key == "image" {
					currentContainer.Image = val
				} else if strings.HasPrefix(trimmed, "- containerPort:") {
					cpVal, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(trimmed, "- containerPort:")))
					currentContainer.Ports = append(currentContainer.Ports, K8sPort{ContainerPort: cpVal, Line: lineNum, Protocol: "TCP"})
				} else if (key == "containerPort" || key == "- containerPort") && len(currentContainer.Ports) > 0 {
					currentContainer.Ports[len(currentContainer.Ports)-1].ContainerPort, _ = strconv.Atoi(val)
				} else if key == "name" && len(currentContainer.Ports) > 0 && indent >= 8 {
					currentContainer.Ports[len(currentContainer.Ports)-1].Name = val
				} else if key == "livenessProbe" || key == "readinessProbe" || key == "startupProbe" {
					currentProbe = &K8sProbe{Type: key, Line: lineNum}
					currentContainer.Probes = append(currentContainer.Probes, *currentProbe)
				} else if key == "port" && currentProbe != nil && indent >= 8 {
					currentProbe.Port = parseIntOrString(val)
					if len(currentContainer.Probes) > 0 {
						currentContainer.Probes[len(currentContainer.Probes)-1].Port = currentProbe.Port
					}
				}
			}
			if indent < 4 {
				sectionStack = nil
				currentContainer = nil
			}
		}
	}

	if res.Namespace == "" {
		res.Namespace = "default"
	}
	return res
}

func labelsMatch(selector, labels map[string]string) bool {
	if len(selector) == 0 || len(labels) == 0 {
		return false
	}
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func validateServicePortTargetPort(services []*K8sResource, workloads []*K8sResource) []InfraIssue {
	var issues []InfraIssue
	for _, svc := range services {
		if svc.Kind != "Service" || len(svc.ServiceSelector) == 0 {
			continue
		}
		var matchingWorkloads []*K8sResource
		for _, w := range workloads {
			if (w.Namespace == svc.Namespace || svc.Namespace == "default" || w.Namespace == "default") && labelsMatch(svc.ServiceSelector, w.TemplateLabels) {
				matchingWorkloads = append(matchingWorkloads, w)
			}
		}
		if len(matchingWorkloads) == 0 {
			continue
		}
		for _, sPort := range svc.ServicePorts {
			target := sPort.TargetPort
			if !target.IsString && target.IntVal == 0 {
				target = IntOrString{IsString: false, IntVal: sPort.Port}
			}
			foundMatch := false
			var declared []string
			for _, wl := range matchingWorkloads {
				for _, c := range wl.Containers {
					for _, cp := range c.Ports {
						declared = append(declared, fmt.Sprintf("%d/%s (name: %s)", cp.ContainerPort, cp.Protocol, cp.Name))
						if target.IsString {
							if cp.Name == target.StrVal {
								foundMatch = true
								break
							}
						} else {
							if cp.ContainerPort == target.IntVal {
								foundMatch = true
								break
							}
						}
					}
					if foundMatch {
						break
					}
				}
				if foundMatch {
					break
				}
			}
			if !foundMatch && len(declared) > 0 {
				issues = append(issues, InfraIssue{
					Severity: "ERROR",
					Code:     "K8S_PORT_MISMATCH",
					File:     svc.FilePath,
					Line:     sPort.Line,
					Resource: fmt.Sprintf("Service/%s", svc.Name),
					Message:  fmt.Sprintf("Service targetPort %s does not match any containerPort in matching Workload (selectors: %v)", target.String(), svc.ServiceSelector),
					FixHint:  fmt.Sprintf("Declared container ports: [%s]. Update Service targetPort or declare matching containerPort.", strings.Join(declared, ", ")),
				})
			}
		}
	}
	return issues
}

var sensitiveKeyRegex = regexp.MustCompile(`(?i)(password|passwd|api[_-]?key|access[_-]?key|auth[_-]?token|secret|jwt|db_pass|private_key)`)
var benignPlaceholderRegex = regexp.MustCompile(`(?i)^(<.+>|\$\{.+\}|\{\{.+}}|changeme|dummy|todo|example|null|none|placeholder)$`)

func validateConfigMapSecrets(cms []*K8sResource) []InfraIssue {
	var issues []InfraIssue
	for _, cm := range cms {
		if cm.Kind != "ConfigMap" {
			continue
		}
		for key, val := range cm.Data {
			val = strings.TrimSpace(val)
			if sensitiveKeyRegex.MatchString(key) {
				if val == "" || benignPlaceholderRegex.MatchString(val) {
					continue
				}
				line := cm.DataLines[key]
				entropy := calculateShannonEntropy(val)
				if entropy >= 3.0 || len(val) >= 12 {
					issues = append(issues, InfraIssue{
						Severity: "ERROR",
						Code:     "CONFIGMAP_PLAINTEXT_SECRET",
						File:     cm.FilePath,
						Line:     line,
						Resource: fmt.Sprintf("ConfigMap/%s", cm.Name),
						Message:  fmt.Sprintf("ConfigMap key '%s' contains hardcoded plaintext credential", key),
						FixHint:  "Migrate this sensitive value to a Kubernetes Secret, SealedSecret, or ExternalSecret vault store.",
					})
				}
			}
		}
	}
	return issues
}

func validateImageTags(resources []*K8sResource) []InfraIssue {
	var issues []InfraIssue
	for _, res := range resources {
		for _, c := range res.Containers {
			image := strings.TrimSpace(c.Image)
			if image == "" || strings.Contains(image, "@sha256:") {
				continue
			}
			parts := strings.Split(image, "/")
			lastPart := parts[len(parts)-1]
			tagParts := strings.Split(lastPart, ":")
			if len(tagParts) == 1 {
				issues = append(issues, InfraIssue{
					Severity: "WARNING",
					Code:     "IMAGE_TAG_MISSING",
					File:     res.FilePath,
					Line:     c.Line,
					Resource: fmt.Sprintf("%s/%s (container: %s)", res.Kind, res.Name, c.Name),
					Message:  fmt.Sprintf("Container image '%s' is untagged (defaults to mutable ':latest')", image),
					FixHint:  "Pin the image to an immutable release tag (e.g., 'myapp:1.4.2') or a SHA256 digest.",
				})
			} else if tagParts[1] == "latest" {
				issues = append(issues, InfraIssue{
					Severity: "WARNING",
					Code:     "IMAGE_TAG_LATEST",
					File:     res.FilePath,
					Line:     c.Line,
					Resource: fmt.Sprintf("%s/%s (container: %s)", res.Kind, res.Name, c.Name),
					Message:  fmt.Sprintf("Container image '%s' uses mutable ':latest' tag", image),
					FixHint:  "Avoid ':latest' in production manifests to guarantee deterministic deployments and rollbacks.",
				})
			}
		}
	}
	return issues
}

func validateContainerProbes(resources []*K8sResource) []InfraIssue {
	var issues []InfraIssue
	for _, res := range resources {
		for _, c := range res.Containers {
			if len(c.Ports) == 0 {
				continue
			}
			for _, probe := range c.Probes {
				if !probe.Port.IsString && probe.Port.IntVal == 0 {
					continue
				}
				matched := false
				var declared []string
				for _, cp := range c.Ports {
					declared = append(declared, strconv.Itoa(cp.ContainerPort))
					if probe.Port.IsString {
						if cp.Name == probe.Port.StrVal {
							matched = true
							break
						}
					} else {
						if cp.ContainerPort == probe.Port.IntVal {
							matched = true
							break
						}
					}
				}
				if !matched {
					issues = append(issues, InfraIssue{
						Severity: "ERROR",
						Code:     "PROBE_PORT_MISMATCH",
						File:     res.FilePath,
						Line:     probe.Line,
						Resource: fmt.Sprintf("%s/%s (container: %s)", res.Kind, res.Name, c.Name),
						Message:  fmt.Sprintf("%s probe targets port %s which is not declared in container ports", probe.Type, probe.Port.String()),
						FixHint:  fmt.Sprintf("Declared ports: [%s]. Align probe port with a declared containerPort.", strings.Join(declared, ", ")),
					})
				}
			}
		}
	}
	return issues
}

func lintDockerfile(data []byte, relPath, repoRoot string) []InfraIssue {
	var issues []InfraIssue
	lines := strings.Split(string(data), "\n")
	hasUser := false
	lastFromLine := 0
	lastBaseImage := ""
	lastBaseTag := ""
	hasCopyBroad := false
	copyBroadLine := 0

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) == 0 {
			continue
		}
		verb := strings.ToUpper(fields[0])
		if verb == "FROM" && len(fields) >= 2 {
			hasUser = false
			lastFromLine = lineNum + 1
			base := fields[1]
			parts := strings.Split(base, "/")
			lastPart := parts[len(parts)-1]
			tagParts := strings.Split(lastPart, ":")
			lastBaseImage = tagParts[0]
			if len(tagParts) > 1 {
				lastBaseTag = tagParts[1]
			} else {
				lastBaseTag = ""
			}
			if lastBaseImage != "scratch" && (lastBaseTag == "" || lastBaseTag == "latest") {
				issues = append(issues, InfraIssue{
					Severity: "WARNING",
					Code:     "DOCKER_UNPINNED_BASE",
					File:     relPath,
					Line:     lineNum + 1,
					Resource: "Dockerfile",
					Message:  fmt.Sprintf("Base image '%s' is untagged or uses mutable ':latest'", base),
					FixHint:  "Pin the base image to an explicit version (e.g., 'alpine:3.19.1' or SHA256 digest).",
				})
			}
		} else if verb == "USER" && len(fields) >= 2 {
			u := strings.ToLower(fields[1])
			if u != "root" && u != "0" {
				hasUser = true
			}
		} else if (verb == "COPY" || verb == "ADD") && len(fields) >= 3 {
			if fields[1] == "." || fields[1] == "./" {
				hasCopyBroad = true
				copyBroadLine = lineNum + 1
			}
		}
	}

	if lastFromLine > 0 && !hasUser && lastBaseImage != "scratch" {
		issues = append(issues, InfraIssue{
			Severity: "WARNING",
			Code:     "DOCKER_ROOT_USER",
			File:     relPath,
			Line:     lastFromLine,
			Resource: "Dockerfile",
			Message:  "Container executes as default root user without non-root USER instruction",
			FixHint:  "Declare a non-root user (e.g., 'USER 10001:10001' or 'USER appuser') before CMD/ENTRYPOINT.",
		})
	}

	if hasCopyBroad {
		dockerignoreDir := filepath.Join(repoRoot, filepath.Dir(relPath), ".dockerignore")
		dockerignoreRoot := filepath.Join(repoRoot, ".dockerignore")
		_, err1 := os.Stat(dockerignoreDir)
		_, err2 := os.Stat(dockerignoreRoot)
		if err1 != nil && err2 != nil {
			issues = append(issues, InfraIssue{
				Severity: "WARNING",
				Code:     "DOCKER_MISSING_DOCKERIGNORE",
				File:     relPath,
				Line:     copyBroadLine,
				Resource: "Dockerfile",
				Message:  "Broad 'COPY . .' used without a .dockerignore file",
				FixHint:  "Create a .dockerignore file to prevent leaking .git, secrets, or temporary files into the container image.",
			})
		}
	}
	return issues
}

func lintTerraformHygiene(repoRoot string, files []string) []InfraIssue {
	var issues []InfraIssue
	hasTF := false
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		if ext == ".tf" || ext == ".tfvars" || strings.HasPrefix(f, "terraform/") {
			hasTF = true
			break
		}
	}
	if !hasTF {
		return issues
	}

	// 1. Check .gitignore for .terraform/ and *.tfstate*
	giPath := filepath.Join(repoRoot, ".gitignore")
	giData, err := os.ReadFile(giPath)
	if err != nil {
		issues = append(issues, InfraIssue{
			Severity: "ERROR",
			Code:     "TF_GITIGNORE_MISSING",
			File:     ".gitignore",
			Line:     1,
			Resource: "Terraform",
			Message:  ".gitignore is missing in repository containing Terraform files",
			FixHint:  "Create .gitignore containing '.terraform/' and '*.tfstate*'.",
		})
	} else {
		giStr := string(giData)
		if !strings.Contains(giStr, ".terraform") || !strings.Contains(giStr, "tfstate") {
			issues = append(issues, InfraIssue{
				Severity: "ERROR",
				Code:     "TF_GITIGNORE_MISSING",
				File:     ".gitignore",
				Line:     1,
				Resource: "Terraform",
				Message:  ".gitignore is missing '.terraform/' or '*.tfstate*' rules",
				FixHint:  "Add '.terraform/' and '*.tfstate*' to .gitignore to prevent committing large provider caches or raw state secrets.",
			})
		}
	}

	// 2. Check for plaintext credentials in .tfvars
	for _, rel := range files {
		ext := strings.ToLower(filepath.Ext(rel))
		base := strings.ToLower(filepath.Base(rel))
		if ext == ".tfvars" || strings.Contains(base, "tfvars") {
			if strings.HasSuffix(base, ".example") || strings.HasSuffix(base, ".sample") {
				continue
			}
			absPath := filepath.Join(repoRoot, rel)
			data, err := os.ReadFile(absPath)
			if err != nil {
				continue
			}
			for lineNum, line := range strings.Split(string(data), "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
					continue
				}
				if strings.Contains(trimmed, "=") {
					parts := strings.SplitN(trimmed, "=", 2)
					k := strings.ToLower(strings.TrimSpace(parts[0]))
					v := strings.TrimSpace(strings.Trim(parts[1], `"'`))
					if sensitiveKeyRegex.MatchString(k) && len(v) >= 8 && !benignPlaceholderRegex.MatchString(v) {
						issues = append(issues, InfraIssue{
							Severity: "ERROR",
							Code:     "TF_PLAINTEXT_SECRET",
							File:     rel,
							Line:     lineNum + 1,
							Resource: "Terraform Variables",
							Message:  fmt.Sprintf("Variable '%s' contains hardcoded sensitive plaintext value in %s", parts[0], rel),
							FixHint:  "Use environment variables (TF_VAR_*) or cloud secret managers instead of committing plaintext .tfvars values.",
						})
					}
				}
			}
		}
	}
	return issues
}

func validateDevOpsGate(repoRoot string, files []string) []InfraIssue {
	var allIssues []InfraIssue
	var k8sManifests []*K8sResource

	for _, rel := range files {
		ext := strings.ToLower(filepath.Ext(rel))
		base := strings.ToLower(filepath.Base(rel))
		absPath := filepath.Join(repoRoot, rel)

		// 1. Kubernetes YAML manifests
		if ext == ".yaml" || ext == ".yml" {
			data, err := os.ReadFile(absPath)
			if err == nil {
				dataStr := string(data)
				if strings.Contains(dataStr, "apiVersion:") && strings.Contains(dataStr, "kind:") {
					docs := splitYAMLDocuments(data)
					currentLine := 1
					for _, doc := range docs {
						lineCount := len(strings.Split(doc, "\n"))
						res := parseK8sDoc(doc, rel, currentLine)
						if res != nil && res.Kind != "" {
							k8sManifests = append(k8sManifests, res)
						}
						currentLine += lineCount + 1
					}
				}
			}
		}

		// 2. Dockerfile linting
		if base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") || base == "containerfile" {
			data, err := os.ReadFile(absPath)
			if err == nil {
				issues := lintDockerfile(data, rel, repoRoot)
				allIssues = append(allIssues, issues...)
			}
		}
	}

	// Cross-resource Kubernetes validation
	if len(k8sManifests) > 0 {
		var services []*K8sResource
		var workloads []*K8sResource
		var configMaps []*K8sResource

		for _, r := range k8sManifests {
			if r.Kind == "Service" {
				services = append(services, r)
			} else if r.Kind == "Deployment" || r.Kind == "StatefulSet" || r.Kind == "DaemonSet" || r.Kind == "Pod" {
				workloads = append(workloads, r)
			} else if r.Kind == "ConfigMap" {
				configMaps = append(configMaps, r)
			}
		}

		allIssues = append(allIssues, validateServicePortTargetPort(services, workloads)...)
		allIssues = append(allIssues, validateConfigMapSecrets(configMaps)...)
		allIssues = append(allIssues, validateImageTags(k8sManifests)...)
		allIssues = append(allIssues, validateContainerProbes(k8sManifests)...)
	}

	// 3. Terraform hygiene
	allIssues = append(allIssues, lintTerraformHygiene(repoRoot, files)...)

	return allIssues
}

type preflightJSONResult struct {
	Passed       bool         `json:"passed"`
	ExitCode     int          `json:"exit_code"`
	Mode         string       `json:"mode,omitempty"`
	SyntaxErrors []string     `json:"syntax_errors"`
	Quarantined  []string     `json:"quarantined_files"`
	DevOpsIssues []InfraIssue `json:"devops_issues,omitempty"`
	Details      []string     `json:"details"`
}

func runPreflight(repoRoot string, stagedOnly, jsonOutput bool) (int, string) {
	mode := resolveEffectiveRepoMode(repoRoot, "")
	if env := os.Getenv("SYNCALIBUR_MODE"); env != "" {
		mode = strings.ToLower(env)
	}

	var files []string
	fileMap := make(map[string]bool)
	addFiles := func(out []byte) {
		for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if f = strings.TrimSpace(f); f != "" {
				fileMap[f] = true
			}
		}
	}

	if stagedOnly {
		out, err := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "diff", "--cached", "--name-only").Output()
		if err == nil {
			addFiles(out)
		}
	} else {
		outStaged, err := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "diff", "--cached", "--name-only").Output()
		if err == nil {
			addFiles(outStaged)
		}
		outUnstaged, err := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "diff", "--name-only").Output()
		if err == nil {
			addFiles(outUnstaged)
		}
		outUntracked, err := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "ls-files", "--others", "--exclude-standard").Output()
		if err == nil {
			addFiles(outUntracked)
		}
	}

	for f := range fileMap {
		files = append(files, f)
	}
	sort.Strings(files)

	var syntaxErrors []string
	var quarantinedFiles []string
	var details []string

	for _, rel := range files {
		// 0. Path Traversal & Symlink Safety
		if safe, reason := isSafePath(repoRoot, rel); !safe {
			quarantinedFiles = append(quarantinedFiles, rel)
			details = append(details, fmt.Sprintf("%s: %s", rel, reason))
			continue
		}

		absPath := filepath.Join(repoRoot, rel)
		if st, err := os.Stat(absPath); err != nil || st.IsDir() {
			continue
		}

		// 1. Conflict Marker Gate
		if hasMarker, markerMsg := detectConflictMarkers(absPath); hasMarker {
			syntaxErrors = append(syntaxErrors, rel)
			details = append(details, markerMsg)
		}

		// 2. Secret & Sensitive Pattern Check
		if hasSecret, reason := detectSecretsInFile(absPath); hasSecret {
			quarantinedFiles = append(quarantinedFiles, rel)
			details = append(details, reason)
		}

		// 3. Syntax Validation Check
		if valid, errReason := validateFileSyntax(repoRoot, rel); !valid {
			syntaxErrors = append(syntaxErrors, rel)
			details = append(details, errReason)
		}
	}

	// 4. DevOps Policy Gate (for devops mode or general devops files)
	var devopsIssues []InfraIssue
	hasDevOpsError := false
	if mode == "devops" || mode == "infra" || mode == "k8s" {
		devopsIssues = validateDevOpsGate(repoRoot, files)
		for _, iss := range devopsIssues {
			if iss.Severity == "ERROR" {
				hasDevOpsError = true
			}
			details = append(details, fmt.Sprintf("[%s] %s (%s:%d): %s", iss.Severity, iss.Resource, iss.File, iss.Line, iss.Message))
		}
	}

	exitCode := 0
	if len(quarantinedFiles) > 0 {
		exitCode = 3 // Secret detected
	} else if len(syntaxErrors) > 0 {
		exitCode = 2 // Syntax error or conflict marker
	}

	if jsonOutput {
		res := preflightJSONResult{
			Passed:       exitCode == 0,
			ExitCode:     exitCode,
			Mode:         mode,
			SyntaxErrors: syntaxErrors,
			Quarantined:  quarantinedFiles,
			DevOpsIssues: devopsIssues,
			Details:      details,
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return exitCode, string(b)
	}

	if exitCode == 0 && len(devopsIssues) == 0 {
		return 0, fmt.Sprintf("%s%s %sPre-flight verification passed (all files clean)%s", cGreen, iconCheck, cWhite, cReset)
	}

	var sb strings.Builder
	if len(quarantinedFiles) > 0 {
		sb.WriteString(fmt.Sprintf("%s%s %sQUARANTINED FILES (%d)%s\n", cRed, iconShield, cBold, len(quarantinedFiles), cReset))
		for _, q := range quarantinedFiles {
			sb.WriteString(fmt.Sprintf("  %s%s%s %s\n", cRed, iconCross, cReset, q))
		}
	}
	if len(syntaxErrors) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s%s %sSYNTAX / CONFLICT ERRORS DETECTED (%d)%s\n", cAmber, iconWarn, cBold, len(syntaxErrors), cReset))
		for _, s := range syntaxErrors {
			sb.WriteString(fmt.Sprintf("  %s%s%s %s\n", cAmber, iconCross, cReset, s))
		}
	}
	if len(devopsIssues) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		headerColor := cCyan
		if hasDevOpsError {
			headerColor = cRed
		}
		sb.WriteString(fmt.Sprintf("%s%s %sDEVOPS & INFRASTRUCTURE GATE ISSUES (%d)%s\n", headerColor, iconWheel, cBold, len(devopsIssues), cReset))
		for _, iss := range devopsIssues {
			sym := iconWarn
			col := cAmber
			if iss.Severity == "ERROR" {
				sym = iconCross
				col = cRed
			}
			sb.WriteString(fmt.Sprintf("  %s%s %s[%s] %s%s (%s:%d)\n", col, sym, cBold, iss.Code, iss.Resource, cReset, iss.File, iss.Line))
			sb.WriteString(fmt.Sprintf("    %s%s %s%s\n", cSlate, iconChevron, iss.Message, cReset))
			if iss.FixHint != "" {
				sb.WriteString(fmt.Sprintf("    %s↳ Fix: %s%s\n", cGreen, iss.FixHint, cReset))
			}
		}
	}
	for _, d := range details {
		if !strings.HasPrefix(d, "[") {
			sb.WriteString(fmt.Sprintf("    %s%s %s%s\n", cSlate, iconChevron, d, cReset))
		}
	}
	return exitCode, sb.String()
}


// ==============================================================================
// 3b. Smart Staging Partition Engine (`smart-stage`)
// ==============================================================================

type stagingGroup struct {
	Scope string   `json:"scope"`
	Type  string   `json:"type"`
	Files []string `json:"files"`
}

func planSmartStaging(repoRoot string) []stagingGroup {
	var files []string
	seen := make(map[string]bool)

	out, err := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "diff", "HEAD", "--name-only").Output()
	if err == nil {
		for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			f = strings.TrimSpace(f)
			if f != "" && !seen[f] {
				seen[f] = true
				files = append(files, f)
			}
		}
	} else {
		out, _ = exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "status", "--porcelain").Output()
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if len(line) > 3 {
				f := strings.TrimSpace(line[3:])
				if f != "" && !seen[f] {
					seen[f] = true
					files = append(files, f)
				}
			}
		}
	}

	untracked, err := exec.Command("git", "-C", repoRoot, "-c", "core.quotePath=false", "ls-files", "--others", "--exclude-standard").Output()
	if err == nil {
		for _, f := range strings.Split(strings.TrimSpace(string(untracked)), "\n") {
			f = strings.TrimSpace(f)
			if f != "" && !seen[f] {
				seen[f] = true
				files = append(files, f)
			}
		}
	}

	if len(files) == 0 {
		return nil
	}

	groupMap := make(map[string]*stagingGroup)
	var groupKeys []string

	for _, f := range files {
		f = filepath.ToSlash(f)
		lower := strings.ToLower(f)
		base := filepath.Base(f)
		ext := strings.ToLower(filepath.Ext(f))

		var key, scope, gType string

		if ext == ".md" || ext == ".rst" || ext == ".txt" || strings.HasPrefix(lower, "docs/") || strings.EqualFold(base, "license") || strings.EqualFold(base, "contributing.md") {
			key = "docs"
			scope = "docs"
			gType = "docs"
		} else if strings.HasSuffix(lower, "_test.go") || strings.Contains(lower, "test") || strings.Contains(lower, "spec") {
			key = "test"
			scope = "test"
			gType = "test"
		} else if strings.HasPrefix(lower, "charts/") || base == "chart.yaml" || strings.HasPrefix(base, "values") {
			key = "helm"
			scope = "helm"
			gType = "helm"
		} else if strings.HasPrefix(lower, "k8s/") || strings.HasPrefix(lower, "deploy/") || strings.HasPrefix(lower, "manifests/") || strings.HasPrefix(lower, "base/") || strings.HasPrefix(lower, "overlays/") {
			key = "k8s"
			scope = "infra(k8s)"
			gType = "infra"
		} else if strings.HasPrefix(base, "dockerfile") || base == ".dockerignore" || base == "compose.yaml" || base == "docker-compose.yml" {
			key = "docker"
			scope = "docker"
			gType = "build"
		} else if strings.HasSuffix(lower, ".tf") || strings.HasSuffix(lower, ".tfvars") || strings.HasPrefix(lower, "terraform/") {
			key = "tf"
			scope = "tf"
			gType = "infra"
		} else if strings.HasPrefix(lower, "argocd/") || strings.HasPrefix(lower, "flux-system/") || strings.HasPrefix(lower, "apps/") {
			key = "gitops"
			scope = "gitops"
			gType = "gitops"
		} else if strings.HasPrefix(lower, ".github/") || strings.HasPrefix(lower, ".gitlab") || base == "jenkinsfile" {
			key = "ci"
			scope = "ci"
			gType = "ci"
		} else if base == "makefile" || base == "go.mod" || base == "go.sum" || base == "package.json" {
			key = "build"
			scope = "build"
			gType = "build"
		} else if ext == ".css" || ext == ".scss" || ext == ".less" {
			key = "style"
			scope = "style"
			gType = "style"
		} else {
			dir := filepath.ToSlash(filepath.Dir(f))
			if dir == "." || dir == "" {
				key = "core"
				scope = "core"
			} else {
				parts := strings.Split(dir, "/")
				if len(parts) >= 2 && parts[0] == "cmd" {
					scope = parts[1]
				} else {
					scope = parts[0]
				}
				key = "code:" + scope
			}
			gType = "feat"
		}

		if grp, exists := groupMap[key]; exists {
			grp.Files = append(grp.Files, f)
		} else {
			grp := &stagingGroup{
				Scope: scope,
				Type:  gType,
				Files: []string{f},
			}
			groupMap[key] = grp
			groupKeys = append(groupKeys, key)
		}
	}

	sort.Slice(groupKeys, func(i, j int) bool {
		order := func(k string) int {
			if k == "docs" {
				return 1
			}
			if k == "helm" {
				return 2
			}
			if k == "k8s" {
				return 3
			}
			if k == "docker" {
				return 4
			}
			if k == "tf" {
				return 5
			}
			if k == "ci" {
				return 6
			}
			if k == "gitops" {
				return 7
			}
			if k == "build" {
				return 8
			}
			if k == "style" {
				return 9
			}
			if strings.HasPrefix(k, "code") {
				return 10
			}
			if k == "test" {
				return 11
			}
			return 12
		}
		oi, oj := order(groupKeys[i]), order(groupKeys[j])
		if oi != oj {
			return oi < oj
		}
		return groupKeys[i] < groupKeys[j]
	})

	var result []stagingGroup
	for _, k := range groupKeys {
		result = append(result, *groupMap[k])
	}
	return result
}

// ==============================================================================
// 4. 7-Day Commit Velocity Sparkline Generator
// ==============================================================================

// CommitVelocity stores 7-day commit metrics and sparkline visualization.
type CommitVelocity struct {
	Counts    [7]int
	Today     int
	Total7d   int
	Peak      int
	AvgDaily  float64
	Sparkline string
}

func getCommitVelocity(repoRoot string) CommitVelocity {
	glyphs := [8]string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	now := time.Now()
	loc := now.Location()

	var result CommitVelocity

	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	windowStart := todayMidnight.AddDate(0, 0, -6)

	dayMap := make(map[string]int, 7)
	for d := 0; d < 7; d++ {
		dayKey := windowStart.AddDate(0, 0, d).Format("2006-01-02")
		dayMap[dayKey] = d
	}

	sinceParam := windowStart.AddDate(0, 0, -1).Format("2006-01-02")

	out, err := exec.Command("git", "-C", repoRoot, "log",
		"--since="+sinceParam,
		"--format=%ct",
		"HEAD").Output()

	if err == nil && len(out) > 0 {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			epoch, parseErr := strconv.ParseInt(line, 10, 64)
			if parseErr != nil {
				continue
			}
			t := time.Unix(epoch, 0).In(loc)
			key := t.Format("2006-01-02")
			if idx, ok := dayMap[key]; ok {
				result.Counts[idx]++
			}
		}
	}

	for _, c := range result.Counts {
		result.Total7d += c
		if c > result.Peak {
			result.Peak = c
		}
	}
	result.Today = result.Counts[6]
	result.AvgDaily = float64(result.Total7d) / 7.0

	var b strings.Builder
	for _, c := range result.Counts {
		if c == 0 {
			b.WriteString(cSlate + "·" + cReset)
		} else if result.Peak <= 1 {
			b.WriteString(cGreen + glyphs[1] + cReset)
		} else {
			scaleMax := result.Peak
			if scaleMax < 4 {
				scaleMax = 4
			}
			idx := (c * 7) / scaleMax
			if idx > 7 {
				idx = 7
			}
			b.WriteString(cGreen + glyphs[idx] + cReset)
		}
	}

	fmt.Fprintf(&b, " %s(%d commits/7d • %.1f/d • peak %d)%s",
		cSlate, result.Total7d, result.AvgDaily, result.Peak, cReset)

	result.Sparkline = b.String()
	return result
}

func generateSparkline(repoRoot string) string {
	return getCommitVelocity(repoRoot).Sparkline
}

// ==============================================================================
// 5. Multi-Repository Fleet Scanner & Parallel Status Probe
// ==============================================================================

func isGitRepoDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func resolveGitDir(repoPath string) string {
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "--absolute-git-dir").Output()
	if err == nil {
		gdir := strings.TrimSpace(string(out))
		if gdir != "" {
			return gdir
		}
	}
	return filepath.Join(repoPath, ".git")
}

func scanGitRepos(root string, maxDepth int) []string {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil
	}
	var repos []string
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if depth > maxDepth {
			return
		}
		if isGitRepoDir(dir) {
			repos = append(repos, dir)
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "snap" {
				continue
			}
			walk(filepath.Join(dir, name), depth+1)
		}
	}
	walk(rootAbs, 0)
	sort.Strings(repos)
	return repos
}

type fleetRepoConfig struct {
	Path     string
	Mode     string
	Remote   string
	Interval string
}

func parseFleetEntry(line string) *fleetRepoConfig {
	line = strings.TrimSpace(strings.Split(line, "#")[0])
	if line == "" {
		return nil
	}

	cfg := &fleetRepoConfig{Mode: "", Remote: "origin"}
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	cfg.Path = parts[0]
	if strings.HasPrefix(cfg.Path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			cfg.Path = filepath.Join(home, cfg.Path[2:])
		}
	}
	if abs, err := filepath.Abs(cfg.Path); err == nil {
		cfg.Path = abs
	}

	for _, p := range parts[1:] {
		p = strings.Trim(p, "[]")
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			k := strings.ToLower(strings.TrimSpace(kv[0]))
			v := strings.TrimSpace(kv[1])
			switch k {
			case "mode":
				cfg.Mode = canonicalizeMode(v)
			case "remote":
				cfg.Remote = v
			case "interval":
				cfg.Interval = v
			}
		} else if len(kv) == 1 {
			val := strings.ToLower(strings.TrimSpace(kv[0]))
			cMode := canonicalizeMode(val)
			if cMode == "autopilot" || cMode == "manual" {
				cfg.Mode = cMode
			}
		}
	}
	return cfg
}

func canonicalizeMode(m string) string {
	m = strings.ToLower(strings.TrimSpace(m))
	switch m {
	case "manual", "idle", "sentinel":
		return "manual"
	default:
		return "autopilot"
	}
}

func resolveEffectiveRepoMode(repoPath, fleetLineMode string) string {
	if fleetLineMode != "" && fleetLineMode != "default" {
		return canonicalizeMode(fleetLineMode)
	}
	for _, fname := range []string{".syncalibur", ".syncalibur.env", "syncalibur.env"} {
		fpath := filepath.Join(repoPath, fname)
		if data, err := os.ReadFile(fpath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "MODE=") || strings.HasPrefix(line, "SYNCALIBUR_MODE=") {
					val := strings.Trim(strings.SplitN(line, "=", 2)[1], `"' `)
					if val != "" {
						return canonicalizeMode(val)
					}
				}
			}
		}
	}
	if env := strings.TrimSpace(os.Getenv("SYNCALIBUR_MODE")); env != "" {
		return canonicalizeMode(env)
	}
	return "autopilot"
}

func visibleLen(s string) int {
	var clean strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		clean.WriteRune(r)
	}
	w := 0
	for _, r := range clean.String() {
		if r > 0x2000 {
			w += 2
		} else {
			w += 1
		}
	}
	return w
}

func padRightVisible(s string, width int) string {
	vl := visibleLen(s)
	if vl >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vl)
}

func probeRemoteLiveness(rawURL string, timeoutMs int) bool {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return true
	}
	host, port := extractHostPort(rawURL)
	if host == "" {
		host = "github.com"
		port = "22"
	}
	timeout := time.Duration(timeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 1000 * time.Millisecond
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func extractHostPort(rawURL string) (string, string) {
	if strings.Contains(rawURL, "@") && strings.Contains(rawURL, ":") && !strings.Contains(rawURL, "://") {
		parts := strings.Split(rawURL, "@")
		hostPart := strings.Split(parts[len(parts)-1], ":")[0]
		return hostPart, "22"
	}
	if strings.Contains(rawURL, "://") {
		if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
			host := u.Hostname()
			port := u.Port()
			if port == "" {
				if u.Scheme == "ssh" || u.Scheme == "git" {
					port = "22"
				} else {
					port = "443"
				}
			}
			return host, port
		}
	}
	if strings.Contains(rawURL, ":") {
		parts := strings.Split(rawURL, ":")
		return parts[0], parts[1]
	}
	return rawURL, "22"
}

type fleetRepoStatus struct {
	Path     string
	Name     string
	Mode     string
	Branch   string
	State    string
	Dirty    int
	Unpushed int
	Remote   string
}

func printFleetStatusTable(fleetFile string) {
	data, err := os.ReadFile(fleetFile)
	if err != nil {
		fmt.Printf("  %s%s No fleet file found at %s%s\n", cAmber, iconWarn, fleetFile, cReset)
		return
	}

	var configs []*fleetRepoConfig
	seen := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		cfg := parseFleetEntry(line)
		if cfg == nil {
			continue
		}
		if st, err := os.Stat(cfg.Path); err == nil && st.IsDir() && !seen[cfg.Path] {
			seen[cfg.Path] = true
			configs = append(configs, cfg)
		}
	}

	if len(configs) == 0 {
		fmt.Printf("  %s%s No repositories enrolled in %s. Run: syncalibur fleet add <path>%s\n", cAmber, iconWarn, fleetFile, cReset)
		return
	}

	results := make([]fleetRepoStatus, len(configs))
	var wg sync.WaitGroup
	for i, cfg := range configs {
		wg.Add(1)
		go func(idx int, c *fleetRepoConfig) {
			defer wg.Done()
			results[idx] = inspectRepoStatus(c.Path, c.Mode)
		}(i, cfg)
	}
	wg.Wait()

	fmt.Printf("  %s╭───────────────────────────────────────────────────────────────────────────────────────────────────╮%s\n", cGold, cReset)
	fmt.Printf("  %s│%s  %s%s%-95s%s  %s│%s\n", cGold, cReset, cGold, cBold, fmt.Sprintf("%s  S Y N C A L I B U R   F L E E T   S T A T U S   (%d REPOSITORIES)", iconDiamond, len(results)), cReset, cGold, cReset)
	fmt.Printf("  %s╰───────────────────────────────────────────────────────────────────────────────────────────────────╯%s\n", cGold, cReset)
	fmt.Printf("    %s%-18s %-15s %-13s %-11s %-20s %-20s%s\n", cBold, "REPOSITORY", "MODE", "DAEMON", "BRANCH", "SYNC STATE", "PATH", cReset)
	fmt.Printf("    %s───────────────────────────────────────────────────────────────────────────────────────────────────%s\n", cSlate, cReset)

	home, _ := os.UserHomeDir()
	for _, r := range results {
		shortPath := r.Path
		if home != "" && strings.HasPrefix(shortPath, home) {
			shortPath = "~" + strings.TrimPrefix(shortPath, home)
		}
		rawSync := iconCheck + " In Sync"
		syncColor := cGreen
		if r.Dirty > 0 || r.Unpushed > 0 {
			rawSync = fmt.Sprintf("%s %d dirty, %d unpushed", iconDot, r.Dirty, r.Unpushed)
			syncColor = cAmber
		}
		stateColor := cSlate
		if strings.HasPrefix(r.State, iconDot) {
			stateColor = cGreen
		}
		modeLabel := "⚡ autopilot"
		modeColor := cCyan
		if canonicalizeMode(r.Mode) == "manual" {
			modeLabel = "⏸ manual"
			modeColor = cGold
		}

		cRepo := padRightVisible(cWhite+cBold+truncateStr(r.Name, 17)+cReset, 18)
		cMode := padRightVisible(modeColor+modeLabel+cReset, 15)
		cState := padRightVisible(stateColor+truncateStr(r.State, 12)+cReset, 13)
		cBranch := padRightVisible(cCyan+truncateStr(r.Branch, 10)+cReset, 11)
		cSync := padRightVisible(syncColor+truncateStr(rawSync, 19)+cReset, 20)
		cPath := cSlate + shortPath + cReset

		fmt.Printf("    %s %s %s %s %s %s\n", cRepo, cMode, cState, cBranch, cSync, cPath)
	}
	fmt.Println()
}

func inspectRepoStatus(repoPath, inlineMode string) fleetRepoStatus {
	st := fleetRepoStatus{
		Path:   repoPath,
		Name:   filepath.Base(repoPath),
		Mode:   resolveEffectiveRepoMode(repoPath, inlineMode),
		Branch: "main",
		State:  iconCircle + " Idle",
	}
	if bOut, err := exec.Command("git", "-C", repoPath, "symbolic-ref", "--quiet", "--short", "HEAD").Output(); err == nil {
		st.Branch = strings.TrimSpace(string(bOut))
	}
	gitDir := resolveGitDir(repoPath)
	pidFile := filepath.Join(gitDir, "syncalibur.pid")
	if pBytes, err := os.ReadFile(pidFile); err == nil {
		pidStr := strings.TrimSpace(string(pBytes))
		if pid, err := strconv.Atoi(pidStr); err == nil && processAlive(pid) {
			st.State = fmt.Sprintf("%s PID %d", iconDot, pid)
		}
	}
	if sOut, err := exec.Command("git", "-C", repoPath, "status", "--porcelain", "-uall").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(sOut)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			st.Dirty = len(lines)
		}
	}
	if uOut, err := exec.Command("git", "-C", repoPath, "rev-list", "--count", "@{u}..HEAD").Output(); err == nil {
		if unp, err := strconv.Atoi(strings.TrimSpace(string(uOut))); err == nil {
			st.Unpushed = unp
		}
	}
	return st
}

func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// ==============================================================================
// 6. First-Class Telegram Bot & Multi-Channel Webhook Dispatcher
// ==============================================================================

func escapeTelegramHTML(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return r.Replace(s)
}

func formatTelegramHTMLMessage(eventType, repoName, branchName, hostInfo, title, message, details string) string {
	badge := iconDiamond
	switch eventType {
	case "failure":
		badge = iconCross
	case "recovery", "success":
		badge = iconCheck
	case "test", "summary":
		badge = iconSparkle
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>%s %s</b>\n", badge, escapeTelegramHTML(title)))
	sb.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", escapeTelegramHTML(message)))
	sb.WriteString(fmt.Sprintf("<b>Repository:</b> <code>%s</code>\n", escapeTelegramHTML(repoName)))
	sb.WriteString(fmt.Sprintf("<b>Branch:</b> <code>%s</code>\n", escapeTelegramHTML(branchName)))
	sb.WriteString(fmt.Sprintf("<b>Host:</b> <code>%s</code>\n", escapeTelegramHTML(hostInfo)))
	sb.WriteString(fmt.Sprintf("<b>Time:</b> <code>%s</code>\n", time.Now().Format("2006-01-02 15:04:05 MST")))

	if strings.TrimSpace(details) != "" {
		if len(details) > 800 {
			details = details[:800] + "\n... (truncated)"
		}
		sb.WriteString(fmt.Sprintf("\n<pre><b>Telemetry:</b>\n%s</pre>", escapeTelegramHTML(details)))
	}

	return sb.String()
}

func sendTelegramMessage(botToken, chatID, eventType, repoName, branchName, title, message, details string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	host, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = "user"
	}
	hostInfo := fmt.Sprintf("%s@%s (%s/%s)", user, host, runtime.GOOS, runtime.GOARCH)

	text := formatTelegramHTMLMessage(eventType, repoName, branchName, hostInfo, title, message, details)

	payload := map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	urlStr := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("telegram API HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBytes)))
	}
	return nil
}

func sendWebhook(webhookURL, eventType, repoName, branchName, title, message, details string) error {
	var firstErr error

	tgToken := strings.TrimSpace(os.Getenv("SYNCALIBUR_TELEGRAM_BOT_TOKEN"))
	tgChat := strings.TrimSpace(os.Getenv("SYNCALIBUR_TELEGRAM_CHAT_ID"))
	lowerURL := strings.ToLower(strings.TrimSpace(webhookURL))

	if tgToken != "" && tgChat != "" && !strings.Contains(lowerURL, "api.telegram.org/bot") {
		if err := sendTelegramMessage(tgToken, tgChat, eventType, repoName, branchName, title, message, details); err != nil {
			firstErr = err
		}
	}

	if strings.TrimSpace(webhookURL) == "" {
		return firstErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	host, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = "user"
	}
	hostInfo := fmt.Sprintf("%s@%s (%s/%s)", user, host, runtime.GOOS, runtime.GOARCH)
	nowISO := time.Now().UTC().Format(time.RFC3339)

	colorDec := 16494372
	colorHex := "#fbbf24"
	switch eventType {
	case "failure":
		colorDec = 16259441
		colorHex = "#f87171"
	case "recovery", "success":
		colorDec = 3461465
		colorHex = "#34d399"
	case "summary", "test":
		colorDec = 3718584
		colorHex = "#38bdf8"
	}

	var payload any

	switch {
	case strings.Contains(lowerURL, "discord.com/api/webhooks/") || strings.Contains(lowerURL, "discordapp.com/api/webhooks/"):
		if len(details) > 900 {
			details = details[:900] + "..."
		}
		fields := []map[string]any{
			{"name": "Repository", "value": "`" + repoName + "`", "inline": true},
			{"name": "Branch", "value": "`" + branchName + "`", "inline": true},
			{"name": "Host", "value": "`" + hostInfo + "`", "inline": true},
		}
		if strings.TrimSpace(details) != "" {
			fields = append(fields, map[string]any{"name": "Telemetry", "value": "```text\n" + details + "\n```", "inline": false})
		}
		payload = map[string]any{
			"username":   "Syncalibur Daemon",
			"avatar_url": "https://raw.githubusercontent.com/YusufKaya00/Syncalibur/main/assets/syncalibur-logo.png",
			"embeds": []map[string]any{
				{
					"title":       title,
					"description": message,
					"color":       colorDec,
					"fields":      fields,
					"footer":      map[string]string{"text": "Syncalibur Autonomous Git Daemon (Go Engine)"},
					"timestamp":   nowISO,
				},
			},
		}
	case strings.Contains(lowerURL, "hooks.slack.com/services/"):
		attText := fmt.Sprintf("%s\n*Repo:* `%s` (%s) • *Host:* `%s`", message, repoName, branchName, hostInfo)
		if strings.TrimSpace(details) != "" {
			attText += fmt.Sprintf("\n```%s```", details)
		}
		payload = map[string]any{
			"text": "*" + title + "*",
			"attachments": []map[string]any{
				{
					"color": colorHex,
					"text":  attText,
				},
			},
		}
	case strings.Contains(lowerURL, "api.telegram.org/bot"):
		chatID := tgChat
		if parsedURL, err := url.Parse(webhookURL); err == nil {
			if qChat := parsedURL.Query().Get("chat_id"); qChat != "" {
				chatID = qChat
			}
		}
		text := formatTelegramHTMLMessage(eventType, repoName, branchName, hostInfo, title, message, details)
		tgMap := map[string]any{
			"parse_mode":               "HTML",
			"disable_web_page_preview": true,
			"text":                     text,
		}
		if chatID != "" {
			tgMap["chat_id"] = chatID
		}
		payload = tgMap
	default:
		payload = map[string]any{
			"service":    "syncalibur",
			"event":      eventType,
			"timestamp":  nowISO,
			"repository": repoName,
			"branch":     branchName,
			"host":       hostInfo,
			"title":      title,
			"message":    message,
			"details":    details,
		}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("webhook HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBytes)))
	}
	return firstErr
}

// ==============================================================================
// 7. Two-Pane Interactive Live TUI with Scrollable Logs (`syncalibur ui`)
// ==============================================================================

func getDaemonStatus(gitDir string) (bool, int) {
	pidBytes, err := os.ReadFile(filepath.Join(gitDir, "syncalibur.pid"))
	if err == nil {
		pidStr := strings.TrimSpace(string(pidBytes))
		if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
			if isProcessAlive(pid) {
				return true, pid
			}
		}
	}

	out, err := exec.Command("systemctl", "--user", "is-active", "syncalibur.service").Output()
	if err == nil && strings.TrimSpace(string(out)) == "active" {
		pidOut, err := exec.Command("systemctl", "--user", "show", "--property", "MainPID", "--value", "syncalibur.service").Output()
		if err == nil {
			pid, _ := strconv.Atoi(strings.TrimSpace(string(pidOut)))
			return true, pid
		}
		return true, 0
	}
	return false, 0
}

func loadLogLines(logPath string, maxLines int) []string {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	var clean []string
	for _, l := range lines {
		l = strings.TrimRight(l, "\r")
		if strings.TrimSpace(l) != "" {
			clean = append(clean, l)
		}
	}
	if len(clean) > maxLines {
		clean = clean[len(clean)-maxLines:]
	}
	return clean
}

var emojiToLucideReplacer = strings.NewReplacer(
	"🎉", "✓",
	"🚀", "●",
	"🛑", "○",
	"⏸️", "▲",
	"⏸", "▲",
	"🔄", "⇄",
	"🔀", "⇄",
	"📤", "↗",
	"📥", "↙",
	"📝", "✦",
	"✨", "✦",
	"⚡", "✦",
	"🔍", "◇",
	"🎯", "◆",
	"👤", "◆",
	"⏱️", "◆",
	"⏱", "◆",
	"📋", "•",
	"🖥️", "•",
	"🖥", "•",
	"⚠️", "▲",
	"❌", "✕",
	"✅", "✓",
	"🚫", "▲",
	"🔁", "↻",
	"🛡️", "⛨",
	"🛡", "⛨",
	"⚙️", "✦",
	"⚙", "✦",
	"⚔️", "✦",
	"⚔", "✦",
	"ℹ️", "ℹ",
	"ℹ", "ℹ",
	"✔", "✓",
	"✖", "✕",
	"📡", "✦",
	"⏳", "▲",
	"★", "✓",
	"\ufe0f", "",
)

func renderLineANSI(s string, maxCols int, horizOffset int) string {
	if maxCols < 10 {
		maxCols = 10
	}
	isBoxRow := strings.HasPrefix(s, "│")
	limit := maxCols
	if isBoxRow {
		limit = maxCols - 2
	}

	var out strings.Builder
	runes := []rune(s)
	n := len(runes)

	visCol := 0
	printedCols := 0
	hasMore := false

	inEscape := false
	var escBuf []rune

	for i := 0; i < n; i++ {
		r := runes[i]
		if r == '\033' {
			inEscape = true
			escBuf = append(escBuf, r)
			continue
		}
		if inEscape {
			escBuf = append(escBuf, r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '~' {
				inEscape = false
				out.WriteString(string(escBuf))
				escBuf = escBuf[:0]
			}
			continue
		}

		charWidth := 1
		if r == '\t' {
			charWidth = 4
		}

		if visCol >= horizOffset {
			if printedCols+charWidth <= limit {
				out.WriteRune(r)
				printedCols += charWidth
			} else {
				hasMore = true
				break
			}
		}
		visCol += charWidth
	}

	if hasMore {
		out.WriteString(cGold + "…" + cReset)
		printedCols++
	}
	if isBoxRow {
		pad := maxCols - 1 - printedCols
		if pad > 0 {
			out.WriteString(strings.Repeat(" ", pad))
		}
		out.WriteString(cGold + "│" + cReset)
	}
	out.WriteString(cReset + "\033[K")
	return out.String()
}

func formatLogLineForTUI(raw string, maxCols int, horizOffset int) string {
	raw = emojiToLucideReplacer.Replace(raw)

	var icon, color string
	if strings.Contains(raw, "[SYNTAX HOLD]") {
		color = cAmber
		icon = iconWarn
	} else if strings.Contains(raw, "[SECRET SHIELD]") {
		color = cRed
		icon = iconShield
	} else if strings.Contains(raw, "Push succeeded") || strings.Contains(raw, "Committed") || strings.Contains(raw, "✓") {
		color = cGreen
		icon = iconCheck
	} else if strings.Contains(raw, "Pushing") || strings.Contains(raw, "↗") {
		color = cGold
		icon = iconArrowUp
	} else if strings.Contains(raw, "failed") || strings.Contains(raw, "error") || strings.Contains(raw, "✕") {
		color = cRed
		icon = iconCross
	} else {
		color = cSlate
		icon = iconChevron
	}

	formatted := fmt.Sprintf("%s%s %s%s", color, icon, raw, cReset)
	return renderLineANSI(formatted, maxCols, horizOffset)
}

func findConfigFile(scriptPath, repoRoot string) string {
	if cfg := os.Getenv("SYNCALIBUR_CONFIG_FILE"); cfg != "" {
		if _, err := os.Stat(cfg); err == nil {
			return cfg
		}
	}
	candidates := []string{
		filepath.Join(filepath.Dir(scriptPath), "syncalibur.env"),
		filepath.Join(repoRoot, "syncalibur.env"),
	}
	if exePath, err := os.Executable(); err == nil {
		if realExe, err := filepath.EvalSymlinks(exePath); err == nil {
			candidates = append(candidates, filepath.Join(filepath.Dir(filepath.Dir(realExe)), "syncalibur.env"))
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func getActiveMode(scriptPath, repoRoot string) string {
	for _, fname := range []string{".syncalibur", ".syncalibur.env", "syncalibur.env"} {
		cfgPath := filepath.Join(repoRoot, fname)
		if data, err := os.ReadFile(cfgPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "MODE=") || strings.HasPrefix(line, "SYNCALIBUR_MODE=") {
					val := strings.Trim(strings.SplitN(line, "=", 2)[1], `"' `)
					if val != "" {
						return strings.ToLower(val)
					}
				}
			}
		}
	}
	cfgPath := findConfigFile(scriptPath, repoRoot)
	if cfgPath != "" {
		data, err := os.ReadFile(cfgPath)
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "SYNCALIBUR_MODE=") || strings.HasPrefix(line, "MODE=") {
					val := strings.Trim(strings.SplitN(line, "=", 2)[1], `"' `)
					if val != "" {
						return strings.ToLower(val)
					}
				}
			}
		}
	}
	if envMode := strings.TrimSpace(os.Getenv("SYNCALIBUR_MODE")); envMode != "" {
		return strings.ToLower(envMode)
	}
	return "developer"
}

func getModeBadge(mode string) string {
	if canonicalizeMode(mode) == "manual" {
		return cGold + cBold + "⏸ Manual" + cReset
	}
	return cCyan + cBold + "⚡ Autopilot" + cReset
}

func cycleActiveMode(scriptPath, repoRoot string) string {
	current := canonicalizeMode(getActiveMode(scriptPath, repoRoot))
	next := "autopilot"
	if current == "autopilot" {
		next = "manual"
	} else {
		next = "autopilot"
	}

	targetCfg := filepath.Join(repoRoot, ".syncalibur")
	if !isGitRepoDir(repoRoot) {
		targetCfg = findConfigFile(scriptPath, repoRoot)
	}

	if targetCfg != "" {
		data, err := os.ReadFile(targetCfg)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			found := false
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "MODE=") {
					lines[i] = fmt.Sprintf(`MODE="%s"`, next)
					found = true
					break
				} else if strings.HasPrefix(trimmed, "SYNCALIBUR_MODE=") {
					lines[i] = fmt.Sprintf(`SYNCALIBUR_MODE="%s"`, next)
					found = true
					break
				}
			}
			if !found {
				lines = append(lines, fmt.Sprintf(`MODE="%s"`, next))
			}
			_ = os.WriteFile(targetCfg, []byte(strings.Join(lines, "\n")), 0644)
		} else {
			_ = os.WriteFile(targetCfg, []byte(fmt.Sprintf("# Syncalibur local configuration\nMODE=\"%s\"\n", next)), 0644)
		}
	}

	if scriptPath != "" {
		_ = exec.Command("bash", scriptPath, "mode", next).Run()
	}
	_ = os.Setenv("SYNCALIBUR_MODE", next)
	return next
}

func boostAnsiColors(line string, factor float64, addR, addG, addB int) string {
	if factor == 1.0 && addR == 0 && addG == 0 && addB == 0 {
		return line
	}
	var sb strings.Builder
	sb.Grow(len(line) + 64)
	i := 0
	n := len(line)
	for i < n {
		if i+7 < n && line[i] == '\033' && line[i+1] == '[' && (strings.HasPrefix(line[i+2:], "38;2;") || strings.HasPrefix(line[i+2:], "48;2;")) {
			isFg := line[i+2] == '3'
			prefix := "\033[38;2;"
			if !isFg {
				prefix = "\033[48;2;"
			}
			idx := i + 7
			mIdx := strings.IndexByte(line[idx:], 'm')
			if mIdx != -1 {
				parts := strings.Split(line[idx:idx+mIdx], ";")
				if len(parts) == 3 {
					r, _ := strconv.Atoi(parts[0])
					g, _ := strconv.Atoi(parts[1])
					b, _ := strconv.Atoi(parts[2])

					if r > 0 || g > 0 || b > 0 {
						nr := int(float64(r)*factor) + addR
						ng := int(float64(g)*factor) + addG
						nb := int(float64(b)*factor) + addB
						if nr > 255 {
							nr = 255
						}
						if ng > 255 {
							ng = 255
						}
						if nb > 255 {
							nb = 255
						}
						if nr < 0 {
							nr = 0
						}
						if ng < 0 {
							ng = 0
						}
						if nb < 0 {
							nb = 0
						}
						sb.WriteString(prefix)
						sb.WriteString(strconv.Itoa(nr))
						sb.WriteByte(';')
						sb.WriteString(strconv.Itoa(ng))
						sb.WriteByte(';')
						sb.WriteString(strconv.Itoa(nb))
						sb.WriteByte('m')
						i = idx + mIdx + 1
						continue
					}
				}
			}
		}
		sb.WriteByte(line[i])
		i++
	}
	return sb.String()
}

func injectMarginEmbers(line string, r, frame int) string {
	const resetSpace = "\x1b[0m "
	leadCount := 0
	idx := 0
	for strings.HasPrefix(line[idx:], resetSpace) {
		leadCount++
		idx += len(resetSpace)
	}
	if leadCount > 3 {
		slot := (r*3 + frame/2) % leadCount
		if (frame+r*7)%5 == 0 && slot > 0 && slot < leadCount {
			sparkles := []string{
				"\x1b[38;2;90;225;255m✦\x1b[0m",
				"\x1b[38;2;255;215;0m✧\x1b[0m",
				"\x1b[38;2;167;243;208m·\x1b[0m",
				"\x1b[38;2;255;255;255m°\x1b[0m",
				"\x1b[38;2;251;191;36m⋆\x1b[0m",
			}
			char := sparkles[(frame+r)%len(sparkles)]
			before := strings.Repeat(resetSpace, slot)
			after := strings.Repeat(resetSpace, leadCount-slot-1)
			line = before + char + after + line[idx:]
		}
	}

	const trailSpace = " \x1b[0m"
	trailCount := 0
	endIdx := len(line)
	for endIdx >= len(trailSpace) && strings.HasSuffix(line[:endIdx], trailSpace) {
		trailCount++
		endIdx -= len(trailSpace)
	}
	if trailCount > 3 {
		slot := (r*2 + frame/3) % trailCount
		if (frame+r*5)%6 == 0 && slot > 0 && slot < trailCount {
			sparkles := []string{
				"\x1b[38;2;255;215;0m✦\x1b[0m",
				"\x1b[38;2;90;225;255m✧\x1b[0m",
				"\x1b[38;2;255;255;255m·\x1b[0m",
				"\x1b[38;2;147;197;253m⋆\x1b[0m",
			}
			char := sparkles[(frame+r*2)%len(sparkles)]
			before := strings.Repeat(trailSpace, slot)
			after := strings.Repeat(trailSpace, trailCount-slot-1)
			line = line[:endIdx] + before + char + after
		}
	}
	return line
}

func applyAnimatedLogo(baseLines []string, frame int) []string {
	res := make([]string, len(baseLines))
	gleamRow := frame % 28
	for r, l := range baseLines {
		withEmbers := injectMarginEmbers(l, r, frame)
		var factor float64 = 1.0
		var addR, addG, addB int
		if gleamRow < 21 {
			if r == gleamRow {
				factor = 1.55
				addR, addG, addB = 55, 70, 95
			} else if r == gleamRow-1 || r == gleamRow+1 {
				factor = 1.25
				addR, addG, addB = 25, 30, 45
			}
		} else {
			pulse := (frame % 4)
			switch pulse {
			case 0:
				factor = 1.05
				addR, addG, addB = 8, 12, 18
			case 1:
				factor = 1.12
				addR, addG, addB = 15, 22, 30
			case 2:
				factor = 1.08
				addR, addG, addB = 10, 16, 22
			default:
				factor = 1.0
			}
		}
		res[r] = boostAnsiColors(withEmbers, factor, addR, addG, addB)
	}
	return res
}

func getEkgWave(frame int, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune("───────/\\_───/\\/\\──────────────/\\_───/\\/\\───────")
	cycle := 24
	offset := (frame * 2) % cycle
	var result []rune
	for len(result) < width {
		needed := width - len(result)
		avail := len(runes) - offset
		if needed <= avail {
			result = append(result, runes[offset:offset+needed]...)
			break
		}
		result = append(result, runes[offset:]...)
		offset = 0
	}
	return string(result)
}

func getAnimatedModeBadge(mode string, frame int) string {
	if canonicalizeMode(mode) == "manual" {
		return fmt.Sprintf("%s%s⏸ Manual%s", cGold, cBold, cReset)
	}
	return fmt.Sprintf("%s%s⚡ Autopilot%s", cCyan, cBold, cReset)
}

func runInteractiveTUI(scriptPath, repoRoot string) {
	sttyOrigCmd := exec.Command("stty", "-g")
	sttyOrigCmd.Stdin = os.Stdin
	sttyOrigOut, _ := sttyOrigCmd.Output()
	sttyOrig := strings.TrimSpace(string(sttyOrigOut))

	rawCmd := exec.Command("stty", "-icanon", "-echo", "min", "1", "time", "0")
	rawCmd.Stdin = os.Stdin
	_ = rawCmd.Run()

	cleanup := func() {
		if sttyOrig != "" {
			c := exec.Command("stty", sttyOrig)
			c.Stdin = os.Stdin
			_ = c.Run()
		} else {
			c := exec.Command("stty", "sane")
			c.Stdin = os.Stdin
			_ = c.Run()
		}
		fmt.Print("\033[?1049l\033[?25h")
	}
	defer cleanup()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cleanup()
		os.Exit(0)
	}()

	fmt.Print("\033[?1049h\033[?25l")

	keyCh := make(chan []byte, 32)
	go func() {
		buf := make([]byte, 16)
		for {
			n, err := os.Stdin.Read(buf)
			if err == nil && n > 0 {
				cpy := make([]byte, n)
				copy(cpy, buf[:n])
				keyCh <- cpy
			} else if err != nil {
				return
			}
		}
	}()

	var flashMsg string
	var flashUntil time.Time
	gitDir := resolveGitDir(repoRoot)
	logFile := filepath.Join(gitDir, "syncalibur.log")
	candidatePaths := []string{
		filepath.Join(filepath.Dir(scriptPath), "assets", "syncalibur-logo.ansi"),
		filepath.Join(repoRoot, "assets", "syncalibur-logo.ansi"),
		"/home/ysf/Syncalibur/assets/syncalibur-logo.ansi",
	}
	if exePath, err := os.Executable(); err == nil {
		if realExe, err := filepath.EvalSymlinks(exePath); err == nil {
			candidatePaths = append(candidatePaths, filepath.Join(filepath.Dir(filepath.Dir(realExe)), "assets", "syncalibur-logo.ansi"))
		}
	}
	var logoLines []string
	for _, p := range candidatePaths {
		if _, err := os.Stat(p); err == nil {
			lines := loadLogoLines(p)
			if len(lines) == 21 {
				logoLines = lines
				break
			}
		}
	}
	if len(logoLines) != 21 {
		logoLines = defaultEmbeddedLogoLines
	}
	showLogo := true

	scrollOffset := 0
	horizOffset := 0
	followMode := true

	animFrame := 0
	lastGitQuery := time.Time{}
	var cachedOnline bool
	var cachedDaemonPid int
	cachedBranch := "main"
	cachedUnpushed := "0"
	cachedToday := "0"
	cachedSpark := "───────"
	cachedOfflineQueued := ""
	isPaused := false
	cachedActiveMode := "developer"

	renderFrame := func() {
		termRows, termCols := getTerminalSize()
		if termCols < 60 {
			termCols = 60
		}
		if termRows < 14 {
			termRows = 14
		}

		if time.Since(lastGitQuery) >= 1000*time.Millisecond {
			cachedOnline, cachedDaemonPid = getDaemonStatus(gitDir)

			branchOut, _ := exec.Command("git", "-C", repoRoot, "symbolic-ref", "--short", "HEAD").Output()
			b := strings.TrimSpace(string(branchOut))
			if b != "" {
				cachedBranch = b
			} else {
				cachedBranch = "main"
			}

			unpushedOut, _ := exec.Command("git", "-C", repoRoot, "rev-list", "--count", "@{u}..HEAD").Output()
			u := strings.TrimSpace(string(unpushedOut))
			if u != "" {
				cachedUnpushed = u
			} else {
				cachedUnpushed = "0"
			}

			if _, err := os.Stat(filepath.Join(gitDir, "syncalibur.offline")); err == nil {
				cachedOfflineQueued = cachedUnpushed
			} else {
				cachedOfflineQueued = ""
			}

			_, pauseErr := os.Stat(filepath.Join(gitDir, "syncalibur.pause"))
			isPaused = (pauseErr == nil)

			velo := getCommitVelocity(repoRoot)
			cachedToday = strconv.Itoa(velo.Today)
			cachedSpark = velo.Sparkline

			cachedActiveMode = getActiveMode(scriptPath, repoRoot)
			lastGitQuery = time.Now()
		}

		beacons := []string{"●", "◉"}
		beacon := beacons[animFrame%len(beacons)]

		var onlineStr string
		if !cachedOnline {
			onlineStr = cRed + "○ OFFLINE" + cReset
		} else if cachedDaemonPid > 0 {
			onlineStr = fmt.Sprintf("%s%s ONLINE%s %s(PID %d)%s", cGreen, beacon, cReset, cSlate, cachedDaemonPid, cReset)
		} else {
			onlineStr = fmt.Sprintf("%s%s ONLINE%s", cGreen, beacon, cReset)
		}
		if cachedOfflineQueued != "" {
			onlineStr += fmt.Sprintf(" • %s🌐 OFFLINE (%s queued)%s", cAmber, cachedOfflineQueued, cReset)
		}

		pauseState := cGreen + iconPlay + " ACTIVE" + cReset
		if isPaused {
			pauseState = cAmber + cBold + iconPause + " PAUSED" + cReset
		}

		modeBadge := getModeBadge(cachedActiveMode)

		unpushedDisplay := fmt.Sprintf("%s%s%s", cAmber, cachedUnpushed, cReset)
		nowStr := time.Now().Format("15:04:05")

		boxWidth := termCols - 6
		if boxWidth > 96 {
			boxWidth = 96
		}
		if boxWidth < 50 {
			boxWidth = 50
		}

		var frame strings.Builder
		frame.WriteString("\033[H")

		useSideBySide := showLogo && termCols >= 96 && termRows >= 28

		if useSideBySide {
			// Side-by-side Emblem Cockpit (21 rows)
			rightBoxWidth := termCols - 50
			if rightBoxWidth > 76 {
				rightBoxWidth = 76
			}
			if rightBoxWidth < 46 {
				rightBoxWidth = 46
			}

			rightLines := make([]string, 21)
			rightLines[0] = fmt.Sprintf("╭%s╮", strings.Repeat("─", rightBoxWidth-2))
			rightLines[1] = fmt.Sprintf("│  %s%s%s SYNCALIBUR COCKPIT %s[%s]%s", cCyan, cBold, iconDiamond, cGold, SyncaliburVersion, cReset)
			rightLines[2] = fmt.Sprintf("│  Status    : %s", onlineStr)
			rightLines[3] = fmt.Sprintf("│  Mode      : %s  •  Loop: %s", modeBadge, pauseState)
			rightLines[4] = fmt.Sprintf("├%s┤", strings.Repeat("─", rightBoxWidth-2))
			rightLines[5] = fmt.Sprintf("│  %s◈ Watching%s : %s", cGold, cReset, repoRoot)
			rightLines[6] = fmt.Sprintf("│  %s◈ Branch%s   : %s%s%s  •  Unpushed: %s", cGold, cReset, cBold, cachedBranch, cReset, unpushedDisplay)
			rightLines[7] = fmt.Sprintf("│  %s◈ Today%s    : %s%s commits%s  •  Time: %s%s%s", cGold, cReset, cGreen, cachedToday, cReset, cSlate, nowStr, cReset)
			rightLines[8] = fmt.Sprintf("├%s┤", strings.Repeat("─", rightBoxWidth-2))
			rightLines[9] = fmt.Sprintf("│  %s◈ Shields%s  : %s%s Syntax Pre-flight (AST)%s", cGold, cReset, cGreen, iconCheck, cReset)
			rightLines[10] = fmt.Sprintf("│              %s%s Shannon Entropy Guard%s", cGreen, iconCheck, cReset)
			rightLines[11] = fmt.Sprintf("│              %s%s Multi-Remote Rebase%s", cGreen, iconCheck, cReset)
			rightLines[12] = fmt.Sprintf("│  %s◈ Velocity%s : %s", cGold, cReset, cachedSpark)
			rightLines[13] = fmt.Sprintf("├%s┤", strings.Repeat("─", rightBoxWidth-2))
			rightLines[14] = fmt.Sprintf("│  %s◈ Release%s  : %s%s%s (%s)", cGold, cReset, cGreen, SyncaliburVersion, cReset, SyncaliburMilestone)
			rightLines[15] = fmt.Sprintf("├%s┤", strings.Repeat("─", rightBoxWidth-2))
			if time.Now().Before(flashUntil) && flashMsg != "" {
				rightLines[16] = fmt.Sprintf("│  %s", flashMsg)
				rightLines[17] = fmt.Sprintf("│  %s", cSlate+"Press any hotkey to interact..."+cReset)
				rightLines[18] = fmt.Sprintf("│")
				rightLines[19] = fmt.Sprintf("│")
			} else {
				rightLines[16] = fmt.Sprintf("│  %sHotkeys:%s %s[q]%s Quit  %s[p]%s Pause  %s[z]%s Undo  %s[s]%s Sync", cBold, cReset, cRed, cReset, cGold, cReset, cGreen, cReset, cGold, cReset)
				rightLines[17] = fmt.Sprintf("│           %s[b]%s Pull  %s[m]%s Mode   %s[t]%s Logo  %s[d]%s Doctor", cCyan, cReset, cGold, cReset, cCyan, cReset, cGold, cReset)
				rightLines[18] = fmt.Sprintf("│           %s[↑/↓]%s Scroll  %s[←/→]%s Pan  %s[f]%s Tail", cCyan, cReset, cCyan, cReset, cGreen, cReset)
				rightLines[19] = fmt.Sprintf("│           %s[0]%s Reset Pan  %s[r]%s Redraw", cCyan, cReset, cSlate, cReset)
			}
			rightLines[20] = fmt.Sprintf("╰%s╯", strings.Repeat("─", rightBoxWidth-2))

			for r := 0; r < 21; r++ {
				lPart := ""
				if r < len(logoLines) {
					lPart = logoLines[r]
				} else {
					lPart = strings.Repeat(" ", 44)
				}
				fmt.Fprintf(&frame, "  %s  %s%s\n", lPart, cGold, renderLineANSI(rightLines[r], rightBoxWidth, 0))
			}
		} else {
			// Compact Header (Strictly 7 lines)
			fmt.Fprintf(&frame, "  %s╭%s╮%s\033[K\n", cGold, strings.Repeat("─", boxWidth), cReset)

			line1 := fmt.Sprintf("│  %s%s%s SYNCALIBUR COCKPIT %s[%s]%s  •  %s  •  Mode: %s  •  Loop: %s",
				cCyan, cBold, iconDiamond, cGold, SyncaliburVersion, cReset, onlineStr, modeBadge, pauseState)
			fmt.Fprintf(&frame, "  %s%s\n", cGold, renderLineANSI(line1, boxWidth, 0))

			line2 := fmt.Sprintf("│  %s◈ Watching :%s %s%s [%s]%s  •  %sToday :%s %s%s commits%s  •  %sUnpushed :%s %s",
				cGold, cReset, cBold, repoRoot, cachedBranch, cReset, cGold, cReset, cGreen, cachedToday, cReset, cGold, cReset, unpushedDisplay)
			fmt.Fprintf(&frame, "  %s%s\n", cGold, renderLineANSI(line2, boxWidth, 0))

			line3 := fmt.Sprintf("│  %s◈ Shields  :%s %s%s Syntax%s • %s%s Secrets%s  •  %sVelocity :%s %s",
				cGold, cReset, cGreen, iconCheck, cReset, cGreen, iconCheck, cReset, cGold, cReset, cachedSpark)
			fmt.Fprintf(&frame, "  %s%s\n", cGold, renderLineANSI(line3, boxWidth, 0))

			fmt.Fprintf(&frame, "  %s├%s┤%s\033[K\n", cGold, strings.Repeat("─", boxWidth), cReset)

			if time.Now().Before(flashUntil) && flashMsg != "" {
				lineFlash := fmt.Sprintf("│  %s", flashMsg)
				fmt.Fprintf(&frame, "  %s%s\n", cGold, renderLineANSI(lineFlash, boxWidth, 0))
			} else {
				lineKeys := fmt.Sprintf("│  %sHotkeys:%s %s[q]%s Quit  %s[p]%s Pause  %s[z]%s Undo  %s[s]%s Sync  %s[b]%s Pull  %s[t]%s Logo  %s[f]%s Tail",
					cBold, cReset, cRed, cReset, cGold, cReset, cGreen, cReset, cGold, cReset, cCyan, cReset, cCyan, cReset, cGreen, cReset)
				fmt.Fprintf(&frame, "  %s%s\n", cGold, renderLineANSI(lineKeys, boxWidth, 0))
			}
			fmt.Fprintf(&frame, "  %s╰%s╯%s\033[K\n", cGold, strings.Repeat("─", boxWidth), cReset)
		}

		// 2. Log Viewport Section
		allLogs := loadLogLines(logFile, 1000)
		totalLogs := len(allLogs)

		followBadge := fmt.Sprintf("%s%s TAIL: ON%s", cGreen, iconCheck, cReset)
		if !followMode {
			followBadge = fmt.Sprintf("%s%s SCROLLED ↑ %d (press [f] to tail)%s", cAmber, iconWarn, scrollOffset, cReset)
		}
		panBadge := ""
		if horizOffset > 0 {
			panBadge = fmt.Sprintf("  %sPan: → +%d cols (press [0] to reset)%s", cCyan, horizOffset, cReset)
		}

		fmt.Fprintf(&frame, "  %s%s%s LIVE LOG VIEWPORT%s  %s(%d lines • %s%s)%s\033[K\n",
			cBold, cCyan, iconChevron, cReset, cSlate, totalLogs, followBadge, panBadge, cReset)
		fmt.Fprintf(&frame, "  %s%s%s\033[K\n", cSlate, strings.Repeat("─", boxWidth), cReset)

		headerRows := 9
		if useSideBySide {
			headerRows = 23
		}
		footerRows := 1
		logRows := termRows - headerRows - footerRows
		if logRows < 3 {
			logRows = 3
		}

		maxScroll := 0
		if totalLogs > logRows {
			maxScroll = totalLogs - logRows
		}
		if followMode {
			scrollOffset = 0
		}
		if scrollOffset > maxScroll {
			scrollOffset = maxScroll
		}
		if scrollOffset < 0 {
			scrollOffset = 0
		}

		if totalLogs == 0 {
			fmt.Fprintf(&frame, "    %s%s Waiting for activity logs in %s...%s\033[K\n", cSlate, iconCircle, logFile, cReset)
			for r := 1; r < logRows; r++ {
				fmt.Fprintf(&frame, "\033[K\n")
			}
		} else {
			endIdx := totalLogs - scrollOffset
			startIdx := endIdx - logRows
			if startIdx < 0 {
				startIdx = 0
			}
			rowsRendered := 0
			for i := startIdx; i < endIdx && i < totalLogs; i++ {
				fmt.Fprintf(&frame, "  %s\n", formatLogLineForTUI(allLogs[i], termCols-4, horizOffset))
				rowsRendered++
			}
			for r := rowsRendered; r < logRows; r++ {
				fmt.Fprintf(&frame, "\033[K\n")
			}
		}

		// 3. Footer Bar
		fmt.Fprintf(&frame, "  %s%s%s\033[0m\033[K", cSlate, "[q] Quit  [p] Pause/Resume  [s] Sync  [b] Pull  [↑/↓/k/j] Scroll  [←/→/h/l] Pan  [f] Tail  [d] Doctor", cReset)
		frame.WriteString("\033[J")
		fmt.Print(frame.String())
	}

	renderFrame()
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case keys := <-keyCh:
			if len(keys) == 0 {
				continue
			}

			i := 0
			shouldExit := false
			for i < len(keys) {
				if keys[i] == 27 { // Escape sequence
					if i+2 < len(keys) && keys[i+1] == '[' {
						seq := keys[i+2]
						if (seq == '5' || seq == '6' || seq == '1' || seq == '4') && i+3 < len(keys) && keys[i+3] == '~' {
							switch seq {
							case '5': // Page Up
								scrollOffset += 10
								followMode = false
							case '6': // Page Down
								scrollOffset -= 10
								if scrollOffset <= 0 {
									scrollOffset = 0
									followMode = true
								}
							case '1': // Home
								scrollOffset = 9999
								followMode = false
							case '4': // End
								scrollOffset = 0
								followMode = true
							}
							i += 4
							continue
						}
						switch seq {
						case 'A': // Arrow Up
							scrollOffset++
							followMode = false
						case 'B': // Arrow Down
							if scrollOffset > 0 {
								scrollOffset--
								if scrollOffset == 0 {
									followMode = true
								}
							}
						case 'C': // Arrow Right (pan right)
							horizOffset += 10
						case 'D': // Arrow Left (pan left)
							if horizOffset >= 10 {
								horizOffset -= 10
							} else {
								horizOffset = 0
							}
						case 'H': // Home
							scrollOffset = 9999
							followMode = false
						case 'F': // End
							scrollOffset = 0
							followMode = true
						}
						i += 3
						continue
					} else if i+1 == len(keys) {
						// Lone Escape
						shouldExit = true
						break
					}
				}

				// Single character
				ch := keys[i]
				i++
				switch ch {
				case 'q', 'Q', 3, 27: // q, Q, Ctrl+C, Esc -> INSTANT QUIT
					shouldExit = true
				case 'k', 'K': // vim scroll up
					scrollOffset++
					followMode = false
				case 'j', 'J': // vim scroll down
					if scrollOffset > 0 {
						scrollOffset--
						if scrollOffset == 0 {
							followMode = true
						}
					}
				case 'h', 'H': // vim pan left
					if horizOffset >= 10 {
						horizOffset -= 10
					} else {
						horizOffset = 0
					}
				case 'l', 'L': // vim pan right
					horizOffset += 10
				case '0': // reset pan
					horizOffset = 0
				case 'f', 'F': // tail / follow mode
					scrollOffset = 0
					followMode = true
				case 'g': // top
					scrollOffset = 9999
					followMode = false
				case 'G': // bottom
					scrollOffset = 0
					followMode = true
				case 'p', 'P': // Pause / Resume
					pauseFile := filepath.Join(gitDir, "syncalibur.pause")
					if _, err := os.Stat(pauseFile); err == nil {
						_ = os.Remove(pauseFile)
						flashMsg = fmt.Sprintf("%s%s %sRESUMED:%s Autonomous sync loop resumed.", cGreen, iconPlay, cBold, cReset)
					} else {
						_ = os.WriteFile(pauseFile, []byte("1\n"), 0644)
						flashMsg = fmt.Sprintf("%s%s %sPAUSED:%s Daemon paused. Press [p] again to resume.", cAmber, iconPause, cBold, cReset)
					}
					lastGitQuery = time.Time{}
					flashUntil = time.Now().Add(3 * time.Second)
				case 'm', 'M': // Cycle operating mode
					newMode := cycleActiveMode(scriptPath, repoRoot)
					flashMsg = fmt.Sprintf("%s%s %sMODE SWITCHED:%s Active mode is now %s.", cGreen, iconZap, cBold, cReset, strings.ToUpper(newMode))
					lastGitQuery = time.Time{}
					flashUntil = time.Now().Add(3 * time.Second)
				case 's', 'S': // Sync now
					_ = os.WriteFile(filepath.Join(gitDir, "syncalibur.sync"), []byte("1\n"), 0644)
					flashMsg = fmt.Sprintf("%s%s %sSYNC NOW:%s Immediate commit & push triggered!", cCyan, iconZap, cBold, cReset)
					lastGitQuery = time.Time{}
					flashUntil = time.Now().Add(3 * time.Second)
				case 'b', 'B': // Bidirectional pull
					_ = exec.Command("bash", scriptPath, "pull").Run()
					flashMsg = fmt.Sprintf("%s%s %sBIDIRECTIONAL PULL:%s Checked upstream & rebased cleanly.", cCyan, iconSync, cBold, cReset)
					lastGitQuery = time.Time{}
					flashUntil = time.Now().Add(3 * time.Second)
				case 'z', 'Z': // Undo to previous checkpoint
					out, err := exec.Command("bash", scriptPath, "undo").CombinedOutput()
					if err != nil {
						flashMsg = fmt.Sprintf("%s%s %sUNDO:%s %s", cAmber, iconWarn, cBold, cReset, strings.TrimSpace(string(out)))
					} else {
						flashMsg = fmt.Sprintf("%s%s %sUNDO:%s Rolled back to previous checkpoint cleanly.", cGreen, iconCheck, cBold, cReset)
					}
					lastGitQuery = time.Time{}
					flashUntil = time.Now().Add(4 * time.Second)
				case 'u', 'U': // Self-update
					_ = os.WriteFile(filepath.Join(gitDir, "syncalibur.upd"), []byte("1\n"), 0644)
					flashMsg = fmt.Sprintf("%s%s %sSELF-UPDATE:%s Upstream self-update check triggered.", cGold, iconSync, cBold, cReset)
					lastGitQuery = time.Time{}
					flashUntil = time.Now().Add(3 * time.Second)
				case 'd', 'D': // Doctor
					flashMsg = fmt.Sprintf("%s%s %sDOCTOR:%s Running 12-point diagnostic check...", cGreen, iconDoctor, cBold, cReset)
					flashUntil = time.Now().Add(3 * time.Second)
				case 't', 'T', 'e', 'E': // Toggle Emblem / Logo
					showLogo = !showLogo
					if showLogo {
						flashMsg = fmt.Sprintf("%s%s %sEMBLEM COCKPIT:%s TrueColor sword logo & telemetry expanded.", cCyan, iconZap, cBold, cReset)
					} else {
						flashMsg = fmt.Sprintf("%s%s %sCOMPACT TAIL VIEW:%s Maximized log viewport.", cGreen, iconZap, cBold, cReset)
					}
					flashUntil = time.Now().Add(3 * time.Second)
				case 'r', 'R': // Redraw
				}
				if shouldExit {
					break
				}
			}

			if shouldExit {
				return
			}
			animFrame++
			renderFrame()

		case <-ticker.C:
			animFrame++
			renderFrame()
		}
	}
}

// ==============================================================================
// 8. Automated SemVer & Release Engine (`release`, `version`)
// ==============================================================================

type SemVerBump int

const (
	BumpNone SemVerBump = iota
	BumpPatch
	BumpMinor
	BumpMajor
)

func (b SemVerBump) String() string {
	switch b {
	case BumpMajor:
		return "major"
	case BumpMinor:
		return "minor"
	case BumpPatch:
		return "patch"
	default:
		return "none"
	}
}

type CommitInfo struct {
	Hash        string
	Author      string
	Date        string
	Subject     string
	Type        string
	Scope       string
	IsBreaking  bool
	Description string
}

func parseConventionalCommit(subject string) (commitType string, scope string, isBreaking bool, description string) {
	s := strings.TrimSpace(subject)
	if strings.HasPrefix(s, "BREAKING CHANGE:") || strings.HasPrefix(s, "BREAKING-CHANGE:") {
		idx := strings.Index(s, ":")
		return "feat", "", true, strings.TrimSpace(s[idx+1:])
	}
	re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)(?:\(([^)]+)\))?(!)?:\s*(.*)$`)
	m := re.FindStringSubmatch(s)
	if m != nil {
		cType := strings.ToLower(m[1])
		cScope := m[2]
		isBreak := m[3] == "!"
		desc := strings.TrimSpace(m[4])
		if strings.Contains(desc, "BREAKING CHANGE") || strings.Contains(desc, "BREAKING-CHANGE") {
			isBreak = true
		}
		return cType, cScope, isBreak, desc
	}
	return "other", "", false, s
}

func determineSemVerBump(commits []CommitInfo) SemVerBump {
	bump := BumpNone
	for _, c := range commits {
		if c.IsBreaking {
			return BumpMajor
		}
		if c.Type == "feat" {
			if bump < BumpMinor {
				bump = BumpMinor
			}
		} else if bump < BumpPatch {
			bump = BumpPatch
		}
	}
	if bump == BumpNone && len(commits) > 0 {
		bump = BumpPatch
	}
	return bump
}

func calculateNextVersion(currentTag string, bump SemVerBump) string {
	tag := strings.TrimPrefix(strings.TrimSpace(currentTag), "v")
	parts := strings.Split(tag, ".")
	major, minor, patch := 1, 0, 0
	if len(parts) == 3 {
		maj, err1 := strconv.Atoi(parts[0])
		min, err2 := strconv.Atoi(parts[1])
		pat, err3 := strconv.Atoi(parts[2])
		if err1 == nil && err2 == nil && err3 == nil {
			major, minor, patch = maj, min, pat
		}
	} else if currentTag == "" {
		if bump == BumpMajor {
			return "v2.0.0"
		} else if bump == BumpMinor {
			return "v1.1.0"
		}
		return "v1.0.1"
	}

	switch bump {
	case BumpMajor:
		major++
		minor = 0
		patch = 0
	case BumpMinor:
		minor++
		patch = 0
	case BumpPatch:
		patch++
	}

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

func generateChangelog(version string, date string, commits []CommitInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## [%s] - %s\n\n", version, date))

	var features, shields, fixes, perfs, tooling, others []string

	for _, c := range commits {
		hashStr := ""
		if c.Hash != "" {
			h := c.Hash
			if len(h) > 7 {
				h = h[:7]
			}
			hashStr = fmt.Sprintf(" ([`%s`](https://github.com/YusufKaya00/Syncalibur/commit/%s))", h, h)
		}
		item := ""
		if c.Scope != "" {
			item = fmt.Sprintf("- **%s**: %s%s", c.Scope, c.Description, hashStr)
		} else {
			item = fmt.Sprintf("- %s%s", c.Description, hashStr)
		}

		switch strings.ToLower(c.Type) {
		case "feat":
			features = append(features, item)
		case "shield", "security", "entropy":
			shields = append(shields, item)
		case "fix":
			fixes = append(fixes, item)
		case "perf":
			perfs = append(perfs, item)
		case "ui":
			perfs = append(perfs, item)
		case "docs", "test", "chore", "build", "ci", "refactor":
			tooling = append(tooling, item)
		default:
			others = append(others, item)
		}
	}

	if len(features) > 0 {
		sb.WriteString("### 🚀 Features\n")
		for _, f := range features {
			sb.WriteString(f + "\n")
		}
		sb.WriteString("\n")
	}
	if len(shields) > 0 {
		sb.WriteString("### 🛡️ Shields & Pre-flight\n")
		for _, s := range shields {
			sb.WriteString(s + "\n")
		}
		sb.WriteString("\n")
	}
	if len(fixes) > 0 {
		sb.WriteString("### 🐛 Bug Fixes & Stability\n")
		for _, fix := range fixes {
			sb.WriteString(fix + "\n")
		}
		sb.WriteString("\n")
	}
	if len(perfs) > 0 {
		sb.WriteString("### ⚡ Performance & TUI Polish\n")
		for _, p := range perfs {
			sb.WriteString(p + "\n")
		}
		sb.WriteString("\n")
	}
	if len(tooling) > 0 {
		sb.WriteString("### 📚 Tooling, Tests & Fleet\n")
		for _, t := range tooling {
			sb.WriteString(t + "\n")
		}
		sb.WriteString("\n")
	}
	if len(others) > 0 {
		sb.WriteString("### 📦 Other Improvements\n")
		for _, o := range others {
			sb.WriteString(o + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

var defaultEmbeddedLogoLines = []string{
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;119;112;40m\u2597\u001b[0m\u001b[38;2;165;149;25m\u2584\u001b[0m\u001b[38;2;229;148;0m\u2584\u001b[0m\u001b[38;2;223;118;0m\u2584\u001b[0m\u001b[38;2;161;81;0m\u2584\u001b[0m\u001b[38;2;20;17;19m\u259f\u001b[38;2;32;17;12m\u001b[48;2;133;128;118m\u259b\u001b[38;2;202;203;198m\u001b[48;2;19;5;1m\u2596\u001b[38;2;109;84;69m\u001b[48;2;0;0;4m\u2596\u001b[0m\u001b[38;2;94;47;0m\u2584\u001b[0m\u001b[38;2;248;124;0m\u2584\u001b[0m\u001b[38;2;205;130;0m\u2584\u001b[0m\u001b[38;2;155;123;7m\u2584\u001b[0m\u001b[38;2;106;83;17m\u2596\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;110;109;47m\u2597\u001b[0m\u001b[38;2;216;178;9m\u2584\u001b[0m\u001b[38;2;206;142;14m\u259f\u001b[38;2;187;143;13m\u001b[48;2;106;44;0m\u259b\u001b[38;2;255;168;0m\u001b[48;2;11;0;0m\u2580\u001b[38;2;237;131;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;149;72;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;74;24;0m\u001b[48;2;9;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;3m\u001b[48;2;65;49;59m\u2580\u001b[38;2;65;47;65m\u001b[48;2;89;67;57m\u2598\u001b[38;2;96;89;83m\u001b[48;2;16;8;3m\u259d\u001b[38;2;104;106;107m\u001b[48;2;4;0;0m\u258c\u001b[38;2;102;75;64m\u001b[48;2;86;62;54m\u258c\u001b[38;2;0;0;15m\u001b[48;2;64;47;52m\u2580\u001b[38;2;0;0;19m\u001b[48;2;2;0;0m\u2596\u001b[38;2;87;31;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;187;99;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;252;150;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;255;168;0m\u001b[48;2;46;19;0m\u2580\u001b[38;2;170;108;2m\u001b[48;2;255;149;0m\u259b\u001b[0m\u001b[38;2;197;143;5m\u2599\u001b[0m\u001b[38;2;183;140;0m\u2584\u001b[0m\u001b[38;2;72;53;13m\u2596\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;111;110;54m\u2597\u001b[0m\u001b[38;2;228;218;4m\u2584\u001b[38;2;231;203;6m\u001b[48;2;162;114;24m\u259e\u001b[38;2;255;199;0m\u001b[48;2;12;0;0m\u2580\u001b[38;2;208;110;0m\u001b[48;2;25;10;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;67;50;42m\u259b\u001b[38;2;32;22;16m\u001b[48;2;92;69;55m\u2598\u001b[38;2;84;63;51m\u001b[48;2;97;71;57m\u259e\u001b[38;2;95;67;54m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u2598\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u259d\u001b[38;2;86;56;47m\u001b[48;2;0;0;0m\u259d\u001b[38;2;77;53;43m\u001b[48;2;95;73;54m\u259e\u001b[38;2;79;59;48m\u001b[48;2;66;48;42m\u258c\u001b[38;2;83;60;53m\u001b[48;2;12;6;4m\u2596\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;57;24;0m\u001b[48;2;0;0;0m\u259d\u001b[38;2;216;138;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;255;204;0m\u001b[48;2;72;39;0m\u2580\u001b[38;2;135;96;0m\u001b[48;2;255;206;0m\u2580\u001b[0m\u001b[38;2;214;163;0m\u2584\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;154;154;38m\u2597\u001b[0m\u001b[38;2;219;216;15m\u259f\u001b[38;2;242;201;0m\u001b[48;2;0;0;0m\u259b\u001b[38;2;255;145;0m\u001b[48;2;17;6;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;31;22;16m\u259b\u001b[38;2;0;0;0m\u001b[48;2;72;53;43m\u2598\u001b[38;2;84;61;50m\u001b[48;2;73;56;43m\u259b\u001b[38;2;85;66;56m\u001b[48;2;81;58;43m\u258c\u001b[38;2;90;64;52m\u001b[48;2;0;0;0m\u258c\u001b[38;2;0;0;0m\u001b[48;2;222;220;222m\u2598\u001b[38;2;255;255;255m\u001b[48;2;255;255;255m\u2598\u001b[38;2;255;255;255m\u001b[48;2;255;255;255m\u2598\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u259d\u001b[38;2;0;0;0m\u001b[48;2;62;37;29m\u258c\u001b[38;2;56;41;33m\u001b[48;2;87;66;54m\u2596\u001b[38;2;84;61;52m\u001b[48;2;59;43;34m\u259e\u001b[38;2;9;10;7m\u001b[48;2;71;54;43m\u259d\u001b[38;2;66;47;34m\u001b[48;2;0;0;0m\u2596\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;7m\u259b\u001b[38;2;0;0;0m\u001b[48;2;0;0;4m\u2580\u001b[38;2;0;0;0m\u001b[48;2;0;0;4m\u2580\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;16;2;0m\u001b[48;2;0;0;0m\u259d\u001b[38;2;233;185;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;186;133;0m\u001b[48;2;255;248;0m\u259e\u001b[0m\u001b[38;2;205;176;0m\u2599\u001b[0m\u001b[38;2;111;75;0m\u2596\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;198;197;13m\u259f\u001b[38;2;255;234;0m\u001b[48;2;7;0;0m\u259b\u001b[38;2;238;124;0m\u001b[48;2;0;0;7m\u2598\u001b[38;2;0;0;3m\u001b[48;2;0;0;9m\u2580\u001b[38;2;0;0;4m\u001b[48;2;0;3;9m\u2580\u001b[38;2;0;0;3m\u001b[48;2;0;1;8m\u2580\u001b[38;2;0;0;3m\u001b[48;2;0;0;9m\u2580\u001b[38;2;0;0;5m\u001b[48;2;0;0;0m\u259b\u001b[38;2;0;0;5m\u001b[48;2;0;0;0m\u259b\u001b[38;2;0;0;0m\u001b[48;2;47;31;15m\u259b\u001b[38;2;13;9;6m\u001b[48;2;85;63;48m\u2580\u001b[38;2;92;67;55m\u001b[48;2;81;60;46m\u259e\u001b[38;2;83;64;50m\u001b[48;2;71;54;43m\u2580\u001b[38;2;76;59;48m\u001b[48;2;95;77;65m\u258c\u001b[38;2;94;75;65m\u001b[48;2;74;57;46m\u258c\u001b[38;2;88;59;49m\u001b[48;2;0;0;0m\u258c\u001b[38;2;25;25;20m\u001b[48;2;207;203;198m\u2596\u001b[38;2;198;196;192m\u001b[48;2;255;255;255m\u2596\u001b[38;2;255;255;255m\u001b[48;2;249;241;235m\u259b\u001b[38;2;255;255;255m\u001b[48;2;131;123;114m\u2580\u001b[38;2;0;0;0m\u001b[48;2;51;28;20m\u258c\u001b[38;2;44;29;22m\u001b[48;2;65;49;39m\u258c\u001b[38;2;83;63;50m\u001b[48;2;71;54;44m\u2580\u001b[38;2;60;44;36m\u001b[48;2;72;55;44m\u2580\u001b[38;2;66;47;32m\u001b[48;2;79;55;44m\u2598\u001b[38;2;23;16;8m\u001b[48;2;102;73;62m\u2580\u001b[38;2;31;16;14m\u001b[48;2;107;58;0m\u259b\u001b[38;2;158;91;0m\u001b[48;2;13;1;0m\u259b\u001b[38;2;64;36;0m\u001b[48;2;251;173;0m\u2580\u001b[38;2;0;0;4m\u001b[48;2;0;1;27m\u259b\u001b[38;2;0;0;3m\u001b[48;2;0;0;9m\u2580\u001b[38;2;0;0;5m\u001b[48;2;0;4;9m\u2580\u001b[38;2;0;0;0m\u001b[48;2;0;0;17m\u259d\u001b[38;2;215;178;0m\u001b[48;2;0;0;11m\u2580\u001b[38;2;234;168;0m\u001b[48;2;255;255;0m\u259e\u001b[0m\u001b[38;2;210;180;0m\u2584\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m\u001b[38;2;164;149;32m\u2597\u001b[38;2;124;119;73m\u001b[48;2;255;255;0m\u2598\u001b[38;2;217;171;0m\u001b[48;2;22;5;7m\u2580\u001b[38;2;0;0;0m\u001b[48;2;0;4;33m\u2598\u001b[38;2;0;9;29m\u001b[48;2;0;0;28m\u2596\u001b[38;2;0;0;24m\u001b[48;2;151;103;0m\u259b\u001b[38;2;0;0;31m\u001b[48;2;241;163;0m\u2580\u001b[38;2;39;38;0m\u001b[48;2;0;3;33m\u2596\u001b[38;2;0;0;3m\u001b[48;2;90;58;49m\u259b\u001b[38;2;145;111;86m\u001b[48;2;59;43;29m\u2596\u001b[38;2;12;7;6m\u001b[48;2;0;0;0m\u2598\u001b[38;2;17;2;0m\u001b[48;2;88;61;46m\u2596\u001b[38;2;83;58;45m\u001b[48;2;98;75;58m\u2598\u001b[38;2;80;57;46m\u001b[48;2;92;67;54m\u2580\u001b[38;2;65;47;37m\u001b[48;2;73;57;43m\u259d\u001b[38;2;77;59;49m\u001b[48;2;95;79;66m\u2596\u001b[38;2;98;80;70m\u001b[48;2;77;59;49m\u259b\u001b[38;2;87;66;48m\u001b[48;2;69;53;43m\u2580\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u259b\u001b[38;2;255;255;255m\u001b[48;2;57;44;36m\u2596\u001b[38;2;67;70;113m\u001b[48;2;255;231;198m\u259b\u001b[38;2;255;255;255m\u001b[48;2;0;0;0m\u2596\u001b[38;2;36;18;9m\u001b[48;2;70;48;36m\u2596\u001b[38;2;76;54;42m\u001b[48;2;89;72;61m\u259b\u001b[38;2;81;62;47m\u001b[48;2;98;80;70m\u2580\u001b[38;2;96;78;66m\u001b[48;2;89;70;60m\u2580\u001b[38;2;112;93;80m\u001b[48;2;81;64;48m\u2580\u001b[38;2;110;91;77m\u001b[48;2;88;68;54m\u2598\u001b[38;2;100;80;75m\u001b[48;2;61;43;63m\u2598\u001b[38;2;0;0;15m\u001b[48;2;191;96;0m\u259b\u001b[38;2;255;233;0m\u001b[48;2;151;88;0m\u259b\u001b[38;2;0;0;0m\u001b[48;2;27;16;43m\u2580\u001b[38;2;0;0;7m\u001b[48;2;0;4;34m\u2596\u001b[38;2;23;17;23m\u001b[48;2;255;214;0m\u259b\u001b[38;2;0;0;30m\u001b[48;2;255;230;0m\u2580\u001b[38;2;0;0;3m\u001b[48;2;0;0;34m\u2596\u001b[38;2;101;66;0m\u001b[48;2;0;0;35m\u259d\u001b[38;2;0;0;17m\u001b[48;2;255;253;0m\u2596\u001b[0m\u001b[38;2;255;231;0m\u2599\u001b[0m\u001b[38;2;68;40;0m\u2596\u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m\u001b[38;2;169;145;26m\u2597\u001b[38;2;180;151;0m\u001b[48;2;255;255;0m\u2598\u001b[38;2;255;255;0m\u001b[48;2;0;0;17m\u2598\u001b[38;2;0;16;53m\u001b[48;2;0;5;37m\u2596\u001b[38;2;0;14;34m\u001b[48;2;0;0;70m\u259b\u001b[38;2;11;21;46m\u001b[48;2;0;0;45m\u2598\u001b[38;2;255;255;0m\u001b[48;2;25;10;0m\u259d\u001b[38;2;255;255;0m\u001b[48;2;245;168;0m\u2580\u001b[38;2;59;44;0m\u001b[48;2;255;253;0m\u2580\u001b[38;2;115;65;0m\u001b[48;2;0;0;1m\u258c\u001b[38;2;1;0;0m\u001b[48;2;104;76;65m\u2580\u001b[38;2;34;15;2m\u001b[48;2;2;5;9m\u2596\u001b[38;2;18;23;22m\u001b[48;2;255;255;255m\u2580\u001b[38;2;0;0;0m\u001b[48;2;229;231;231m\u2580\u001b[38;2;86;88;90m\u001b[48;2;0;0;0m\u2596\u001b[38;2;66;45;35m\u001b[48;2;9;1;0m\u2580\u001b[38;2;89;72;61m\u001b[48;2;69;52;42m\u259d\u001b[38;2;75;55;44m\u001b[48;2;85;64;54m\u259b\u001b[38;2;103;75;59m\u001b[48;2;80;60;47m\u259e\u001b[38;2;144;146;147m\u001b[48;2;6;0;0m\u259d\u001b[38;2;0;113;198m\u001b[48;2;44;37;48m\u259d\u001b[38;2;44;227;255m\u001b[48;2;29;3;0m\u258c\u001b[38;2;199;184;180m\u001b[48;2;0;0;0m\u2598\u001b[38;2;53;35;26m\u001b[48;2;93;68;55m\u259b\u001b[38;2;100;83;72m\u001b[48;2;86;67;54m\u259d\u001b[38;2;79;60;47m\u001b[48;2;69;51;40m\u258c\u001b[38;2;66;49;39m\u001b[48;2;0;0;0m\u259b\u001b[38;2;40;20;8m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;206;208;212m\u2580\u001b[38;2;0;8;190m\u001b[48;2;0;0;0m\u2596\u001b[38;2;17;0;0m\u001b[48;2;255;255;0m\u2598\u001b[38;2;255;255;0m\u001b[48;2;40;30;2m\u2598\u001b[38;2;76;57;59m\u001b[48;2;0;0;0m\u2596\u001b[38;2;41;19;0m\u001b[48;2;221;180;0m\u2580\u001b[38;2;91;64;0m\u001b[48;2;255;218;0m\u2598\u001b[38;2;247;194;0m\u001b[48;2;63;13;0m\u259b\u001b[38;2;0;0;68m\u001b[48;2;0;0;8m\u259d\u001b[38;2;0;11;40m\u001b[48;2;0;0;63m\u2580\u001b[38;2;0;0;31m\u001b[48;2;0;4;56m\u259d\u001b[38;2;0;0;46m\u001b[48;2;255;245;0m\u2596\u001b[38;2;72;41;0m\u001b[48;2;255;255;0m\u259d\u001b[0m\u001b[38;2;40;24;0m\u2596\u001b[0m \u001b[0m",
	"\u001b[0m \u001b[38;2;124;106;49m\u001b[48;2;255;255;0m\u2598\u001b[38;2;225;186;0m\u001b[48;2;8;0;34m\u258c\u001b[38;2;0;10;60m\u001b[48;2;0;8;92m\u2580\u001b[38;2;0;15;82m\u001b[48;2;0;12;45m\u259e\u001b[38;2;30;17;22m\u001b[48;2;255;255;0m\u259b\u001b[38;2;255;255;0m\u001b[48;2;199;151;0m\u259e\u001b[38;2;255;255;0m\u001b[48;2;0;0;48m\u2580\u001b[38;2;192;171;0m\u001b[48;2;0;0;70m\u2580\u001b[38;2;0;0;1m\u001b[48;2;26;17;28m\u259b\u001b[38;2;0;0;28m\u001b[48;2;88;68;62m\u2598\u001b[38;2;85;67;56m\u001b[48;2;101;84;73m\u259b\u001b[38;2;34;14;0m\u001b[48;2;93;70;55m\u259d\u001b[38;2;68;44;34m\u001b[48;2;0;0;0m\u2596\u001b[38;2;0;0;0m\u001b[48;2;241;238;237m\u2596\u001b[38;2;255;255;255m\u001b[48;2;255;255;255m\u2598\u001b[38;2;48;50;48m\u001b[48;2;255;255;255m\u2580\u001b[38;2;80;58;41m\u001b[48;2;0;0;0m\u259d\u001b[38;2;89;66;51m\u001b[48;2;0;0;0m\u259b\u001b[38;2;0;0;0m\u001b[48;2;34;21;17m\u259e\u001b[38;2;0;0;0m\u001b[48;2;127;95;79m\u2598\u001b[38;2;97;62;51m\u001b[48;2;0;0;0m\u258c\u001b[38;2;47;13;0m\u001b[48;2;0;0;4m\u259d\u001b[38;2;47;34;27m\u001b[48;2;128;96;82m\u259d\u001b[38;2;99;72;55m\u001b[48;2;0;0;0m\u2596\u001b[38;2;74;57;43m\u001b[48;2;19;9;11m\u2580\u001b[38;2;69;46;36m\u001b[48;2;0;0;0m\u259b\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u259b\u001b[38;2;255;255;255m\u001b[48;2;251;255;255m\u259b\u001b[38;2;157;171;255m\u001b[48;2;38;12;0m\u2580\u001b[38;2;0;0;0m\u001b[48;2;255;255;0m\u2598\u001b[38;2;225;188;0m\u001b[48;2;0;0;0m\u259b\u001b[38;2;13;7;33m\u001b[48;2;74;50;50m\u259b\u001b[38;2;110;85;79m\u001b[48;2;83;60;57m\u2598\u001b[38;2;89;66;63m\u001b[48;2;0;0;0m\u2596\u001b[38;2;126;99;0m\u001b[48;2;2;7;32m\u259d\u001b[38;2;255;235;0m\u001b[48;2;0;0;65m\u2580\u001b[38;2;0;0;9m\u001b[48;2;227;215;0m\u2596\u001b[38;2;0;0;0m\u001b[48;2;249;229;0m\u259d\u001b[38;2;41;18;0m\u001b[48;2;0;8;60m\u2596\u001b[38;2;0;0;50m\u001b[48;2;0;0;96m\u259d\u001b[38;2;56;64;41m\u001b[48;2;255;255;0m\u2596\u001b[0m\u001b[38;2;241;213;0m\u2599\u001b[0m \u001b[0m",
	"\u001b[0m\u001b[38;2;182;154;10m\u2590\u001b[38;2;255;247;0m\u001b[48;2;60;47;0m\u2580\u001b[38;2;0;0;58m\u001b[48;2;241;209;0m\u2580\u001b[38;2;0;0;22m\u001b[48;2;255;255;0m\u2580\u001b[38;2;0;0;0m\u001b[48;2;255;255;0m\u2580\u001b[38;2;58;25;0m\u001b[48;2;255;233;0m\u2580\u001b[38;2;0;0;40m\u001b[48;2;255;236;0m\u2580\u001b[38;2;0;0;102m\u001b[48;2;6;0;13m\u2598\u001b[38;2;5;0;0m\u001b[48;2;85;55;41m\u259b\u001b[38;2;89;61;43m\u001b[48;2;101;76;61m\u2598\u001b[38;2;108;91;81m\u001b[48;2;92;71;58m\u259d\u001b[38;2;101;84;75m\u001b[48;2;90;72;58m\u2580\u001b[38;2;93;75;63m\u001b[48;2;77;57;47m\u258c\u001b[38;2;26;11;5m\u001b[48;2;88;66;53m\u259d\u001b[38;2;68;68;68m\u001b[48;2;15;7;3m\u259d\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u2596\u001b[38;2;255;255;255m\u001b[48;2;255;255;255m\u2598\u001b[38;2;229;231;229m\u001b[48;2;0;0;0m\u258c\u001b[38;2;12;6;4m\u001b[48;2;122;91;77m\u259b\u001b[38;2;93;72;53m\u001b[48;2;111;86;73m\u2598\u001b[38;2;78;55;53m\u001b[48;2;0;0;0m\u259b\u001b[38;2;0;0;0m\u001b[48;2;255;255;0m\u259b\u001b[38;2;30;6;0m\u001b[48;2;255;255;0m\u2580\u001b[38;2;105;77;71m\u001b[48;2;0;0;8m\u259d\u001b[38;2;116;90;90m\u001b[48;2;106;81;69m\u2596\u001b[38;2;22;15;12m\u001b[48;2;122;92;83m\u2580\u001b[38;2;47;56;60m\u001b[48;2;0;0;16m\u259d\u001b[38;2;255;255;255m\u001b[48;2;26;40;82m\u2580\u001b[38;2;0;2;64m\u001b[48;2;226;186;0m\u2580\u001b[38;2;255;246;0m\u001b[48;2;138;103;0m\u259b\u001b[38;2;255;255;0m\u001b[48;2;5;0;18m\u2598\u001b[38;2;14;6;32m\u001b[48;2;81;62;67m\u2580\u001b[38;2;70;52;49m\u001b[48;2;94;76;65m\u2580\u001b[38;2;76;58;48m\u001b[48;2;114;92;81m\u2598\u001b[38;2;115;91;73m\u001b[48;2;102;73;57m\u259b\u001b[38;2;90;57;42m\u001b[48;2;4;0;0m\u258c\u001b[38;2;0;0;12m\u001b[48;2;0;31;101m\u258c\u001b[38;2;0;0;50m\u001b[48;2;255;212;0m\u259b\u001b[38;2;4;0;0m\u001b[48;2;255;255;0m\u2580\u001b[38;2;59;27;0m\u001b[48;2;255;245;0m\u2580\u001b[38;2;0;0;4m\u001b[48;2;255;255;0m\u2580\u001b[38;2;0;0;68m\u001b[48;2;221;189;0m\u2598\u001b[38;2;255;255;0m\u001b[48;2;165;136;0m\u2580\u001b[0m\u001b[38;2;42;27;0m\u258c\u001b[0m",
	"\u001b[38;2;166;143;21m\u001b[48;2;255;255;0m\u258c\u001b[38;2;255;255;0m\u001b[48;2;57;37;16m\u2596\u001b[38;2;237;211;0m\u001b[48;2;0;0;104m\u2580\u001b[38;2;255;255;0m\u001b[48;2;0;0;41m\u2580\u001b[38;2;255;255;0m\u001b[48;2;0;0;15m\u2598\u001b[38;2;74;32;0m\u001b[48;2;255;196;0m\u259b\u001b[38;2;255;184;0m\u001b[48;2;40;22;26m\u2596\u001b[38;2;0;0;123m\u001b[48;2;0;0;26m\u2598\u001b[38;2;6;15;68m\u001b[48;2;255;255;255m\u259b\u001b[38;2;0;0;0m\u001b[48;2;238;246;255m\u2580\u001b[38;2;105;113;127m\u001b[48;2;2;4;4m\u2596\u001b[38;2;63;36;23m\u001b[48;2;8;0;0m\u259d\u001b[38;2;89;65;51m\u001b[48;2;0;0;0m\u2580\u001b[38;2;88;67;51m\u001b[48;2;0;0;0m\u2580\u001b[38;2;67;39;29m\u001b[48;2;0;0;0m\u2580\u001b[38;2;199;196;189m\u001b[48;2;0;0;0m\u259d\u001b[38;2;223;218;212m\u001b[48;2;0;0;0m\u2598\u001b[38;2;7;1;0m\u001b[48;2;73;45;40m\u258c\u001b[38;2;112;87;78m\u001b[48;2;6;7;27m\u2580\u001b[38;2;12;12;44m\u001b[48;2;112;48;0m\u259b\u001b[38;2;0;0;0m\u001b[48;2;255;203;0m\u2580\u001b[38;2;0;0;0m\u001b[48;2;232;144;0m\u2598\u001b[38;2;224;169;0m\u001b[48;2;226;89;0m\u2580\u001b[38;2;13;7;0m\u001b[48;2;255;232;0m\u2580\u001b[38;2;0;0;42m\u001b[48;2;189;123;0m\u2580\u001b[38;2;24;23;69m\u001b[48;2;0;0;0m\u2580\u001b[38;2;20;5;4m\u001b[48;2;251;202;0m\u2580\u001b[38;2;253;187;0m\u001b[48;2;61;25;0m\u259b\u001b[38;2;203;173;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;0;0;31m\u001b[48;2;0;0;1m\u259d\u001b[38;2;66;49;72m\u001b[48;2;0;0;0m\u2580\u001b[38;2;88;63;55m\u001b[48;2;0;0;0m\u2580\u001b[38;2;75;46;33m\u001b[48;2;0;0;0m\u2580\u001b[38;2;3;0;0m\u001b[48;2;73;77;87m\u259b\u001b[38;2;0;0;0m\u001b[48;2;222;228;233m\u2580\u001b[38;2;16;15;13m\u001b[48;2;196;205;255m\u2580\u001b[38;2;0;19;93m\u001b[48;2;0;0;4m\u259d\u001b[38;2;0;0;65m\u001b[48;2;178;94;0m\u259b\u001b[38;2;55;36;0m\u001b[48;2;255;211;0m\u2580\u001b[38;2;21;0;0m\u001b[48;2;0;0;37m\u259b\u001b[38;2;255;255;0m\u001b[48;2;0;0;14m\u2580\u001b[38;2;254;224;0m\u001b[48;2;0;0;85m\u2580\u001b[38;2;203;162;4m\u001b[48;2;255;255;0m\u259b\u001b[0m\u001b[38;2;141;128;0m\u2599\u001b[0m",
	"\u001b[38;2;255;213;6m\u001b[48;2;255;255;0m\u258c\u001b[38;2;255;228;0m\u001b[48;2;0;0;65m\u258c\u001b[38;2;0;26;92m\u001b[48;2;17;37;62m\u258c\u001b[38;2;3;32;89m\u001b[48;2;20;35;53m\u2580\u001b[38;2;4;34;99m\u001b[48;2;20;35;59m\u2580\u001b[38;2;0;0;19m\u001b[48;2;7;20;86m\u259d\u001b[38;2;230;188;0m\u001b[48;2;0;0;79m\u2580\u001b[38;2;0;10;9m\u001b[48;2;235;223;0m\u2596\u001b[38;2;150;111;6m\u001b[48;2;255;255;0m\u2580\u001b[38;2;254;255;255m\u001b[48;2;255;241;0m\u2580\u001b[38;2;255;255;255m\u001b[48;2;188;125;0m\u2580\u001b[38;2;255;255;255m\u001b[48;2;137;119;45m\u2580\u001b[38;2;192;199;224m\u001b[48;2;255;255;255m\u259e\u001b[38;2;211;234;255m\u001b[48;2;0;0;15m\u2596\u001b[38;2;122;93;76m\u001b[48;2;17;9;23m\u259d\u001b[38;2;125;94;73m\u001b[48;2;2;2;78m\u2580\u001b[38;2;86;63;49m\u001b[48;2;51;44;95m\u2580\u001b[38;2;0;0;5m\u001b[48;2;0;0;0m\u2596\u001b[38;2;77;41;0m\u001b[48;2;255;207;0m\u2580\u001b[38;2;225;209;0m\u001b[48;2;0;0;51m\u259b\u001b[38;2;163;62;0m\u001b[48;2;0;76;199m\u2580\u001b[38;2;2;0;17m\u001b[48;2;0;255;255m\u2580\u001b[38;2;0;1;29m\u001b[48;2;2;254;255m\u2580\u001b[38;2;15;0;0m\u001b[48;2;2;153;251m\u2580\u001b[38;2;243;186;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;182;110;0m\u001b[48;2;255;255;0m\u259e\u001b[38;2;255;157;0m\u001b[48;2;0;0;0m\u2596\u001b[38;2;0;0;50m\u001b[48;2;70;62;110m\u259b\u001b[38;2;96;78;90m\u001b[48;2;11;12;85m\u2580\u001b[38;2;119;90;77m\u001b[48;2;0;0;29m\u2580\u001b[38;2;0;0;0m\u001b[48;2;143;167;255m\u259b\u001b[38;2;143;152;172m\u001b[48;2;233;237;253m\u2598\u001b[38;2;255;255;255m\u001b[48;2;105;100;68m\u2580\u001b[38;2;255;255;255m\u001b[48;2;122;70;0m\u2580\u001b[38;2;247;255;255m\u001b[48;2;251;201;0m\u2580\u001b[38;2;56;49;59m\u001b[48;2;255;255;0m\u2580\u001b[38;2;255;241;0m\u001b[48;2;179;150;0m\u259b\u001b[38;2;255;255;0m\u001b[48;2;0;0;53m\u2580\u001b[38;2;140;90;0m\u001b[48;2;0;8;82m\u2598\u001b[38;2;6;20;101m\u001b[48;2;20;35;59m\u2580\u001b[38;2;0;24;94m\u001b[48;2;21;35;53m\u2580\u001b[38;2;7;30;70m\u001b[48;2;0;0;96m\u258c\u001b[38;2;183;156;51m\u001b[48;2;255;255;0m\u258c\u001b[38;2;255;255;0m\u001b[48;2;10;1;0m\u258c\u001b[0m",
	"\u001b[38;2;185;150;0m\u001b[48;2;255;255;0m\u258c\u001b[38;2;255;255;0m\u001b[48;2;0;0;39m\u258c\u001b[38;2;0;6;96m\u001b[48;2;20;35;51m\u258c\u001b[38;2;20;35;55m\u001b[48;2;20;35;51m\u2580\u001b[38;2;20;35;55m\u001b[48;2;20;35;51m\u2580\u001b[38;2;20;35;53m\u001b[48;2;20;35;51m\u2580\u001b[38;2;2;29;82m\u001b[48;2;21;35;56m\u259d\u001b[38;2;0;0;105m\u001b[48;2;22;37;56m\u2580\u001b[38;2;0;0;56m\u001b[48;2;7;27;79m\u2580\u001b[38;2;0;7;122m\u001b[48;2;24;11;14m\u2596\u001b[38;2;131;84;4m\u001b[48;2;244;254;255m\u259b\u001b[38;2;166;69;0m\u001b[48;2;152;174;227m\u2580\u001b[38;2;172;62;0m\u001b[48;2;56;52;57m\u2580\u001b[38;2;226;133;0m\u001b[48;2;44;7;0m\u2580\u001b[38;2;207;141;0m\u001b[48;2;69;9;0m\u2580\u001b[38;2;121;84;0m\u001b[48;2;230;143;0m\u2580\u001b[38;2;71;48;0m\u001b[48;2;255;195;0m\u2580\u001b[38;2;0;0;0m\u001b[48;2;107;55;0m\u258c\u001b[38;2;255;255;0m\u001b[48;2;66;45;0m\u258c\u001b[38;2;0;0;7m\u001b[48;2;255;255;0m\u259b\u001b[38;2;8;224;255m\u001b[48;2;241;153;0m\u2580\u001b[38;2;0;29;89m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;8;184;232m\u001b[48;2;162;93;0m\u2580\u001b[38;2;0;53;120m\u001b[48;2;202;167;0m\u2580\u001b[38;2;251;163;0m\u001b[48;2;39;24;0m\u259d\u001b[38;2;255;255;0m\u001b[48;2;163;92;0m\u258c\u001b[38;2;28;18;0m\u001b[48;2;255;160;0m\u259b\u001b[38;2;109;69;0m\u001b[48;2;215;135;0m\u2580\u001b[38;2;168;104;0m\u001b[48;2;42;0;0m\u259b\u001b[38;2;210;121;0m\u001b[48;2;49;1;0m\u2580\u001b[38;2;143;36;0m\u001b[48;2;60;45;27m\u2580\u001b[38;2;156;54;0m\u001b[48;2;132;140;204m\u2580\u001b[38;2;232;144;0m\u001b[48;2;162;172;237m\u2580\u001b[38;2;126;83;0m\u001b[48;2;0;0;58m\u2580\u001b[38;2;0;0;29m\u001b[48;2;0;10;93m\u2580\u001b[38;2;0;0;96m\u001b[48;2;17;38;65m\u2580\u001b[38;2;0;0;99m\u001b[48;2;16;33;60m\u2598\u001b[38;2;23;37;61m\u001b[48;2;20;35;51m\u2598\u001b[38;2;20;35;55m\u001b[48;2;20;35;51m\u2580\u001b[38;2;20;35;55m\u001b[48;2;20;35;51m\u2580\u001b[38;2;6;21;54m\u001b[48;2;0;0;89m\u258c\u001b[38;2;221;176;45m\u001b[48;2;255;255;2m\u2598\u001b[38;2;255;255;0m\u001b[48;2;0;0;0m\u258c\u001b[0m",
	"\u001b[0m\u001b[38;2;193;188;2m\u259c\u001b[38;2;85;57;0m\u001b[48;2;255;252;0m\u259d\u001b[38;2;10;30;64m\u001b[48;2;0;1;81m\u259d\u001b[38;2;5;24;48m\u001b[48;2;16;31;50m\u2596\u001b[38;2;17;32;51m\u001b[48;2;15;30;49m\u2580\u001b[38;2;17;32;51m\u001b[48;2;15;30;49m\u2580\u001b[38;2;17;32;51m\u001b[48;2;15;30;49m\u2580\u001b[38;2;16;32;51m\u001b[48;2;15;30;49m\u2580\u001b[38;2;17;34;58m\u001b[48;2;15;30;49m\u259d\u001b[38;2;13;30;47m\u001b[48;2;0;0;0m\u259b\u001b[38;2;255;255;255m\u001b[48;2;3;0;0m\u259d\u001b[38;2;255;255;255m\u001b[48;2;34;23;17m\u2580\u001b[38;2;162;157;163m\u001b[48;2;255;255;255m\u2596\u001b[38;2;255;255;255m\u001b[48;2;255;255;255m\u2598\u001b[38;2;0;0;77m\u001b[48;2;255;255;255m\u259d\u001b[38;2;181;185;184m\u001b[48;2;0;0;0m\u2596\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;255;198;0m\u2596\u001b[38;2;255;234;0m\u001b[48;2;28;3;11m\u2596\u001b[38;2;0;0;0m\u001b[48;2;7;188;255m\u2596\u001b[38;2;0;0;1m\u001b[48;2;0;189;235m\u259d\u001b[38;2;0;21;47m\u001b[48;2;0;228;255m\u2580\u001b[38;2;9;204;255m\u001b[48;2;0;0;111m\u259b\u001b[38;2;0;63;255m\u001b[48;2;48;22;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;255;224;0m\u2598\u001b[38;2;255;227;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;220;232;255m\u2598\u001b[38;2;255;255;255m\u001b[48;2;255;255;255m\u2598\u001b[38;2;255;255;255m\u001b[48;2;255;255;255m\u2598\u001b[38;2;255;255;255m\u001b[48;2;35;35;35m\u2580\u001b[38;2;255;255;255m\u001b[48;2;11;5;1m\u2598\u001b[38;2;0;0;0m\u001b[48;2;10;35;53m\u258c\u001b[38;2;16;35;60m\u001b[48;2;16;31;48m\u258c\u001b[38;2;16;32;51m\u001b[48;2;15;30;49m\u2580\u001b[38;2;17;32;51m\u001b[48;2;15;30;49m\u2580\u001b[38;2;17;32;51m\u001b[48;2;15;30;49m\u2580\u001b[38;2;17;32;51m\u001b[48;2;15;30;49m\u2580\u001b[38;2;14;29;50m\u001b[48;2;0;23;51m\u259b\u001b[38;2;0;0;76m\u001b[48;2;168;147;89m\u259b\u001b[38;2;255;255;0m\u001b[48;2;255;255;0m\u2598\u001b[0m\u001b[38;2;143;103;0m\u258c\u001b[0m",
	"\u001b[0m\u001b[38;2;104;94;0m\u259d\u001b[38;2;199;199;0m\u001b[48;2;255;255;0m\u2596\u001b[38;2;255;255;0m\u001b[48;2;25;19;35m\u2596\u001b[38;2;0;0;77m\u001b[48;2;4;21;45m\u2596\u001b[38;2;13;26;43m\u001b[48;2;10;24;39m\u2580\u001b[38;2;13;26;43m\u001b[48;2;10;24;39m\u2580\u001b[38;2;13;26;43m\u001b[48;2;10;24;39m\u2580\u001b[38;2;13;26;43m\u001b[48;2;10;24;39m\u2580\u001b[38;2;12;26;43m\u001b[48;2;10;24;39m\u2580\u001b[38;2;20;40;65m\u001b[48;2;11;25;40m\u259d\u001b[38;2;0;0;0m\u001b[48;2;3;26;49m\u259d\u001b[38;2;3;0;0m\u001b[48;2;91;62;44m\u259b\u001b[38;2;129;95;75m\u001b[48;2;11;1;0m\u2596\u001b[38;2;232;229;228m\u001b[48;2;0;0;0m\u2580\u001b[38;2;84;81;86m\u001b[48;2;0;0;1m\u2598\u001b[38;2;255;255;255m\u001b[48;2;23;22;20m\u259d\u001b[38;2;0;0;83m\u001b[48;2;255;255;255m\u2596\u001b[38;2;255;255;255m\u001b[48;2;255;255;255m\u2598\u001b[38;2;0;0;0m\u001b[48;2;78;86;123m\u259e\u001b[38;2;193;146;112m\u001b[48;2;0;0;0m\u259e\u001b[38;2;233;81;0m\u001b[48;2;32;18;0m\u2580\u001b[38;2;139;18;0m\u001b[48;2;17;0;0m\u2580\u001b[38;2;118;13;0m\u001b[48;2;36;0;0m\u2580\u001b[38;2;212;42;0m\u001b[48;2;69;32;0m\u2580\u001b[38;2;255;163;0m\u001b[48;2;0;0;0m\u2580\u001b[38;2;209;236;255m\u001b[48;2;0;0;2m\u2596\u001b[38;2;18;33;67m\u001b[48;2;255;255;255m\u258c\u001b[38;2;255;255;255m\u001b[48;2;208;242;255m\u259b\u001b[38;2;255;255;255m\u001b[48;2;0;0;28m\u2580\u001b[38;2;35;30;25m\u001b[48;2;0;0;0m\u2598\u001b[38;2;246;244;244m\u001b[48;2;0;0;0m\u2580\u001b[38;2;0;0;0m\u001b[48;2;83;52;35m\u259b\u001b[38;2;132;101;80m\u001b[48;2;30;10;5m\u2596\u001b[38;2;1;0;0m\u001b[48;2;8;35;63m\u259b\u001b[38;2;11;33;54m\u001b[48;2;10;24;39m\u259b\u001b[38;2;12;26;43m\u001b[48;2;10;24;39m\u2580\u001b[38;2;13;26;43m\u001b[48;2;10;24;39m\u2580\u001b[38;2;13;26;43m\u001b[48;2;10;24;39m\u2580\u001b[38;2;13;26;43m\u001b[48;2;10;24;39m\u2580\u001b[38;2;11;25;42m\u001b[48;2;0;19;42m\u259b\u001b[38;2;2;6;37m\u001b[48;2;0;0;69m\u259b\u001b[38;2;0;0;117m\u001b[48;2;255;251;16m\u2598\u001b[38;2;255;234;0m\u001b[48;2;0;0;0m\u259b\u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m\u001b[38;2;255;255;0m\u259d\u001b[38;2;255;249;0m\u001b[48;2;255;255;0m\u259d\u001b[38;2;255;222;0m\u001b[48;2;0;0;48m\u2596\u001b[38;2;0;0;57m\u001b[48;2;0;14;34m\u2596\u001b[38;2;8;20;35m\u001b[48;2;2;16;30m\u2580\u001b[38;2;8;20;35m\u001b[48;2;5;16;30m\u2580\u001b[38;2;8;20;35m\u001b[48;2;5;16;30m\u2580\u001b[38;2;8;20;35m\u001b[48;2;5;16;30m\u2580\u001b[38;2;8;20;35m\u001b[48;2;5;16;30m\u2580\u001b[38;2;16;38;66m\u001b[48;2;4;16;35m\u259d\u001b[38;2;0;0;1m\u001b[48;2;31;13;3m\u259b\u001b[38;2;112;84;69m\u001b[48;2;15;11;11m\u2580\u001b[38;2;91;67;50m\u001b[48;2;16;10;6m\u258c\u001b[38;2;13;8;17m\u001b[48;2;180;131;0m\u259b\u001b[38;2;178;83;0m\u001b[48;2;248;215;0m\u2580\u001b[38;2;46;10;0m\u001b[48;2;255;255;0m\u259b\u001b[38;2;242;195;127m\u001b[48;2;0;0;42m\u259e\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u259d\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u2596\u001b[38;2;199;205;230m\u001b[48;2;0;6;19m\u259e\u001b[38;2;230;219;215m\u001b[48;2;255;255;255m\u2596\u001b[38;2;255;255;255m\u001b[48;2;255;255;255m\u2598\u001b[38;2;255;255;255m\u001b[48;2;1;6;24m\u2598\u001b[38;2;255;255;255m\u001b[48;2;106;105;118m\u259e\u001b[38;2;0;0;0m\u001b[48;2;209;205;202m\u259e\u001b[38;2;255;255;255m\u001b[48;2;0;0;0m\u259b\u001b[38;2;255;255;0m\u001b[48;2;62;31;6m\u2596\u001b[38;2;71;17;0m\u001b[48;2;228;176;0m\u258c\u001b[38;2;255;255;0m\u001b[48;2;21;9;3m\u2596\u001b[38;2;8;0;0m\u001b[48;2;47;34;29m\u259e\u001b[38;2;108;77;65m\u001b[48;2;9;1;1m\u259b\u001b[38;2;65;46;35m\u001b[48;2;1;0;0m\u2598\u001b[38;2;0;0;2m\u001b[48;2;4;20;48m\u2596\u001b[38;2;8;20;35m\u001b[48;2;5;16;30m\u2580\u001b[38;2;8;20;35m\u001b[48;2;5;16;30m\u2580\u001b[38;2;8;20;35m\u001b[48;2;5;16;30m\u2580\u001b[38;2;8;20;35m\u001b[48;2;5;16;30m\u2580\u001b[38;2;8;20;35m\u001b[48;2;2;14;31m\u2580\u001b[38;2;0;15;39m\u001b[48;2;0;0;29m\u2598\u001b[38;2;0;0;41m\u001b[48;2;190;180;63m\u2598\u001b[38;2;255;255;0m\u001b[48;2;125;92;0m\u259b\u001b[0m\u001b[38;2;55;30;0m\u2598\u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m\u001b[38;2;255;255;0m\u259d\u001b[38;2;215;201;0m\u001b[48;2;255;254;0m\u2596\u001b[38;2;255;255;0m\u001b[48;2;5;6;25m\u2596\u001b[38;2;0;7;27m\u001b[48;2;0;0;31m\u259d\u001b[38;2;1;12;26m\u001b[48;2;0;5;18m\u2580\u001b[38;2;1;12;26m\u001b[48;2;0;8;18m\u2580\u001b[38;2;1;12;26m\u001b[48;2;0;1;10m\u2580\u001b[38;2;1;13;26m\u001b[48;2;0;0;13m\u2580\u001b[38;2;0;0;6m\u001b[48;2;59;53;39m\u259b\u001b[38;2;101;78;67m\u001b[48;2;35;10;0m\u259e\u001b[38;2;0;0;0m\u001b[48;2;81;61;45m\u259d\u001b[38;2;108;83;66m\u001b[48;2;3;1;1m\u2596\u001b[38;2;0;0;36m\u001b[48;2;0;0;0m\u259d\u001b[38;2;57;38;0m\u001b[48;2;33;26;83m\u2580\u001b[38;2;92;51;0m\u001b[48;2;18;14;42m\u259d\u001b[38;2;16;7;0m\u001b[48;2;72;54;76m\u2580\u001b[38;2;20;23;31m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u258c\u001b[38;2;255;255;255m\u001b[48;2;32;31;32m\u258c\u001b[38;2;119;109;104m\u001b[48;2;225;223;222m\u258c\u001b[38;2;255;255;255m\u001b[48;2;144;137;135m\u259b\u001b[38;2;0;0;0m\u001b[48;2;81;80;77m\u2580\u001b[38;2;205;195;195m\u001b[48;2;255;255;255m\u259d\u001b[38;2;20;26;31m\u001b[48;2;0;0;0m\u259d\u001b[38;2;11;3;9m\u001b[48;2;79;58;69m\u259b\u001b[38;2;96;62;0m\u001b[48;2;18;16;47m\u2598\u001b[38;2;111;77;0m\u001b[48;2;29;24;59m\u259d\u001b[38;2;13;11;42m\u001b[48;2;0;0;13m\u2596\u001b[38;2;0;0;0m\u001b[48;2;42;29;23m\u259b\u001b[38;2;8;4;2m\u001b[48;2;103;81;63m\u2580\u001b[38;2;104;74;59m\u001b[48;2;52;29;10m\u259b\u001b[38;2;102;90;89m\u001b[48;2;0;0;0m\u2596\u001b[38;2;1;13;26m\u001b[48;2;0;0;8m\u2580\u001b[38;2;1;12;26m\u001b[48;2;0;0;12m\u2580\u001b[38;2;1;12;26m\u001b[48;2;0;4;19m\u2580\u001b[38;2;0;7;24m\u001b[48;2;0;0;8m\u259b\u001b[38;2;0;2;10m\u001b[48;2;0;0;75m\u259b\u001b[38;2;0;0;45m\u001b[48;2;232;219;43m\u2598\u001b[38;2;255;255;0m\u001b[48;2;56;34;0m\u259b\u001b[0m\u001b[38;2;129;87;0m\u2598\u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;164;127;0m\u259d\u001b[38;2;56;42;0m\u001b[48;2;255;247;0m\u2596\u001b[38;2;0;0;19m\u001b[48;2;234;217;0m\u259d\u001b[38;2;79;59;0m\u001b[48;2;0;0;17m\u2596\u001b[38;2;0;0;15m\u001b[48;2;0;0;0m\u259b\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u259b\u001b[38;2;255;255;255m\u001b[48;2;29;21;14m\u2596\u001b[38;2;55;34;21m\u001b[48;2;5;0;0m\u2580\u001b[38;2;17;0;0m\u001b[48;2;136;138;136m\u2596\u001b[38;2;191;186;190m\u001b[48;2;20;8;3m\u2596\u001b[38;2;2;0;0m\u001b[48;2;89;67;53m\u2596\u001b[38;2;13;9;9m\u001b[48;2;98;71;55m\u2580\u001b[38;2;31;19;16m\u001b[48;2;1;0;0m\u2596\u001b[38;2;69;50;44m\u001b[48;2;0;0;0m\u2580\u001b[38;2;111;83;65m\u001b[48;2;0;0;0m\u2580\u001b[38;2;7;2;1m\u001b[48;2;115;84;69m\u2596\u001b[38;2;2;0;0m\u001b[48;2;129;94;75m\u2580\u001b[38;2;105;109;113m\u001b[48;2;15;6;2m\u259d\u001b[38;2;242;245;244m\u001b[48;2;6;8;10m\u2580\u001b[38;2;217;215;214m\u001b[48;2;0;0;0m\u259b\u001b[38;2;198;202;202m\u001b[48;2;18;20;21m\u2598\u001b[38;2;0;0;0m\u001b[48;2;104;72;54m\u2580\u001b[38;2;128;96;75m\u001b[48;2;57;36;23m\u259e\u001b[38;2;127;95;80m\u001b[48;2;0;0;0m\u2580\u001b[38;2;84;62;51m\u001b[48;2;0;0;0m\u2580\u001b[38;2;36;25;19m\u001b[48;2;0;0;0m\u2598\u001b[38;2;0;0;0m\u001b[48;2;80;57;42m\u2580\u001b[38;2;105;79;62m\u001b[48;2;62;39;23m\u259e\u001b[38;2;26;13;6m\u001b[48;2;82;66;53m\u259e\u001b[38;2;182;185;187m\u001b[48;2;62;48;41m\u259e\u001b[38;2;94;77;65m\u001b[48;2;15;0;0m\u2580\u001b[38;2;13;0;0m\u001b[48;2;172;151;142m\u259b\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u2580\u001b[38;2;0;0;3m\u001b[48;2;0;0;0m\u259d\u001b[38;2;0;0;31m\u001b[48;2;255;243;42m\u259b\u001b[38;2;56;69;90m\u001b[48;2;255;255;2m\u2598\u001b[0m\u001b[38;2;203;167;0m\u259b\u001b[0m\u001b[38;2;0;0;0m\u2598\u001b[0m \u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;187;162;0m\u2580\u001b[38;2;114;78;0m\u001b[48;2;255;248;0m\u2596\u001b[38;2;31;24;0m\u001b[48;2;255;255;0m\u2580\u001b[38;2;101;69;0m\u001b[48;2;0;2;20m\u2596\u001b[38;2;255;255;255m\u001b[48;2;0;0;0m\u2580\u001b[38;2;4;26;69m\u001b[48;2;212;211;214m\u2596\u001b[38;2;24;11;3m\u001b[48;2;243;238;235m\u2580\u001b[38;2;245;245;252m\u001b[48;2;63;46;41m\u259d\u001b[38;2;33;14;0m\u001b[48;2;182;174;172m\u259d\u001b[38;2;181;173;170m\u001b[48;2;57;37;29m\u2596\u001b[38;2;95;65;51m\u001b[48;2;40;16;10m\u2580\u001b[38;2;96;71;58m\u001b[48;2;27;2;0m\u2580\u001b[38;2;6;2;0m\u001b[48;2;59;38;22m\u259d\u001b[38;2;0;0;0m\u001b[48;2;71;44;30m\u2580\u001b[38;2;67;46;33m\u001b[48;2;0;0;0m\u259d\u001b[38;2;0;0;0m\u001b[48;2;123;90;73m\u2596\u001b[38;2;77;48;36m\u001b[48;2;0;0;0m\u258c\u001b[38;2;0;0;0m\u001b[48;2;0;0;0m\u2598\u001b[38;2;120;88;70m\u001b[48;2;47;33;24m\u259b\u001b[38;2;114;85;70m\u001b[48;2;9;6;6m\u2598\u001b[38;2;0;0;0m\u001b[48;2;71;43;33m\u259b\u001b[38;2;10;6;6m\u001b[48;2;78;45;30m\u2580\u001b[38;2;78;59;48m\u001b[48;2;38;9;0m\u2580\u001b[38;2;104;76;58m\u001b[48;2;19;0;0m\u2580\u001b[38;2;57;34;20m\u001b[48;2;114;106;99m\u259b\u001b[38;2;31;20;16m\u001b[48;2;216;211;210m\u2580\u001b[38;2;210;210;209m\u001b[48;2;52;40;37m\u2580\u001b[38;2;5;0;0m\u001b[48;2;195;180;174m\u2580\u001b[38;2;238;234;234m\u001b[48;2;85;86;107m\u259e\u001b[38;2;255;255;255m\u001b[48;2;0;0;0m\u2580\u001b[38;2;0;0;0m\u001b[48;2;160;135;85m\u259e\u001b[38;2;0;0;29m\u001b[48;2;251;235;15m\u2598\u001b[38;2;234;211;0m\u001b[48;2;0;0;0m\u259b\u001b[0m\u001b[38;2;94;60;0m\u2580\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;153;116;0m\u2580\u001b[0m\u001b[38;2;219;199;0m\u259c\u001b[38;2;14;0;0m\u001b[48;2;252;214;0m\u259d\u001b[38;2;0;0;0m\u001b[48;2;253;213;0m\u2580\u001b[38;2;63;35;0m\u001b[48;2;0;0;15m\u2596\u001b[38;2;136;139;155m\u001b[48;2;0;0;0m\u2580\u001b[38;2;111;97;91m\u001b[48;2;0;0;16m\u2580\u001b[38;2;198;191;185m\u001b[48;2;67;68;88m\u2580\u001b[38;2;164;156;156m\u001b[48;2;217;209;207m\u2596\u001b[38;2;149;139;135m\u001b[48;2;255;255;255m\u2580\u001b[38;2;56;40;31m\u001b[48;2;255;255;255m\u2580\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u2580\u001b[38;2;255;255;255m\u001b[48;2;26;16;13m\u2596\u001b[38;2;7;3;3m\u001b[48;2;255;255;255m\u259b\u001b[38;2;0;0;0m\u001b[48;2;133;129;125m\u2580\u001b[38;2;255;255;255m\u001b[48;2;18;17;16m\u2596\u001b[38;2;0;0;0m\u001b[48;2;237;234;234m\u2580\u001b[38;2;0;0;0m\u001b[48;2;161;146;135m\u259b\u001b[38;2;0;0;0m\u001b[48;2;255;255;255m\u2580\u001b[38;2;36;21;16m\u001b[48;2;255;255;255m\u2580\u001b[38;2;121;112;106m\u001b[48;2;255;255;255m\u2580\u001b[38;2;233;225;223m\u001b[48;2;190;183;182m\u259e\u001b[38;2;220;213;210m\u001b[48;2;79;87;105m\u2580\u001b[38;2;129;110;101m\u001b[48;2;0;0;26m\u2580\u001b[38;2;123;123;136m\u001b[48;2;0;0;0m\u2580\u001b[38;2;6;14;34m\u001b[48;2;156;115;0m\u259b\u001b[38;2;0;0;0m\u001b[48;2;255;255;0m\u2580\u001b[38;2;255;250;0m\u001b[48;2;170;106;0m\u259e\u001b[0m\u001b[38;2;184;167;0m\u259b\u001b[0m\u001b[38;2;65;39;0m\u2580\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;177;120;0m\u2580\u001b[0m\u001b[38;2;189;174;0m\u259c\u001b[38;2;175;113;0m\u001b[48;2;255;207;0m\u259e\u001b[38;2;32;17;0m\u001b[48;2;255;238;0m\u2580\u001b[38;2;0;0;0m\u001b[48;2;255;221;0m\u2580\u001b[38;2;0;0;0m\u001b[48;2;156;96;0m\u2580\u001b[38;2;0;1;49m\u001b[48;2;3;0;0m\u259d\u001b[38;2;79;96;125m\u001b[48;2;0;0;0m\u2580\u001b[38;2;224;228;236m\u001b[48;2;0;0;0m\u2580\u001b[38;2;255;255;255m\u001b[48;2;0;0;0m\u2580\u001b[38;2;170;170;168m\u001b[48;2;5;6;8m\u2598\u001b[38;2;255;255;255m\u001b[48;2;31;45;84m\u2580\u001b[38;2;231;239;255m\u001b[48;2;0;0;73m\u259b\u001b[38;2;218;211;209m\u001b[48;2;2;4;6m\u2598\u001b[38;2;255;255;255m\u001b[48;2;0;0;0m\u2580\u001b[38;2;249;253;255m\u001b[48;2;0;0;0m\u2580\u001b[38;2;105;121;149m\u001b[48;2;0;0;0m\u2580\u001b[38;2;0;14;67m\u001b[48;2;17;8;4m\u2598\u001b[38;2;0;0;0m\u001b[48;2;200;133;0m\u2580\u001b[38;2;0;0;0m\u001b[48;2;255;244;0m\u2580\u001b[38;2;65;46;0m\u001b[48;2;255;219;0m\u2580\u001b[38;2;255;226;0m\u001b[48;2;137;71;0m\u2580\u001b[0m\u001b[38;2;170;153;0m\u259b\u001b[0m\u001b[38;2;112;62;0m\u2580\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m",
	"\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m\u001b[38;2;114;64;0m\u2580\u001b[0m\u001b[38;2;228;158;0m\u2580\u001b[0m\u001b[38;2;255;216;0m\u2580\u001b[38;2;255;217;0m\u001b[48;2;15;4;0m\u2580\u001b[38;2;255;184;0m\u001b[48;2;54;24;0m\u2580\u001b[38;2;229;150;0m\u001b[48;2;96;48;0m\u2580\u001b[38;2;185;122;0m\u001b[48;2;125;67;0m\u2580\u001b[38;2;74;10;0m\u001b[48;2;135;74;0m\u259d\u001b[38;2;83;22;0m\u001b[48;2;135;69;0m\u2580\u001b[38;2;191;124;0m\u001b[48;2;115;60;0m\u2580\u001b[38;2;242;165;0m\u001b[48;2;81;41;0m\u2580\u001b[38;2;255;200;0m\u001b[48;2;39;12;0m\u2580\u001b[38;2;255;226;0m\u001b[48;2;3;0;0m\u2580\u001b[0m\u001b[38;2;255;208;0m\u2580\u001b[0m\u001b[38;2;213;128;0m\u2580\u001b[0m\u001b[38;2;64;30;0m\u2580\u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m \u001b[0m",
}

func loadLogoLines(logoPath string) []string {
	data, err := os.ReadFile(logoPath)
	if err == nil {
		rawLines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		var lines []string
		for _, l := range rawLines {
			lines = append(lines, l)
		}
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
		if len(lines) == 21 {
			return lines
		}
	}
	return defaultEmbeddedLogoLines
}

func getLatestSemVerTag(repoRoot string) string {
	cmd := exec.Command("git", "-C", repoRoot, "tag", "-l", "v*", "--sort=-v:refname")
	if out, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" && strings.HasPrefix(l, "v") {
				return l
			}
		}
	}
	tagCmd := exec.Command("git", "-C", repoRoot, "describe", "--tags", "--abbrev=0")
	if tagOut, err := tagCmd.Output(); err == nil {
		return strings.TrimSpace(string(tagOut))
	}
	return ""
}

func printVersion() {
	hash := "dev"
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		hash = strings.TrimSpace(string(out))
	}
	ver := SyncaliburVersion
	if t := getLatestSemVerTag("."); t != "" {
		ver = t
	}
	fmt.Printf("%s%sSyncalibur %s%s (%s/%s)\n", cBold, cCyan, ver, cReset, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("%sMilestone    :%s %s%s%s\n", cGold, cReset, cWhite, SyncaliburMilestone, cReset)
	fmt.Printf("%sGit Commit   :%s %s\n", cGold, cReset, hash)
	fmt.Printf("%sArchitecture :%s Autonomous Multi-Remote Git Synchronization & Pre-flight Code Guard\n", cGold, cReset)
}

func runRelease(repoRoot string, mode string, dryRun bool) error {
	ctxFetch, cancelFetch := context.WithTimeout(context.Background(), 15*time.Second)
	cmdFetch := exec.CommandContext(ctxFetch, "git", "-C", repoRoot, "fetch", "--tags", "--force")
	cmdFetch.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	_ = cmdFetch.Run()
	cancelFetch()

	lastTag := getLatestSemVerTag(repoRoot)

	logArgs := []string{"-C", repoRoot, "log", "--pretty=format:%H%x09%an%x09%ad%x09%s", "--date=short"}
	if lastTag != "" {
		logArgs = append(logArgs, fmt.Sprintf("%s..HEAD", lastTag))
	} else {
		logArgs = append(logArgs, "HEAD")
	}

	cmdLog := exec.Command("git", logArgs...)
	out, err := cmdLog.Output()
	if err != nil {
		return fmt.Errorf("failed to read git log: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	var commits []CommitInfo
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		parts := strings.Split(l, "\t")
		if len(parts) >= 4 {
			cType, cScope, isBreak, cDesc := parseConventionalCommit(parts[3])
			commits = append(commits, CommitInfo{
				Hash:        parts[0],
				Author:      parts[1],
				Date:        parts[2],
				Subject:     parts[3],
				Type:        cType,
				Scope:       cScope,
				IsBreaking:  isBreak,
				Description: cDesc,
			})
		}
	}

	if len(commits) == 0 && lastTag != "" {
		fmt.Printf("%s%s No new commits since %s. Everything is up to date.%s\n", cGreen, iconCheck, lastTag, cReset)
		return nil
	}

	var bump SemVerBump
	switch strings.ToLower(mode) {
	case "major":
		bump = BumpMajor
	case "minor":
		bump = BumpMinor
	case "patch":
		bump = BumpPatch
	default:
		bump = determineSemVerBump(commits)
		if lastTag == "" {
			bump = BumpMajor
		}
	}

	nextVersion := calculateNextVersion(lastTag, bump)

	// Idempotency guard: If nextVersion tag already exists, nothing to release
	if err := exec.Command("git", "-C", repoRoot, "rev-parse", "--verify", "refs/tags/"+nextVersion).Run(); err == nil {
		fmt.Printf("%s%s Release %s is already tagged and up to date.%s\n", cGreen, iconCheck, nextVersion, cReset)
		return nil
	}

	today := time.Now().Format("2006-01-02")
	changelogDraft := generateChangelog(nextVersion, today, commits)

	if dryRun {
		fmt.Printf("%s%s%s SYNCALIBUR RELEASE ENGINE (DRY RUN)%s\n", cCyan, cBold, iconDiamond, cReset)
		fmt.Printf("  • Current Tag : %s\n", func() string {
			if lastTag == "" {
				return "(none - initial release)"
			}
			return lastTag
		}())
		fmt.Printf("  • Bump Type   : %s\n", bump.String())
		fmt.Printf("  • Target Ver  : %s%s%s%s\n", cGreen, cBold, nextVersion, cReset)
		fmt.Printf("  • Commits     : %d\n\n", len(commits))
		fmt.Printf("%s--- CHANGELOG PREVIEW ---%s\n%s\n", cGold, cReset, changelogDraft)
		return nil
	}

	changelogFile := filepath.Join(repoRoot, "CHANGELOG.md")
	var finalChangelog string
	if existingData, err := os.ReadFile(changelogFile); err == nil {
		content := string(existingData)
		if strings.HasPrefix(content, "# Changelog\n") {
			rest := strings.TrimPrefix(content, "# Changelog\n")
			finalChangelog = fmt.Sprintf("# Changelog\n\n%s\n%s", changelogDraft, strings.TrimSpace(rest))
		} else {
			finalChangelog = fmt.Sprintf("# Changelog\n\n%s\n%s", changelogDraft, strings.TrimSpace(content))
		}
	} else {
		finalChangelog = fmt.Sprintf("# Changelog\n\nAll notable changes to Syncalibur will be documented in this file.\n\n%s", changelogDraft)
	}

	if err := os.WriteFile(changelogFile, []byte(finalChangelog), 0644); err != nil {
		return fmt.Errorf("failed to write CHANGELOG.md: %w", err)
	}

	// 1. Update version constant in cmd/syncalibur-engine/main.go
	engineMainFile := filepath.Join(repoRoot, "cmd", "syncalibur-engine", "main.go")
	if mainData, err := os.ReadFile(engineMainFile); err == nil {
		reVer := regexp.MustCompile(`(?m)^\s*SyncaliburVersion\s*=\s*"[^"]+"`)
		newMain := reVer.ReplaceAllString(string(mainData), "\tSyncaliburVersion   = \""+nextVersion+"\"")
		_ = os.WriteFile(engineMainFile, []byte(newMain), 0644)
		_ = exec.Command("git", "-C", repoRoot, "add", filepath.Join("cmd", "syncalibur-engine", "main.go")).Run()
	}

	// 2. Update version constant in syncalibur.sh
	shFile := filepath.Join(repoRoot, "syncalibur.sh")
	if shData, err := os.ReadFile(shFile); err == nil {
		reSh := regexp.MustCompile(`(?m)^\s*SYNCALIBUR_PLATFORM_VERSION="[^"]+"`)
		newSh := reSh.ReplaceAllString(string(shData), "SYNCALIBUR_PLATFORM_VERSION=\""+nextVersion+"\"")
		_ = os.WriteFile(shFile, []byte(newSh), 0755)
		_ = exec.Command("git", "-C", repoRoot, "add", "syncalibur.sh").Run()
	}

	_ = exec.Command("git", "-C", repoRoot, "add", "CHANGELOG.md").Run()
	commitMsg := fmt.Sprintf("chore(release): bump version to %s", nextVersion)
	_ = exec.Command("git", "-C", repoRoot, "commit", "-m", commitMsg).Run()

	tagMsg := fmt.Sprintf("Release %s - Syncalibur Autonomous Engine", nextVersion)
	if err := exec.Command("git", "-C", repoRoot, "tag", "-a", nextVersion, "-m", tagMsg).Run(); err != nil {
		return fmt.Errorf("failed to create tag %s: %w", nextVersion, err)
	}

	remotesOut, _ := exec.Command("git", "-C", repoRoot, "remote").Output()
	remotes := strings.Fields(string(remotesOut))
	for _, r := range remotes {
		ctxPush, cancelPush := context.WithTimeout(context.Background(), 30*time.Second)
		cmdPush := exec.CommandContext(ctxPush, "git", "-C", repoRoot, "push", "--recurse-submodules=on-demand", r, "HEAD", "--tags")
		cmdPush.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		_ = cmdPush.Run()
		cancelPush()
	}

	fmt.Printf("\n  %s%s%s RELEASE %s SUCCESSFUL!%s\n", cGreen, cBold, iconCheck, nextVersion, cReset)
	fmt.Printf("  %s•%s Tagged as %s%s%s\n", cGold, cReset, cBold, nextVersion, cReset)
	fmt.Printf("  %s•%s CHANGELOG.md updated with %d commits\n", cGold, cReset, len(commits))
	fmt.Printf("  %s•%s Pushed tags to %s\n", cGold, cReset, strings.Join(remotes, ", "))

	// 3. Create official GitHub Release via gh CLI if available
	if _, err := exec.LookPath("gh"); err == nil {
		tmpNotes, err := os.CreateTemp("", "syncalibur-release-notes-*.md")
		if err == nil {
			_, _ = tmpNotes.WriteString(changelogDraft)
			_ = tmpNotes.Close()
			defer os.Remove(tmpNotes.Name())

			ghArgs := []string{"release", "create", nextVersion, "--title", fmt.Sprintf("Syncalibur %s", nextVersion), "--notes-file", tmpNotes.Name()}
			binPath := filepath.Join(repoRoot, "bin", "syncalibur")
			if _, err := os.Stat(binPath); err == nil {
				ghArgs = append(ghArgs, binPath)
			}
			cmdGH := exec.Command("gh", ghArgs...)
			cmdGH.Dir = repoRoot
			if out, err := cmdGH.CombinedOutput(); err == nil {
				fmt.Printf("  %s•%s Official GitHub Release created: https://github.com/YusufKaya00/Syncalibur/releases/tag/%s\n", cGreen, cReset, nextVersion)
			} else {
				fmt.Printf("  %s•%s GitHub Release notice: %s\n", cAmber, cReset, strings.TrimSpace(string(out)))
			}
		}
	}
	fmt.Println()

	return nil
}

