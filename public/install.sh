#!/usr/bin/env bash
set -euo pipefail

# force bash for the execution (should avoid issues on codespaces and executing from non-bash shells)
if [ -z "$BASH_VERSION" ]; then
  exec bash "$0" "$@"
fi

# Reset
NoColor=''

# Regular Colors
Red=''
Green=''
Magenta=''
Cyan=''
Yellow=''

# Bold
BoldWhite=''
Bold_Green=''
BoldMagenta=''

if [[ -t 1 ]]; then
    # Reset
    NoColor='\033[0m' # Text Reset

    # Regular Colors
    Red='\033[0;31m'
    Green='\033[0;32m'
    Magenta='\033[35m'
    Cyan='\033[36m'
    Yellow='\033[0;33m'

    BoldMagenta='\033[1;35m'
    BoldWhite='\033[1m'
fi

LOG_LEVEL=${LOG_LEVEL:-INFO}

log() {
    local level=$1
    shift
    local levelColor=$1
    shift
    echo -e "$(date +%Y-%m-%dT%H:%M:%S%z) ${levelColor}${level}${NoColor} $@"
}
debug() {
    if [ "$LOG_LEVEL" = "DEBUG" ]; then
        log 'DBG' $Magenta $@
    fi
}
error() {
    log 'ERR' ${Red} "$@" >&2; exit 1
}
info() {
    log 'INF' $Cyan $@
}
warning() {
    log 'WRN' $Yellow "${BoldWhite}$@ ${NoColor}"
}
success() {
    log 'OK!' $Green $@
}

TARGET_OS=$(uname -s)
TARGET_ARCH=$(uname -m)

# Configuration
install_env=INSTALL_DIR
INSTALL_DIR=${!install_env:-/usr/local/bin}
REPO="dontrm"
TEMP_DIR="$REPO-installation-tmp"
BIN_NAME=$REPO
CURL_FLAGS='-s'

if [ "$LOG_LEVEL" = "DEBUG" ]; then
    CURL_FLAGS=''
fi

command -v $BIN_NAME >/dev/null && warning "$BIN_NAME already installed, script will update"

dependencies=(curl jq tar)
for dependency in "${dependencies[@]}"; do
    command -v $dependency >/dev/null || error "$dependency is required to install $BIN_NAME"
done

debug " - OS: $TARGET_OS"
debug " - ARCH: $TARGET_ARCH"
debug " - curl flags: $CURL_FLAGS"
debug " - Temporal directory: ${BoldWhite}${TEMP_DIR}${NoColor}"
debug " - Install to directory: ${BoldWhite}${INSTALL_DIR}${NoColor}"
debug " - Binary name: ${BoldWhite}${BIN_NAME}${NoColor}"

info "Installation will be available at ${BoldMagenta}$INSTALL_DIR/$BIN_NAME${NoColor}"

mkdir -p $TEMP_DIR
cd $TEMP_DIR
function cleanup {
    info "Tearing down temporary directory"
    cd .. && rm -rf $TEMP_DIR
}
trap cleanup EXIT

info "Fetching ${BoldWhite}latest${NoColor} release information"
response_file=releases-latest.json
curl $CURL_FLAGS "https://api.github.com/repos/Fuabioo/$BIN_NAME/releases/latest" > ${response_file}

tag_name=$(jq -r '.tag_name' < ${response_file})

asset_url=$(jq -r --arg os "$TARGET_OS" --arg arch "$TARGET_ARCH" '
    .assets[]
    | select(.name | test($os; "i") and test($arch; "i"))
    | .url
' < ${response_file})
asset_name=$(jq -r --arg os "$TARGET_OS" --arg arch "$TARGET_ARCH" '
    .assets[]
    | select(.name | test($os; "i") and test($arch; "i"))
    | .name
' < ${response_file})

if [ -z "$asset_name" ] || [ -z "$asset_url" ]; then
    error "Could not find a compatible release asset for $TARGET_OS $TARGET_ARCH."
    exit 1
fi

info "Downloading ${BoldWhite}${asset_name}${NoColor} (${BoldWhite}${tag_name}${NoColor})"
curl $CURL_FLAGS \
    -L \
    -H "Accept: application/octet-stream" \
    -o "${asset_name}" \
    "${asset_url}"

if [ ! -f "${asset_name}" ]; then
    error "Failed to download ${asset_name}"
    exit 1
fi

info "Extracting ${BoldWhite}${asset_name}${NoColor}"
tar -xzf "${asset_name}"

if [ ! -f $BIN_NAME ]; then
    error "Failed to extract ${asset_name}"
    exit 1
fi

info "Setting executable permissions for $BIN_NAME..."
chmod +x $BIN_NAME

warning "Moving $BIN_NAME to $INSTALL_DIR..."
sudo mv $BIN_NAME $INSTALL_DIR/$BIN_NAME

if command -v $BIN_NAME &> /dev/null; then
    success "Installation complete!"
else
    error "Installation failed."
    exit 1
fi
