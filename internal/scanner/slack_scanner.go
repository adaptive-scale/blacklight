package scanner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/adaptive-scale/blacklight/internal/model"
	"github.com/slack-go/slack"
)

// SlackScanner represents a scanner for Slack messages and files
type SlackScanner struct {
	client *slack.Client
	config *SlackConfig
}

// SlackConfig holds configuration for Slack scanning
type SlackConfig struct {
	Token           string   // Slack Bot User OAuth Token
	Channels        []string // List of channel IDs to scan
	DaysToScan      int      // Number of days of history to scan
	IncludeThreads  bool     // Whether to scan message threads
	IncludeFiles    bool     // Whether to scan file contents
	ExcludeArchived bool     // Whether to exclude archived channels
}

// NewSlackScanner creates a new Slack scanner instance
func NewSlackScanner(config *SlackConfig) (*SlackScanner, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("slack token is required")
	}

	client := slack.New(config.Token)
	return &SlackScanner{
		client: client,
		config: config,
	}, nil
}

// Scan implements the Scanner interface for Slack
func (s *SlackScanner) Scan(ctx context.Context, rules []model.Configuration) ([]model.Violation, error) {
	var violations []model.Violation

	// If no specific channels provided, get all accessible channels
	channels := s.config.Channels
	if len(channels) == 0 {
		var err error
		channels, err = s.getAccessibleChannels()
		if err != nil {
			return nil, fmt.Errorf("failed to get channels: %w", err)
		}
	}

	// Calculate the oldest timestamp to scan
	oldest := time.Now().AddDate(0, 0, -s.config.DaysToScan).Unix()

	// Scan each channel
	for _, channelID := range channels {
		// Get channel info
		channel, err := s.client.GetConversationInfo(&slack.GetConversationInfoInput{
			ChannelID: channelID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get channel info for %s: %w", channelID, err)
		}

		// Skip archived channels if configured
		if s.config.ExcludeArchived && channel.IsArchived {
			continue
		}

		// Get channel messages
		params := slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Oldest:    fmt.Sprintf("%d", oldest),
		}
		
		history, err := s.client.GetConversationHistory(&params)
		if err != nil {
			return nil, fmt.Errorf("failed to get history for channel %s: %w", channelID, err)
		}

		// Scan messages
		for _, msg := range history.Messages {
			// Scan message text
			msgViolations := s.scanText(msg.Text, rules, channelID, msg.Timestamp)
			violations = append(violations, msgViolations...)

			// Scan thread if enabled
			if s.config.IncludeThreads && msg.ThreadTimestamp != "" {
				threadViolations, err := s.scanThread(ctx, channelID, msg.ThreadTimestamp, rules)
				if err != nil {
					return nil, fmt.Errorf("failed to scan thread: %w", err)
				}
				violations = append(violations, threadViolations...)
			}

			// Scan files if enabled
			if s.config.IncludeFiles && len(msg.Files) > 0 {
				fileViolations, err := s.scanFiles(ctx, msg.Files, rules)
				if err != nil {
					return nil, fmt.Errorf("failed to scan files: %w", err)
				}
				violations = append(violations, fileViolations...)
			}
		}
	}

	return violations, nil
}

// getAccessibleChannels returns a list of channel IDs the bot has access to
func (s *SlackScanner) getAccessibleChannels() ([]string, error) {
	var channels []string
	
	// List all channels
	params := &slack.GetConversationsParameters{
		ExcludeArchived: s.config.ExcludeArchived,
		Types:          []string{"public_channel", "private_channel"},
	}
	
	for {
		convs, cursor, err := s.client.GetConversations(params)
		if err != nil {
			return nil, err
		}

		for _, channel := range convs {
			channels = append(channels, channel.ID)
		}

		if cursor == "" {
			break
		}
		params.Cursor = cursor
	}

	return channels, nil
}

// scanText scans a text string for rule violations
func (s *SlackScanner) scanText(text string, rules []model.Configuration, channelID, timestamp string) []model.Violation {
	var violations []model.Violation

	for _, rule := range rules {
		if rule.Disabled {
			continue
		}

		if rule.CompiledRegex == nil {
			fmt.Printf("Warning: Skipping rule %s - regex not compiled\n", rule.Name)
			continue
		}

		matches := rule.CompiledRegex.FindAllString(text, -1)
		for _, match := range matches {
			violation := model.Violation{
				Rule:     rule,
				Match:    match,
				Location: fmt.Sprintf("slack://channel/%s/message/%s", channelID, timestamp),
				Context:  s.getMessageContext(text, match),
			}
			violations = append(violations, violation)
		}
	}

	return violations
}

// scanThread scans a message thread for violations
func (s *SlackScanner) scanThread(ctx context.Context, channelID, threadTS string, rules []model.Configuration) ([]model.Violation, error) {
	var violations []model.Violation

	params := slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
	}

	msgs, _, _, err := s.client.GetConversationReplies(&params)
	if err != nil {
		return nil, err
	}

	for _, reply := range msgs {
		msgViolations := s.scanText(reply.Text, rules, channelID, reply.Timestamp)
		violations = append(violations, msgViolations...)

		if s.config.IncludeFiles && len(reply.Files) > 0 {
			fileViolations, err := s.scanFiles(ctx, reply.Files, rules)
			if err != nil {
				return nil, err
			}
			violations = append(violations, fileViolations...)
		}
	}

	return violations, nil
}

// scanFiles scans file contents for violations
func (s *SlackScanner) scanFiles(ctx context.Context, files []slack.File, rules []model.Configuration) ([]model.Violation, error) {
	var violations []model.Violation

	for _, file := range files {
		// Skip files larger than 10MB to prevent memory issues
		if file.Size > 10*1024*1024 {
			continue
		}

		// Download file content
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, file.URLPrivateDownload, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for file %s: %w", file.ID, err)
		}
		req.Header.Set("Authorization", "Bearer "+s.config.Token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to download file %s: %w", file.ID, err)
		}
		defer resp.Body.Close()

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s content: %w", file.ID, err)
		}

		// Scan file content
		for _, rule := range rules {
			if rule.Disabled {
				continue
			}

			if rule.CompiledRegex == nil {
				fmt.Printf("Warning: Skipping rule %s - regex not compiled\n", rule.Name)
				continue
			}

			matches := rule.CompiledRegex.FindAllString(string(content), -1)
			for _, match := range matches {
				violation := model.Violation{
					Rule:     rule,
					Match:    match,
					Location: fmt.Sprintf("slack://file/%s", file.ID),
					Context:  fmt.Sprintf("File: %s", file.Name),
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations, nil
}

// getMessageContext returns a snippet of text around the matched string
func (s *SlackScanner) getMessageContext(text, match string) string {
	const contextLength = 50 // characters before and after the match

	start := strings.Index(text, match)
	if start == -1 {
		return ""
	}

	// Calculate context boundaries
	contextStart := start - contextLength
	if contextStart < 0 {
		contextStart = 0
	}
	contextEnd := start + len(match) + contextLength
	if contextEnd > len(text) {
		contextEnd = len(text)
	}

	// Add ellipsis if context is truncated
	prefix := ""
	if contextStart > 0 {
		prefix = "..."
	}
	suffix := ""
	if contextEnd < len(text) {
		suffix = "..."
	}

	return prefix + text[contextStart:contextEnd] + suffix
} 