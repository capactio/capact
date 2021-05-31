# 2. Provision AWS RDS for PostgreSQL

To create Amazon Relational Database Service for PostgreSQL, follow the steps:

1. Create file with parameters:

    ```bash
    cat > /tmp/rds-params.yaml << ENDOFFILE
    superuser:
      username: capact
      password: capact-pass123
    region: "eu-west-1"
    tier: "db.t3.small"
    multi_az: true
    ingress_rule_cidr_blocks: "0.0.0.0/0"
    publicly_accessible: true
    performance_insights_enabled: true
    engine_version: "11.10"
    major_engine_version: "11"
    ENDOFFILE
    ```

1. Create Action:

    ```bash
    capact act create cap.interface.aws.rds.postgresql.provision --name rds --parameters-from-file /tmp/rds-params.yaml
    ```

1. Run Action:

    ```bash
    capact act run rds
    ```

1. Wait until the Action is ready to run:

    ```bash
    capact act get rds
    ```   

1. Watch Action:

    ```bash
    capact act watch rds
    ```

1. Once the Action is succeeded, view Output TypeInstances:

   ```bash
   capact act status rds
   ```
    
   Note the `postgresql` TypeInstance ID, as it will be used for Atlassian application installation.

1. Export environmental variable with PostgreSQL TypeInstance ID noted from previous step:

   ```bash
   export POSTGRESQL_TI_ID={postgresql_typeinstance_id}
   ```

**Next steps:** Navigate back to the [Introduction](./0-intro.md) and follow next steps.
