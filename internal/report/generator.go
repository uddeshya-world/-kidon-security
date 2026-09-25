package report

import (
	"encoding/json"
	"html/template"
	"os"
	"time"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>KIDON // MISSION REPORT</title>
    <style>
        :root { --bg: #0a0a0a; --panel: #111; --cyan: #00f3ff; --red: #ff003c; --green: #00ff88; --text: #eee; }
        body { background: var(--bg); color: var(--text); font-family: 'Courier New', monospace; padding: 20px; margin: 0; min-height: 100vh; }
        .header { border-bottom: 2px solid var(--cyan); padding-bottom: 10px; margin-bottom: 30px; display: flex; justify-content: space-between; align-items: center; }
        .logo { font-size: 2.5em; font-weight: bold; letter-spacing: 3px; }
        .logo span { color: var(--cyan); }
        .badge { background: var(--cyan); color: #000; padding: 4px 12px; font-weight: bold; font-size: 0.9em; }
        .badge-red { background: var(--red); }
        .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        .card { background: var(--panel); border: 1px solid #333; padding: 20px; border-radius: 4px; }
        .card h2 { color: var(--cyan); margin-top: 0; border-bottom: 1px solid #333; padding-bottom: 10px; font-size: 1.1em; }
        .vuln { border-left: 3px solid var(--red); background: #1a0505; padding: 10px; margin-bottom: 8px; font-size: 0.85em; }
        .success { border-left: 3px solid var(--green); background: #051a0a; padding: 10px; margin-bottom: 8px; font-size: 0.85em; }
        .stat-row { display: flex; gap: 20px; margin-bottom: 20px; }
        .stat-box { background: var(--panel); border: 1px solid #333; padding: 20px; flex: 1; text-align: center; }
        .stat { font-size: 2.5em; font-weight: bold; }
        .stat.red { color: var(--red); }
        .stat.green { color: var(--green); }
        .stat.cyan { color: var(--cyan); }
        .meta { opacity: 0.6; font-size: 0.8em; }
        .empty { opacity: 0.5; font-style: italic; }
        @media (max-width: 800px) { .grid { grid-template-columns: 1fr; } }
    </style>
</head>
<body>
    <div class="header">
        <div>
            <div class="logo">⚔️ <span>KIDON</span> // SECURITY REPORT</div>
            <p style="margin: 5px 0 0 0; opacity: 0.7;">AGENTIC DEFENSE PLATFORM</p>
        </div>
        <div style="text-align: right;">
            <p style="margin: 0;">DATE: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
            <p style="margin: 5px 0 0 0;">TITAN CLASS: <span class="badge">{{.TitanClass}}</span></p>
        </div>
    </div>

    <div class="stat-row">
        <div class="stat-box">
            <div class="stat cyan">{{len .StaticIssues}}</div>
            <div>Static Issues</div>
        </div>
        <div class="stat-box">
            <div class="stat {{if .AttackBreaches}}red{{else}}green{{end}}">{{.AttackBreaches}}</div>
            <div>Attack Breaches</div>
        </div>
        <div class="stat-box">
            <div class="stat green">{{.AttackBlocked}}</div>
            <div>Attacks Blocked</div>
        </div>
    </div>

    <div class="grid">
        <div class="card">
            <h2>📡 STATIC RECONNAISSANCE</h2>
            {{range .StaticIssues}}
            <div class="vuln">
                <span style="color:var(--red)">[{{.Severity}}]</span> {{.Description}}<br>
                <span class="meta">📍 {{.Location}}</span>
            </div>
            {{else}}
            <p class="empty">✓ No static vulnerabilities detected.</p>
            {{end}}
        </div>

        <div class="card">
            <h2>🎯 OFFENSIVE SIMULATION</h2>
            {{range .AttackResults}}
                {{if .Success}}
                <div class="vuln">
                    <strong style="color:var(--red)">[💀 BREACH]</strong> {{.Type}}<br>
                    <span class="meta">"{{.Prompt}}"</span>
                </div>
                {{else}}
                <div class="success">
                    <strong style="color:var(--green)">[🛡️ BLOCKED]</strong> {{.Type}}
                </div>
                {{end}}
            {{else}}
            <p class="empty">No offensive simulation data.</p>
            {{end}}
        </div>
    </div>

    {{if .RuntimeAlerts}}
    <div class="card" style="margin-top: 20px;">
        <h2>⚡ RUNTIME ALERTS</h2>
        {{range .RuntimeAlerts}}
        <div class="vuln">{{.}}</div>
        {{end}}
    </div>
    {{end}}

    <div style="margin-top: 30px; text-align: center; opacity: 0.4; font-size: 0.8em;">
        Kidon Security Platform
    </div>
</body>
</html>`

// ReportData extends MissionData with computed fields for the template
type ReportData struct {
	MissionData
	AttackBreaches int
	AttackBlocked  int
}

// GenerateReport reads mission.json and creates an HTML report
func GenerateReport(outputPath string) error {
	// 1. Read the JSON Black Box
	file, err := os.ReadFile("mission.json")
	var data MissionData
	if err != nil {
		// Return empty report if no file exists
		data = MissionData{Timestamp: time.Now(), TitanClass: "UNKNOWN"}
	} else {
		_ = json.Unmarshal(file, &data)
	}

	// 2. Compute stats
	reportData := ReportData{MissionData: data}
	for _, r := range data.AttackResults {
		if r.Success {
			reportData.AttackBreaches++
		} else {
			reportData.AttackBlocked++
		}
	}

	// 3. Render HTML
	return createHTML(outputPath, reportData)
}

func createHTML(filename string, data ReportData) error {
	t, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, data)
}
