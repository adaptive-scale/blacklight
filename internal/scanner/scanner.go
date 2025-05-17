package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adaptive-scale/blacklight/internal/model"

	"gopkg.in/yaml.v3"
)

type SecretScanner struct {
	patterns    []model.Configuration
	ignoreDirs  []string
	verbose     bool
}

func NewSecretScanner() *SecretScanner {
	return &SecretScanner{
		ignoreDirs: []string{".git", "node_modules", "vendor"},
	}
}

func (s *SecretScanner) loadRules() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %v", err)
	}

	configDir := filepath.Join(home, ".blacklight")
	configFile := filepath.Join(configDir, "config.yaml")

	// Load user-defined rules
	if data, err := os.ReadFile(configFile); err == nil {
		var configs []model.Configuration
		if err := yaml.Unmarshal(data, &configs); err != nil {
			return fmt.Errorf("error parsing config file: %v", err)
		}
		// Compile regex patterns
		for i := range configs {
			if pattern, err := regexp.Compile(configs[i].Regex); err == nil {
				configs[i].CompiledRegex = pattern
			} else {
				fmt.Printf("Warning: Invalid regex pattern in rule %s: %v\n", configs[i].Name, err)
			}
		}
		s.patterns = append(s.patterns, configs...)
	}

	// Load built-in rules if no user rules found
	if len(s.patterns) == 0 {
		// Compile regex patterns for built-in rules
		for i := range Regex {
			if pattern, err := regexp.Compile(Regex[i].Regex); err == nil {
				Regex[i].CompiledRegex = pattern
			} else {
				fmt.Printf("Warning: Invalid regex pattern in built-in rule %s: %v\n", Regex[i].Name, err)
			}
		}
		s.patterns = append(s.patterns, Regex...)
	}

	return nil
}

func (s *SecretScanner) SetVerbose(verbose string) {
	s.verbose = verbose != ""
}

func (s *SecretScanner) AddIgnoreDir(dir string) {
	s.ignoreDirs = append(s.ignoreDirs, dir)
}

func (s *SecretScanner) AddPattern(patterns ...model.Configuration) {
	s.patterns = append(s.patterns, patterns...)
}

func (s *SecretScanner) shouldIgnore(path string) bool {
	for _, dir := range s.ignoreDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}
	return false
}

func (s *SecretScanner) StartScan(path string) []model.Violation {
	var violations []model.Violation

	// Check if path is a file
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error accessing path %s: %v\n", path, err)
		return violations
	}

	if !info.IsDir() {
		// Single file scan
		fileViolations := s.scanFile(path)
		violations = append(violations, fileViolations...)
		return violations
	}

	// Directory scan
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", filePath, err)
			return nil
		}

		if info.IsDir() {
			if s.shouldIgnore(filePath) {
				return filepath.SkipDir
			}
			return nil
		}

		if s.shouldIgnore(filePath) {
			return nil
		}

		fileViolations := s.scanFile(filePath)
		violations = append(violations, fileViolations...)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory %s: %v\n", path, err)
	}

	return violations
}

func (s *SecretScanner) scanFile(filePath string) []model.Violation {
	var violations []model.Violation

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filePath, err)
		return violations
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, pattern := range s.patterns {
			if pattern.CompiledRegex != nil && pattern.CompiledRegex.MatchString(line) {
				// Get context around the match
				match := pattern.CompiledRegex.FindString(line)
				context := getMatchContext(line, match)

				violations = append(violations, model.Violation{
					Rule:     pattern,
					Match:    match,
					Location: filePath,
					LineNum:  lineNum,
					Context:  context,
				})

				if s.verbose {
					fmt.Printf("Found potential %s in %s:%d\n", pattern.Name, filePath, lineNum)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning file %s: %v\n", filePath, err)
	}

	return violations
}

// getMatchContext returns a snippet of text around the matched string
func getMatchContext(line, match string) string {
	const contextLength = 50 // characters before and after the match

	start := strings.Index(line, match)
	if start == -1 {
		return line
	}

	// Calculate context boundaries
	contextStart := start - contextLength
	if contextStart < 0 {
		contextStart = 0
	}
	contextEnd := start + len(match) + contextLength
	if contextEnd > len(line) {
		contextEnd = len(line)
	}

	// Add ellipsis if context is truncated
	prefix := ""
	if contextStart > 0 {
		prefix = "..."
	}
	suffix := ""
	if contextEnd < len(line) {
		suffix = "..."
	}

	return prefix + line[contextStart:contextEnd] + suffix
}

// isBinary checks if a file is likely binary
func isBinary(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	buf := make([]byte, 8000)
	n, err := file.Read(buf)
	if err != nil {
		return true
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}
