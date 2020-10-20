#!/usr/bin/env bash

set -e

srcPaths=()
POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"
    case ${key} in
        --linter-args)
            linterArgs="$2"
            shift # past argument
            shift # past value
        ;;
        --src)
            srcPaths+=("$2")
            shift
            shift
        ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters


for path in "${srcPaths[@]}"
do
  graphql-schema-linter "${linterArgs}" "${path}"
done
