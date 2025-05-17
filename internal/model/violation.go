package model

// Violation represents a detected secret or sensitive information
type Violation struct {
	Rule      Configuration // The rule that triggered this violation
	Match     string       // The actual text that matched the rule
	Location  string       // Where the violation was found (file path, URL, etc.)
	Context   string       // Surrounding context of the violation
	LineNum   int         // Line number in the file (if applicable)
	StartCol  int         // Starting column of the match (if applicable)
	EndCol    int         // Ending column of the match (if applicable)
}
