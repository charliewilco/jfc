#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-dev}"
dist_dir="${DIST_DIR:-dist}"
host_goos="$(go env GOOS)"
host_goarch="$(go env GOARCH)"
targets=(
	"darwin/amd64"
	"darwin/arm64"
	"linux/amd64"
	"linux/arm64"
	"windows/amd64"
	"windows/arm64"
)

rm -rf "$dist_dir"
mkdir -p "$dist_dir"

checksum_file="$dist_dir/checksums.txt"
: > "$checksum_file"

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1"
	else
		shasum -a 256 "$1"
	fi
}

for target in "${targets[@]}"; do
	goos="${target%/*}"
	goarch="${target#*/}"
	binary="jfc"
	if [ "$goos" = "windows" ]; then
		binary="jfc.exe"
	fi

	package_name="jfc_${version}_${goos}_${goarch}"
	work_dir="$(mktemp -d)"
	package_dir="$work_dir/$package_name"
	mkdir -p "$package_dir/man"

	env GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -trimpath -ldflags "-X github.com/charliewilco/jfc/internal/jfc.Version=$version" -o "$package_dir/$binary" .
	cp man/jfc.1 "$package_dir/man/jfc.1"

	if [ "$goos" = "$host_goos" ] && [ "$goarch" = "$host_goarch" ]; then
		actual_version="$("$package_dir/$binary" --version)"
		if [ "$actual_version" != "$version" ]; then
			echo "expected $binary --version to print $version, got $actual_version" >&2
			exit 1
		fi
	fi

	archive="$dist_dir/$package_name.tar.gz"
	tar -C "$work_dir" -czf "$archive" "$package_name"
	rm -rf "$work_dir"

	sha256 "$archive" >> "$checksum_file"

	archive_listing="$(tar -tzf "$archive")"
	case "$archive_listing" in
		*"$package_name/$binary"*"$package_name/man/jfc.1"*|*"$package_name/man/jfc.1"*"$package_name/$binary"*)
			;;
		*)
			echo "archive $archive does not contain $binary and man/jfc.1" >&2
			exit 1
			;;
	esac
done

expected_count="$((${#targets[@]} + 1))"
actual_count="$(find "$dist_dir" -maxdepth 1 -type f | wc -l | tr -d ' ')"
if [ "$actual_count" != "$expected_count" ]; then
	echo "expected $expected_count release files, found $actual_count" >&2
	find "$dist_dir" -maxdepth 1 -type f >&2
	exit 1
fi

checksum_count="$(wc -l < "$checksum_file" | tr -d ' ')"
if [ "$checksum_count" != "${#targets[@]}" ]; then
	echo "expected ${#targets[@]} checksums, found $checksum_count" >&2
	exit 1
fi

echo "release artifacts verified in $dist_dir"
