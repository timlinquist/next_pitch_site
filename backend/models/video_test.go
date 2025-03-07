package models

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

// MockStorageClient implements a mock storage client for testing
type MockStorageClient struct {
	uploadFunc           func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error)
	getTemporaryLinkFunc func(arg *files.GetTemporaryLinkArg) (*files.GetTemporaryLinkResult, error)
}

func (m *MockStorageClient) Upload(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(arg, content)
	}
	return nil, nil
}

func (m *MockStorageClient) GetTemporaryLink(arg *files.GetTemporaryLinkArg) (*files.GetTemporaryLinkResult, error) {
	if m.getTemporaryLinkFunc != nil {
		return m.getTemporaryLinkFunc(arg)
	}
	return nil, nil
}

func TestNewVideo(t *testing.T) {
	tests := []struct {
		name        string
		envToken    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful initialization",
			envToken: "test-token",
			wantErr:  false,
		},
		{
			name:        "missing token",
			envToken:    "",
			wantErr:     true,
			errContains: "DROPBOX_ACCESS_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("DROPBOX_ACCESS_TOKEN", tt.envToken)
			defer os.Unsetenv("DROPBOX_ACCESS_TOKEN")

			// Test initialization
			got, err := NewVideo()
			if tt.wantErr {
				if err == nil {
					t.Error("NewVideo() error = nil, wantErr true")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("NewVideo() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("NewVideo() error = %v, wantErr false", err)
				return
			}
			if got == nil {
				t.Error("NewVideo() returned nil without error")
			}
		})
	}
}

func TestVideo_Upload(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		filename    string
		uploadFunc  func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error)
		getLinkFunc func(arg *files.GetTemporaryLinkArg) (*files.GetTemporaryLinkResult, error)
		wantPath    string
		wantLink    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful upload",
			content:  "test video content",
			filename: "test.mp4",
			uploadFunc: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
				return nil, nil
			},
			getLinkFunc: func(arg *files.GetTemporaryLinkArg) (*files.GetTemporaryLinkResult, error) {
				return &files.GetTemporaryLinkResult{Link: "https://test.com/video.mp4"}, nil
			},
			wantPath: "/videos/test.mp4",
			wantLink: "https://test.com/video.mp4",
			wantErr:  false,
		},
		{
			name:     "upload failure",
			content:  "test video content",
			filename: "test.mp4",
			uploadFunc: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
				return nil, errors.New("upload failed")
			},
			wantErr:     true,
			errContains: "failed to upload to Dropbox",
		},
		{
			name:     "get link failure",
			content:  "test video content",
			filename: "test.mp4",
			uploadFunc: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
				return nil, nil
			},
			getLinkFunc: func(arg *files.GetTemporaryLinkArg) (*files.GetTemporaryLinkResult, error) {
				return nil, errors.New("get link failed")
			},
			wantErr:     true,
			errContains: "failed to generate download link",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &MockStorageClient{
				uploadFunc:           tt.uploadFunc,
				getTemporaryLinkFunc: tt.getLinkFunc,
			}

			// Create video instance with mock client
			v := &Video{
				dbx: mockClient,
			}

			// Create test reader
			reader := strings.NewReader(tt.content)

			// Test upload
			gotPath, gotLink, err := v.Upload(reader, tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Error("Upload() error = nil, wantErr true")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Upload() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("Upload() error = %v, wantErr false", err)
				return
			}
			if gotPath != tt.wantPath {
				t.Errorf("Upload() path = %v, want %v", gotPath, tt.wantPath)
			}
			if gotLink != tt.wantLink {
				t.Errorf("Upload() link = %v, want %v", gotLink, tt.wantLink)
			}
		})
	}
}
