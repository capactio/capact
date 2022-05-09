#!/usr/bin/env bash
#
# This script compresses binary file built by goreleaser.
#
# Requires passing path to the binary file as an argument.
# Requires UPX to be installed.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

main() {
    number_of_arguments=$1
    file_path=$2

    if [[ $number_of_arguments -lt 1 ]]; then
        echo "Path to the binary file not provided"
        exit 1
    fi

    # Do not compress darwin arm64 binary as it causes UPX to output broken file
    if ! ( grep -q "darwin" <<< "$file_path" && grep -q "arm64" <<< "$file_path" ); then 
        upx -9 "$file_path"
    fi
}

main $# "${1:-""}"
