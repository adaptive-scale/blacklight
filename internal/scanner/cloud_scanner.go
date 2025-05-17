package scanner

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/adaptive-scale/blacklight/internal/model"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// CloudStorageType represents the type of cloud storage
type CloudStorageType string

const (
	GoogleDrive CloudStorageType = "gdrive"
	OneDrive    CloudStorageType = "onedrive"
	Dropbox     CloudStorageType = "dropbox"
	Box         CloudStorageType = "box"
)

// CloudConfig holds configuration for cloud storage scanning
type CloudConfig struct {
	Type          CloudStorageType
	Token         string   // OAuth token or API key
	FolderPaths   []string // List of folder paths to scan
	DaysToScan    int      // Number of days of history to scan
	IncludeShared bool     // Whether to scan shared files
	MaxFileSize   int64    // Maximum file size to scan in bytes (default 10MB)
}

// CloudScanner represents a scanner for cloud storage services
type CloudScanner struct {
	config   *CloudConfig
	patterns []model.Configuration
}

// NewCloudScanner creates a new cloud storage scanner instance
func NewCloudScanner(config *CloudConfig) (*CloudScanner, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	if config.MaxFileSize == 0 {
		config.MaxFileSize = 10 * 1024 * 1024 // 10MB default
	}

	return &CloudScanner{
		config: config,
	}, nil
}

func (s *CloudScanner) AddPattern(patterns ...model.Configuration) {
	s.patterns = append(s.patterns, patterns...)
}

func (s *CloudScanner) parseCloudURL(cloudURL string) (CloudStorageType, string, error) {
	u, err := url.Parse(cloudURL)
	if err != nil {
		return "", "", err
	}

	switch u.Scheme {
	case "gdrive":
		return GoogleDrive, strings.TrimPrefix(u.Path, "/"), nil
	case "onedrive":
		return OneDrive, strings.TrimPrefix(u.Path, "/"), nil
	case "dropbox":
		return Dropbox, strings.TrimPrefix(u.Path, "/"), nil
	case "box":
		return Box, strings.TrimPrefix(u.Path, "/"), nil
	default:
		return "", "", fmt.Errorf("unsupported cloud storage type: %s", u.Scheme)
	}
}

func (s *CloudScanner) ScanStorage(ctx context.Context, cloudURL string) ([]model.Violation, error) {
	storageType, path, err := s.parseCloudURL(cloudURL)
	if err != nil {
		return nil, err
	}

	switch storageType {
	case GoogleDrive:
		return s.scanGoogleDrive(ctx, path)
	case OneDrive:
		return s.scanOneDrive(ctx, path)
	case Dropbox:
		return s.scanDropbox(ctx, path)
	case Box:
		return s.scanBox(ctx, path)
	default:
		return nil, fmt.Errorf("unsupported cloud storage type: %s", storageType)
	}
}

func (s *CloudScanner) scanGoogleDrive(ctx context.Context, path string) ([]model.Violation, error) {
	config, err := google.ConfigFromJSON([]byte(s.config.Token), drive.DriveReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file: %v", err)
	}

	client := config.Client(ctx, &oauth2.Token{})
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Drive client: %v", err)
	}

	var violations []model.Violation
	query := fmt.Sprintf("'%s' in parents and mimeType != 'application/vnd.google-apps.folder'", path)
	
	err = srv.Files.List().Q(query).Pages(ctx, func(list *drive.FileList) error {
		for _, file := range list.Files {
			if file.Size > s.config.MaxFileSize {
				continue
			}

			resp, err := srv.Files.Get(file.Id).Download()
			if err != nil {
				fmt.Printf("Error downloading file %s: %v\n", file.Name, err)
				continue
			}
			defer resp.Body.Close()

			violations = append(violations, s.scanContent(resp.Body, file.Name, fmt.Sprintf("gdrive://%s/%s", path, file.Id))...)
		}
		return nil
	})

	return violations, err
}

func (s *CloudScanner) scanDropbox(ctx context.Context, path string) ([]model.Violation, error) {
	config := dropbox.Config{
		Token: s.config.Token,
	}
	client := files.New(config)

	arg := files.NewListFolderArg(path)
	result, err := client.ListFolder(arg)
	if err != nil {
		return nil, fmt.Errorf("unable to list Dropbox folder: %v", err)
	}

	var violations []model.Violation
	for _, entry := range result.Entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			if uint64(f.Size) > uint64(s.config.MaxFileSize) {
				continue
			}

			arg := files.NewDownloadArg(f.PathLower)
			_, resp, err := client.Download(arg)
			if err != nil {
				fmt.Printf("Error downloading file %s: %v\n", f.Name, err)
				continue
			}
			defer resp.Close()

			violations = append(violations, s.scanContent(resp, f.Name, fmt.Sprintf("dropbox://%s", f.PathLower))...)
		}
	}

	return violations, nil
}

func (s *CloudScanner) scanOneDrive(ctx context.Context, path string) ([]model.Violation, error) {
	// OneDrive implementation using Microsoft Graph API
	// This is a placeholder - actual implementation would use the Microsoft Graph SDK
	return nil, fmt.Errorf("OneDrive scanning not yet implemented")
}

func (s *CloudScanner) scanBox(ctx context.Context, path string) ([]model.Violation, error) {
	// Box implementation using Box SDK
	// This is a placeholder - actual implementation would use the Box SDK
	return nil, fmt.Errorf("Box scanning not yet implemented")
}

func (s *CloudScanner) scanContent(reader io.Reader, fileName, location string) []model.Violation {
	var violations []model.Violation

	// Extract text content based on file format
	lines, err := extractTextFromFile(fileName, reader)
	if err != nil {
		fmt.Printf("Warning: Error processing file %s: %v\n", fileName, err)
		return violations
	}

	// Process each extracted line
	for lineNum, line := range lines {
		for _, pattern := range s.patterns {
			if pattern.CompiledRegex == nil {
				fmt.Printf("Warning: Skipping rule %s - regex not compiled\n", pattern.Name)
				continue
			}

			if pattern.CompiledRegex.MatchString(line) {
				// Create violation with context
				violation := model.Violation{
					Rule:     pattern,
					Match:    line,
					Location: location,
					LineNum:  lineNum + 1,
					Context:  fmt.Sprintf("File: %s - %s", fileName, line),
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// getCloudMatchContext returns a snippet of text around the matched string
func getCloudMatchContext(line, match string) string {
	const contextLength = 50 // characters before and after the match

	start := strings.Index(line, match)
	if start == -1 {
		return line
	}

	contextStart := start - contextLength
	if contextStart < 0 {
		contextStart = 0
	}
	contextEnd := start + len(match) + contextLength
	if contextEnd > len(line) {
		contextEnd = len(line)
	}

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