// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package org

import (
	"net/http"

	"code.gitea.io/gitea/models/db"
	issues_model "code.gitea.io/gitea/models/issues"
	"code.gitea.io/gitea/modules/optional"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/templates"
	"code.gitea.io/gitea/modules/web"
	"code.gitea.io/gitea/routers/common"
	"code.gitea.io/gitea/services/context"
	"code.gitea.io/gitea/services/forms"
)

const (
	tplOrgMilestones   templates.TplName = "org/milestones"
	tplOrgMilestoneNew templates.TplName = "org/milestone_new"
)

func orgMilestonesLink(ctx *context.Context) string {
	return ctx.ContextUser.HomeLink() + "/-/milestones"
}

// OrgMilestones renders the org milestone list page
func OrgMilestones(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.milestones")
	ctx.Data["PageIsViewOrgMilestones"] = true

	isShowClosed := ctx.FormString("state") == "closed"
	keyword := ctx.FormTrim("q")
	sortType := ctx.FormString("sort")
	page := max(ctx.FormInt("page"), 1)

	miles, total, err := db.FindAndCount[issues_model.Milestone](ctx, issues_model.FindMilestoneOptions{
		ListOptions: db.ListOptions{
			Page:     page,
			PageSize: setting.UI.IssuePagingNum,
		},
		OrgID:    ctx.ContextUser.ID,
		IsClosed: optional.Some(isShowClosed),
		Name:     keyword,
		SortType: sortType,
	})
	if err != nil {
		ctx.ServerError("FindAndCount milestones", err)
		return
	}

	openCount, err := db.Count[issues_model.Milestone](ctx, issues_model.FindMilestoneOptions{
		OrgID: ctx.ContextUser.ID, IsClosed: optional.Some(false), Name: keyword,
	})
	if err != nil {
		ctx.ServerError("Count open milestones", err)
		return
	}
	closedCount, err := db.Count[issues_model.Milestone](ctx, issues_model.FindMilestoneOptions{
		OrgID: ctx.ContextUser.ID, IsClosed: optional.Some(true), Name: keyword,
	})
	if err != nil {
		ctx.ServerError("Count closed milestones", err)
		return
	}

	if isShowClosed {
		ctx.Data["State"] = "closed"
	} else {
		ctx.Data["State"] = "open"
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["SortType"] = sortType
	ctx.Data["IsShowClosed"] = isShowClosed
	ctx.Data["OpenCount"] = openCount
	ctx.Data["ClosedCount"] = closedCount
	ctx.Data["Milestones"] = miles
	ctx.Data["MilestoneLink"] = orgMilestonesLink(ctx)

	pager := context.NewPagination(total, setting.UI.IssuePagingNum, page, 5)
	pager.AddParamFromRequest(ctx.Req)
	ctx.Data["Page"] = pager

	ctx.HTML(http.StatusOK, tplOrgMilestones)
}

// NewOrgMilestone renders the create milestone form
func NewOrgMilestone(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.milestones.new")
	ctx.Data["PageIsViewOrgMilestones"] = true
	ctx.Data["MilestoneLink"] = orgMilestonesLink(ctx)
	ctx.Data["Link"] = orgMilestonesLink(ctx) + "/new"
	ctx.HTML(http.StatusOK, tplOrgMilestoneNew)
}

// NewOrgMilestonePost handles milestone creation
func NewOrgMilestonePost(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.CreateMilestoneForm)
	ctx.Data["Title"] = ctx.Tr("org.milestones.new")
	ctx.Data["PageIsViewOrgMilestones"] = true
	ctx.Data["MilestoneLink"] = orgMilestonesLink(ctx)
	ctx.Data["Link"] = orgMilestonesLink(ctx) + "/new"

	if ctx.HasError() {
		ctx.HTML(http.StatusOK, tplOrgMilestoneNew)
		return
	}

	deadlineUnix, err := common.ParseDeadlineDateToEndOfDay(form.Deadline)
	if err != nil {
		ctx.Data["Err_Deadline"] = true
		ctx.RenderWithErrDeprecated(ctx.Tr("repo.milestones.invalid_due_date_format"), tplOrgMilestoneNew, &form)
		return
	}

	if err := issues_model.NewMilestone(ctx, &issues_model.Milestone{
		OrgID:        ctx.ContextUser.ID,
		Name:         form.Title,
		Content:      form.Content,
		DeadlineUnix: deadlineUnix,
	}); err != nil {
		ctx.ServerError("NewMilestone", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("repo.milestones.create_success", form.Title))
	ctx.Redirect(orgMilestonesLink(ctx))
}

// EditOrgMilestone renders the edit milestone form
func EditOrgMilestone(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("org.milestones.edit")
	ctx.Data["PageIsViewOrgMilestones"] = true
	ctx.Data["PageIsEditMilestone"] = true
	ctx.Data["MilestoneLink"] = orgMilestonesLink(ctx)

	m, err := issues_model.GetMilestoneByOrgID(ctx, ctx.ContextUser.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if issues_model.IsErrMilestoneNotExist(err) {
			ctx.NotFound(nil)
		} else {
			ctx.ServerError("GetMilestoneByOrgID", err)
		}
		return
	}
	ctx.Data["title"] = m.Name
	ctx.Data["content"] = m.Content
	ctx.Data["Link"] = orgMilestonesLink(ctx) + "/" + ctx.PathParam("id") + "/edit"
	if len(m.DeadlineString) > 0 {
		ctx.Data["deadline"] = m.DeadlineString
	}
	ctx.HTML(http.StatusOK, tplOrgMilestoneNew)
}

// EditOrgMilestonePost handles milestone update
func EditOrgMilestonePost(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.CreateMilestoneForm)
	ctx.Data["Title"] = ctx.Tr("org.milestones.edit")
	ctx.Data["PageIsViewOrgMilestones"] = true
	ctx.Data["PageIsEditMilestone"] = true
	ctx.Data["MilestoneLink"] = orgMilestonesLink(ctx)
	ctx.Data["Link"] = orgMilestonesLink(ctx) + "/" + ctx.PathParam("id") + "/edit"

	if ctx.HasError() {
		ctx.HTML(http.StatusOK, tplOrgMilestoneNew)
		return
	}

	deadlineUnix, err := common.ParseDeadlineDateToEndOfDay(form.Deadline)
	if err != nil {
		ctx.Data["Err_Deadline"] = true
		ctx.RenderWithErrDeprecated(ctx.Tr("repo.milestones.invalid_due_date_format"), tplOrgMilestoneNew, &form)
		return
	}

	m, err := issues_model.GetMilestoneByOrgID(ctx, ctx.ContextUser.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if issues_model.IsErrMilestoneNotExist(err) {
			ctx.NotFound(nil)
		} else {
			ctx.ServerError("GetMilestoneByOrgID", err)
		}
		return
	}
	m.Name = form.Title
	m.Content = form.Content
	m.DeadlineUnix = deadlineUnix
	if err = issues_model.UpdateMilestone(ctx, m, m.IsClosed); err != nil {
		ctx.ServerError("UpdateMilestone", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("repo.milestones.edit_success", m.Name))
	ctx.Redirect(orgMilestonesLink(ctx))
}

// ChangeOrgMilestoneStatus handles opening/closing a milestone
func ChangeOrgMilestoneStatus(ctx *context.Context) {
	var toClose bool
	switch ctx.PathParam("action") {
	case "open":
		toClose = false
	case "close":
		toClose = true
	default:
		ctx.JSONRedirect(orgMilestonesLink(ctx))
		return
	}

	m, err := issues_model.GetMilestoneByOrgID(ctx, ctx.ContextUser.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if issues_model.IsErrMilestoneNotExist(err) {
			ctx.NotFound(nil)
		} else {
			ctx.ServerError("GetMilestoneByOrgID", err)
		}
		return
	}
	if err := issues_model.ChangeMilestoneStatus(ctx, m, toClose); err != nil {
		ctx.ServerError("ChangeMilestoneStatus", err)
		return
	}
	ctx.JSONRedirect(orgMilestonesLink(ctx) + "?state=" + ctx.PathParam("action"))
}

// DeleteOrgMilestone deletes a milestone from the organization
func DeleteOrgMilestone(ctx *context.Context) {
	if err := issues_model.DeleteMilestoneByOrgID(ctx, ctx.ContextUser.ID, ctx.FormInt64("id")); err != nil {
		ctx.Flash.Error("DeleteMilestoneByOrgID: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("repo.milestones.deletion_success"))
	}
	ctx.JSONRedirect(orgMilestonesLink(ctx))
}
