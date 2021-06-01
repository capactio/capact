# Source

This directory contains source Terraform module which is used in the `cap.implementation.terraform.gcp.cloudsql.postgresql.install:0.2.0` Implementation manifest.

### Update Terraform content

1. Prepare `tgz` directory with the 
    
   ```bash
    cd ./install-0.2.0-module && tar -zcvf /tmp/module.tgz . && cd -
    ```

1. Set environmental variables:
   ```bash
   export BUCKET="capactio-terraform-modules"
   export MANIFEST_PATH="terraform.gcp.cloudsql.postgresql.install"
   export MANIFEST_REVISION="0.2.0"
   ```
   
1. Upload `tgz` directory to GCS bucket:
    
   ```bash
   gsutil cp /tmp/module.tgz gs://${BUCKET}/${MANIFEST_PATH}/${MANIFEST_REVISION}/module.tgz
   ```

