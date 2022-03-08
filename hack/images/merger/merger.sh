#!/usr/bin/env bash

SRC=${SRC:-"/yamls"}
OUT=${OUT:-"/merged.yaml"}

for filename in "${SRC}"/*; do
  filename=$(basename -- "$filename")
  prefix="${filename%.*}"

  # remove value key if exists
  if [[ $(yq e 'has("value")' "${SRC}"/"${filename}") == "true" ]]; then
    yq e '.value' -i "${SRC}"/"${filename}"
  fi 

  # prefix each file with its filename
  yq e -i "{\"${prefix}\": . }" "${SRC}"/"${filename}"
done

# merge all yaml files into one
# shellcheck disable=SC2016
yq ea '. as $item ireduce ({}; . * $item )' "${SRC}"/*.yaml >"${OUT}"
