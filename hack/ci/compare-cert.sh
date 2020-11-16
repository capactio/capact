#!/bin/bash
export CREATED=$(gsutil ls -la gs://${BUCKET} |sort -k2 -r |grep -v ^TOTAL |head -n 1 |awk -F"#" '{ print substr($2,1,10) }')
echo "CREATED=${CREATED}" >> $GITHUB_ENV

if [ -z "${CREATED}" ] 
then 
  printf "\n***Certs not existing in the bucket. New one needed to be generated***\n";
  exit
fi

if (( $(echo $((($(date +%s) - ${CREATED} ) / 86400))) > ${CERT_MAX_AGE} )); then
   echo -e "\n- New cert needs to be generated -"
else
  echo -e "\n- Cert still valid. Restoring -"
  gsutil cp $(gsutil ls gs://${BUCKET}  |sort -k2 -r |grep -v ^TOTAL |head -n 1) .
  kubectl create ns ${NAMESPACE} || true
  kubectl apply -f secret-*.yaml --namespace ${NAMESPACE}
  export CERT_RESTORE=1
fi
