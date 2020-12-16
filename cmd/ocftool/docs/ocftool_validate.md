## ocftool validate

Validate OCF manifests

```
ocftool validate [flags]
```

### Examples

```
# Validate interface-group.yaml file with OCF specification in default location
ocftool validate ocf-spec/0.0.1/examples/interface-group.yaml

# Validate multiple files inside test_manifests directory
ocftool validate pkg/ocftool/test_manifests/*.yaml

# Validate interface-group.yaml file with custom OCF specification location 
ocftool validate -s my/ocf/spec/directory ocf-spec/0.0.1/examples/interface-group.yaml

# Validate all OCH manifests
ocftool validate ./och-content/**/*.yaml
```

### Options

```
  -h, --help             help for validate
  -s, --schemas string   Path to the local directory with OCF JSONSchemas. If not provided, built-in JSONSchemas are used.
```

### SEE ALSO

* [ocftool](ocftool.md)	 - CLI tool for working with OCF manifest files

