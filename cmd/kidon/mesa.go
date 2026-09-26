package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/uddeshya-world/kidon-security/internal/mesa"
)

// mesaCmd consumes MESA formats: it never defines them.
var mesaCmd = &cobra.Command{
	Use:   "mesa",
	Short: "MESA integration: scan a committed topology, replay witness paths in the lab",
}

var mesaScanCmd = &cobra.Command{
	Use:   "scan <topology.json>",
	Short: "Run mesa-ibi-scan (lattice mode) and print INV01 and NEAR_MISS findings",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bin, _ := cmd.Flags().GetString("scanner")
		r, err := mesa.RunScanner(bin, args[0])
		if err != nil {
			return err
		}
		fmt.Print(mesa.Summary(r))
		return nil
	},
}

var mesaReplayCmd = &cobra.Command{
	Use:   "replay <scanner-report.json> <agent> <dimension>",
	Short: "Kidon Strike: plant a canary on the witness path in the lab and check whether it reaches egress (SC-12)",
	Long: `Lab only. The lab file maps vertex ids to lab URLs and lists the only hosts Strike may contact:
{"surfaces": {"svc:cache": "http://lab-cache.internal/put"}, "observer": "http://lab-egress.internal/seen",
 "allowed_hosts": ["lab-cache.internal", "lab-egress.internal"], "wait_seconds": 30}`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		labPath, _ := cmd.Flags().GetString("lab")
		if labPath == "" {
			return fmt.Errorf("--lab is required: Strike only replays against an allowlisted lab")
		}
		raw, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		r, err := mesa.ParseReport(raw)
		if err != nil {
			return err
		}
		var cfg struct {
			Surfaces     map[string]string `json:"surfaces"`
			Observer     string            `json:"observer"`
			AllowedHosts []string          `json:"allowed_hosts"`
			WaitSeconds  int               `json:"wait_seconds"`
		}
		b, err := os.ReadFile(labPath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &cfg); err != nil {
			return err
		}
		lab := mesa.Lab{Surfaces: cfg.Surfaces, Observer: cfg.Observer, AllowedHosts: cfg.AllowedHosts,
			Wait: time.Duration(cfg.WaitSeconds) * time.Second}
		for _, a := range r.Agents {
			if a.AgentID == args[1] {
				out, _ := json.MarshalIndent(lab.ReplayWitness(a, args[2]), "", "  ")
				fmt.Println(string(out))
				return nil
			}
		}
		return fmt.Errorf("agent %s not in report", args[1])
	},
}

func init() {
	mesaScanCmd.Flags().String("scanner", "mesa-ibi-scan", "path to mesa-ibi-scan (v0.4+)")
	mesaReplayCmd.Flags().String("lab", "", "lab configuration JSON (required)")
	mesaCmd.AddCommand(mesaScanCmd, mesaReplayCmd)
	rootCmd.AddCommand(mesaCmd)
}
