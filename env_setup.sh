echo "GO_VERSION=^1.15.2" >>$GITHUB_ENV
echo "PROJECT_ID=projectvoltron" >>$GITHUB_ENV
echo "DOCKER_TAG=$(echo ${GITHUB_SHA:0:7})" >>$GITHUB_ENV
#echo "PR_NUMBER=${{ github.event.number }}" >>GITHUB_ENV
echo "APPS = gateway k8s-engine och" >>GITHUB_ENV