#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 4 ]]; then
  echo "usage: $0 <release-tag> <assets-dir> <homebrew-repo> <scoop-repo>" >&2
  exit 2
fi

release_tag=$1
assets_dir=$2
homebrew_repo=$3
scoop_repo=$4
release_repository=${RELEASE_REPOSITORY:-dl-alexandre/Grokipedia-CLI}

if [[ ! $release_tag =~ ^v[0-9]+\.[0-9]+\.[0-9]+([+-][0-9A-Za-z.-]+)?$ ]]; then
  echo "invalid release tag: $release_tag" >&2
  exit 2
fi

checksums_file="$assets_dir/checksums.txt"
if [[ ! -f $checksums_file ]]; then
  echo "checksums file not found: $checksums_file" >&2
  exit 1
fi

version=${release_tag#v}

hash_for() {
  local asset_name=$1
  local hash

  if [[ ! -f "$assets_dir/$asset_name" ]]; then
    echo "release asset not found: $assets_dir/$asset_name" >&2
    return 1
  fi

  hash=$(awk -v asset="$asset_name" '$2 == asset { print $1; exit }' "$checksums_file")
  if [[ ! $hash =~ ^[0-9a-fA-F]{64}$ ]]; then
    echo "checksum not found for $asset_name" >&2
    return 1
  fi

  printf '%s' "$hash"
}

darwin_amd64_hash=$(hash_for grokipedia-darwin-amd64.tar.gz)
darwin_arm64_hash=$(hash_for grokipedia-darwin-arm64.tar.gz)
linux_amd64_hash=$(hash_for grokipedia-linux-amd64.tar.gz)
linux_arm64_hash=$(hash_for grokipedia-linux-arm64.tar.gz)
windows_amd64_hash=$(hash_for grokipedia-windows-amd64.zip)

mkdir -p "$homebrew_repo/Formula" "$scoop_repo/bucket"

cat > "$homebrew_repo/Formula/grokipedia.rb" <<EOF
# typed: false
# frozen_string_literal: true

class Grokipedia < Formula
  desc "Unofficial command-line interface for the Grokipedia API"
  homepage "https://github.com/$release_repository"
  version "$version"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/$release_repository/releases/download/$release_tag/grokipedia-darwin-amd64.tar.gz"
      sha256 "$darwin_amd64_hash"

      define_method(:install) do
        bin.install "grokipedia"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/$release_repository/releases/download/$release_tag/grokipedia-darwin-arm64.tar.gz"
      sha256 "$darwin_arm64_hash"

      define_method(:install) do
        bin.install "grokipedia"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel? && Hardware::CPU.is_64_bit?
      url "https://github.com/$release_repository/releases/download/$release_tag/grokipedia-linux-amd64.tar.gz"
      sha256 "$linux_amd64_hash"

      define_method(:install) do
        bin.install "grokipedia"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/$release_repository/releases/download/$release_tag/grokipedia-linux-arm64.tar.gz"
      sha256 "$linux_arm64_hash"

      define_method(:install) do
        bin.install "grokipedia"
      end
    end
  end

  test do
    system "#{bin}/grokipedia", "--help"
  end
end
EOF

cat > "$scoop_repo/bucket/grokipedia.json" <<EOF
{
    "version": "$version",
    "architecture": {
        "64bit": {
            "url": "https://github.com/$release_repository/releases/download/$release_tag/grokipedia-windows-amd64.zip",
            "bin": [
                "grokipedia.exe"
            ],
            "hash": "$windows_amd64_hash"
        }
    },
    "homepage": "https://github.com/$release_repository",
    "license": "MIT",
    "description": "CLI for Wikipedia/Grok integration"
}
EOF

echo "Updated Homebrew and Scoop metadata for $release_tag"
