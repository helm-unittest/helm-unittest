#!/usr/bin/env sh

# borrowed from https://github.com/technosophos/helm-template

PROJECT_NAME="helm-unittest"
PROJECT_GH="quintush/$PROJECT_NAME"
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
  local supported="linux-arm64\nlinux-amd64\nmacos-amd64\nwindows-amd64\nmacos-arm64"
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
  # Use the GitHub API to find the latest version for this project.
  local latest_url="https://api.github.com/repos/$PROJECT_GH/releases/latest"
  if [ -z "$HELM_PLUGIN_UPDATE" ]; then
    local version=$(git describe --tags --exact-match 2>/dev/null)
    if [ -n "$version" ]; then
      latest_url="https://api.github.com/repos/$PROJECT_GH/releases/tags/$version"
    fi
  fi
  echo "Retrieving $latest_url"
  if type "curl" >/dev/null 2>&1; then
    DOWNLOAD_URL=$(curl -sL "$latest_url" | grep "$OS-$ARCH" | awk '/\"browser_download_url\":/{gsub(/[,\"]/,"", $2); print $2}' 2>/dev/null)
    # Backward compatibility when arch type is not yet used.
    if [ -z "$DOWNLOAD_URL" ]; then
      echo "No download_url found only searching for $OS"
      DOWNLOAD_URL=$(curl -sL "$latest_url" | grep "$OS" | awk '/\"browser_download_url\":/{gsub(/[,\"]/,"", $2); print $2}' 2>/dev/null)
    fi
    PROJECT_CHECKSUM=$(curl -sL "$latest_url" | grep "checksum" | awk '/\"browser_download_url\":/{gsub(/[,\"]/,"", $2); print $2}' 2>/dev/null)
  elif type "wget" >/dev/null 2>&1; then
    DOWNLOAD_URL=$(wget -q -O - "$latest_url" | grep "$OS-$ARCH" | awk '/\"browser_download_url\":/{gsub(/[,\"]/,"", $2); print $2}' 2>/dev/null)
    # Backward compatibility when arch type is not yet used.
    if [ -z "$DOWNLOAD_URL" ]; then
      echo "No download_url found only searching for $OS"
      DOWNLOAD_URL=$(wget -q -O - "$latest_url" | grep "$OS" | awk '/\"browser_download_url\":/{gsub(/[,\"]/,"", $2); print $2}' 2>/dev/null)
    fi
    PROJECT_CHECKSUM=$(wget -q -O - "$latest_url" | grep "checksum" | awk '/\"browser_download_url\":/{gsub(/[,\"]/,"", $2); print $2}' 2>/dev/null)
  fi
}

# downloadFile downloads the latest binary package and also the checksum
# for that binary.
downloadFile() {
  PLUGIN_TMP_FOLDER="/tmp/_dist/"
  [ -d "$PLUGIN_TMP_FOLDER" ] && rm -r "$PLUGIN_TMP_FOLDER" >/dev/null
  mkdir -p "$PLUGIN_TMP_FOLDER"
  echo "Downloading "$DOWNLOAD_URL" to location $PLUGIN_TMP_FOLDER"
  if type "curl" >/dev/null 2>&1; then
      (cd "$PLUGIN_TMP_FOLDER" && curl -LO "$DOWNLOAD_URL")
  elif type "wget" >/dev/null 2>&1; then
      wget -P "$PLUGIN_TMP_FOLDER" "$DOWNLOAD_URL"
  fi
}

# installFile verifies the SHA256 for the file, then unpacks and
# installs it.
installFile() {
  cd "/tmp"
  DOWNLOAD_FILE=$(find ./_dist -name "*.tgz")
  if [ -n "$PROJECT_CHECKSUM" ]; then
    echo Validating Checksum.
    if type "curl" >/dev/null 2>&1; then
      if type "shasum" >/dev/null 2>&1; then
        curl -s -L "$PROJECT_CHECKSUM" | grep "$DOWNLOAD_FILE" | shasum -a 256 -c -s
      elif type "sha256sum" >/dev/null 2>&1; then
        if grep -q "ID=alpine" /etc/os-release; then
          curl -s -L "$PROJECT_CHECKSUM" | grep "$DOWNLOAD_FILE" | sha256sum -c -s
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
          wget -q -O - "$PROJECT_CHECKSUM" | grep "$DOWNLOAD_FILE" | sha256sum -c -s
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
  HELM_TMP="/tmp/$PROJECT_NAME"
  mkdir -p "$HELM_TMP"
  tar xf "$DOWNLOAD_FILE" -C "$HELM_TMP"
  HELM_TMP_BIN="$HELM_TMP/untt"
  echo "Preparing to install into ${HELM_PLUGIN_PATH}"
  # Use * to also copy the file with the exe suffix on Windows
  cp "$HELM_TMP_BIN"* "$HELM_PLUGIN_PATH"
  rm -r "$HELM_TMP"
  rm -r "$PLUGIN_TMP_FOLDER"
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

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e
initArch
initOS
verifySupported
getDownloadURL
downloadFile
installFile
testVersion
