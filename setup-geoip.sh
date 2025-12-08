#!/bin/bash

# GeoIP Database Setup Script
# Downloads compressed GeoIP databases (.tar.gz) from cloud storage and extracts them
# Supports multiple database files: GeoLite2-City.mmdb and GeoLite2-ASN.mmdb
# CI/CD friendly with non-interactive mode

set -e  # Exit on error

# Configuration
GEOIP_DIR="${GEOIP_DIR:-pkg/geoip}"
DOWNLOAD_URL="${GEOIP_DOWNLOAD_URL:-}"
FORCE_DOWNLOAD="${FORCE_DOWNLOAD:-false}"
CI_MODE="${CI:-false}"

# Database files to look for
REQUIRED_FILES=("GeoLite2-City.mmdb" "GeoLite2-ASN.mmdb")

# Colors for output (disabled in CI)
if [ "$CI_MODE" = "true" ]; then
    GREEN=''
    YELLOW=''
    RED=''
    NC=''
else
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    RED='\033[0;31m'
    NC='\033[0m'
fi

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${YELLOW}$1${NC}"
    echo "================================"
}

# Check if URL is configured
check_url_configured() {
    if [ -z "$DOWNLOAD_URL" ]; then
        log_error "DOWNLOAD_URL not configured!"
        echo ""
        echo "Set the URL via environment variable:"
        echo "  export GEOIP_DOWNLOAD_URL='https://your-url.com/geoip-databases.tar.gz'"
        echo ""
        echo "Or pass it as an argument:"
        echo "  $0 https://your-url.com/geoip-databases.tar.gz"
        echo ""
        echo "Supported cloud storage options:"
        echo "  • GitHub Releases: https://github.com/user/repo/releases/download/v1.0/geoip-databases.tar.gz"
        echo "  • Google Drive: https://drive.google.com/uc?export=download&id=FILE_ID"
        echo "  • Dropbox: https://www.dropbox.com/s/FILE_ID/geoip-databases.tar.gz?dl=1"
        echo "  • AWS S3: https://bucket.s3.amazonaws.com/geoip-databases.tar.gz"
        echo "  • Any public HTTPS URL"
        exit 1
    fi
}

# Create directory if needed
create_directory() {
    if [ ! -d "$GEOIP_DIR" ]; then
        log_info "Creating directory: $GEOIP_DIR"
        mkdir -p "$GEOIP_DIR"
    fi
}

# Check if all required files exist
check_existing_files() {
    local all_exist=true

    for db_file in "${REQUIRED_FILES[@]}"; do
        if [ ! -f "$GEOIP_DIR/$db_file" ]; then
            all_exist=false
            break
        fi
    done

    if [ "$all_exist" = true ]; then
        if [ "$FORCE_DOWNLOAD" = "true" ]; then
            log_warn "Databases exist but FORCE_DOWNLOAD=true, re-downloading..."
            for db_file in "${REQUIRED_FILES[@]}"; do
                [ -f "$GEOIP_DIR/$db_file" ] && rm "$GEOIP_DIR/$db_file"
            done
            return 0
        fi

        if [ "$CI_MODE" = "true" ]; then
            log_info "All databases already exist, skipping download (use FORCE_DOWNLOAD=true to override)"
            for db_file in "${REQUIRED_FILES[@]}"; do
                if [ -f "$GEOIP_DIR/$db_file" ]; then
                    FILE_SIZE_HR=$(du -h "$GEOIP_DIR/$db_file" | cut -f1)
                    log_info "  ✓ $db_file ($FILE_SIZE_HR)"
                fi
            done
            exit 0
        else
            log_warn "All databases already exist:"
            for db_file in "${REQUIRED_FILES[@]}"; do
                if [ -f "$GEOIP_DIR/$db_file" ]; then
                    FILE_SIZE_HR=$(du -h "$GEOIP_DIR/$db_file" | cut -f1)
                    echo "  • $db_file ($FILE_SIZE_HR)"
                fi
            done
            read -p "Do you want to re-download? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_info "Setup cancelled."
                exit 0
            fi
            for db_file in "${REQUIRED_FILES[@]}"; do
                [ -f "$GEOIP_DIR/$db_file" ] && rm "$GEOIP_DIR/$db_file"
            done
        fi
    fi
}

# Download and extract the files
download_and_extract() {
    log_info "Downloading compressed GeoIP databases..."
    log_info "Source: $DOWNLOAD_URL"

    local temp_dir=$(mktemp -d)
    local temp_file="$temp_dir/geoip-databases.tar.gz"

    # Download the compressed file
    log_info "Downloading .tar.gz archive..."
    if command -v curl &> /dev/null; then
        if [ "$CI_MODE" = "true" ]; then
            curl -fsSL "$DOWNLOAD_URL" -o "$temp_file"
        else
            curl -L --progress-bar "$DOWNLOAD_URL" -o "$temp_file"
        fi
    elif command -v wget &> /dev/null; then
        if [ "$CI_MODE" = "true" ]; then
            wget -q "$DOWNLOAD_URL" -O "$temp_file"
        else
            wget --show-progress "$DOWNLOAD_URL" -O "$temp_file"
        fi
    else
        log_error "Neither curl nor wget is installed."
        log_error "Please install curl or wget to download the databases."
        rm -rf "$temp_dir"
        exit 1
    fi

    # Verify the downloaded file
    if [ ! -f "$temp_file" ]; then
        log_error "Download failed - file not created"
        rm -rf "$temp_dir"
        exit 1
    fi

    local archive_size=$(stat -f%z "$temp_file" 2>/dev/null || stat -c%s "$temp_file" 2>/dev/null)
    log_info "Downloaded archive size: ${archive_size} bytes"

    if [ "$archive_size" -lt 100000 ]; then
        log_error "Downloaded file is too small (${archive_size} bytes)"
        log_error "This might be an error page or invalid file."
        rm -rf "$temp_dir"
        exit 1
    fi

    # Extract the archive
    log_info "Extracting .tar.gz archive..."
    tar -xzf "$temp_file" -C "$temp_dir"

    # Find all .mmdb files in the extracted contents
    log_info "Looking for database files..."
    local found_files=0

    for db_file in "${REQUIRED_FILES[@]}"; do
        MMDB_FILE=$(find "$temp_dir" -name "$db_file" | head -n 1)

        if [ -n "$MMDB_FILE" ]; then
            log_info "  ✓ Found: $db_file"
            mv "$MMDB_FILE" "$GEOIP_DIR/$db_file"
            found_files=$((found_files + 1))
        else
            log_warn "  ✗ Not found: $db_file"
        fi
    done

    # Cleanup temp directory
    rm -rf "$temp_dir"

    if [ "$found_files" -eq 0 ]; then
        log_error "No .mmdb files found in the archive"
        exit 1
    fi

    log_info "Extraction complete! Found $found_files database(s)"
}

# Verify extracted files
verify_files() {
    local verified=0
    local failed=0

    log_info "Verifying extracted databases..."

    for db_file in "${REQUIRED_FILES[@]}"; do
        if [ -f "$GEOIP_DIR/$db_file" ]; then
            # Check file size
            FILE_SIZE=$(stat -f%z "$GEOIP_DIR/$db_file" 2>/dev/null || stat -c%s "$GEOIP_DIR/$db_file" 2>/dev/null)

            if [ "$FILE_SIZE" -lt 100000 ]; then
                log_error "  ✗ $db_file is too small (${FILE_SIZE} bytes)"
                failed=$((failed + 1))
            else
                FILE_SIZE_HR=$(du -h "$GEOIP_DIR/$db_file" | cut -f1)
                log_info "  ✓ $db_file ($FILE_SIZE_HR)"
                verified=$((verified + 1))

                # Verify it's a valid MMDB file (magic bytes check)
                if command -v file &> /dev/null; then
                    FILE_TYPE=$(file "$GEOIP_DIR/$db_file")
                    if [[ ! "$FILE_TYPE" =~ "data" ]] && [[ ! "$FILE_TYPE" =~ "MaxMind" ]]; then
                        log_warn "    Warning: File type might not be correct"
                    fi
                fi
            fi
        else
            log_warn "  ✗ $db_file not found"
            failed=$((failed + 1))
        fi
    done

    if [ "$failed" -gt 0 ]; then
        log_warn "Some databases could not be verified"
        if [ "$verified" -eq 0 ]; then
            log_error "No valid databases found"
            exit 1
        fi
    fi
}

# Print success message
print_success() {
    echo ""
    log_section "✓ Setup complete!"
    log_info "Database location: $GEOIP_DIR/"
    log_info "Databases installed:"

    for db_file in "${REQUIRED_FILES[@]}"; do
        if [ -f "$GEOIP_DIR/$db_file" ]; then
            FILE_SIZE_HR=$(du -h "$GEOIP_DIR/$db_file" | cut -f1)
            echo "  • $db_file ($FILE_SIZE_HR)"
        fi
    done
    echo "================================"
}

# Main execution
main() {
    log_section "GeoIP Databases Setup (Compressed)"

    # Allow URL to be passed as first argument
    if [ -n "$1" ]; then
        DOWNLOAD_URL="$1"
    fi

    check_url_configured
    create_directory
    check_existing_files
    download_and_extract
    verify_files
    print_success
}

# Run main function
main "$@"