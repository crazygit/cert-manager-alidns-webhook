#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <version>" >&2
  echo "Example: $0 0.1.3" >&2
  exit 1
fi

version="${1#v}"
chart_path="deploy/cert-manager-alidns-webhook/Chart.yaml"

if [[ ! -f "${chart_path}" ]]; then
  echo "ERROR: ${chart_path} not found." >&2
  exit 1
fi

# Update chart version and appVersion in place.
tmp_file="$(mktemp)"
awk -v ver="${version}" '
  /^version:/ { print "version: " ver; next }
  /^appVersion:/ { print "appVersion: \"" ver "\""; next }
  { print }
' "${chart_path}" > "${tmp_file}"
mv "${tmp_file}" "${chart_path}"

./scripts/check-version.sh "${version}"

echo "OK: Chart.yaml updated to ${version}."
echo "Next: git add ${chart_path} && git commit -m \"release: v${version}\" && git tag -a v${version} -m \"release: v${version}\" && git push --follow-tags"
