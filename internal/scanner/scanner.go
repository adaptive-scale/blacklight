package scanner

import (
	"blacklight/internal/model"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SecretScanner struct {
	allPatterns     []model.Configuration
	ignoreDir       map[string]*struct{}
	ignoreSelective map[string]string
	verbose         bool
}

func NewSecretScanner() *SecretScanner {
	return &SecretScanner{
		allPatterns:     []model.Configuration{},
		ignoreDir:       map[string]*struct{}{},
		ignoreSelective: map[string]string{},
	}
}

func (s *SecretScanner) AddPattern(pattern ...model.Configuration) {
	s.allPatterns = append(s.allPatterns, pattern...)
}

func (s *SecretScanner) AddIgnoreDir(dir string) {
	s.ignoreDir[dir] = &struct{}{}
}

func (s *SecretScanner) AddIgnoreSelective(path, pattern string) {
	s.ignoreSelective[path] = pattern
}

func (s *SecretScanner) SetVerbose(verbose string) {
	if verbose == "true" {
		s.verbose = true
	} else {
		s.verbose = false
	}
}

func (s *SecretScanner) StartScan(path string) []model.Violation {
	return s.scanDir(path)
}

func (s *SecretScanner) scanFile(path string) []model.Violation {

	var vt []model.Violation
	file, err := os.Open(path)
	if err != nil {
		return vt
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		for _, re := range s.allPatterns {
			if re.RegexVal.MatchString(line) {
				msg := fmt.Sprintf("❌ Possible %s in %s:%d\n", re.Name, path, lineNum)
				fmt.Print(msg)
				vt = append(vt, model.Violation{
					RuleID:     re.Name,
					LineNumber: lineNum,
					FilePath:   path,
					Line:       line,
					Level:      re.Severity,
					Message:    msg,
				})
			}
		}
		lineNum++
	}

	return vt

}

func (s *SecretScanner) scanDir(root string) []model.Violation {

	var violation []model.Violation

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return nil
		}

		if s.ignoreDir[path] != nil {
			if s.verbose {
				fmt.Println("Ignoring directory/file:", path)
			}
			return nil
		}

		for s2, _ := range s.ignoreDir {
			if strings.HasPrefix(path, s2) {
				if s.verbose {
					fmt.Println("Ignoring directory/file:", path)
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}
		// Skip binaries or large files
		if info.Size() > 1*1024*1024 {
			return nil
		}
		// Basic binary detection
		if isBinary(path) {
			return nil
		}

		if s.verbose {
			fmt.Println("Scanning file:", path)
		}
		vt := s.scanFile(path)

		if len(vt) > 0 {
			violation = append(violation, vt...)
		}
		return nil
	})

	fmt.Println("ret")

	return violation
}

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
