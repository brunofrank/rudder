#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SILENT=false
ALIAS_NAME=""

# Function to print colored messages
print_message() {
    if [ "$SILENT" = false ]; then
        echo -e "${GREEN}[Rudder]${NC} $1"
    fi
}

print_error() {
    echo -e "${RED}[Error]${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}[Warning]${NC} $1" >&2
}

print_usage() {
    echo "Usage: install.sh [-s|--silent] [-a|--alias <name>]"
    echo "  -s, --silent        Install without prompts, printing only warnings and errors"
    echo "  -a, --alias <name>  Create an alias for rudder without asking"
}

# The alias is written to the shell rc file, so only allow safe names
is_valid_alias() {
    case "$1" in
        ""|-*|*[!A-Za-z0-9_-]*) return 1 ;;
    esac
}

create_alias() {
    if [ -z "$SHELL_RC" ]; then
        print_warning "Could not create alias '$1': unsupported shell"
    elif grep -qsF "alias $1=\"$BIN_DIR/rudder\"" "$SHELL_RC"; then
        print_warning "Alias '$1' already exists"
    elif echo "alias $1=\"$BIN_DIR/rudder\"" >> "$SHELL_RC"; then
        print_message "Alias '$1' created successfully"
        print_message "Please restart your terminal or run 'source $SHELL_RC' to use the alias"
    else
        print_error "Failed to write alias '$1' to $SHELL_RC"
        exit 1
    fi
}

# Parse options
while [ $# -gt 0 ]; do
    case "$1" in
        -s|--silent)
            SILENT=true
            ;;
        -a|--alias|--alias=*)
            case "$1" in
                --alias=*)
                    ALIAS_NAME="${1#--alias=}"
                    ;;
                *)
                    shift
                    ALIAS_NAME="${1:-}"
                    ;;
            esac
            if ! is_valid_alias "$ALIAS_NAME"; then
                print_error "--alias requires a name made of letters, numbers, '_' or '-'"
                exit 1
            fi
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            print_usage >&2
            exit 1
            ;;
    esac
    shift
done

CURL_OPTS="-L"
if [ "$SILENT" = true ]; then
    CURL_OPTS="-sSL"
fi

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

# Map OS to GitHub release asset name
case "$OS" in
    "Darwin")
        if [ "$ARCH" = "arm64" ]; then
            ASSET_NAME="rudder_Darwin_arm64.tar.gz"
        else
            ASSET_NAME="rudder_Darwin_x86_64.tar.gz"
        fi
        ;;
    "Linux")
        if [ "$ARCH" = "aarch64" ]; then
            ASSET_NAME="rudder_Linux_arm64.tar.gz"
        else
            ASSET_NAME="rudder_Linux_x86_64.tar.gz"
        fi
        ;;
    *)
        print_error "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Create installation directory
INSTALL_DIR="$HOME/.rudder"
BIN_DIR="$INSTALL_DIR/bin"

print_message "Creating installation directory at $INSTALL_DIR"
mkdir -p "$BIN_DIR"

# Get the latest version
LATEST_VERSION=$(curl -s https://api.github.com/repos/brunofrank/rudder/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    print_error "Failed to get latest version"
    exit 1
fi

# Download the latest release
LATEST_RELEASE_URL="https://github.com/brunofrank/rudder/releases/latest/download/$ASSET_NAME"
TEMP_FILE="$INSTALL_DIR/temp.tar.gz"

print_message "Downloading Rudder $LATEST_VERSION..."
if ! curl "$CURL_OPTS" "$LATEST_RELEASE_URL" -o "$TEMP_FILE"; then
    print_error "Failed to download Rudder"
    exit 1
fi

# Extract the archive
print_message "Extracting Rudder..."
if ! tar -xzf "$TEMP_FILE" -C "$BIN_DIR"; then
    print_error "Failed to extract Rudder"
    exit 1
fi

# Clean up
rm "$TEMP_FILE"

# Make the binary executable
chmod +x "$BIN_DIR/rudder"

# Save version information
echo "$LATEST_VERSION" > "$INSTALL_DIR/version"

# Add to PATH if not already present
SHELL_RC=""
case "$SHELL" in
    */zsh)
        SHELL_RC="$HOME/.zshrc"
        ;;
    */bash)
        SHELL_RC="$HOME/.bashrc"
        ;;
    *)
        print_warning "Unsupported shell: $SHELL"
        ;;
esac

if [ -n "$SHELL_RC" ]; then
    if ! grep -qs "export PATH=\"\$PATH:$BIN_DIR\"" "$SHELL_RC"; then
        print_message "Adding Rudder to PATH in $SHELL_RC"
        echo "export PATH=\"\$PATH:$BIN_DIR\"" >> "$SHELL_RC"
        print_message "Please restart your terminal or run 'source $SHELL_RC' to update your PATH"
    fi
fi

# Create the alias given by --alias, or ask for one unless running silently
if [ -n "$ALIAS_NAME" ]; then
    create_alias "$ALIAS_NAME"
elif [ "$SILENT" = false ]; then
    print_message "Would you like to create an alias for Rudder? (y/n)"
    read -r wants_alias

    if [ "$wants_alias" = "y" ] || [ "$wants_alias" = "Y" ]; then
        print_message "What alias would you like to use? (e.g., 'rd')"
        read -r alias_name

        if [ -z "$alias_name" ]; then
            print_warning "No alias name provided, skipping alias creation"
        elif ! is_valid_alias "$alias_name"; then
            print_warning "Invalid alias name '$alias_name', skipping alias creation"
        else
            create_alias "$alias_name"
        fi
    fi
fi

print_message "Installation complete! Rudder $LATEST_VERSION has been installed to $BIN_DIR"
print_message "You can now use Rudder by running 'rudder' or your custom alias (after restarting your terminal)"
print_message "To check for updates, run 'rudder --update'"
