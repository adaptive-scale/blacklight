package model

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type SarifLog struct {
	Version string     `json:"version"`
	Runs    []SarifRun `json:"runs"`
}

type SarifRun struct {
	Tool    SarifTool     `json:"tool"`
	Results []SarifResult `json:"results"`
}

type SarifTool struct {
	Driver SarifDriver `json:"driver"`
}

type SarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri,omitempty"`
	Rules          []SarifRule `json:"rules,omitempty"`
}

type SarifRule struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	ShortDesc  SarifMessage `json:"shortDescription,omitempty"`
	FullDesc   SarifMessage `json:"fullDescription,omitempty"`
	HelpURI    string       `json:"helpUri,omitempty"`
	Properties RuleProps    `json:"properties,omitempty"`
}

type SarifMessage struct {
	Text string `json:"text"`
}

type RuleProps struct {
	Severity string   `json:"severity"`
	Tags     []string `json:"tags,omitempty"`
}

type SarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   SarifMessage    `json:"message"`
	Locations []SarifLocation `json:"locations,omitempty"`
}

type SarifLocation struct {
	PhysicalLocation SarifPhysicalLocation `json:"physicalLocation"`
}

type SarifPhysicalLocation struct {
	ArtifactLocation SarifArtifactLocation `json:"artifactLocation"`
	Region           SarifRegion           `json:"region"`
}

type SarifArtifactLocation struct {
	URI string `json:"uri"`
}

type SarifRegion struct {
	StartLine int `json:"startLine"`
}

func NewSarifLog() *SarifLog {
	return &SarifLog{
		Version: "2.1.0",
		Runs:    []SarifRun{},
	}
}

func GenerateSARIF(violations []Violation) {
	var results []SarifResult

	for _, v := range violations {
		results = append(results, SarifResult{
			RuleID: v.RuleID, // use a unique ID like "aws-access-key"
			Level:  severityToLevel(v.Level),
			Message: SarifMessage{
				Text: v.Message,
			},
			Locations: []SarifLocation{
				{
					PhysicalLocation: SarifPhysicalLocation{
						ArtifactLocation: SarifArtifactLocation{
							URI: v.FilePath,
						},
						Region: SarifRegion{
							StartLine: v.LineNumber,
						},
					},
				},
			},
		})
	}

	sarifLog := SarifLog{
		Version: "2.1.0",
		Runs: []SarifRun{
			{
				Tool: SarifTool{
					Driver: SarifDriver{
						Name: "GoSecretScanner",
					},
				},
				Results: results,
			},
		},
	}

	file, err := os.Create(time.Now().Format(time.RFC3339) + "_results.sarif")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(sarifLog); err != nil {
		log.Fatal(err)
	}

}

func severityToLevel(severity int) string {
	switch severity {
	case 3:
		return "error"
	case 2:
		return "warning"
	default:
		return "note"
	}
}
