#!/usr/bin/env bash

SRC=${SRC:-"/yamls"}
OUT=${OUT:-"/merged.yaml"}

# prefix each file with its filename
for filename in "${SRC}"/*.yaml; do
  filename=$(basename -- "$filename")
  prefix="${filename%.*}"
  yq e -i "{\"${prefix}\": . }" "${SRC}"/"${filename}"
done

# merge all yaml files into one
# shellcheck disable=SC2016
yq ea '. as $item ireduce ({}; . * $item )' "${SRC}"/*.yaml >"${OUT}"
