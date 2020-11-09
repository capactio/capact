#!/bin/bash
CREATED=$(gsutil ls -la gs://projectvoltron_le |sort -h |grep -v ^TOTAL |head -n 1 |awk -F"#" '{ print substr($2,1,10) }')
if (( $(echo $((($(date +%s) - ${CREATED} ) / 86400))) > ${CERT_MAX_AGE} )); then
echo -e "\n- New cert needs to be generated -"
else
echo -e "\n- Cert still valid. Restoring -"
gsutil cp $(gsutil ls gs://projectvoltron_le |sort -h -r |head -n 1) .
kubectl apply -f secret-*.yaml --namespace ${NAMESPACE}
export CERT_RESTORE=1
fi
