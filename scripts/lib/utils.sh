#!/bin/bash

# Common utilities for uber project scripts
# This file contains shared functions used across multiple scripts

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
  echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
  echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
  echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
  echo -e "${RED}❌ $1${NC}"
}

print_header() {
  echo -e "${CYAN}$1${NC}"
}

# Function to change to project root
cd_project_root() {
  # Use UBER_PROJECT_ROOT if available (when run through uber)
  # Otherwise fall back to calculating it from script location
  if [ -n "$UBER_PROJECT_ROOT" ]; then
    PROJECT_ROOT="$UBER_PROJECT_ROOT"
    print_info "Using uber project root: $PROJECT_ROOT"
  else
    # Fallback for direct execution
    # Get the calling script's directory and go up to project root
    local calling_script="${BASH_SOURCE[1]}"
    if [ -n "$calling_script" ]; then
      local script_dir="$(cd "$(dirname "$calling_script")" && pwd)"
      PROJECT_ROOT="$(cd "$script_dir/.." && pwd)"
    else
      # Fallback if we can't determine the calling script
      SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
      PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
    fi
    print_warning "UBER_PROJECT_ROOT not set, using calculated project root: $PROJECT_ROOT"
  fi

  # Change to project root
  cd "$PROJECT_ROOT"
}

# Function to check prerequisites (Go installation)
check_go_prerequisites() {
  # Check if Go is available
  if ! command -v go &>/dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
  fi
}

# Function to check if a directory is in PATH
is_in_path() {
  local dir="$1"
  if [[ ":$PATH:" == *":$dir:"* ]]; then
    return 0
  else
    return 1
  fi
}

# Function to check if a directory is writable
is_writable() {
  local dir="$1"
  if [ -w "$dir" ] 2>/dev/null; then
    return 0
  else
    return 1
  fi
}

# Function to show help with custom usage
show_help() {
  local script_name="$1"
  local usage="$2"
  local description="$3"
  local options="$4"
  local examples="$5"

  echo "Usage: $script_name $usage"
  echo ""
  if [ -n "$description" ]; then
    echo "$description"
    echo ""
  fi
  if [ -n "$options" ]; then
    echo "Options:"
    echo "$options"
    echo ""
  fi
  if [ -n "$examples" ]; then
    echo "Examples:"
    echo "$examples"
    echo ""
  fi
  exit 0
}

# Function to parse common command line arguments
parse_common_args() {
  local verbose_var="$1"
  local help_var="$2"
  shift 2

  while [[ $# -gt 0 ]]; do
    case $1 in
    -v | --verbose)
      eval "$verbose_var=true"
      shift
      ;;
    -h | --help)
      eval "$help_var=true"
      shift
      ;;
    *)
      # Unknown option, let the calling script handle it
      return 1
      ;;
    esac
  done
  return 0
}

# Function to validate file exists
file_exists() {
  local file="$1"
  if [ -f "$file" ]; then
    return 0
  else
    return 1
  fi
}

# Function to validate directory exists
dir_exists() {
  local dir="$1"
  if [ -d "$dir" ]; then
    return 0
  else
    return 1
  fi
}

# Function to create directory if it doesn't exist
ensure_dir() {
  local dir="$1"
  local verbose="${2:-false}"

  if [ ! -d "$dir" ]; then
    if [ "$verbose" = true ]; then
      print_info "Creating directory: $dir"
    fi
    mkdir -p "$dir"
  fi
}

# Function to get script directory
get_script_dir() {
  echo "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
}

# Function to get project root from script location
get_project_root_from_script() {
  local script_path="$1"
  local script_dir="$(cd "$(dirname "$script_path")" && pwd)"
  echo "$(cd "$script_dir/.." && pwd)"
}
