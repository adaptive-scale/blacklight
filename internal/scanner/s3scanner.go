package scanner

import (
	"blacklight/internal/model"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Scanner struct {
	patterns []model.Configuration
	client   *s3.Client
}

func NewS3Scanner() (*S3Scanner, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	client := s3.NewFromConfig(cfg)
	return &S3Scanner{
		client: client,
	}, nil
}

func (s *S3Scanner) AddPattern(patterns ...model.Configuration) {
	s.patterns = append(s.patterns, patterns...)
}

// parseBucketURL parses s3://bucket/prefix into bucket and prefix
func (s *S3Scanner) parseBucketURL(s3URL string) (string, string, error) {
	if !strings.HasPrefix(s3URL, "s3://") {
		return "", "", fmt.Errorf("invalid S3 URL format. Must start with s3://")
	}

	path := strings.TrimPrefix(s3URL, "s3://")
	parts := strings.SplitN(path, "/", 2)
	
	bucket := parts[0]
	prefix := ""
	if len(parts) > 1 {
		prefix = parts[1]
	}

	return bucket, prefix, nil
}

func (s *S3Scanner) ScanBucket(s3URL string) ([]model.Violation, error) {
	bucket, prefix, err := s.parseBucketURL(s3URL)
	if err != nil {
		return nil, err
	}

	var violations []model.Violation

	// List all objects in the bucket with the given prefix
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %v", err)
		}

		for _, obj := range page.Contents {
			// Skip directories (objects ending with /)
			if strings.HasSuffix(*obj.Key, "/") {
				continue
			}

			// Get object content
			result, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
				Bucket: &bucket,
				Key:    obj.Key,
			})
			if err != nil {
				fmt.Printf("Error getting object %s: %v\n", *obj.Key, err)
				continue
			}

			// Read the object content
			content, err := io.ReadAll(result.Body)
			result.Body.Close()
			if err != nil {
				fmt.Printf("Error reading object %s: %v\n", *obj.Key, err)
				continue
			}

			// Split content into lines for better violation reporting
			lines := strings.Split(string(content), "\n")
			for lineNum, line := range lines {
				// Check each pattern against the line
				for _, pattern := range s.patterns {
					if pattern.RegexVal.MatchString(line) {
						violations = append(violations, model.Violation{
							RuleID:     pattern.Name,
							Level:      1,
							Message:    fmt.Sprintf("Found potential %s in S3 object", pattern.Name),
							Line:       line,
							LineNumber: lineNum + 1,
							FilePath:   fmt.Sprintf("s3://%s/%s", bucket, *obj.Key),
						})
					}
				}
			}
		}
	}

	return violations, nil
} 