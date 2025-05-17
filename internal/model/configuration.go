package model

import "regexp"

// Configuration represents a rule for secret scanning
type Configuration struct {
	ID            string         `yaml:"id"`
	Name          string         `yaml:"name"`
	Description   string         `yaml:"description"`
	Regex         string         `yaml:"regex"`
	CompiledRegex *regexp.Regexp `yaml:"-"` // Compiled version of the regex pattern
	Severity      int            `yaml:"severity"`
	Type          string         `yaml:"type"`
	Disabled      bool           `yaml:"disabled"`
}

type Parser struct {
	Parser         string `yaml:"name"`
	ParserFunction string `yaml:"function"`
	Internal       bool   `yaml:"internal"`
	FileExt        string `yaml:"ext"`
}
