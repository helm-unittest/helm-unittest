#!/usr/bin/env sh

# borrowed from https://github.com/technosophos/helm-template

PROJECT_NAME="helm-unittest"
# PROJECT_GH can be overridden via HELM_UNITTEST_PROJECT_GH so forks that
# publish their own prebuilt releases (e.g. owner/helm-unittest) are picked
# up instead of upstream.
PROJECT_GH="${HELM_UNITTEST_PROJECT_GH:-helm-unittest/$PROJECT_NAME}"
PROJECT_CHECKSUM_FILE="$PROJECT_NAME-checksum.sha"
HELM_PLUGIN_PATH="$HELM_PLUGIN_DIR"

# Convert the HELM_PLUGIN_PATH to unix if cygpath is
# available. This is the case when using MSYS2 or Cygwin
# on Windows where helm returns a Windows path but we
# need a Unix path
if type cygpath >/dev/null 2>&1; then
  echo Use Sygpath
  HELM_PLUGIN_PATH=$(cygpath -u "$HELM_PLUGIN_PATH")
fi

if [ "$SKIP_BIN_INSTALL" = "1" ]; then
  echo "Skipping binary install"
  exit
fi

if [ "$SKIP_BIN_DOWNLOAD" = "1" ]; then
  echo "Preparing to install into ${HELM_PLUGIN_PATH}"
  cp -f plugin.yaml $HELM_PLUGIN_PATH/plugin.yaml
  cp -f untt $HELM_PLUGIN_PATH/untt
  chmod +x $HELM_PLUGIN_PATH/untt
  echo "$PROJECT_NAME installed into $HELM_PLUGIN_PATH"
  exit
fi

# buildFromSource compiles the plugin binary from the repository checkout
# Helm has cloned into $HELM_PLUGIN_PATH. Used for fork branches / custom
# tags that don't have a matching prebuilt release.
buildFromSource() {
  echo "Building $PROJECT_NAME from source in $HELM_PLUGIN_PATH"
  if ! type "go" >/dev/null 2>&1; then
    echo "Cannot build from source: 'go' is not on PATH."
    echo "Install Go 1.24+ or set HELM_UNITTEST_PROJECT_GH to a fork that publishes releases."
    return 1
  fi
  if [ ! -f "$HELM_PLUGIN_PATH/go.mod" ]; then
    echo "Cannot build from source: no go.mod found in $HELM_PLUGIN_PATH."
    return 1
  fi
  (cd "$HELM_PLUGIN_PATH" && go build -o untt ./cmd/helm-unittest) || return 1
  chmod +x "$HELM_PLUGIN_PATH/untt"
  echo "$PROJECT_NAME built from source into $HELM_PLUGIN_PATH"
  return 0
}

# Explicit opt-in: build straight from source without trying the download
# path. Useful in CI when you know the install must use the checked-out
# branch and not whatever happens to be tagged in plugin.yaml.
if [ "$BUILD_FROM_SOURCE" = "1" ] || [ "$BUILD_FROM_SOURCE" = "true" ]; then
  buildFromSource && exit 0
  exit 1
fi

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case "$ARCH" in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="armv7";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(uname | tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Msys support
    msys*) OS='windows';;
    # Minimalist GNU for Windows
    mingw*) OS='windows';;
    darwin) OS='macos';;
  esac
}

# verifySupported checks that the os/arch combination is supported for
# binary builds.
verifySupported() {
  local supported="linux-amd64\nlinux-arm64\nlinux-s390x\nlinux-ppc64le\nmacos-amd64\nmacos-arm64\nwindows-amd64\nwindows_nt-amd64"
  if ! echo "$supported" | grep -q "$OS-$ARCH"; then
    echo "No prebuild binary for $OS-$ARCH."
    exit 1
  fi

  if ! type "curl" >/dev/null 2>&1 && ! type "wget" >/dev/null 2>&1; then
    echo "Either curl or wget is required"
    exit 1
  fi
  echo "Support $OS-$ARCH"
}

# getDownloadURL checks the latest available version.
getDownloadURL() {
  # Determine last tag based on VCS download
  local version=$(git describe --tags --abbrev=0 >/dev/null 2>&1)
  # If no version found (because of no git), try fetch from plugin
  if [ -z "$version" ]; then
    echo "No version found"
    version=v$(sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)  
  fi

  # Setup Download Url
  DOWNLOAD_URL="https://github.com/$PROJECT_GH/releases/download/${version}/$PROJECT_NAME-$OS-$ARCH-${version#v}.tgz"
  # Setup Checksum Url
  PROJECT_CHECKSUM="https://github.com/$PROJECT_GH/releases/download/${version}/$PROJECT_CHECKSUM_FILE"
}

# downloadFile downloads the latest binary package and also the checksum
# for that binary.
downloadFile() {
  PLUGIN_TMP_FOLDER="/tmp/_dist/"
  [ -d "$PLUGIN_TMP_FOLDER" ] && rm -r "$PLUGIN_TMP_FOLDER" >/dev/null
  mkdir -p "$PLUGIN_TMP_FOLDER"
  echo "Downloading "$DOWNLOAD_URL" to location $PLUGIN_TMP_FOLDER"
  if type "curl" >/dev/null 2>&1; then
      # -f makes curl exit non-zero on HTTP errors (404 etc.) instead of
      # silently writing an HTML error page to disk; the auto-fallback to a
      # source build depends on that.
      (cd "$PLUGIN_TMP_FOLDER" && curl -f -LO "$DOWNLOAD_URL")
  elif type "wget" >/dev/null 2>&1; then
      wget -P "$PLUGIN_TMP_FOLDER" "$DOWNLOAD_URL"
  fi
}

# installFile verifies the SHA256 for the file, then unpacks and
# installs it.
installFile() {
  cd /tmp
  DOWNLOAD_FILE=$(find ./_dist -name "*.tgz")
  echo "found: $DOWNLOAD_FILE"
  if [ -n "$PROJECT_CHECKSUM" ]; then
    echo Validating Checksum.
    if type "curl" >/dev/null 2>&1; then
      if type "shasum" >/dev/null 2>&1; then
        curl -s -L "$PROJECT_CHECKSUM" | grep "$DOWNLOAD_FILE" | shasum -a 256 -c -s
      elif type "sha256sum" >/dev/null 2>&1; then
        if grep -q "ID=alpine" /etc/os-release; then
          curl -s -L "$PROJECT_CHECKSUM" | grep "$DOWNLOAD_FILE" | (sha256sum -c -s || sha256sum -c --status)
        else
          curl -s -L "$PROJECT_CHECKSUM" | grep "$DOWNLOAD_FILE" | sha256sum -c --status
        fi
      else
        echo No Checksum as there is no shasum or sha256sum found.
      fi
    elif type "wget" >/dev/null 2>&1; then
      if type "shasum" >/dev/null 2>&1; then
        wget -q -O - "$PROJECT_CHECKSUM" | grep "$DOWNLOAD_FILE" | shasum -a 256 -c -s
      elif type "sha256sum" >/dev/null 2>&1; then
        if grep -q "ID=alpine" /etc/os-release; then
          wget -q -O - "$PROJECT_CHECKSUM" | grep "$DOWNLOAD_FILE" | (sha256sum -c -s || sha256sum -c --status)
        else
          wget -q -O - "$PROJECT_CHECKSUM" | grep "$DOWNLOAD_FILE" | sha256sum -c --status
        fi
      else
        echo No Checksum as there is no shasum or sha256sum found.
      fi
    fi
  else
    echo No Checksum validated.
  fi
  echo "Preparing to install into ${HELM_PLUGIN_PATH}"
  tar xzf "$DOWNLOAD_FILE" -C "$HELM_PLUGIN_PATH"
  rm -rf "/tmp/_dist"
  echo "$PROJECT_NAME installed into $HELM_PLUGIN_PATH"
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    echo "Failed to install $PROJECT_NAME"
    echo "For support, go to https://github.com/kubernetes/helm"
  fi
  exit $result
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  # To avoid to keep track of the Windows suffix,
  # call the plugin assuming it is in the PATH
  PATH=$PATH:$HELM_PLUGIN_PATH
  untt -h
}

# detectFork returns 0 if the cloned repo's `origin` remote is NOT the
# canonical upstream and the user has not overridden PROJECT_GH. This is the
# `helm plugin install https://github.com/<fork>/helm-unittest.git ...` case:
# downloading a release tarball from upstream would silently install a
# different binary than the user's fork-branch contains, so we build from
# source instead.
detectFork() {
  # If the user explicitly told us where to download from, trust them.
  if [ -n "$HELM_UNITTEST_PROJECT_GH" ]; then
    return 1
  fi
  if [ ! -d "$HELM_PLUGIN_PATH/.git" ]; then
    return 1
  fi
  local remoteUrl
  remoteUrl=$(cd "$HELM_PLUGIN_PATH" && git config --get remote.origin.url 2>/dev/null || echo "")
  case "$remoteUrl" in
    ""|*helm-unittest/helm-unittest*) return 1 ;;
    *) return 0 ;;
  esac
}

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e
initArch
initOS

# Fork installs (where origin points at a non-upstream repo) build from
# source so the user's branch contents are what actually runs.
if detectFork; then
  echo "Detected fork install: building from source so the checked-out branch is used."
  buildFromSource && testVersion && exit 0
  exit 1
fi

# Try the prebuilt-release path first; if any step in it errors, fall back to
# a source build when a Go toolchain is available.
if (set -e; verifySupported && getDownloadURL && downloadFile && installFile); then
  testVersion
else
  echo
  echo "Prebuilt download was unavailable (likely because plugin.yaml's version does not match a published release on $PROJECT_GH)."
  echo "Falling back to building from source."
  buildFromSource && testVersion
fi
