// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package convert

import (
	project_model "code.gitea.io/gitea/models/project"
	api "code.gitea.io/gitea/modules/structs"
)

// ToAPIProject converts a Project to API format
func ToAPIProject(p *project_model.Project) *api.Project {
	apiProject := &api.Project{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		IsClosed:    p.IsClosed,
		Created:     p.CreatedUnix.AsTime(),
		Updated:     p.UpdatedUnix.AsTime(),
	}
	if p.ClosedDateUnix > 0 {
		apiProject.Closed = p.ClosedDateUnix.AsTimePtr()
	}
	return apiProject
}
