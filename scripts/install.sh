#!/bin/sh
# pandaprobe CLI installer for macOS and Linux.
#
#   curl -fsSL https://cli.pandaprobe.com/install.sh | sh
#
# Environment overrides:
#   PANDAPROBE_VERSION      install a specific version (e.g. v0.2.0); default: latest
#   PANDAPROBE_INSTALL_DIR  install location; default: /usr/local/bin (falls back to
#                           $HOME/.local/bin when /usr/local/bin is not writable)
#   PANDAPROBE_BASE_URL     release download root for mirrors/testing;
#                           default: https://github.com/<repo>/releases/download
#
# The same can be passed as the first positional argument: install.sh v0.2.0
set -eu

REPO="chirpz-ai/pandaprobe-cli"
PROJECT="pandaprobe-cli"
BINARY="pandaprobe"

info()  { printf '==> %s\n' "$1"; }
warn()  { printf 'warning: %s\n' "$1" >&2; }
error() { printf 'error: %s\n' "$1" >&2; exit 1; }

# --- detect download tool ---
if command -v curl >/dev/null 2>&1; then
	dl() { curl -fsSL "$1" -o "$2"; }
	dl_stdout() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
	dl() { wget -qO "$2" "$1"; }
	dl_stdout() { wget -qO - "$1"; }
else
	error "curl or wget is required to install pandaprobe"
fi

# --- detect platform ---
os=$(uname -s)
case "$os" in
	Linux)  os="linux" ;;
	Darwin) os="darwin" ;;
	*)      error "unsupported OS: $os (only Linux and macOS are supported; on Windows use install.ps1)" ;;
esac

arch=$(uname -m)
case "$arch" in
	x86_64 | amd64)  arch="amd64" ;;
	aarch64 | arm64) arch="arm64" ;;
	*)               error "unsupported architecture: $arch" ;;
esac

# --- resolve version ---
version="${1:-${PANDAPROBE_VERSION:-latest}}"
if [ "$version" = "latest" ]; then
	info "Resolving latest release"
	tag=$(dl_stdout "https://api.github.com/repos/${REPO}/releases/latest" \
		| grep '"tag_name"' | head -n1 | sed 's/.*"tag_name"[^"]*"\([^"]*\)".*/\1/')
	[ -n "$tag" ] || error "could not determine the latest release; set PANDAPROBE_VERSION explicitly"
else
	tag="$version"
fi
# Archive names use the version without a leading "v".
ver_no_v=$(printf '%s' "$tag" | sed 's/^v//')

asset="${PROJECT}_${ver_no_v}_${os}_${arch}.tar.gz"
base_url="${PANDAPROBE_BASE_URL:-https://github.com/${REPO}/releases/download}/${tag}"

# --- download ---
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

info "Downloading ${asset} (${tag})"
dl "${base_url}/${asset}" "${tmp}/${asset}" \
	|| error "download failed for ${base_url}/${asset}"

# --- verify checksum (best effort) ---
if dl "${base_url}/checksums.txt" "${tmp}/checksums.txt" 2>/dev/null; then
	expected=$(grep " ${asset}\$" "${tmp}/checksums.txt" | awk '{print $1}')
	if [ -n "$expected" ]; then
		if command -v sha256sum >/dev/null 2>&1; then
			actual=$(sha256sum "${tmp}/${asset}" | awk '{print $1}')
		elif command -v shasum >/dev/null 2>&1; then
			actual=$(shasum -a 256 "${tmp}/${asset}" | awk '{print $1}')
		else
			actual=""
			warn "no sha256 tool found; skipping checksum verification"
		fi
		if [ -n "$actual" ] && [ "$actual" != "$expected" ]; then
			error "checksum mismatch for ${asset} (expected ${expected}, got ${actual})"
		fi
		[ -n "$actual" ] && info "Checksum verified"
	fi
else
	warn "checksums.txt not found; skipping verification"
fi

# --- extract ---
tar -xzf "${tmp}/${asset}" -C "$tmp" || error "failed to extract ${asset}"
[ -f "${tmp}/${BINARY}" ] || error "binary ${BINARY} not found in archive"
chmod +x "${tmp}/${BINARY}"

# --- choose install dir ---
install_dir="${PANDAPROBE_INSTALL_DIR:-/usr/local/bin}"
sudo=""
if [ ! -d "$install_dir" ] || [ ! -w "$install_dir" ]; then
	if [ "$install_dir" = "/usr/local/bin" ]; then
		if command -v sudo >/dev/null 2>&1 && [ -d "$install_dir" ]; then
			sudo="sudo"
		else
			install_dir="${HOME}/.local/bin"
		fi
	fi
fi
mkdir -p "$install_dir" 2>/dev/null || $sudo mkdir -p "$install_dir"

info "Installing to ${install_dir}/${BINARY}"
$sudo mv "${tmp}/${BINARY}" "${install_dir}/${BINARY}"

# --- report ---
installed_version=$("${install_dir}/${BINARY}" version 2>/dev/null | grep -o '"version"[^,]*' || true)
info "Installed pandaprobe (${tag}) to ${install_dir}/${BINARY}"
[ -n "$installed_version" ] && info "${installed_version}"

case ":${PATH}:" in
	*":${install_dir}:"*) ;;
	*)
		printf '\n'
		warn "${install_dir} is not on your PATH."
		printf 'Add it by appending this line to your shell profile:\n\n    export PATH="%s:$PATH"\n\n' "$install_dir"
		;;
esac

printf 'Run `pandaprobe --help` to get started.\n'
