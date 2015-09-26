#!/bin/bash

set -o errexit

echoerr() { echo "$@" 1>&2; }

if [ $# -lt 3 ]; then
  echoerr "Usage: release.sh <repo> <version> <files>"
  exit 1
fi

repo=$1
new_version=$2
shift; shift
files=$@

if ! [[ "$new_version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echoerr "Usage: version must follow pattern vN.N.N: $new_version"
  exit 1
fi

check_github_release_config() {
  if ! github-release edit \
    --user andersjanmyr \
    --repo route53copy \
    --tag v1.0.0 \
    --name "Release 1.0.0" ; then
    echoerr "github release must be installed and configured"
    echoerr "https://github.com/aktau/github-release#how-to-install"
    exit 1
  fi
}
check_github_release_config

echo "Create Release $new_version"
github-release release \
    --user andersjanmyr \
    --repo $repo \
    --tag $new_version \
    --name "Release $new_version" \
    --description "$description"
echo "Upload assets $files"
for f in $files; do
  echo "Upload $f as ${f#*/}"
  github-release upload \
      --user andersjanmyr \
      --repo $repo \
      --tag $new_version \
      --name "${f#*/}" \
      --file "$f"
done

echo "Update Homebrew formula"

