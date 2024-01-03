// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/lfs"
	"code.gitea.io/gitea/modules/structs"
	files_service "code.gitea.io/gitea/services/repository/files"
)

func GetDirInfos(ctx *context.APIContext) {
	var err error
	branch := ctx.Req.URL.Query().Get("branch")
	if len(branch) == 0 {
		branch = "main"
	}
	ctx.Repo.Commit, err = ctx.Repo.GitRepo.GetBranchCommit(branch)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "failed to get branch commit, error: "+err.Error())
		return
	}
	path := ctx.Req.URL.Query().Get("path")
	entries, err := getDirectoryEntries(ctx, branch, path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "failed to get directry entries, error: "+err.Error())
		return
	}
	ctx.JSON(http.StatusOK, entries)
}

func getDirectoryEntries(ctx *context.APIContext, branch, path string) ([]structs.GitEntry, error) {
	var (
		allEntries git.Entries
		commits    []git.CommitInfo
	)
	entry, err := ctx.Repo.Commit.GetTreeEntryByPath(path)
	if err != nil {
		return nil, err
	}

	if entry.Type() != "tree" {
		allEntries = append(allEntries, entry)
		commit, err := ctx.Repo.Commit.GetCommitByPath(path)
		if err != nil {
			return nil, err
		}

		commitInfo := git.CommitInfo{
			Entry:  entry,
			Commit: commit,
		}
		commits = append(commits, commitInfo)
		path = filepath.Dir(path)

	} else {
		tree, err := ctx.Repo.Commit.SubTree(path)
		if err != nil {
			return nil, fmt.Errorf("failed to exec SubTree, cause:%w", err)
		}

		allEntries, err := tree.ListEntries()
		if err != nil {
			return nil, fmt.Errorf("failed to exec ListEntries, cause:%w", err)
		}
		allEntries.CustomSort(base.NaturalSortLess)

		commits, _, err = allEntries.GetCommitsInfo(ctx, ctx.Repo.Commit, path)
		if err != nil {
			return nil, fmt.Errorf("failed to exec GetCommitsInfo, cause:%w", err)
		}
	}

	var ges = make([]structs.GitEntry, 0, len(commits))
	for _, c := range commits {
		e := structs.GitEntry{
			Name:          c.Entry.Name(),
			Path:          filepath.Join(path, c.Entry.Name()),
			Mode:          c.Entry.Mode().String(),
			Type:          c.Entry.Type(),
			Size:          c.Entry.Size(),
			SHA:           c.Commit.ID.String(),
			CommitMsg:     c.Commit.CommitMessage,
			CommitterDate: c.Commit.Committer.When,
		}
		e.URL = ctx.Repo.Repository.HTMLURL() + "/raw/branch/" + url.PathEscape(branch) + "/" + url.PathEscape(strings.TrimPrefix(e.Path, "/"))
		e.DownloadUrl = e.URL
		//lfs pointer size is less than 1024
		if c.Entry.Size() <= 1024 {
			content, _ := c.Entry.Blob().GetBlobContent(1024)
			p, _ := lfs.ReadPointerFromBuffer([]byte(content))
			if p.IsValid() {
				e.Size = p.Size
				e.IsLfs = true
				e.LfsRelativePath = p.RelativePath()
				e.DownloadUrl = ctx.Repo.Repository.HTMLURL() + "/media/branch/" + url.PathEscape(branch) + "/" + url.PathEscape(strings.TrimPrefix(e.Path, "/"))
			}
		}
		ges = append(ges, e)
	}

	return ges, nil
}

// GetTree get the tree of a repository.
func GetTree(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/git/trees/{sha} repository GetTree
	// ---
	// summary: Gets the tree of a repository.
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
	// - name: sha
	//   in: path
	//   description: sha of the commit
	//   type: string
	//   required: true
	// - name: recursive
	//   in: query
	//   description: show all directories and files
	//   required: false
	//   type: boolean
	// - name: page
	//   in: query
	//   description: page number; the 'truncated' field in the response will be true if there are still more items after this page, false if the last page
	//   required: false
	//   type: integer
	// - name: per_page
	//   in: query
	//   description: number of items per page
	//   required: false
	//   type: integer
	// responses:
	//   "200":
	//     "$ref": "#/responses/GitTreeResponse"
	//   "400":
	//     "$ref": "#/responses/error"
	//   "404":
	//     "$ref": "#/responses/notFound"

	sha := ctx.Params(":sha")
	if len(sha) == 0 {
		ctx.Error(http.StatusBadRequest, "", "sha not provided")
		return
	}
	if tree, err := files_service.GetTreeBySHA(ctx, ctx.Repo.Repository, ctx.Repo.GitRepo, sha, ctx.FormInt("page"), ctx.FormInt("per_page"), ctx.FormBool("recursive")); err != nil {
		ctx.Error(http.StatusBadRequest, "", err.Error())
	} else {
		ctx.SetTotalCountHeader(int64(tree.TotalCount))
		ctx.JSON(http.StatusOK, tree)
	}
}
