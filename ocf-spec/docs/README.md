# JSONSchema 2 Markdown generator test

### Adobe `jsonschema2md`

Repo: https://github.com/adobe/jsonschema2md

Example: [adobe-jsonschema2md](./adobe-jsonschema2md)

Command:

```bash
jsonschema2md -d ocf-spec/0.0.1/schema/ -o ocf-spec/docs/adobe-jsonschema2md --schema-extension=json --example-format=yaml --skip  typesection -n --schema-out=-
```

Issues:
- Generates 60+ files for our schemas, and we are not able to say that each schema should be in a single file via flags/config. As a result, we have files named like `implementation-properties-metadata-allof-1-properties-license-oneof-0-properties-name.md`.
- We need to add title property to our schemas, otherwise it generates is as `Untitled schema`.
- Not all files are generated, but we also do not get any error, e.g. for [interfaces.md](./adobe-jsonschema2md/interface.md) the `interface-properties-spec.md` was not generated. Needs to be investigated further.
- Information duplication, e.g. [atrribute.md](./adobe-jsonschema2md/attribute.md) has full definition for the `ocfVersion` type, but we also have [attribute-properties-ocfversion.md](./adobe-jsonschema2md/attribute-properties-ocfversion.md) which duplicates this information.
- The shared metadata is generated as a single .md file. But others files refer to `attribute-properties-ocf-metadata.md` instead of `metadata.md`. Needs to be investigated further.  
