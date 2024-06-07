#!/bin/bash
CURRENT_DIR=$(pwd)

if [[ $STRING == *"/scripts" ]]; then
  cd ..
fi

NAME=parallels
BINARY=packer-plugin-${NAME}
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')
PLUGIN_FQN=$(grep -E '^module' <go.mod | sed -E 's/module *//' | tr '[:upper:]' '[:lower:]' )

PLUGIN_PATH=$(echo "${PLUGIN_FQN}" | sed 's/packer-plugin-//')
PACKER_FOLDER=$(dirname $(which packer))
PACKER_PACKAGE_FOLDER=$(echo ${PACKER_FOLDER}/${PLUGIN_PATH})
VERSION=$(grep -E '^\tVersion = ' <./version/version.go | sed -E 's/\tVersion = *//' | tr -d \")
FILENAME="${BINARY}_v${VERSION}_x5.0_${OS}_${ARCH}"
FILENAME_SHA256SUM="${FILENAME}_SHA256SUM"
FULL_FILENAME_PATH="${PACKER_PACKAGE_FOLDER}/${FILENAME}"
FULL_FILENAME_SHA256SUM_PATH="${PACKER_PACKAGE_FOLDER}/${FILENAME_SHA256SUM}"
BINARY_CHECKSUM=$(sha256sum "$BINARY" | cut -d ' ' -f 1)

# Calculate file SHA256 checksum

echo "Building ${BINARY} for ${OS}..."
go mod tidy
go build -ldflags="-X ${PLUGIN_FQN}/version.VersionPrerelease=dev" -o ${BINARY}

echo "Installing ${BINARY}..."
if [ ! -d "$PACKER_PACKAGE_FOLDER" ]; then
  echo "Creating folder $PACKER_PACKAGE_FOLDER"
  mkdir -p "$PACKER_PACKAGE_FOLDER"
else
  echo "Folder $PACKER_PACKAGE_FOLDER already exists, removing old binary"
  rm -f "$PACKER_PACKAGE_FOLDER"/*
fi

cp $BINARY "$PACKER_PACKAGE_FOLDER/$FILENAME"
touch "$FULL_FILENAME_SHA256SUM_PATH"
echo "$BINARY_CHECKSUM" > "$FULL_FILENAME_SHA256SUM_PATH"

echo "Checking if the binary was installed successfully..."
RESULT=$(packer plugins installed | grep -c "${FULL_FILENAME_PATH}")
if [ "$RESULT" -ne 1 ]; then
  echo "Binary installation failed. Exiting..."
  exit 1
fi

WHEREAMI=$(pwd)
if [ "$WHEREAMI" != "$CURRENT_DIR" ]; then
  cd "$CURRENT_DIR" || exit 1
fi
