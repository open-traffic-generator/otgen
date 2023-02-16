#!/usr/bin/env bash

# Copyright © 2023 Open Traffic Generator
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The install script is based off of the Apache 2.0 script from Helm,
# https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3

: ${BINARY_NAME:="otgen"}
: ${ORG_URL:="https://github.com/open-traffic-generator"}
: ${REPO_URL:="${ORG_URL}/${BINARY_NAME}"}
: ${USE_SUDO:="true"}
: ${DEBUG:="false"}
: ${VERIFY_CHECKSUM:="true"}
: ${VERIFY_SIGNATURES:="false"}
: ${INSTALL_DIR:="/usr/local/bin"}
: ${GPG_PUBRING:="pubring.kbx"}

HAS_CURL="$(type "curl" &> /dev/null && echo true || echo false)"
HAS_WGET="$(type "wget" &> /dev/null && echo true || echo false)"
HAS_OPENSSL="$(type "openssl" &> /dev/null && echo true || echo false)"
HAS_GPG="$(type "gpg" &> /dev/null && echo true || echo false)"
HAS_GIT="$(type "git" &> /dev/null && echo true || echo false)"

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="i386";;
    x86_64) ARCH="x86_64";;
    i686) ARCH="i386";;
    i386) ARCH="i386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(uname)

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*|cygwin*) OS='Windows';;
  esac
}

# runs the given command as root (detects if we are root already)
runAsRoot() {
  if [ $EUID -ne 0 -a "$USE_SUDO" = "true" ]; then
    sudo "${@}"
  else
    "${@}"
  fi
}

# verifySupported checks that the os/arch combination is supported for
# binary builds, as well whether or not necessary tools are present.
verifySupported() {
  local supported="Darwin_x86_64\nDarwin_arm64\nLinux_x86_64\nLinux_arm64"
  if ! echo "${supported}" | grep -q "${OS}_${ARCH}"; then
    echo "No prebuilt binary for ${OS}_${ARCH}."
    echo "To build from source, go to ${REPO_URL}"
    exit 1
  fi

  if [ "${HAS_CURL}" != "true" ] && [ "${HAS_WGET}" != "true" ]; then
    echo "Either curl or wget is required"
    exit 1
  fi

  if [ "${VERIFY_CHECKSUM}" == "true" ] && [ "${HAS_OPENSSL}" != "true" ]; then
    echo "In order to verify checksum, openssl must first be installed."
    echo "Please install openssl or set VERIFY_CHECKSUM=false in your environment."
    exit 1
  fi

  if [ "${VERIFY_SIGNATURES}" == "true" ]; then
    if [ "${HAS_GPG}" != "true" ]; then
      echo "In order to verify signatures, gpg must first be installed."
      echo "Please install gpg or set VERIFY_SIGNATURES=false in your environment."
      exit 1
    fi
    if [ "${OS}" != "linux" ]; then
      echo "Signature verification is currently only supported on Linux."
      echo "Please set VERIFY_SIGNATURES=false or verify the signatures manually."
      exit 1
    fi
  fi

  if [ "${HAS_GIT}" != "true" ]; then
    echo "[WARNING] Could not find git. It is required for plugin installation."
  fi
}

# checkDesiredVersion checks if the desired version is available.
checkDesiredVersion() {
  if [ "x$DESIRED_VERSION" == "x" ]; then
    # Get tag from release URL
    local latest_release_url="${REPO_URL}/releases"
    if [ "${HAS_CURL}" == "true" ]; then
      TAG=$(curl -Ls $latest_release_url | grep 'href="/open-traffic-generator/otgen/releases/tag/v[0-9]*.[0-9]*.[0-9]*\"' | sed -E 's/.*\/open-traffic-generator\/otgen\/releases\/tag\/(v[0-9\.]+)".*/\1/g' | head -1)
    elif [ "${HAS_WGET}" == "true" ]; then
      TAG=$(wget $latest_release_url -O - 2>&1 | grep 'href="/open-traffic-generator/otgen/releases/tag/v[0-9]*.[0-9]*.[0-9]*\"' | sed -E 's/.*\/open-traffic-generator\/otgen\/releases\/tag\/(v[0-9\.]+)".*/\1/g' | head -1)
    fi
  else
    TAG="v$(echo ${DESIRED_VERSION} | sed 's/^v//')"
  fi
  VERSION="$(echo $TAG | sed 's/^v//')"
}

# checkInstalledVersion checks which version of the binary is installed and
# if it needs to be changed.
checkInstalledVersion() {
  if [[ -f "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
    local ver=$("${INSTALL_DIR}/${BINARY_NAME}" version | grep version | awk '{print $2}')
    if [[ "$ver" == "$VERSION" ]]; then
      echo "${BINARY_NAME} ${ver} is already ${DESIRED_VERSION:-latest}"
      return 0
    else
      echo "${BINARY_NAME} ${VERSION} is available. Changing from version ${ver}."
      return 1
    fi
  else
    return 1
  fi
}

# downloadFile downloads the latest binary package and also the checksum
# for that binary.
downloadFile() {
  DIST="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
  DOWNLOAD_URL="${REPO_URL}/releases/download/${TAG}/${DIST}"
  CHECKSUM_URL="${REPO_URL}/releases/download/${TAG}/checksums.txt"
  TMP_ROOT="$(mktemp -dt ${BINARY_NAME}-installer-XXXXXX)"
  TMP_FILE="$TMP_ROOT/$DIST"
  SUM_FILE="$TMP_ROOT/checksums.txt"
  echo "Downloading $DOWNLOAD_URL"
  if [ "${HAS_CURL}" == "true" ]; then
    curl -SsL "$CHECKSUM_URL" -o "$SUM_FILE"
    curl -SsL "$DOWNLOAD_URL" -o "$TMP_FILE"
  elif [ "${HAS_WGET}" == "true" ]; then
    wget -q -O "$SUM_FILE" "$CHECKSUM_URL"
    wget -q -O "$TMP_FILE" "$DOWNLOAD_URL"
  fi
}

# verifyFile verifies the SHA256 checksum of the binary package
# and the GPG signatures for both the package and checksum file
# (depending on settings in environment).
verifyFile() {
  if [ "${VERIFY_CHECKSUM}" == "true" ]; then
    verifyChecksum
  fi
  if [ "${VERIFY_SIGNATURES}" == "true" ]; then
    verifySignatures
  fi
}

# installFile installs the ${BINARY_NAME} binary.
installFile() {
  TMP_INSTALL_DIR="$TMP_ROOT/$BINARY_NAME"
  mkdir -p "$TMP_INSTALL_DIR"
  tar xf "$TMP_FILE" -C "$TMP_INSTALL_DIR"
  TMP_BIN="$TMP_INSTALL_DIR/${BINARY_NAME}"
  echo "Preparing to install $BINARY_NAME into ${INSTALL_DIR}"
  runAsRoot mkdir -p "${INSTALL_DIR}"
  runAsRoot cp "$TMP_BIN" "$INSTALL_DIR/$BINARY_NAME"
  echo "$BINARY_NAME installed into $INSTALL_DIR/$BINARY_NAME"
}

# verifyChecksum verifies the SHA256 checksum of the binary package.
verifyChecksum() {
  printf "Verifying checksum... "
  local sum=$(openssl sha1 -sha256 ${TMP_FILE} | awk '{print $2}')
  local expected_sum=$(cat ${SUM_FILE} | grep ${DIST} | awk '{print $1}')
  if [ "$sum" != "$expected_sum" ]; then
    echo "SHA sum of ${TMP_FILE} does not match. Aborting."
    exit 1
  fi
  echo "Done."
}

# verifySignatures obtains the latest KEYS file from GitHub main branch
# as well as the signature .asc files from the specific GitHub release,
# then verifies that the release artifacts were signed by a maintainer's key.
verifySignatures() {
  printf "Verifying signatures... "
  local keys_filename="KEYS"
  local github_keys_url="https://raw.githubusercontent.com/open-traffic-generator/otgen/main/${keys_filename}"
  if [ "${HAS_CURL}" == "true" ]; then
    curl -SsL "${github_keys_url}" -o "${TMP_ROOT}/${keys_filename}"
  elif [ "${HAS_WGET}" == "true" ]; then
    wget -q -O "${TMP_ROOT}/${keys_filename}" "${github_keys_url}"
  fi
  local gpg_keyring="${TMP_ROOT}/keyring.gpg"
  local gpg_homedir="${TMP_ROOT}/gnupg"
  mkdir -p -m 0700 "${gpg_homedir}"
  local gpg_stderr_device="/dev/null"
  if [ "${DEBUG}" == "true" ]; then
    gpg_stderr_device="/dev/stderr"
  fi
  gpg --batch --quiet --homedir="${gpg_homedir}" --import "${TMP_ROOT}/${keys_filename}" 2> "${gpg_stderr_device}"
  gpg --batch --no-default-keyring --keyring "${gpg_homedir}/${GPG_PUBRING}" --export > "${gpg_keyring}"
  local github_release_url="${REPO_URL}/releases/download/${TAG}"
  if [ "${HAS_CURL}" == "true" ]; then
    curl -SsL "${github_release_url}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc" -o "${TMP_ROOT}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc"
    curl -SsL "${github_release_url}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.asc" -o "${TMP_ROOT}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.asc"
  elif [ "${HAS_WGET}" == "true" ]; then
    wget -q -O "${TMP_ROOT}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc" "${github_release_url}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc"
    wget -q -O "${TMP_ROOT}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.asc" "${github_release_url}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.asc"
  fi
  local error_text="If you think this might be a potential security issue,"
  error_text="${error_text}\nplease see here: ${ORG_URL}/community/blob/main/SECURITY.md"
  local num_goodlines_sha=$(gpg --verify --keyring="${gpg_keyring}" --status-fd=1 "${TMP_ROOT}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc" 2> "${gpg_stderr_device}" | grep -c -E '^\[GNUPG:\] (GOODSIG|VALIDSIG)')
  if [[ ${num_goodlines_sha} -lt 2 ]]; then
    echo "Unable to verify the signature of ${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.sha256!"
    echo -e "${error_text}"
    exit 1
  fi
  local num_goodlines_tar=$(gpg --verify --keyring="${gpg_keyring}" --status-fd=1 "${TMP_ROOT}/${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz.asc" 2> "${gpg_stderr_device}" | grep -c -E '^\[GNUPG:\] (GOODSIG|VALIDSIG)')
  if [[ ${num_goodlines_tar} -lt 2 ]]; then
    echo "Unable to verify the signature of ${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz!"
    echo -e "${error_text}"
    exit 1
  fi
  echo "Done."
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    if [[ -n "$INPUT_ARGUMENTS" ]]; then
      echo "Failed to install $BINARY_NAME with the arguments provided: $INPUT_ARGUMENTS"
      help
    else
      echo "Failed to install $BINARY_NAME"
    fi
    echo -e "\tFor support, go to ${REPO_URL}."
  fi
  cleanup
  exit $result
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  set +e
  BIN_FULL_NAME="$(command -v $BINARY_NAME)"
  if [ "$?" = "1" ]; then
    echo "$BINARY_NAME not found. Is $INSTALL_DIR on your "'$PATH?'
    exit 1
  fi
  set -e
}

# help provides possible cli installation arguments
help () {
  echo "Accepted cli arguments are:"
  echo -e "\t[--help|-h ] ->> prints this help"
  echo -e "\t[--version|-v <desired_version>] . When not defined it fetches the latest release from GitHub"
  echo -e "\te.g. --version 0.4.0"
  echo -e "\t[--no-sudo]  ->> install without sudo"
}

# cleanup temporary files
cleanup() {
  if [[ -d "${TMP_ROOT:-}" ]]; then
    rm -rf "$TMP_ROOT"
  fi
}

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e

# Set debug if desired
if [ "${DEBUG}" == "true" ]; then
  set -x
fi

# Parsing input arguments (if any)
export INPUT_ARGUMENTS="${@}"
set -u
while [[ $# -gt 0 ]]; do
  case $1 in
    '--version'|-v)
       shift
       if [[ $# -ne 0 ]]; then
           export DESIRED_VERSION="${1}"
       else
           echo -e "Please provide the desired version. e.g. --version 0.4.0"
           exit 0
       fi
       ;;
    '--no-sudo')
       USE_SUDO="false"
       ;;
    '--help'|-h)
       help
       exit 0
       ;;
    *) exit 1
       ;;
  esac
  shift
done
set +u

initArch
initOS
verifySupported
checkDesiredVersion
if ! checkInstalledVersion; then
  downloadFile
  verifyFile
  installFile
fi
testVersion
cleanup
