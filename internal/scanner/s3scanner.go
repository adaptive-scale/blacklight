package scanner

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/adaptive-scale/blacklight/internal/model"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Scanner struct {
	client   *s3.Client
	patterns []model.Configuration
}

func NewS3Scanner() (*S3Scanner, error) {
	// Load AWS configuration from environment variables or credentials file
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %v", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	return &S3Scanner{
		client: client,
	}, nil
}

func (s *S3Scanner) AddPattern(patterns ...model.Configuration) {
	s.patterns = append(s.patterns, patterns...)
}

func (s *S3Scanner) parseS3URL(s3URL string) (string, string, error) {
	u, err := url.Parse(s3URL)
	if err != nil {
		return "", "", err
	}

	if u.Scheme != "s3" {
		return "", "", fmt.Errorf("invalid S3 URL scheme: %s", u.Scheme)
	}

	bucket := u.Host
	prefix := strings.TrimPrefix(u.Path, "/")

	return bucket, prefix, nil
}

func (s *S3Scanner) ScanBucket(s3URL string) ([]model.Violation, error) {
	bucket, prefix, err := s.parseS3URL(s3URL)
	if err != nil {
		return nil, err
	}

	var violations []model.Violation

	// List objects in the bucket
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	})

	// Process each page of results
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("error listing objects: %v", err)
		}

		// Process each object in the page
		for _, obj := range page.Contents {
			// Skip directories
			if strings.HasSuffix(*obj.Key, "/") {
				continue
			}

			// Get object content
			result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
				Bucket: &bucket,
				Key:    obj.Key,
			})
			if err != nil {
				fmt.Printf("Error getting object %s: %v\n", *obj.Key, err)
				continue
			}

			// Scan object content
			scanner := bufio.NewScanner(result.Body)
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()

				// Check each pattern
				for _, pattern := range s.patterns {
					if pattern.CompiledRegex == nil {
						fmt.Printf("Warning: Skipping rule %s - regex not compiled\n", pattern.Name)
						continue
					}
					if pattern.CompiledRegex.MatchString(line) {
						violations = append(violations, model.Violation{
							Rule:     pattern,
							Match:    line,
							Location: fmt.Sprintf("s3://%s/%s", bucket, *obj.Key),
							LineNum:  lineNum,
							Context:  line,
						})
					}
				}
			}

			result.Body.Close()

			if err := scanner.Err(); err != nil {
				fmt.Printf("Error scanning object %s: %v\n", *obj.Key, err)
			}
		}
	}

	return violations, nil
} 