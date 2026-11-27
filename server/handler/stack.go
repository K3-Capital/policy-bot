// Copyright 2026 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"context"

	"github.com/google/go-github/v93/github"
	"github.com/palantir/policy-bot/pull"
	"github.com/pkg/errors"
)

// maxStackDepth bounds the walk up a stack and protects against cycles.
const maxStackDepth = 50

// stackedContext reports the root base branch of a stack as the base branch of
// the pull request, so that the policy and status context match those of the
// branch the stack will eventually merge into.
type stackedContext struct {
	pull.Context
	base string
}

func (c *stackedContext) Branches() (string, string) {
	_, head := c.Context.Branches()
	return c.base, head
}

// resolveStackRoot follows a chain of open pull requests, where each pull
// request targets the head branch of another open pull request in the same
// repository, and returns the base branch at the bottom of the chain. If the
// base branch is not the head of an open pull request, it is returned as is.
func resolveStackRoot(ctx context.Context, client *github.Client, owner, repo, base string) (string, error) {
	seen := map[string]bool{base: true}
	for i := 0; i < maxStackDepth; i++ {
		prs, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
			State:       "open",
			Head:        owner + ":" + base,
			ListOptions: github.ListOptions{PerPage: 1},
		})
		if err != nil {
			return "", errors.Wrapf(err, "failed to list pull requests with head %q", base)
		}
		if len(prs) == 0 {
			return base, nil
		}
		parent := prs[0].GetBase().GetRef()
		if parent == "" || seen[parent] {
			return base, nil
		}
		seen[parent] = true
		base = parent
	}
	return base, nil
}
