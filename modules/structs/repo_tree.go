// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import "time"

// GitEntry represents a git tree
type GitEntry struct {
	Name            string    `json:"name"`
	Path            string    `json:"path"`
	Mode            string    `json:"mode"`
	Type            string    `json:"type"`
	Size            int64     `json:"size"`
	SHA             string    `json:"sha"`
	URL             string    `json:"url"`
	CommitMsg       string    `json:"commit_msg"`
	CommitterDate   time.Time `json:"committer_date"`
	IsLfs           bool      `json:"is_lfs"`
	LfsRelativePath string    `json:"lfs_relative_path"`
	DownloadUrl     string    `json:"download_url"`
}

// GitTreeResponse returns a git tree
type GitTreeResponse struct {
	SHA        string     `json:"sha"`
	URL        string     `json:"url"`
	Entries    []GitEntry `json:"tree"`
	Truncated  bool       `json:"truncated"`
	Page       int        `json:"page"`
	TotalCount int        `json:"total_count"`
}
