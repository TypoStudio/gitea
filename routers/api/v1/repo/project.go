// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"net/http"

	"code.gitea.io/gitea/models/db"
	issues_model "code.gitea.io/gitea/models/issues"
	project_model "code.gitea.io/gitea/models/project"
	"code.gitea.io/gitea/modules/optional"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/routers/api/v1/utils"
	"code.gitea.io/gitea/services/context"
	"code.gitea.io/gitea/services/convert"
)

// ListRepoProjects returns a list of projects in a repository
func ListRepoProjects(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/projects project repoListProjects
	// ---
	// summary: Get all of a repository's projects
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: page
	//   in: query
	//   description: page number of results to return (1-based)
	//   type: integer
	// - name: limit
	//   in: query
	//   description: page size of results
	//   type: integer
	// responses:
	//   "200":
	//     "$ref": "#/responses/ProjectList"
	//   "404":
	//     "$ref": "#/responses/notFound"

	listOptions := utils.GetListOptions(ctx)
	opts := project_model.SearchOptions{
		ListOptions: listOptions,
		RepoID:      ctx.Repo.Repository.ID,
		IsClosed:    optional.None[bool](),
		Type:        project_model.TypeRepository,
	}

	projects, err := db.Find[project_model.Project](ctx, opts)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	count, err := db.Count[project_model.Project](ctx, opts)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	apiProjects := make([]*api.Project, len(projects))
	for i := range projects {
		apiProjects[i] = convert.ToAPIProject(projects[i])
	}

	ctx.SetTotalCountHeader(count)
	ctx.JSON(http.StatusOK, apiProjects)
}

// GetProjectIssues returns a list of issues in a project
func GetProjectIssues(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/projects/{project_id}/issues project repoGetProjectIssues
	// ---
	// summary: Get the issues of a project
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: project_id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: page
	//   in: query
	//   description: page number of results to return (1-based)
	//   type: integer
	// - name: limit
	//   in: query
	//   description: page size of results
	//   type: integer
	// responses:
	//   "200":
	//     "$ref": "#/responses/IssueList"
	//   "404":
	//     "$ref": "#/responses/notFound"

	projectID := ctx.PathParamInt64("project_id")

	// Verify the project belongs to this repo
	_, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, projectID)
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	// Get project issue relations
	var projectIssues []project_model.ProjectIssue
	if err := db.GetEngine(ctx).Where("project_id=?", projectID).Find(&projectIssues); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	if len(projectIssues) == 0 {
		ctx.SetTotalCountHeader(0)
		ctx.JSON(http.StatusOK, []*api.Issue{})
		return
	}

	issueIDs := make([]int64, 0, len(projectIssues))
	for _, pi := range projectIssues {
		issueIDs = append(issueIDs, pi.IssueID)
	}

	issues, err := issues_model.GetIssuesByIDs(ctx, issueIDs)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.SetTotalCountHeader(int64(len(issues)))
	ctx.JSON(http.StatusOK, convert.ToAPIIssueList(ctx, ctx.Doer, issues))
}
