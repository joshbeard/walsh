#!/bin/sh

# Default install directory
INSTALL_DIR="${INSTALL_DIR:-$HOME/bin}"

# Parse arguments
while [ "$#" -gt 0 ]; do
  case "$1" in
    -d|--dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    *)
      echo "Unknown parameter passed: $1"
      exit 1
      ;;
  esac
done

# Create install directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Determine platform and architecture
OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  armv6*|armv7*)
    ARCH="arm"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Determine the latest version
LATEST_TAG=$(curl -s https://api.github.com/repos/joshbeard/walsh/releases/latest | grep 'tag_name' | cut -d\" -f4)
if [ -z "$LATEST_TAG" ]; then
  echo "Unable to determine the latest release."
  exit 1
fi

# Download the package
PACKAGE_URL="https://github.com/joshbeard/walsh/releases/download/$LATEST_TAG/walsh_${LATEST_TAG}_${OS}_${ARCH}.tar.gz"
CHECKSUMS_URL="https://github.com/joshbeard/walsh/releases/download/$LATEST_TAG/checksums.txt"

curl -sLO "$PACKAGE_URL"
curl -sLO "$CHECKSUMS_URL"

# Verify checksum
PACKAGE_FILE="walsh_${LATEST_TAG}_${OS}_${ARCH}.tar.gz"
CHECKSUM=$(grep "$PACKAGE_FILE" checksums.txt | awk '{print $1}')

if ! echo "$CHECKSUM $PACKAGE_FILE" | sha256sum -c -; then
  echo "Checksum verification failed."
  exit 1
fi

# Extract the package
tar -xzf "$PACKAGE_FILE"

# Move the binary to the install directory
mv walsh "$INSTALL_DIR"

# Clean up
rm "$PACKAGE_FILE" checksums.txt

echo "walsh has been installed to $INSTALL_DIR"
