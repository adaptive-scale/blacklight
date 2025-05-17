package scanner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileFormat represents supported file formats
type FileFormat string

const (
	FormatText FileFormat = "text"
	FormatJSON FileFormat = "json"
	FormatYAML FileFormat = "yaml"
	FormatXML  FileFormat = "xml"
	FormatINI  FileFormat = "ini"
	FormatEnv  FileFormat = "env"
)

// detectFileFormat determines the file format based on extension and content
func detectFileFormat(fileName string, content []byte) FileFormat {
	ext := strings.ToLower(filepath.Ext(fileName))
	
	switch ext {
	case ".json":
		return FormatJSON
	case ".yaml", ".yml":
		return FormatYAML
	case ".xml":
		return FormatXML
	case ".ini", ".conf", ".config":
		return FormatINI
	case ".env":
		return FormatEnv
	default:
		// Try to detect JSON/YAML from content
		if isJSON(content) {
			return FormatJSON
		}
		if isYAML(content) {
			return FormatYAML
		}
		return FormatText
	}
}

// isJSON checks if content is valid JSON
func isJSON(content []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(content, &js) == nil
}

// isYAML checks if content is valid YAML
func isYAML(content []byte) bool {
	var y interface{}
	return yaml.Unmarshal(content, &y) == nil
}

// extractTextFromFile extracts scannable text based on file format
func extractTextFromFile(fileName string, reader io.Reader) ([]string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading file content: %v", err)
	}

	format := detectFileFormat(fileName, content)
	
	switch format {
	case FormatJSON:
		return extractFromJSON(content)
	case FormatYAML:
		return extractFromYAML(content)
	case FormatXML:
		return extractFromXML(content)
	case FormatINI:
		return extractFromINI(content)
	case FormatEnv:
		return extractFromEnv(content)
	default:
		return extractFromText(content)
	}
}

// extractFromJSON extracts values from JSON content
func extractFromJSON(content []byte) ([]string, error) {
	var result []string
	var data interface{}
	
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	extractValues(&result, data)
	return result, nil
}

// extractFromYAML extracts values from YAML content
func extractFromYAML(content []byte) ([]string, error) {
	var result []string
	var data interface{}
	
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	extractValues(&result, data)
	return result, nil
}

// extractFromXML extracts text content from XML
func extractFromXML(content []byte) ([]string, error) {
	// Simple XML parsing - extract text between tags and attribute values
	var result []string
	text := string(content)
	
	// Extract text between tags
	inTag := false
	var buffer strings.Builder
	
	for _, char := range text {
		if char == '<' {
			if buffer.Len() > 0 {
				result = append(result, strings.TrimSpace(buffer.String()))
				buffer.Reset()
			}
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			buffer.WriteRune(char)
		}
	}

	// Extract attribute values
	parts := strings.Split(text, "\"")
	for i := 1; i < len(parts); i += 2 {
		if len(parts[i]) > 0 {
			result = append(result, parts[i])
		}
	}

	return result, nil
}

// extractFromINI extracts values from INI/config content
func extractFromINI(content []byte) ([]string, error) {
	var result []string
	scanner := bufio.NewScanner(bytes.NewReader(content))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || line == "" {
			continue
		}

		// Extract value part of key=value
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			result = append(result, strings.TrimSpace(parts[1]))
		}
	}

	return result, scanner.Err()
}

// extractFromEnv extracts values from .env file content
func extractFromEnv(content []byte) ([]string, error) {
	var result []string
	scanner := bufio.NewScanner(bytes.NewReader(content))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Extract value part of key=value
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			// Handle quoted values
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
			result = append(result, value)
		}
	}

	return result, scanner.Err()
}

// extractFromText extracts lines from plain text content
func extractFromText(content []byte) ([]string, error) {
	var result []string
	scanner := bufio.NewScanner(bytes.NewReader(content))
	
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return result, scanner.Err()
}

// extractValues recursively extracts values from nested data structures
func extractValues(result *[]string, data interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			extractValues(result, value)
		}
	case []interface{}:
		for _, item := range v {
			extractValues(result, item)
		}
	case string:
		if strings.TrimSpace(v) != "" {
			*result = append(*result, v)
		}
	case float64:
		*result = append(*result, fmt.Sprintf("%v", v))
	case bool:
		*result = append(*result, fmt.Sprintf("%v", v))
	case nil:
		// Skip nil values
	}
} 