#!/usr/bin/env bash
set -euo pipefail

CHART_PATH="deploy/cert-manager-alidns-webhook/Chart.yaml"

version_input="${1:-}"

if [[ -n "${version_input}" ]]; then
  version="${version_input#v}"
elif [[ -n "${GITHUB_REF_NAME:-}" ]]; then
  version="${GITHUB_REF_NAME#v}"
else
  tag="$(git describe --tags --exact-match 2>/dev/null || true)"
  if [[ -z "${tag}" ]]; then
    echo "ERROR: no version provided and no exact git tag found." >&2
    exit 1
  fi
  version="${tag#v}"
fi

chart_version="$(awk -F: '/^version:/{gsub(/[ "]/,"",$2); print $2; exit}' "${CHART_PATH}")"
app_version="$(awk -F: '/^appVersion:/{gsub(/[ "]/,"",$2); print $2; exit}' "${CHART_PATH}")"

if [[ "${chart_version}" != "${version}" ]]; then
  echo "ERROR: Chart version (${chart_version}) != release version (${version})." >&2
  exit 1
fi

if [[ "${app_version}" != "${version}" ]]; then
  echo "ERROR: Chart appVersion (${app_version}) != release version (${version})." >&2
  exit 1
fi

echo "OK: version=${version}, chart.version=${chart_version}, chart.appVersion=${app_version}"
