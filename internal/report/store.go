package report

import (
	"encoding/json"
	"os"
	"time"
)

const missionFile = "mission.json"

// AppendStatic saves static analysis issues to the mission log
func AppendStatic(issues []Issue) error {
	data := load()
	data.StaticIssues = issues
	data.Timestamp = time.Now()
	return save(data)
}

// AppendAttacks saves attack results to the mission log
func AppendAttacks(results []AttackResult) error {
	data := load()
	data.AttackResults = results
	data.Timestamp = time.Now()
	return save(data)
}

// AppendRuntimeAlert adds a runtime alert to the log
func AppendRuntimeAlert(alert string) error {
	data := load()
	data.RuntimeAlerts = append(data.RuntimeAlerts, alert)
	data.Timestamp = time.Now()
	return save(data)
}

// ClearMission resets the mission data
func ClearMission() error {
	return os.Remove(missionFile)
}

func load() MissionData {
	file, err := os.ReadFile(missionFile)
	if err != nil {
		return MissionData{
			// Default names the assessed agent (see configs/kidon_policy.yaml class_2).
			TitanClass: "ARMORED TITAN (Class 2)",
			Timestamp:  time.Now(),
		}
	}
	var data MissionData
	_ = json.Unmarshal(file, &data)
	if data.TitanClass == "" {
		data.TitanClass = "ARMORED TITAN (Class 2)"
	}
	return data
}

func save(data MissionData) error {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(missionFile, file, 0644)
}
