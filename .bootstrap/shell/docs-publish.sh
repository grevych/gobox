#!/usr/bin/env bash
#
# Either creates or updates a gh-pages branch containing docs
#
# Initially based on https://github.com/malept/github-action-gh-pages/blob/v1.3.0/entrypoint.sh
# which is under the Apache 2.0 License
#

set -e

SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

# shellcheck source=./lib/bootstrap.sh
source "$SCRIPTS_DIR/lib/bootstrap.sh"

DOCS_DIR="$(get_repo_directory)/apidocs"
GH_DOCS_DIR="$(get_repo_directory)/docs"
NODEJS_CLIENT_DIR="$(get_repo_directory)/api/clients/node"

GITHUB_ACTOR=outreach-ci
GIT_COMMIT_USER="Outreach CI"
GIT_COMMIT_EMAIL="$GITHUB_ACTOR@users.noreply.github.com"
GH_PAGES_CUSTOM_DOMAIN="$(get_app_name).engdocs.outreach.cloud"
PUBLISH_BRANCH=gh-pages

set -e

if [[ -d "$NODEJS_CLIENT_DIR"/node_modules ]]; then
  rm -r "$NODEJS_CLIENT_DIR"/node_modules
fi

if ! git branch --list --remote | grep --quiet "origin/${PUBLISH_BRANCH}$"; then
  echo "Creating new $PUBLISH_BRANCH..."
  # Create a new branch without any history
  git checkout --orphan $PUBLISH_BRANCH

  # Delete everything
  git rm --force -r .
else
  mv "$DOCS_DIR" staged_docs
  # Revert any inconsistencies generated by protoc and friends
  # For example, Go import statement placement
  git restore api/
  echo "Switching to existing $PUBLISH_BRANCH..."
  git checkout $PUBLISH_BRANCH
  git rm -r "$GH_DOCS_DIR"
  mv staged_docs "$GH_DOCS_DIR"
fi

if [[ -n $GH_PAGES_CUSTOM_DOMAIN ]]; then
  echo "$GH_PAGES_CUSTOM_DOMAIN" >"$GH_DOCS_DIR"/CNAME
fi

git add "$GH_DOCS_DIR"

if test -n "$(git status -s)"; then
  git config user.name "$GIT_COMMIT_USER"
  git config user.email "$GIT_COMMIT_EMAIL"
  git commit --message "docs: rebuild docs [skip ci]"

  echo "machine github.com login $GITHUB_ACTOR password $OUTREACH_GITHUB_TOKEN" >~/.netrc
  chmod 600 ~/.netrc

  git push origin HEAD:$PUBLISH_BRANCH
fi
