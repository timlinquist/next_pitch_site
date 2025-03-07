package models

import (
	"fmt"
	"io"
	"os"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

// StorageClientInterface defines the interface for storage operations
type StorageClientInterface interface {
	Upload(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error)
	GetTemporaryLink(arg *files.GetTemporaryLinkArg) (*files.GetTemporaryLinkResult, error)
}

type Video struct {
	dbx StorageClientInterface
}

func NewVideo() (*Video, error) {
	// Get Dropbox access token from environment
	accessToken := os.Getenv("DROPBOX_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("DROPBOX_ACCESS_TOKEN environment variable is not set")
	}

	// Initialize Dropbox client
	config := dropbox.Config{
		Token: accessToken,
	}
	dbx := files.New(config)

	return &Video{
		dbx: dbx,
	}, nil
}

func (v *Video) Upload(file io.Reader, filename string) (string, string, error) {
	// Upload to Dropbox using streaming
	dropboxPath := fmt.Sprintf("/videos/%s", filename)
	arg := files.NewUploadArg(dropboxPath)
	arg.Mode = &files.WriteMode{Tagged: dropbox.Tagged{Tag: "overwrite"}}

	// Upload the file directly from the reader
	_, err := v.dbx.Upload(arg, file)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to Dropbox: %v", err)
	}

	// Get a temporary link to the uploaded file
	arg2 := files.NewGetTemporaryLinkArg(dropboxPath)
	res, err := v.dbx.GetTemporaryLink(arg2)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate download link: %v", err)
	}

	return dropboxPath, res.Link, nil
}
