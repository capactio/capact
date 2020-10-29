#!/usr/bin/env bash

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

exitCode=0
for path in "${srcPaths[@]}"
do
  echo "- Linting ${path}..."
  graphql-schema-linter "${linterArgs}" "${path}"
  lastExitCode=$?
  if [ $lastExitCode -ne 0 ]; then
    exitCode=${lastExitCode}
  fi
done

exit ${exitCode}
