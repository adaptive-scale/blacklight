package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SARIF represents the root object of a SARIF log file
type SARIF struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []Run  `json:"runs"`
}

// Run represents a single run of an analysis tool
type Run struct {
	Tool        Tool         `json:"tool"`
	Results     []Result     `json:"results"`
	Invocations []Invocation `json:"invocations"`
}

// Tool represents the analysis tool that was run
type Tool struct {
	Driver Driver `json:"driver"`
}

// Driver represents the analysis tool driver
type Driver struct {
	Name           string  `json:"name"`
	Version        string  `json:"version"`
	InformationURI string  `json:"informationUri"`
	Rules          []Rule  `json:"rules"`
}

// Rule represents a rule that was evaluated during the analysis
type Rule struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	ShortDescription Message           `json:"shortDescription"`
	FullDescription  Message           `json:"fullDescription"`
	Help             Message           `json:"help"`
	Properties       RuleProperties    `json:"properties"`
}

// Message represents a message string or structured message
type Message struct {
	Text string `json:"text"`
}

// RuleProperties represents additional metadata about a rule
type RuleProperties struct {
	Tags    []string `json:"tags"`
	Level   string   `json:"level"`
	Type    string   `json:"type"`
	Enabled bool     `json:"enabled"`
}

// Result represents a single analysis result
type Result struct {
	RuleID    string      `json:"ruleId"`
	RuleIndex int         `json:"ruleIndex"`
	Level     string      `json:"level"`
	Message   Message     `json:"message"`
	Locations []Location  `json:"locations"`
}

// Location represents a location within an analyzed artifact
type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

// PhysicalLocation represents a physical location within an artifact
type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

// ArtifactLocation represents the location of an artifact
type ArtifactLocation struct {
	URI string `json:"uri"`
}

// Region represents a region within an artifact
type Region struct {
	StartLine   int    `json:"startLine"`
	StartColumn int    `json:"startColumn,omitempty"`
	EndLine     int    `json:"endLine,omitempty"`
	EndColumn   int    `json:"endColumn,omitempty"`
	Snippet     Snippet  `json:"snippet,omitempty"`
}

// Snippet represents a snippet of text from the artifact
type Snippet struct {
	Text string `json:"text"`
}

// Invocation represents a single invocation of an analysis tool
type Invocation struct {
	ExecutionSuccessful bool `json:"executionSuccessful"`
}

func NewSarifLog() *SARIF {
	return &SARIF{
		Version: "2.1.0",
		Runs:    []Run{},
	}
}

// GenerateSARIF generates a SARIF report from the violations
func GenerateSARIF(violations []Violation) {
	// Create SARIF report structure
	sarif := SARIF{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:           "Blacklight",
						Version:        "0.1.0",
						InformationURI: "https://github.com/adaptive-scale/blacklight",
					},
				},
				Results:     make([]Result, 0),
				Invocations: []Invocation{{ExecutionSuccessful: true}},
			},
		},
	}

	// Create a map to store unique rules
	ruleMap := make(map[string]Rule)
	ruleIndex := make(map[string]int)
	currentIndex := 0

	// Process violations
	for _, v := range violations {
		// Create or get rule
		if _, exists := ruleMap[v.Rule.ID]; !exists {
			ruleMap[v.Rule.ID] = Rule{
				ID:   v.Rule.ID,
				Name: v.Rule.Name,
				ShortDescription: Message{
					Text: v.Rule.Description,
				},
				Properties: RuleProperties{
					Tags:    []string{v.Rule.Type},
					Level:   getSeverityLevel(v.Rule.Severity),
					Type:    v.Rule.Type,
					Enabled: !v.Rule.Disabled,
				},
			}
			ruleIndex[v.Rule.ID] = currentIndex
			currentIndex++
		}

		// Create result
		result := Result{
			RuleID:    v.Rule.ID,
			RuleIndex: ruleIndex[v.Rule.ID],
			Level:     getSeverityLevel(v.Rule.Severity),
			Message: Message{
				Text: v.Context,
			},
			Locations: []Location{
				{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: ArtifactLocation{
							URI: toFileURI(v.Location),
						},
						Region: Region{
							StartLine: v.LineNum,
							StartColumn: v.StartCol,
							EndColumn: v.EndCol,
							Snippet: Snippet{
								Text: v.Match,
							},
						},
					},
				},
			},
		}

		sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)
	}

	// Add rules to driver
	rules := make([]Rule, 0, len(ruleMap))
	for _, rule := range ruleMap {
		rules = append(rules, rule)
	}
	sarif.Runs[0].Tool.Driver.Rules = rules

	// Write SARIF report to file
	output, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		fmt.Printf("Error generating SARIF report: %v\n", err)
		return
	}

	err = os.WriteFile("blacklight-results.sarif", output, 0644)
	if err != nil {
		fmt.Printf("Error writing SARIF report: %v\n", err)
		return
	}

	fmt.Println("✅ SARIF report generated: blacklight-results.sarif")
}

// getSeverityLevel converts numeric severity to SARIF level
func getSeverityLevel(severity int) string {
	switch severity {
	case 3:
		return "error"
	case 2:
		return "warning"
	case 1:
		return "note"
	default:
		return "warning"
	}
}

// toFileURI converts a file path to a URI
func toFileURI(path string) string {
	// Handle special URIs (e.g., s3://, slack://, table://)
	if strings.HasPrefix(path, "s3://") || 
	   strings.HasPrefix(path, "slack://") || 
	   strings.HasPrefix(path, "table://") {
		return path
	}

	// Convert file path to URI
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return "file://" + filepath.ToSlash(absPath)
}
