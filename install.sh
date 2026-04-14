#!/bin/sh
# SwiftTalon Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/Bhuw1234/swifttalon/main/install.sh | sh
#
# Environment Variables:
#   SWIFTTALON_VERSION        - Version to install (default: latest)
#   SWIFTTALON_INSTALL_DIR    - Installation directory (default: /usr/local/bin)
#   SWIFTTALON_INSTALL_METHOD - Installation method: binary|source|docker (default: binary)
#   SWIFTTALON_NO_INIT        - If set, skip 'swifttalon init' after installation
#   SWIFTTALON_QUICK_START    - If set, auto-run init and show sample command

set -eu

# Version and installation directory can be overridden with environment variables
SWIFTTALON_VERSION="${SWIFTTALON_VERSION:-latest}"
SWIFTTALON_INSTALL_DIR="${SWIFTTALON_INSTALL_DIR:-/usr/local/bin}"
SWIFTTALON_INSTALL_METHOD="${SWIFTTALON_INSTALL_METHOD:-binary}"
SWIFTTALON_QUICK_START="${SWIFTTALON_QUICK_START:-}"

# Repository configuration
GITHUB_REPO="Bhuw1234/swifttalon"
BINARY_NAME="swifttalon"

# Minimum expected binary size (100KB) - helps detect corrupted downloads
MIN_BINARY_SIZE=102400

# Terminal colors using tput with fallbacks
RED=""
GREEN=""
YELLOW=""
BLUE=""
CYAN=""
MAGENTA=""
BOLD=""
RESET=""

# Check if stdout is a terminal
if [ -t 1 ]; then
    # Check if tput is available
    if command -v tput >/dev/null 2>&1; then
        ncolors=$(tput colors 2>/dev/null || echo 0)
        if [ -n "$ncolors" ] && [ "$ncolors" -ge 8 ]; then
            RED=$(tput setaf 1)
            GREEN=$(tput setaf 2)
            YELLOW=$(tput setaf 3)
            BLUE=$(tput setaf 4)
            CYAN=$(tput setaf 6)
            MAGENTA=$(tput setaf 5)
            BOLD=$(tput bold)
            RESET=$(tput sgr0)
        fi
    fi
fi

# Cleanup handler
TEMP_DIR=""
INSTALLED_BINARY=""
cleanup() {
    # If installation failed and we installed a binary, remove it
    if [ -n "$INSTALLED_BINARY" ] && [ -f "$INSTALLED_BINARY" ]; then
        rm -f "$INSTALLED_BINARY" 2>/dev/null || true
    fi
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Print functions with Ollama-style >>> prefix
status() {
    printf "%s>>> %s%s\n" "$BLUE" "$1" "$RESET"
}

success() {
    printf "%s>>> %s%s\n" "$GREEN" "$1" "$RESET"
}

warn() {
    printf "%s>>> Warning: %s%s\n" "$YELLOW" "$1" "$RESET" >&2
}

error() {
    printf "%s>>> Error: %s%s\n" "$RED" "$1" "$RESET" >&2
}

info() {
    printf "    %s\n" "$1"
}

# Check if we're running as root
check_root() {
    if [ "$(id -u)" -eq 0 ]; then
        return 0
    fi
    return 1
}

# Check if sudo is available and passwordless
has_sudo() {
    if command -v sudo >/dev/null 2>&1; then
        # Check if we can run sudo without password
        if sudo -n true 2>/dev/null; then
            return 0
        fi
    fi
    return 1
}

# Run command with sudo if needed
maybe_sudo() {
    if check_root; then
        "$@"
    elif has_sudo; then
        sudo "$@"
    else
        error "Root privileges required to install to $SWIFTTALON_INSTALL_DIR"
        error "Please run with sudo or set SWIFTTALON_INSTALL_DIR to a user-writable directory"
        exit 1
    fi
}

# Detect operating system
detect_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    
    case "$os" in
        linux)
            echo "linux"
            ;;
        darwin)
            echo "darwin"
            ;;
        mingw*|msys*|cygwin*)
            error "Windows is not supported by this installer"
            error "Please use WSL or download the Windows binary manually from GitHub releases"
            exit 1
            ;;
        freebsd)
            error "FreeBSD is not officially supported"
            error "You may need to build from source"
            exit 1
            ;;
        *)
            error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    
    case "$arch" in
        x86_64|amd64)
            echo "amd64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        riscv64)
            echo "riscv64"
            ;;
        armv7l|armv7|armhf)
            echo "arm"
            ;;
        i386|i686)
            error "32-bit x86 is not supported"
            exit 1
            ;;
        *)
            error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

# Get the latest release version from GitHub
get_latest_version() {
    local version
    local api_url="https://api.github.com/repos/$GITHUB_REPO/releases/latest"
    
    # Try to fetch the latest version
    if command -v curl >/dev/null 2>&1; then
        version=$(curl -fsSL "$api_url" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/') || true
    elif command -v wget >/dev/null 2>&1; then
        version=$(wget -qO- "$api_url" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/') || true
    fi
    
    if [ -z "$version" ]; then
        warn "Could not determine latest version from GitHub API"
        warn "Falling back to 'latest' which may not work"
        echo "latest"
    else
        echo "$version"
    fi
}

# Download the binary with progress bar
download_binary() {
    local os="$1"
    local arch="$2"
    local version="$3"
    local output_path="$4"
    local download_url
    
    # Construct download URL
    if [ "$version" = "latest" ]; then
        download_url="https://github.com/$GITHUB_REPO/releases/latest/download/${BINARY_NAME}-${os}-${arch}"
    else
        download_url="https://github.com/$GITHUB_REPO/releases/download/${version}/${BINARY_NAME}-${os}-${arch}"
    fi
    
    status "Downloading ${BINARY_NAME} ${version} for ${os}/${arch}..."
    
    # Download with progress bar
    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL --progress-bar "$download_url" -o "$output_path" 2>&1; then
            error "Failed to download from: $download_url"
            error "Please check your internet connection and try again"
            exit 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget --show-progress -q "$download_url" -O "$output_path" 2>&1; then
            error "Failed to download from: $download_url"
            error "Please check your internet connection and try again"
            exit 1
        fi
    else
        error "Neither curl nor wget is installed"
        error "Please install curl or wget and try again"
        exit 1
    fi
    
    # Verify download succeeded
    if [ ! -f "$output_path" ]; then
        error "Download failed: file not found"
        exit 1
    fi
    
    if [ ! -s "$output_path" ]; then
        error "Download failed: file is empty"
        exit 1
    fi
}

# Verify binary file integrity
verify_binary_file() {
    local binary_path="$1"
    local file_size
    
    # Check file exists
    if [ ! -f "$binary_path" ]; then
        error "Binary not found: $binary_path"
        return 1
    fi
    
    # Check file is not empty
    if [ ! -s "$binary_path" ]; then
        error "Binary file is empty (0 bytes)"
        return 1
    fi
    
    # Check minimum file size (detect truncated downloads)
    file_size=$(wc -c < "$binary_path" 2>/dev/null || echo 0)
    if [ "$file_size" -lt "$MIN_BINARY_SIZE" ]; then
        error "Binary file appears corrupted (size: ${file_size} bytes, expected > ${MIN_BINARY_SIZE})"
        return 1
    fi
    
    # Check file is executable or can be made executable
    if [ ! -x "$binary_path" ]; then
        chmod +x "$binary_path" 2>/dev/null || {
            error "Cannot make binary executable"
            return 1
        }
    fi
    
    return 0
}

# Verify the binary actually runs
verify_binary_execution() {
    local binary_path="$1"
    
    # Try to run --version
    if ! "$binary_path" --version >/dev/null 2>&1; then
        error "Binary execution test failed"
        error "The binary may be incompatible with your system or corrupted"
        
        # Try to provide more diagnostic info
        if file "$binary_path" >/dev/null 2>&1; then
            error "File type: $(file "$binary_path" 2>/dev/null)"
        fi
        
        return 1
    fi
    
    return 0
}

# Full installation verification
verify_installation() {
    local binary_path="$1"
    local version_output
    
    status "Verifying installation..."
    
    # Step 1: Check file integrity
    if ! verify_binary_file "$binary_path"; then
        return 1
    fi
    success "Binary file integrity: OK"
    
    # Step 2: Check execution
    if ! verify_binary_execution "$binary_path"; then
        return 1
    fi
    success "Binary execution: OK"
    
    # Step 3: Get version info
    version_output=$("$binary_path" --version 2>/dev/null || echo "unknown")
    success "Version: ${version_output}"
    
    # Step 4: Test basic functionality (help command)
    if "$binary_path" --help >/dev/null 2>&1; then
        success "Help system: OK"
    fi
    
    return 0
}

# Install the binary to the target directory
install_binary_file() {
    local source_path="$1"
    local target_dir="$2"
    local target_path="${target_dir}/${BINARY_NAME}"
    
    status "Installing ${BINARY_NAME} to ${target_dir}..."
    
    # Create install directory if it doesn't exist
    if [ ! -d "$target_dir" ]; then
        maybe_sudo mkdir -p "$target_dir"
    fi
    
    # Check if we can write to the directory
    if check_root || [ -w "$target_dir" ]; then
        # Direct install with proper permissions
        if command -v install >/dev/null 2>&1; then
            install -o0 -g0 -m755 "$source_path" "$target_path"
        else
            cp "$source_path" "$target_path"
            chmod 755 "$target_path"
        fi
    elif has_sudo; then
        # Use sudo for installation
        sudo mkdir -p "$target_dir" 2>/dev/null || true
        if command -v install >/dev/null 2>&1; then
            sudo install -o0 -g0 -m755 "$source_path" "$target_path"
        else
            sudo cp "$source_path" "$target_path"
            sudo chmod 755 "$target_path"
        fi
    else
        error "Cannot write to $target_dir"
        error "Please run with sudo or set SWIFTTALON_INSTALL_DIR to a writable directory"
        error "Example: curl -fsSL ... | SWIFTTALON_INSTALL_DIR=\$HOME/.local/bin sh"
        exit 1
    fi
    
    # Track installed binary for cleanup on failure
    INSTALLED_BINARY="$target_path"
    
    success "${BINARY_NAME} installed to ${target_path}"
}

# Check if the binary is in PATH
check_path() {
    local install_dir="$1"
    
    case ":$PATH:" in
        *":$install_dir:")
            return 0
            ;;
    esac
    return 1
}

# Suggest adding to PATH
suggest_path() {
    local install_dir="$1"
    local shell_rc=""
    
    if check_path "$install_dir"; then
        return 0
    fi
    
    warn "$install_dir is not in your PATH"
    
    # Detect shell
    local shell_name=""
    if [ -n "${SHELL:-}" ]; then
        shell_name=$(basename "$SHELL")
    fi
    
    case "$shell_name" in
        bash)
            shell_rc="$HOME/.bashrc"
            ;;
        zsh)
            shell_rc="$HOME/.zshrc"
            ;;
        fish)
            shell_rc="$HOME/.config/fish/config.fish"
            ;;
        *)
            shell_rc="$HOME/.profile"
            ;;
    esac
    
    echo ""
    echo "  ${BOLD}Add to PATH:${RESET}"
    echo ""
    if [ "$shell_name" = "fish" ]; then
        echo "    fish_add_path ${install_dir}"
        echo ""
        echo "  Or add to ${shell_rc}:"
        echo "    set -gx PATH ${install_dir} \$PATH"
    else
        echo "    echo 'export PATH=\"${install_dir}:\$PATH\"' >> ${shell_rc}"
        echo ""
        echo "  Then reload: source ${shell_rc}"
    fi
}

# Check Go version (requires 1.23+)
check_go_version() {
    local go_version
    local required_version="1.23"
    
    if ! command -v go >/dev/null 2>&1; then
        error "Go is not installed"
        error "Please install Go 1.23 or later: https://golang.org/dl/"
        exit 1
    fi
    
    go_version=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    
    # Compare versions (simple comparison for major.minor)
    local major required_major
    major=$(echo "$go_version" | cut -d. -f1)
    required_major=$(echo "$required_version" | cut -d. -f1)
    
    if [ "$major" -lt "$required_major" ]; then
        error "Go version $go_version is too old (required: $required_version+)"
        error "Please upgrade Go: https://golang.org/dl/"
        exit 1
    fi
    
    success "Found Go version $go_version"
}

# Print ASCII art banner
print_banner() {
    local version="$1"
    
    echo ""
    echo "${CYAN}${BOLD}"
    cat << 'EOF'
    ╔═══════════════════════════════════════════════════════════╗
    ║                                                           ║
    ║   ███████╗██╗███╗   ██╗██╗   ██╗███████╗████████╗██████╗  ║
    ║   ██╔════╝██║████╗  ██║██║   ██║██╔════╝╚══██╔══╝██╔══██╗ ║
    ║   ███████╗██║██╔██╗ ██║██║   ██║███████╗   ██║   ██████╔╝ ║
    ║   ╚════██║██║██║╚██╗██║╚██╗ ██╔╝╚════██║   ██║   ██╔══██╗ ║
    ║   ███████║██║██║ ╚████║ ╚████╔╝ ███████║   ██║   ██║  ██║ ║
    ║   ╚══════╝╚═╝╚═╝  ╚═══╝  ╚═══╝  ╚══════╝   ╚═╝   ╚═╝  ╚═╝ ║
    ║                                                           ║
    ║                    🐙 Ultra-light AI Assistant            ║
    ╚═══════════════════════════════════════════════════════════╝
EOF
    echo "${RESET}"
    echo ""
    if [ -n "$version" ]; then
        echo "  ${GREEN}${BOLD}Version ${version} installed successfully!${RESET}"
    else
        echo "  ${GREEN}${BOLD}Installation complete!${RESET}"
    fi
    echo ""
}

# Print next steps
print_next_steps() {
    local binary_path="$1"
    local in_path="$2"
    
    echo ""
    echo "  ${BOLD}${MAGENTA}Next Steps:${RESET}"
    echo ""
    
    if [ -n "$SWIFTTALON_QUICK_START" ]; then
        echo "  ${GREEN}Quick Start Mode Enabled${RESET}"
        echo ""
        echo "    ${BOLD}1. Initialize:${RESET}     ${BINARY_NAME} init"
        echo "    ${BOLD}2. Start chatting:${RESET} ${BINARY_NAME} agent"
        echo ""
    else
        echo "    ${BOLD}1.${RESET} Initialize SwiftTalon:"
        echo "       ${CYAN}${BINARY_NAME} init${RESET}"
        echo ""
        echo "    ${BOLD}2.${RESET} Start the interactive TUI:"
        echo "       ${CYAN}${BINARY_NAME} agent${RESET}"
        echo ""
        echo "    ${BOLD}3.${RESET} View available commands:"
        echo "       ${CYAN}${BINARY_NAME} --help${RESET}"
        echo ""
    fi
    
    echo "  ${BOLD}${BLUE}─────────────────────────────────────────${RESET}"
    echo ""
    echo "  ${BOLD}Documentation:${RESET}  https://github.com/${GITHUB_REPO}"
    echo "  ${BOLD}Issues:${RESET}        https://github.com/${GITHUB_REPO}/issues"
    echo "  ${BOLD}Website:${RESET}       https://swifttalon.io"
    echo ""
    
    if [ -n "$SWIFTTALON_QUICK_START" ]; then
        echo ""
        echo "  ${BOLD}${GREEN}Ready to go! Try:${RESET} ${CYAN}${BINARY_NAME} agent${RESET}"
        echo ""
    fi
}

# Print installation summary
print_install_summary() {
    local version="$1"
    local method="$2"
    
    echo ""
    echo "  ${BOLD}${BLUE}Installation Summary:${RESET}"
    echo ""
    echo "    Method:      ${method}"
    echo "    Location:    ${SWIFTTALON_INSTALL_DIR}/${BINARY_NAME}"
    echo "    Version:     ${version}"
    echo ""
}

# Run swifttalon init unless disabled
run_init() {
    local binary_path="$1"
    
    if [ -n "${SWIFTTALON_NO_INIT:-}" ]; then
        status "Skipping initialization (SWIFTTALON_NO_INIT is set)"
        return 0
    fi
    
    if [ -n "$SWIFTTALON_QUICK_START" ]; then
        echo ""
        status "Quick Start: Running '${BINARY_NAME} init'..."
        echo ""
        if "$binary_path" init 2>&1; then
            success "Initialization complete!"
        else
            warn "Initialization failed, you can run it manually: ${BINARY_NAME} init"
        fi
    else
        echo ""
        status "Running '${BINARY_NAME} init' to set up your workspace..."
        echo ""
        if "$binary_path" init 2>&1; then
            success "Workspace initialized"
        else
            warn "Initialization failed, you can run it manually: ${BINARY_NAME} init"
        fi
    fi
}

# Install from source code
install_from_source() {
    status "Installing from source..."
    echo ""
    
    # Check prerequisites
    check_go_version
    
    if ! command -v git >/dev/null 2>&1; then
        error "Git is not installed"
        error "Please install git and try again"
        exit 1
    fi
    success "Found git"
    
    # Create temp directory for cloning
    TEMP_DIR=$(mktemp -d)
    local clone_dir="${TEMP_DIR}/${BINARY_NAME}"
    
    status "Cloning repository..."
    if ! git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$clone_dir" 2>&1; then
        error "Failed to clone repository"
        exit 1
    fi
    success "Repository cloned"
    
    cd "$clone_dir"
    
    status "Downloading dependencies..."
    if ! make deps 2>&1; then
        error "Failed to download dependencies"
        exit 1
    fi
    success "Dependencies downloaded"
    
    status "Building and installing..."
    if ! make install 2>&1; then
        error "Failed to build and install"
        exit 1
    fi
    
    success "Build completed"
    
    # Verify installation
    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        # Check if installed to custom location
        if [ -f "$SWIFTTALON_INSTALL_DIR/$BINARY_NAME" ]; then
            INSTALLED_BINARY="$SWIFTTALON_INSTALL_DIR/$BINARY_NAME"
            success "${BINARY_NAME} installed to ${SWIFTTALON_INSTALL_DIR}"
        else
            error "Installation completed but binary not found in PATH"
            exit 1
        fi
    else
        INSTALLED_BINARY="$(command -v "$BINARY_NAME")"
    fi
    
    # Get version for banner
    local version
    version=$("$INSTALLED_BINARY" --version 2>/dev/null || echo "dev")
    
    # Print banner and success message
    print_banner "$version"
    
    # Check PATH
    if ! check_path "$SWIFTTALON_INSTALL_DIR"; then
        suggest_path "$SWIFTTALON_INSTALL_DIR"
    fi
    
    # Print next steps
    print_next_steps "$INSTALLED_BINARY" "$(check_path "$SWIFTTALON_INSTALL_DIR")"
}

# Install using Docker
install_docker() {
    status "Installing with Docker..."
    echo ""
    
    # Check prerequisites
    if ! command -v docker >/dev/null 2>&1; then
        error "Docker is not installed"
        error "Please install Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    success "Found Docker"
    
    if ! docker compose version >/dev/null 2>&1 && ! docker-compose version >/dev/null 2>&1; then
        error "Docker Compose is not installed"
        error "Please install Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    success "Found Docker Compose"
    
    # Create temp directory for cloning
    TEMP_DIR=$(mktemp -d)
    local clone_dir="${TEMP_DIR}/${BINARY_NAME}"
    
    status "Cloning repository..."
    if ! git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$clone_dir" 2>&1; then
        error "Failed to clone repository"
        exit 1
    fi
    success "Repository cloned to ${clone_dir}"
    
    # Check if config exists, copy example if not
    if [ ! -f "$clone_dir/config/config.json" ] && [ -f "$clone_dir/config/config.example.json" ]; then
        status "Creating configuration from example..."
        cp "$clone_dir/config/config.example.json" "$clone_dir/config/config.json"
        success "Configuration created at ${clone_dir}/config/config.json"
        warn "Please edit the configuration file to add your API keys"
    fi
    
    # Print Docker banner
    print_banner "Docker"
    
    echo ""
    success "Docker setup prepared!"
    echo ""
    echo "  ${BOLD}${MAGENTA}Next Steps:${RESET}"
    echo ""
    echo "    ${BOLD}1.${RESET} Navigate to the repository:"
    echo "       ${CYAN}cd ${clone_dir}${RESET}"
    echo ""
    echo "    ${BOLD}2.${RESET} Edit configuration:"
    echo "       ${CYAN}vim config/config.json${RESET}"
    echo ""
    echo "    ${BOLD}3.${RESET} Start SwiftTalon:"
    echo "       ${CYAN}docker compose --profile gateway up -d${RESET}"
    echo ""
    echo "    ${BOLD}4.${RESET} View logs:"
    echo "       ${CYAN}docker compose logs -f${RESET}"
    echo ""
    echo "  ${BOLD}${BLUE}─────────────────────────────────────────${RESET}"
    echo ""
    echo "  ${BOLD}Note:${RESET} The repository is at ${clone_dir}"
    echo "        Move it to a permanent location if desired."
    echo ""
    
    # Don't remove temp dir for Docker install - user needs it
    TEMP_DIR=""
}

# Main installation flow - Binary method
install_binary() {
    # Detect platform
    local os arch version
    os=$(detect_os)
    arch=$(detect_arch)
    
    status "Detected platform: ${os}/${arch}"
    
    # Get version to install
    if [ "$SWIFTTALON_VERSION" = "latest" ]; then
        version=$(get_latest_version)
    else
        version="$SWIFTTALON_VERSION"
    fi
    
    status "Installing version: ${version}"
    
    # Create temp directory for download
    TEMP_DIR=$(mktemp -d)
    local temp_binary="${TEMP_DIR}/${BINARY_NAME}"
    
    # Download binary
    download_binary "$os" "$arch" "$version" "$temp_binary"
    
    # Make binary executable
    chmod +x "$temp_binary"
    
    # Verify binary before installing
    status "Pre-installation verification..."
    if ! verify_binary_file "$temp_binary"; then
        error "Downloaded binary failed integrity check"
        exit 1
    fi
    if ! verify_binary_execution "$temp_binary"; then
        error "Downloaded binary failed execution test"
        exit 1
    fi
    success "Pre-installation verification passed"
    
    # Install binary
    install_binary_file "$temp_binary" "$SWIFTTALON_INSTALL_DIR"
    
    # Full post-installation verification
    if ! verify_installation "$SWIFTTALON_INSTALL_DIR/$BINARY_NAME"; then
        error "Post-installation verification failed"
        error "Cleaning up failed installation..."
        # Cleanup will happen via trap
        exit 1
    fi
    
    # Get version for banner
    local installed_version
    installed_version=$("$SWIFTTALON_INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null || echo "$version")
    
    # Clear the installed binary tracking since we succeeded
    INSTALLED_BINARY=""
    
    # Print banner and success message
    print_banner "$installed_version"
    
    # Print install summary
    print_install_summary "$installed_version" "Binary"
    
    # Check PATH
    local in_path="yes"
    if ! check_path "$SWIFTTALON_INSTALL_DIR"; then
        in_path="no"
        suggest_path "$SWIFTTALON_INSTALL_DIR"
    fi
    
    # Print next steps
    print_next_steps "$SWIFTTALON_INSTALL_DIR/$BINARY_NAME" "$in_path"
    
    # Run init unless disabled
    run_init "$SWIFTTALON_INSTALL_DIR/$BINARY_NAME"
}

# Main function
main() {
    echo ""
    status "Installing SwiftTalon (${SWIFTTALON_INSTALL_METHOD} method)..."
    echo ""
    
    case "$SWIFTTALON_INSTALL_METHOD" in
        binary)
            install_binary
            ;;
        source)
            install_from_source
            ;;
        docker)
            install_docker
            ;;
        *)
            error "Unknown installation method: $SWIFTTALON_INSTALL_METHOD"
            error "Valid methods: binary, source, docker"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
