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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveStackRoot(t *testing.T) {
	// open PRs keyed by head branch, value is the base branch
	open := map[string]string{
		"stack-3": "stack-2",
		"stack-2": "stack-1",
		"stack-1": "develop",
		"loop-a":  "loop-b",
		"loop-b":  "loop-a",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		head := r.URL.Query().Get("head")
		branch, _ := url.QueryUnescape(head)
		branch = branch[len("org:"):]

		var prs []*github.PullRequest
		if base, ok := open[branch]; ok {
			prs = append(prs, &github.PullRequest{Base: &github.PullRequestBranch{Ref: &base}})
		}
		_ = json.NewEncoder(w).Encode(prs)
	}))
	defer srv.Close()

	client := github.NewClient(nil)
	u, err := url.Parse(srv.URL + "/")
	require.NoError(t, err)
	client.BaseURL = u

	tests := map[string]struct {
		base     string
		expected string
	}{
		"notStacked":     {base: "develop", expected: "develop"},
		"singleParent":   {base: "stack-1", expected: "develop"},
		"deepStack":      {base: "stack-3", expected: "develop"},
		"cycleTerminate": {base: "loop-a", expected: "loop-b"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := resolveStackRoot(context.Background(), client, "org", "repo", tc.base)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
