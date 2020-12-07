# Workflow rendering PoC

Implementations used in this PoC are in [implementations](implementations) directory.

```bash
# provision postgresql
go run docs/investigation/workflow-rendering/main.go 'cap.implementation.bitnami.postgresql.install' | kubectl apply -f -

# provision jira
go run docs/investigation/workflow-rendering/main.go 'cap.implementation.atlassian.jira.install' | kubectl apply -f -
```