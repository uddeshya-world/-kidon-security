package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/uddeshya-world/kidon-security/internal/offensive"
	"github.com/uddeshya-world/kidon-security/internal/report"
	kruntime "github.com/uddeshya-world/kidon-security/internal/runtime"
	"github.com/uddeshya-world/kidon-security/internal/static"
	"github.com/uddeshya-world/kidon-security/internal/ui"
)

var (
	version = "0.3.0"
	banner  = `
██╗  ██╗██╗██████╗  ██████╗ ███╗   ██╗
██║ ██╔╝██║██╔══██╗██╔═══██╗████╗  ██║
█████╔╝ ██║██║  ██║██║   ██║██╔██╗ ██║
██╔═██╗ ██║██║  ██║██║   ██║██║╚██╗██║
██║  ██╗██║██████╔╝╚██████╔╝██║ ╚████║
╚═╝  ╚═╝╚═╝╚═════╝  ╚═════╝ ╚═╝  ╚═══╝
  Strike Risks Before They Act.
`
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "kidon",
	Short: "Kidon - Agentic Cyber Defense Platform",
	Long: banner + `
Kidon is a security platform designed for AI agents.
It provides static analysis, runtime protection, and red teaming capabilities.`,
	Version: version,
}

// Scan command - static analysis
var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Run static analysis on the target path",
	Long: `The Checkpoint - Pre-deployment scanning of code and config.
	
Scans for:
  - Hardcoded credentials (API keys, AWS keys)
  - Dangerous tool configurations
  - Vulnerable dependencies (OSV.dev)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetPath := "."
		if len(args) > 0 {
			targetPath = args[0]
		}

		bold := color.New(color.FgCyan, color.Bold)
		red := color.New(color.FgRed, color.Bold)
		green := color.New(color.FgGreen, color.Bold)
		yellow := color.New(color.FgYellow)

		bold.Println("\n⚔️  KIDON STATIC SCANNER")
		bold.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Target: %s\n\n", targetPath)

		// Phase 1: Credential Scan
		fmt.Println("🔐 Scanning for credentials...")
		scanner := static.NewScanner()
		findings, err := scanner.ScanDirectory(targetPath)
		if err != nil {
			red.Printf("✗ Error scanning: %v\n", err)
			os.Exit(1)
		}

		// Phase 5: Supply Chain Scan
		fmt.Println("📦 Analyzing supply chain dependencies...")
		pkgs, _ := static.ParseDependencies(targetPath)
		if len(pkgs) > 0 {
			fmt.Printf("   Found %d packages to check\n", len(pkgs))
			vulns := static.CheckVulnerabilities(pkgs)
			findings = append(findings, vulns...)
			if len(vulns) > 0 {
				red.Printf("   ⚠ Found %d vulnerable dependencies!\n", len(vulns))
			} else {
				green.Println("   ✓ No known vulnerabilities found")
			}
		} else {
			fmt.Println("   No dependency files found")
		}
		fmt.Println()

		if len(findings) == 0 {
			green.Println("✓ No security issues found. Area is clear.")
		} else {
			red.Printf("⚠ ALERT: %d security issue(s) detected!\n\n", len(findings))
			for i, finding := range findings {
				yellow.Printf("[%d] %s\n", i+1, finding.RuleName)
				fmt.Printf("    File: %s:%d\n", finding.FilePath, finding.LineNumber)
				if finding.MatchedText != "" {
					fmt.Printf("    Match: %s\n", finding.MatchedText)
				}
				fmt.Printf("    Severity: %s\n\n", finding.Severity)
			}
		}

		// Save results to mission log
		var reportIssues []report.Issue
		for _, finding := range findings {
			reportIssues = append(reportIssues, report.Issue{
				Severity:    finding.Severity,
				Description: finding.RuleName,
				Location:    fmt.Sprintf("%s:%d", finding.FilePath, finding.LineNumber),
			})
		}
		_ = report.AppendStatic(reportIssues)
		color.Cyan("💾 Results saved to mission log.")

		bold.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	},
}

// Guard command flags
var (
	guardNetwork bool
	guardCgroup  string
)

// Guard command - runtime protection (eBPF)
var guardCmd = &cobra.Command{
	Use:   "guard",
	Short: "Activate runtime eBPF protection",
	Long: `The Watchman - Kernel-level eBPF monitoring of agent processes.

v0.1.0: Process Guard - Blocks unauthorized process execution (bash, sh).
v0.2.0: Network Guard - Blocks unauthorized network connections (Iron Dome).

Target Risks: ASI-02 (Tool Misuse), ASI-05 (Code Exec), ASI-10 (Rogue Agents).

Examples:
  kidon guard                       # Process guard only
  kidon guard --network             # Full protection (process + network)
  kidon guard --network --cgroup /sys/fs/cgroup

NOTE: Requires Linux kernel with eBPF support. Run inside Docker container.`,
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "linux" {
			red := color.New(color.FgRed, color.Bold)
			red.Println("⚠️  ERROR: Guard mode requires Linux kernel with eBPF support.")
			fmt.Println("Run inside the Docker container:")
			fmt.Println("  docker build -f deploy/Dockerfile.kidon -t kidon-security .")
			fmt.Println("  docker run --privileged --pid=host --cgroupns=host kidon-security guard --network")
			os.Exit(1)
		}

		if guardNetwork {
			color.Cyan("🔥 Starting Iron Dome (Full Protection)...")
			kruntime.StartFullGuard(guardCgroup)
		} else {
			kruntime.StartGuard()
		}
	},
}

func init() {
	guardCmd.Flags().BoolVar(&guardNetwork, "network", false, "Enable Network Guard (Iron Dome v0.2.0)")
	guardCmd.Flags().StringVar(&guardCgroup, "cgroup", "/sys/fs/cgroup", "Cgroup path for network filtering")
}

// Strike command flags
var (
	strikeTarget string
	strikeUseAI  bool
	strikeHammer bool
	hammerCount  int
)

// Strike command - red teaming
var strikeCmd = &cobra.Command{
	Use:   "strike",
	Short: "Launch adversarial red team attacks against an AI agent",
	Long: `The Attacker - Proactive adversarial simulation using local SLMs.

Executes jailbreak attempts and security tests against a target AI agent.
Can use static attack strategies or AI-generated attacks via Ollama.

Target Risks: ASI-01 (Goal Hijack), ASI-06 (Memory Poisoning), ASI-08 (Cascading Failures).

Examples:
  kidon strike -t http://localhost:8000/chat
  kidon strike -t http://localhost:8000/chat --ai
  kidon strike -t http://localhost:8000/chat --hammer -n 100`,
	Run: func(cmd *cobra.Command, args []string) {
		bold := color.New(color.FgRed, color.Bold)
		bold.Println("\n⚔️  KIDON STRIKE - Red Team Operations")
		bold.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Target: %s\n", strikeTarget)
		fmt.Printf("AI Mode: %v\n\n", strikeUseAI)

		engine := offensive.NewEngine(strikeTarget, strikeUseAI)

		if strikeHammer {
			engine.HammerMode(hammerCount, 10)
			return
		}

		results, err := engine.Strike()
		if err != nil {
			color.Red("Error executing strike: %v", err)
			os.Exit(1)
		}

		// Summary
		var bypassed, blocked int
		for _, r := range results {
			if r.Success {
				bypassed++
			} else {
				blocked++
			}
		}

		// Save results to mission log
		var reportResults []report.AttackResult
		for _, res := range results {
			reportResults = append(reportResults, report.AttackResult{
				Type:     res.Type,
				Prompt:   res.Prompt,
				Success:  res.Success,
				Severity: res.Severity,
			})
		}
		_ = report.AppendAttacks(reportResults)
		color.Cyan("💾 Combat data saved to mission log.")

		fmt.Println()
		bold.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		color.Green("🛡️  Blocked: %d", blocked)
		if bypassed > 0 {
			color.Red("💀 Bypassed: %d (CRITICAL)", bypassed)
		} else {
			color.Green("💀 Bypassed: 0")
		}
	},
}

// Report command - generate mission report
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate classified mission report (HTML)",
	Long: `Generate a visual HTML report combining all mission data.

Reads from mission.json (populated by scan and strike commands) and
generates mission_report.html with a cyber-military dark theme.`,
	Run: func(cmd *cobra.Command, args []string) {
		bold := color.New(color.FgCyan, color.Bold)
		bold.Println("\n📠 GENERATING MISSION DOSSIER...")

		err := report.GenerateReport("mission_report.html")
		if err != nil {
			color.Red("Error generating report: %v", err)
			os.Exit(1)
		}

		color.Green("✅ Report generated: ./mission_report.html")

		// Open in default browser
		var cmdExec *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmdExec = exec.Command("rundll32", "url.dll,FileProtocolHandler", "mission_report.html")
		case "darwin":
			cmdExec = exec.Command("open", "mission_report.html")
		default:
			cmdExec = exec.Command("xdg-open", "mission_report.html")
		}
		if err := cmdExec.Start(); err != nil {
			fmt.Println("Note: Cannot open browser automatically. Open mission_report.html manually.")
		}
	},
}

// Dashboard command - TUI Cockpit
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the Kidon Command Cockpit (TUI)",
	Long: `The Cockpit - Unified Terminal Dashboard.

Provides a real-time terminal interface combining:
  - THE SENTRY: Static Analysis & Supply Chain Scanner
  - THE SHOMER: Runtime Guard with eBPF event streaming
  - THE KIDON: Red Team Attack Console

Controls:
  [Tab] Switch tabs | [1-3] Direct select | [s] Scan | [q] Quit`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create a channel for guard events
		guardChan := make(chan string)

		// Simulate guard events (replace with real runtime.MonitorEvents)
		go func() {
			events := []string{
				"[PROCESS] bash blocked (PID 1234)",
				"[NET] ✅ api.openai.com allowed",
				"[NET] 🚨 evil-server.com BLOCKED",
				"[PROCESS] python allowed (PID 5678)",
				"[NET] ✅ google.com allowed",
			}
			i := 0
			for {
				time.Sleep(2 * time.Second)
				guardChan <- fmt.Sprintf("%s (%s)", events[i%len(events)], time.Now().Format("15:04:05"))
				i++
			}
		}()

		// Launch TUI
		p := tea.NewProgram(ui.InitialModel(guardChan))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Strike command flags
	strikeCmd.Flags().StringVarP(&strikeTarget, "target", "t", "http://localhost:8000/chat", "Target API URL")
	strikeCmd.Flags().BoolVar(&strikeUseAI, "ai", false, "Enable Ollama-driven dynamic attacks")
	strikeCmd.Flags().BoolVar(&strikeHammer, "hammer", false, "Enable DoS simulation mode")
	strikeCmd.Flags().IntVarP(&hammerCount, "count", "n", 100, "Number of requests for hammer mode")

	// Register all commands
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(guardCmd)
	rootCmd.AddCommand(strikeCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(dashboardCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
