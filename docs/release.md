# Release

- PR https://github.com/Project-Voltron/go-voltron/pull/282
- Create branch
```bash
git checkout master`
git pull
git checkout -b release-0.3
```

- Replace default tag for Capact chart
    `deploy/kubernetes/charts/capact/values.yaml`
- Replace `master` to `release-0.3`
  `deploy/kubernetes/charts/capact/charts/och-public/values.yaml`

```bash
git add .
git commit -m "Set fixed Capact image tag and Populator source branch"
```

- Release Helm charts
```bash
./hack/release-charts.sh
```

- Release tools
```bash
make build-all-tools-prod

export CAPACT_BINARIES_BUCKET=capactio_binaries
gsutil -m cp ./bin/* gs://${CAPACT_BINARIES_BUCKET}/v0.0.0/
```

- Push tag
```bash
git tag v0.3.0 HEAD
git push origin v0.3.0
git push -u origin release-0.3
```

- Create GitHub release https://github.com/Project-Voltron/go-voltron/releases/new

