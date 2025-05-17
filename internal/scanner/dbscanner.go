package scanner

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/adaptive-scale/blacklight/internal/model"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver
)

type DBScanner struct {
	patterns   []model.Configuration
	sampleSize int
}

func NewDBScanner(sampleSize int) *DBScanner {
	return &DBScanner{
		sampleSize: sampleSize,
	}
}

func (s *DBScanner) AddPattern(patterns ...model.Configuration) {
	s.patterns = append(s.patterns, patterns...)
}

func (s *DBScanner) parseDBURL(dbURL string) (string, string, error) {
	if strings.HasPrefix(dbURL, "postgres://") {
		return "postgres", dbURL, nil
	} else if strings.HasPrefix(dbURL, "mysql://") {
		// Convert MySQL URL to DSN format
		dbURL = strings.TrimPrefix(dbURL, "mysql://")
		return "mysql", dbURL, nil
	}
	return "", "", fmt.Errorf("unsupported database type in URL: %s", dbURL)
}

func (s *DBScanner) ScanDatabase(dbURL string) ([]model.Violation, error) {
	driverName, dataSourceName, err := s.parseDBURL(dbURL)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	var violations []model.Violation

	// Get all tables
	tables, err := s.getTables(db, driverName)
	if err != nil {
		return nil, err
	}

	// Scan each table
	for _, table := range tables {
		tableViolations, err := s.scanTable(db, table)
		if err != nil {
			fmt.Printf("Error scanning table %s: %v\n", table, err)
			continue
		}
		violations = append(violations, tableViolations...)
	}

	return violations, nil
}

func (s *DBScanner) getTables(db *sql.DB, driverName string) ([]string, error) {
	var query string
	switch driverName {
	case "postgres":
		query = `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
	case "mysql":
		query = `SHOW TABLES`
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driverName)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (s *DBScanner) scanTable(db *sql.DB, tableName string) ([]model.Violation, error) {
	// Get column information
	columns, err := db.Query(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = $1`, tableName)
	if err != nil {
		return nil, err
	}
	defer columns.Close()

	var violations []model.Violation
	var columnNames []string
	var columnTypes []string

	// Store column information
	for columns.Next() {
		var name, dataType string
		if err := columns.Scan(&name, &dataType); err != nil {
			return nil, err
		}
		columnNames = append(columnNames, name)
		columnTypes = append(columnTypes, dataType)
	}

	// PCI-sensitive column name patterns
	pciColumnPatterns := []string{
		"(?i)card[_-]?num",
		"(?i)cc[_-]?num",
		"(?i)credit[_-]?card",
		"(?i)card[_-]?holder",
		"(?i)cvv",
		"(?i)cvc",
		"(?i)ccv",
		"(?i)cid",
		"(?i)card[_-]?pin",
		"(?i)card[_-]?expiry",
		"(?i)card[_-]?exp",
		"(?i)merchant[_-]?id",
	}

	// For each column, check if it might contain PCI data based on name
	for i, colName := range columnNames {
		isPCISensitive := false
		for _, pattern := range pciColumnPatterns {
			if matched, _ := regexp.MatchString(pattern, colName); matched {
				isPCISensitive = true
				break
			}
		}

		// Adjust sample size based on column sensitivity
		sampleSize := s.sampleSize
		if isPCISensitive {
			// Double the sample size for PCI-sensitive columns
			sampleSize *= 2
		}

		// Build query with appropriate sampling
		query := fmt.Sprintf("SELECT DISTINCT %s FROM %s WHERE %s IS NOT NULL", 
			colName, tableName, colName)
		
		if sampleSize > 0 {
			switch {
			case strings.HasPrefix(strings.ToLower(columnTypes[i]), "char"),
				strings.HasPrefix(strings.ToLower(columnTypes[i]), "varchar"),
				strings.HasPrefix(strings.ToLower(columnTypes[i]), "text"):
				// For text columns, add length-based filtering to avoid scanning huge text fields
				query = fmt.Sprintf("%s AND length(%s) BETWEEN 8 AND 50", query, colName)
			}
			query = fmt.Sprintf("%s LIMIT %d", query, sampleSize)
		}

		rows, err := db.Query(query)
		if err != nil {
			fmt.Printf("Error querying column %s: %v\n", colName, err)
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var value sql.NullString
			if err := rows.Scan(&value); err != nil {
				fmt.Printf("Error scanning value from column %s: %v\n", colName, err)
				continue
			}

			if !value.Valid {
				continue
			}

			// Check each pattern against the value
			for _, pattern := range s.patterns {
				if pattern.CompiledRegex == nil {
					fmt.Printf("Warning: Skipping rule %s - regex not compiled\n", pattern.Name)
					continue
				}
				if pattern.CompiledRegex.MatchString(value.String) {
					context := fmt.Sprintf("Column '%s' contains sensitive data: %s", colName, value.String)
					violations = append(violations, model.Violation{
						Rule:     pattern,
						Match:    value.String,
						Location: fmt.Sprintf("table://%s/%s", tableName, colName),
						Context:  context,
					})
					
					// Break after first match for this value to avoid duplicate reports
					break
				}
			}
		}
	}

	return violations, nil
} 