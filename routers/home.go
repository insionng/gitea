// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"fmt"

	"github.com/Unknwon/paginater"

	"bytes"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/routers/user"
)

const (
	// tplHome home page template
	tplHome base.TplName = "home"
	// tplExploreRepos explore repositories page template
	tplExploreRepos base.TplName = "explore/repos"
	// tplExploreUsers explore users page template
	tplExploreUsers base.TplName = "explore/users"
	// tplExploreOrganizations explore organizations page template
	tplExploreOrganizations base.TplName = "explore/organizations"
)

// Home render home page
func Home(ctx *context.Context) {
	if ctx.IsSigned {
		if !ctx.User.IsActive && setting.Service.RegisterEmailConfirm {
			ctx.Data["Title"] = ctx.Tr("auth.active_your_account")
			ctx.HTML(200, user.TplActivate)
		} else {
			user.Dashboard(ctx)
		}
		return
	}

	// Check auto-login.
	uname := ctx.GetCookie(setting.CookieUserName)
	if len(uname) != 0 {
		ctx.Redirect(setting.AppSubURL + "/user/login")
		return
	}

	ctx.Data["PageIsHome"] = true
	ctx.HTML(200, tplHome)
}

// RepoSearchOptions when calling search repositories
type RepoSearchOptions struct {
	Counter  func(bool) int64
	Ranger   func(int, int) ([]*models.Repository, error)
	Private  bool
	PageSize int
	OrderBy  string
	TplName  base.TplName
}

var (
	nullByte = []byte{0x00}
)

func isKeywordValid(keyword string) bool {
	return !bytes.Contains([]byte(keyword), nullByte)
}

// RenderRepoSearch render repositories search page
func RenderRepoSearch(ctx *context.Context, opts *RepoSearchOptions) {
	page := ctx.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		repos []*models.Repository
		count int64
		err   error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0 {
		repos, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			ctx.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter(opts.Private)
	} else {
		if isKeywordValid(keyword) {
			repos, count, err = models.SearchRepositoryByName(&models.SearchRepoOptions{
				Keyword:  keyword,
				OrderBy:  opts.OrderBy,
				Private:  opts.Private,
				Page:     page,
				PageSize: opts.PageSize,
			})
			if err != nil {
				ctx.Handle(500, "SearchRepositoryByName", err)
				return
			}
		}
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)

	for _, repo := range repos {
		if err = repo.GetOwner(); err != nil {
			ctx.Handle(500, "GetOwner", fmt.Errorf("%d: %v", repo.ID, err))
			return
		}
	}
	ctx.Data["Repos"] = repos

	ctx.HTML(200, opts.TplName)
}

// ExploreRepos render explore repositories page
func ExploreRepos(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreRepositories"] = true

	RenderRepoSearch(ctx, &RepoSearchOptions{
		Counter:  models.CountRepositories,
		Ranger:   models.GetRecentUpdatedRepositories,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "updated_unix DESC",
		TplName:  tplExploreRepos,
	})
}

// UserSearchOptions options when render search user page
type UserSearchOptions struct {
	Type     models.UserType
	Counter  func() int64
	Ranger   func(int, int) ([]*models.User, error)
	PageSize int
	OrderBy  string
	TplName  base.TplName
}

// RenderUserSearch render user search page
func RenderUserSearch(ctx *context.Context, opts *UserSearchOptions) {
	page := ctx.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	var (
		users []*models.User
		count int64
		err   error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0 {
		users, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			ctx.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter()
	} else {
		if isKeywordValid(keyword) {
			users, count, err = models.SearchUserByName(&models.SearchUserOptions{
				Keyword:  keyword,
				Type:     opts.Type,
				OrderBy:  opts.OrderBy,
				Page:     page,
				PageSize: opts.PageSize,
			})
			if err != nil {
				ctx.Handle(500, "SearchUserByName", err)
				return
			}
		}
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	ctx.Data["Users"] = users

	ctx.HTML(200, opts.TplName)
}

// ExploreUsers render explore users page
func ExploreUsers(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreUsers"] = true

	RenderUserSearch(ctx, &UserSearchOptions{
		Type:     models.UserTypeIndividual,
		Counter:  models.CountUsers,
		Ranger:   models.Users,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "name ASC",
		TplName:  tplExploreUsers,
	})
}

// ExploreOrganizations render explore organizations page
func ExploreOrganizations(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreOrganizations"] = true

	RenderUserSearch(ctx, &UserSearchOptions{
		Type:     models.UserTypeOrganization,
		Counter:  models.CountOrganizations,
		Ranger:   models.Organizations,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "name ASC",
		TplName:  tplExploreOrganizations,
	})
}

// NotFound render 404 page
func NotFound(ctx *context.Context) {
	ctx.Data["Title"] = "Page Not Found"
	ctx.Handle(404, "home.NotFound", nil)
}
