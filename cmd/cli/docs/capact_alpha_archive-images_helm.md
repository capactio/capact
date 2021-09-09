---
title: capact alpha archive-images helm
---

## capact alpha archive-images helm

Archive all the Docker container images used in Capact Helm charts

```
capact alpha archive-images helm [flags]
```

### Examples

```
# Archive images from the stable Capact Helm repository from version 0.5.0
capact alpha archive-images helm --version 0.5.0 --output ./capact-images-0.5.0.tar

# Archive images from  Helm Chart released from the the '0fbf562' commit on the main branch
capact alpha archive-images helm --version 0.4.0-0fbf562 --helm-repo-url @latest --output-stdout > ./capact-images-0.4.0-0fbf562.tar

# You can use gzip to save the image file and make the backup smaller.
capact alpha archive-images helm --version 0.5.0 --output ./capact-images-0.5.0.tar.gz --compress gzip

# You can pipe output to use custom gzip
capact alpha archive-images helm --version 0.5.0 --output-stdout | gzip > myimage_latest.tar.gz

```

### Options

```
      --compress string          Use a given compress algorithm. Allowed values: gzip
      --helm-repo-url string     Capact Helm chart repository URL. Use @latest tag to select repository which holds the latest Helm chart versions. (default "https://storage.googleapis.com/capactio-stable-charts")
  -h, --help                     help for helm
  -o, --output string            Write output to a file, instead of standard output.
      --output-stdout            Write output to a standard output, instead of file.
      --save-component strings   Components names for which Docker images should be saved. Takes comma-separated list. (default [neo4j,ingress-nginx,argo,cert-manager,kubed,monitoring,capact])
      --version string           Capact version. Possible values @latest, @local, 0.3.0, ... (default "@latest")
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - tracing (default 0 - disable)
```

### SEE ALSO

* [capact alpha archive-images](capact_alpha_archive-images.md)	 - Export Capact Docker images to a tar archive

