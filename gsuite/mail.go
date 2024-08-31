package gsuite

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/jlewi/gctl/config"
	"github.com/jlewi/gctl/util"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"time"
)

func NewInbox(cfg config.Config, ts oauth2.TokenSource) (*Inbox, error) {
	// TODO(jeremy): Should we inject the credentials?
	svc, err := gmail.NewService(context.Background(), option.WithCredentialsFile(cfg.OAuthClientFile), option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail client: %v", err)
	}
	return &Inbox{
		config: cfg,
		svc:    svc,
	}, nil
}

type EmailInfo struct {
	ID      string
	From    string
	To      string
	Subject string
	Snippet string
	Date    time.Time
}

type Email struct {
	ID      string
	From    string
	To      string
	Subject string
	Body    string
	Date    time.Time
}

// Inbox is a struct that provides routines for interacting with your inbox.
type Inbox struct {
	config config.Config
	svc    *gmail.Service
}

func (i *Inbox) GetMessage(ctx context.Context, messageID string) (*Email, error) {
	// Replace "path/to/your/credentials.json" with the path to your downloaded client configuration file
	user := "me" // Special value to indicate the authenticated user

	fullMsg, err := i.svc.Users.Messages.Get(user, messageID).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to search Gmail: %v", err)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "Error retrieving message with id %s", messageID)
	}

	msg := &Email{
		Date: parseEpochMillis(fullMsg.InternalDate).Local(),
	}
	for _, header := range fullMsg.Payload.Headers {
		switch header.Name {
		case "From":
			msg.From = header.Value
		case "To":
			msg.To = header.Value
		case "Subject":
			msg.Subject = header.Value
		}
	}

	// The body isn't always set for certain types of messages
	// https://developers.google.com/gmail/api/reference/rest/v1/users.messages#Message.MessagePart
	if fullMsg.Payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(fullMsg.Payload.Body.Data)
		if err != nil {
			return msg, fmt.Errorf("failed to decode base64 URL-encoded string: %v", err)
		}
		msg.Body = string(decoded)
	}

	if len(fullMsg.Payload.Parts) > 0 {
		for _, part := range fullMsg.Payload.Parts {
			if part.MimeType == "text/plain" || part.MimeType == "text/html" {
				decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err != nil {
					return msg, fmt.Errorf("failed to decode base64 URL-encoded string: %v", err)
				}
				msg.Body += string(decoded)
			}
		}
	}

	return msg, nil
}

func (i *Inbox) Search(ctx context.Context, query string, maxResults int64, pageToken string) ([]*EmailInfo, error) {
	log := util.LoggerFromContext(ctx)
	// Replace "path/to/your/credentials.json" with the path to your downloaded client configuration file
	user := "me" // Special value to indicate the authenticated user
	searchRequest := i.svc.Users.Messages.List(user).Q(query).MaxResults(maxResults)
	if pageToken != "" {
		searchRequest.PageToken(pageToken)
	}
	response, err := searchRequest.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to search Gmail: %v", err)
	}

	// Get the full message for each message
	// The search request only returns the id and threadId
	emailInfos := make([]*EmailInfo, 0, len(response.Messages))
	for _, msg := range response.Messages {
		fullMsg, err := i.svc.Users.Messages.Get(user, msg.Id).Format("metadata").MetadataHeaders("From", "To", "Subject", "Date").Do()
		if err != nil {
			log.Error(err, "Error retrieving message", "messageId", msg.Id, "messageThreadId", msg.ThreadId)
			continue

		}

		info := &EmailInfo{
			ID:      fullMsg.Id,
			Snippet: fullMsg.Snippet,
			Date:    parseEpochMillis(fullMsg.InternalDate).Local(),
		}

		for _, header := range fullMsg.Payload.Headers {
			switch header.Name {
			case "From":
				info.From = header.Value
			case "To":
				info.To = header.Value
			case "Subject":
				info.Subject = header.Value
			}
		}
		emailInfos = append(emailInfos, info)
	}

	return emailInfos, nil
}

// parseEpochMillis converts an epoch time in milliseconds to a time.Time
func parseEpochMillis(epochMillis int64) time.Time {
	return time.Unix(0, epochMillis*int64(time.Millisecond))
}
