package model

import "regexp"

type Configuration struct {
	Name     string         `json:"name"`
	Regex    string         `json:"regex"`
	Severity int            `json:"severity,omitempty"`
	Type     string         `json:"type,omitempty"`
	RegexVal *regexp.Regexp `json:"-"`
}
