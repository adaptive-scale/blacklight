package model

type Violation struct {
	RuleID     string `json:"ruleId"`
	Level      int    `json:"level"`
	Message    string `json:"message"`
	Line       string `json:"line"`
	LineNumber int    `json:"lineNum"`
	FilePath   string `json:"file"`
}
