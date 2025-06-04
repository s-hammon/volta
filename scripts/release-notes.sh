#!/bin/bash

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: $0 <semver version> [<changelog file>]"
    exit 1
fi

version="$1"
clog_file="${2:-CHANGELOG.md}"

if [ ! -f "$clog_file" ]; then
    echo "Changelog file not found: $clog_file"
    exit 1
fi

CAPTURE=0
items=""

while IFS= read -r LINE; do
    if [[ "${LINE}" == "##"* ]] && [[ "${CAPTURE}" -eq 1 ]]; then
        break
    fi
    if [[ "${LINE}" == "[Unreleased]"* ]]; then
        break
    fi
    if [[ "${LINE}" == "## [${version}]"* ]] && [[ "${CAPTURE}" -eq 0 ]]; then
        CAPTURE=1
        continue
    fi
    if [[ "${CAPTURE}" -eq 1 ]]; then
        if [[ -z "${LINE}" ]]; then
            continue
        fi
        items+="$(echo "${LINE}" | xargs -0)"
        if [[ -n "$items" ]]; then
            items+=$'\n'
        fi
    fi
done <"${clog_file}"

if [[ -n "$items" ]]; then
    echo "${items%$'\n'}"
else
    echo "No changes found for version ${version}"
fi
