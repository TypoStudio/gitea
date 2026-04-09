// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_26

import "xorm.io/xorm"

// AddOrgIDToMilestone adds org_id column to milestone table for organization-wide milestones.
func AddOrgIDToMilestone(x *xorm.Engine) error {
	type Milestone struct {
		OrgID int64 `xorm:"INDEX DEFAULT 0"`
	}
	return x.Sync(new(Milestone))
}
