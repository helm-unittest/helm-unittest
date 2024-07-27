#!/usr/bin/env sh

# borrowed from https://github.com/technosophos/helm-template

PROJECT_NAME="helm-unittest"
PROJECT_GH="helm-unittest/$PROJECT_NAME"
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
  local supported="linux-arm64\nlinux-amd64\nmacos-amd64\nwindows-amd64\nwindows_nt-amd64\nmacos-arm64"
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
      (cd "$PLUGIN_TMP_FOLDER" && curl -LO "$DOWNLOAD_URL")
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
