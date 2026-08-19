#!/usr/bin/env bash

set -euo pipefail

version="${1:-}"
output="${2:-}"
repository="${REPOSITORY:-michaelfromorg/TILed}"

if [[ ! "${version}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z][0-9A-Za-z.-]*)?$ ]]; then
	echo "usage: $0 <version-tag> <output-path>" >&2
	exit 1
fi
if [[ -z "${output}" ]]; then
	echo "usage: $0 <version-tag> <output-path>" >&2
	exit 1
fi

release_base="https://github.com/${repository}/releases/download/${version}"
temporary_directory="$(mktemp -d)"
trap 'rm -rf -- "${temporary_directory}"' EXIT
checksums_file="${temporary_directory}/checksums.txt"
curl --fail --location --silent --show-error \
	"${release_base}/checksums.txt" \
	-o "${checksums_file}"

checksum_for() {
	local filename="$1"
	local checksum=""
	local candidate=""
	local candidate_checksum=""

	while read -r candidate_checksum candidate; do
		if [[ "${candidate}" == "${filename}" ]]; then
			checksum="${candidate_checksum}"
			break
		fi
	done < "${checksums_file}"

	if [[ ! "${checksum}" =~ ^[0-9a-f]{64}$ ]]; then
		echo "No valid checksum found for ${filename}" >&2
		exit 1
	fi
	printf "%s" "${checksum}"
}

darwin_amd64="til_${version}_darwin_amd64.tar.gz"
darwin_arm64="til_${version}_darwin_arm64.tar.gz"
linux_amd64="til_${version}_linux_amd64.tar.gz"
linux_arm64="til_${version}_linux_arm64.tar.gz"

darwin_amd64_sha="$(checksum_for "${darwin_amd64}")"
darwin_arm64_sha="$(checksum_for "${darwin_arm64}")"
linux_amd64_sha="$(checksum_for "${linux_amd64}")"
linux_arm64_sha="$(checksum_for "${linux_arm64}")"

mkdir -p "$(dirname "${output}")"
cat > "${output}" <<EOF
class Til < Formula
  desc "Keep a lightweight, append-first log of what you learn"
  homepage "https://github.com/${repository}"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "${release_base}/${darwin_arm64}"
      sha256 "${darwin_arm64_sha}"
    else
      url "${release_base}/${darwin_amd64}"
      sha256 "${darwin_amd64_sha}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "${release_base}/${linux_arm64}"
      sha256 "${linux_arm64_sha}"
    else
      url "${release_base}/${linux_amd64}"
      sha256 "${linux_amd64_sha}"
    end
  end

  def install
    bin.install "til"
  end

  test do
    assert_match "til v#{version}", shell_output("#{bin}/til version")
  end
end
EOF

echo "Wrote ${output}"
