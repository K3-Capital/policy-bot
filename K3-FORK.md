# K3 Capital fork of policy-bot

`develop` of this fork is what K3 Capital runs (infra-k3, `production/policy-bot`). It is **the upstream release v1.42.1 plus two
commits of the open upstream pull request
[palantir/policy-bot#1363](https://github.com/palantir/policy-bot/pull/1363) "Add policy_from_default_branch option"**, and the K3
files listed below, nothing else. It is **not** a copy of upstream `develop` any more: upstream `develop` has moved on past the
release and none of that is in here.

## Why

GitHub's native stacked pull requests apply the required checks of the stack's *final target branch* to every pull request of a
stack. policy-bot names its status after the branch a pull request targets (`policy-bot: <base branch>`), so the upper pull requests of
a stack wait for a status that is never posted, and the stack cannot be merged at once
(upstream issue [#1332](https://github.com/palantir/policy-bot/issues/1332)). With `policy_from_default_branch` the policy is loaded from
the default branch and the status is the plain `policy-bot`, which is the same for every pull request. See the README section of the
option for the security trade-offs: one policy source for all base branches (we use a single shared policy anyway), and a short window
after a pull request is retargeted in which the previous status is still on the commit.

The option is off by default. It is turned on with `POLICYBOT_OPTIONS_POLICY_FROM_DEFAULT_BRANCH=true` (infra-k3,
`production/policy-bot`, variable `policy_from_default_branch`).

## What is in the branch

| Commit | |
|---|---|
| `Release v1.42.1` | upstream |
| `Add policy_from_default_branch option`, `Force full re-evaluation when a pull request's base branch changes` | cherry-picked from #1363 (one trivial conflict in `server/handler/eval_options.go`, resolved by keeping `AllowPrivateRemotes` from v1.42.1) |
| `K3: build the test of #1363 against go-github v93` | the test of #1363 imports go-github v90, v1.42.1 vendors v93 |
| `K3: publish the image to ECR` | `docker/Dockerfile.k3`, `.github/workflows/k3-publish-image.yml`, this file |
| `Take develop to the K3 line` | a merge that keeps this content and records the old `develop` (a copy of upstream `develop`) as a parent |

## Image

A push to `develop` runs the tests and, once one of the Production Approvers has approved the run (environment `publish`, the person
who pushed cannot approve), publishes
`648832445185.dkr.ecr.eu-north-1.amazonaws.com/policy-bot:sha-<commit>` (linux/amd64, same layout as the official image, checked by
the workflow). The workflow summary prints the digest; the ECS task is changed to that digest in infra-k3 by a reviewed change, so
nothing is deployed from here.

`develop` is protected by the organisation ruleset `protect-default-branch` (pull request, approval, no force-push): everything that
reaches it, and so everything that can be published, has been reviewed.

## Updating

- **When #1363 is merged or replaced upstream:** move to the upstream release that contains it and drop the cherry-picks. Do it
  with a branch whose content is that release (and whose history includes this `develop`, as the merge above does), through a pull
  request, because `develop` cannot be force-pushed.
- **To follow a new upstream release before that:** branch from its tag, cherry-pick the two #1363 commits (they may need a new
  conflict resolution) and the K3 commits, and bring the result into `develop` the same way.
- This is a security control: read the diff of anything that is taken from upstream, and keep the delta from upstream this small.
