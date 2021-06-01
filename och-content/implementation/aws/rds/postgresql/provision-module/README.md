# Source

This directory contains source Terraform module which is used in the `cap.implementation.terraform.aws.rds.postgresql.provision:0.1.0`  Implementation manifest.

### Update Terraform content

1. Prepare `tgz` directory with the

   ```bash
    cd ./provision-module && tar -zcvf /tmp/module.tgz . && cd -
    ```

1. Set environmental variables:
   ```bash
   export BUCKET="capactio-terraform-modules"
   export MANIFEST_PATH="terraform.aws.rds.postgresql.provision"
   export MANIFEST_REVISION="0.1.0"
   ```

1. Upload `tgz` directory to GCS bucket:

   ```bash
   gsutil cp /tmp/module.tgz gs://${BUCKET}/${MANIFEST_PATH}/${MANIFEST_REVISION}/module.tgz
   ```

