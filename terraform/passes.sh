#!/bin/bash
grep db_pass set-cluster-specific-variables.sh |sed 's/db_pass/TF_VAR_service/g' >set-passes-as-tf-variables.sh
. ./set-passes-as-tf-variables.sh
