// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import "time"

// Project represents a project
// swagger:model
type Project struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// swagger:strfmt date-time
	Created time.Time `json:"created_at"`
	// swagger:strfmt date-time
	Updated time.Time `json:"updated_at"`
	// swagger:strfmt date-time
	Closed *time.Time `json:"closed_at"`
	IsClosed bool      `json:"is_closed"`
}

// IssueProjectOption the option for adding/removing a project to/from an issue
type IssueProjectOption struct {
	// project id
	//
	// required: true
	ProjectID int64 `json:"project_id" binding:"Required"`
	// column id, if not provided, the default column will be used
	ColumnID int64 `json:"column_id"`
}
