#!/usr/bin/env bash
# THIS SCRIPT SHOULD BE REFACTORED ASAP.
# shellcheck disable=SC2154,SC2155,SC2086,SC2046,SC2126,SC2004

if [ -z "${CERT_RESTORE}" ] 
then
  printf "\n***Certs NOT restored, I will backup new.***"
  i=0
  while : 
    do 
      if (( $(kubectl get secret -n ${CERT_SERVICE_NAMESPACE} |grep "kubernetes.io/tls" |wc -l) == ${CERT_NUMBER_TO_BACKUP} ))
        then 
        echo "***All secrets created***"
        break
      fi
      echo "All secrets STILL NOT created"
      sleep 30
      i=$((i+1))
      if (( ${i} == 5 ))
        then 
        echo "Secrets STILL NOT READY, pls. check that."
        exit
      fi
    done
    printf "\n*** Getting secrets & stroing to the bucket ****"
    for SECRET in $(kubectl get secret -n ${CERT_SERVICE_NAMESPACE} |grep "\-tls" |awk '{ print $1 }')
    do  
      kubectl get secret ${SECRET} -n ${CERT_SERVICE_NAMESPACE} -o yaml >secret-${SECRET}-$(date -u +"%Y-%m-%dT%H:%M:%SZ").yaml
    done
    gsutil cp secret*.yaml gs://${RECREATE_CLUSTER_GCS_BUCKET}/le
else 
  printf "\n***Certs restored, not backuping new.***"
fi
