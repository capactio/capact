#!/bin/bash
if [ -z "${CERT_RESTORE}" ] 
then
  printf "\n***Certs NOT restored, I will backup new.***"
  i=0
  while : 
    do 
      if (( $(kubectl get secret -n ${NAMESPACE} |grep "kubernetes.io/tls" |wc -l) == $(echo ${SERVICES} |wc -w) )) 
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
    for SECRET in $(kubectl get secret -n ${NAMESPACE} |grep "\-tls" |awk '{ print $1 }') 
    do  
      kubectl get secret ${SECRET} -n ${NAMESPACE} -o yaml >secret-${SECRET}-$(date -u +"%Y-%m-%dT%H:%M:%SZ").yaml
    done
    gsutil cp secret*.yaml gs://${BUCKET}/le
else 
  printf "\n***Certs restored, not backuping new.***"
fi