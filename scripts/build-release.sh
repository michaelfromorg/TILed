#!/usr/bin/env bash

set -euo pipefail

script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "${script_directory}/.." && pwd)"
cd "${project_root}"

version="${VERSION:-}"
if [[ -z "${version}" ]]; then
	version="$(git describe --tags --exact-match 2>/dev/null || true)"
fi
if [[ ! "${version}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z][0-9A-Za-z.-]*)?$ ]]; then
	echo "VERSION must be a semantic release tag such as v1.1.0" >&2
	exit 1
fi

dist_root="${DIST_DIR:-dist}"
case "${dist_root}" in
	"" | "/" | "." | "..")
		echo "Refusing unsafe DIST_DIR: ${dist_root}" >&2
		exit 1
		;;
esac

rm -rf -- "${dist_root}"
mkdir -p "${dist_root}"
dist_directory="$(cd "${dist_root}" && pwd)"
work_directory="$(mktemp -d)"
trap 'rm -rf -- "${work_directory}"' EXIT

if ! command -v zip >/dev/null 2>&1; then
	echo "zip is required to package Windows releases" >&2
	exit 1
fi

targets=(
	"darwin/amd64"
	"darwin/arm64"
	"linux/amd64"
	"linux/arm64"
	"windows/amd64"
	"windows/arm64"
)
ldflags="-s -w -X github.com/michaelfromorg/tiled/cmd.Version=${version}"

for target in "${targets[@]}"; do
	goos="${target%/*}"
	goarch="${target#*/}"
	archive_base="til_${version}_${goos}_${goarch}"
	package_directory="${work_directory}/${archive_base}"
	mkdir -p "${package_directory}"
	cp LICENSE "${package_directory}/LICENSE"

	binary_name="til"
	if [[ "${goos}" == "windows" ]]; then
		binary_name="til.exe"
	fi

	echo "Building ${goos}/${goarch}"
	CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" \
		go build \
		-trimpath \
		-buildvcs=false \
		-ldflags "${ldflags}" \
		-o "${package_directory}/${binary_name}" \
		./cmd/til

	if [[ "${goos}" == "windows" ]]; then
		(
			cd "${package_directory}"
			zip -q -X "${dist_directory}/${archive_base}.zip" "${binary_name}" LICENSE
		)
	else
		COPYFILE_DISABLE=1 tar \
			-C "${package_directory}" \
			-czf "${dist_directory}/${archive_base}.tar.gz" \
			"${binary_name}" \
			LICENSE
	fi
done

(
	cd "${dist_directory}"
	artifacts=(til_*.tar.gz til_*.zip)
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "${artifacts[@]}" > checksums.txt
	else
		shasum -a 256 "${artifacts[@]}" > checksums.txt
	fi
)

echo "Release artifacts:"
ls -1 "${dist_directory}"
