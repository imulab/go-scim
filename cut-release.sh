#!/bin/bash

# cut-release script for the entitlements repo.

set -ex

BRANCH_NAME="$1"
REPO_NAME="$2"
COMMIT_SHA="$3"

if [ -z "$BRANCH_NAME" ] || [ -z "$REPO_NAME" ] || [ -z "$COMMIT_SHA" ]; then
  echo "possibly manually triggered build, skipping tagging."
  exit 0
fi

if [ $BRANCH_NAME != "main" ]; then
  echo "push not to main branch, skipping tagging."
  exit 0 # no releases to cut except when push is to main.
fi

if [ ! -f VERSIONS ]; then
  echo "VERSIONS file not found, not tagging"
  exit 0
fi

# create all tags in VERSIONS file that don't already exist.
newtaglist=`cat VERSIONS`
tagstopush=()
git fetch --tags
for tag in $newtaglist; do
  if ! git tag | grep -q -e '^'$tag'$'; then
    # this is a new tag
    git tag $tag $COMMIT_SHA
    tagstopush+=($tag)
  fi
done

# push all new tags to origin
if [ ${#tagstopush[@]} -gt 0 ]; then
  git push origin ${tagstopush[@]}
fi
