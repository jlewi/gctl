package gsuite

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/jlewi/gctl/config"
	"github.com/jlewi/gctl/util"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Drive struct {
	svc *drive.Service
}

func NewDrive(cfg config.Config, ts oauth2.TokenSource) (*Drive, error) {
	srv, err := drive.NewService(context.Background(), option.WithTokenSource(ts), option.WithCredentialsFile(cfg.OAuthClientFile))
	if err != nil {
		return nil, fmt.Errorf("unable to create Drive service: %v", err)
	}
	return &Drive{svc: srv}, nil
}

func (d *Drive) ImportToGoogleDoc(ctx context.Context, htmlFilePath, docTitle string, folderID string) (string, error) {
	log := util.LoggerFromContext(ctx)
	// URL to the Google Doc
	url := ""

	// Read the HTML file
	content, err := os.ReadFile(htmlFilePath)
	if err != nil {
		return url, errors.Wrapf(err, "unable to read HTML file: %v", htmlFilePath)
	}

	// Create a new Google Doc
	doc := &drive.File{
		Name:     docTitle,
		MimeType: "application/vnd.google-apps.document",
	}

	// If a folder ID is provided, set it as the parent
	if folderID != "" {
		doc.Parents = []string{folderID}
	}

	file, err := d.svc.Files.Create(doc).Do()
	if err != nil {
		return url, errors.Wrapf(err, "unable to create Google Doc")
	}

	// Import the HTML content into the Google Doc
	_, err = d.svc.Files.Update(file.Id, &drive.File{}).Media(bytes.NewReader(content)).Do()
	if err != nil {
		return url, errors.Wrapf(err, "unable to update Google Doc with HTML content")
	}

	url = fmt.Sprintf("https://docs.google.com/document/d/%s/edit", file.Id)
	log.Info("Successfully imported HTML to Google Doc.", "id", file.Id, "url", url)
	return url, nil
}

// Search Google Drive.
// Important: The query syntax used by the API isn't quite the same as that used in the UI.
// https://docs.google.com/document/d/196tomkYJloQcsVUsS19ozn2F67UiIUVGFwDL9OusYyM/edit
// This currently searches the user corpora.
func (d *Drive) Search(ctx context.Context, query string, maxResults int64, pageToken string) ([]*drive.File, error) {

	var files []*drive.File

	for {
		q := d.svc.Files.List().Q(query).
			Fields("nextPageToken, files(id, name, mimeType, createdTime, modifiedTime, size, webViewLink)").
			PageSize(maxResults).
			OrderBy("modifiedTime desc")

		if pageToken != "" {
			q = q.PageToken(pageToken)
		}

		result, err := q.Do()
		if err != nil {
			gErr, ok := err.(*googleapi.Error)
			if ok {
				if gErr.Body != "" {
					return nil, errors.Wrapf(err, "Unable to search Drive. Google API returned error: %s", gErr.Body)
				}
			}
			return nil, errors.Wrapf(err, "unable to search Drive")
		}

		files = append(files, result.Files...)

		pageToken = result.NextPageToken
		if pageToken == "" || int64(len(files)) >= maxResults {
			break
		}
	}

	if int64(len(files)) > maxResults {
		files = files[:maxResults]
	}

	return files, nil
}
