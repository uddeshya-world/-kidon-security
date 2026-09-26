package mesa

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// AgentFinding is one agent row of `mesa-ibi-scan --format json` in lattice mode.
type AgentFinding struct {
	AgentID      string              `json:"agent_id"`
	Zone         string              `json:"zone"`
	Closure      map[string]int      `json:"closure"`
	Threshold    map[string]int      `json:"threshold"`
	Violation    bool                `json:"violation"`
	NearMiss     bool                `json:"near_miss"`
	Missing      []string            `json:"missing"`
	Witness      map[string][]string `json:"witness"`
	MinCut       []map[string]any    `json:"min_cut"`
	CutDimension *string             `json:"cut_dimension"`
}

// Report is the lattice-mode scanner JSON report.
type Report struct {
	Version        string         `json:"version"`
	Mode           string         `json:"mode"`
	Agents         []AgentFinding `json:"agents"`
	ViolationCount int            `json:"violation_count"`
	NearMissCount  int            `json:"near_miss_count"`
}

// ParseReport decodes scanner output.
func ParseReport(b []byte) (Report, error) {
	var r Report
	if err := json.Unmarshal(b, &r); err != nil {
		return r, err
	}
	if r.Mode != "lattice" {
		return r, fmt.Errorf("expected a lattice-mode report (scanner v0.4+), got mode %q", r.Mode)
	}
	return r, nil
}

// RunScanner calls mesa-ibi-scan on a committed topology. The scanner exits 1 on violation; that is not an error.
func RunScanner(bin, topology string) (Report, error) {
	out, err := exec.Command(bin, "--input", topology, "--lattice", "--format", "json").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
			return Report{}, fmt.Errorf("mesa-ibi-scan: %w", err)
		}
	}
	return ParseReport(out)
}

// Summary renders findings for the TUI: INV01 first, then NEAR_MISS, with witness and cut.
func Summary(r Report) string {
	rows := append([]AgentFinding(nil), r.Agents...)
	rank := func(a AgentFinding) int {
		switch {
		case a.Violation:
			return 0
		case a.NearMiss:
			return 1
		}
		return 2
	}
	sort.SliceStable(rows, func(i, j int) bool { return rank(rows[i]) < rank(rows[j]) })
	var b strings.Builder
	fmt.Fprintf(&b, "scanner %s: %d INV01, %d NEAR_MISS\n\n", r.Version, r.ViolationCount, r.NearMissCount)
	for _, a := range rows {
		status := "ok"
		if a.Violation {
			status = "INV01"
		} else if a.NearMiss {
			status = "NEAR_MISS " + strings.Join(a.Missing, ",")
		}
		fmt.Fprintf(&b, "%s [%s] P%d U%d E%d => %s\n", a.AgentID, a.Zone, a.Closure["P"], a.Closure["U"], a.Closure["E"], status)
		if a.Violation {
			for _, d := range []string{"P", "U", "E"} {
				fmt.Fprintf(&b, "  witness %s: %s\n", d, strings.Join(a.Witness[d], " -> "))
			}
			if len(a.MinCut) > 0 {
				var cuts []string
				for _, c := range a.MinCut {
					cuts = append(cuts, fmt.Sprintf("%v -> %v (%v)", c["src"], c["dst"], c["flow_type"]))
				}
				fmt.Fprintf(&b, "  minimal cut: %s\n", strings.Join(cuts, "; "))
			}
		}
	}
	return b.String()
}
