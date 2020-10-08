# Open Capability Format

The Open Capability Format (OCF) is a standard way of representing cloud-native application capabilities and prerequisites. By design, cloud-agnostic.

## The specification

The current version should be treated as unstable. Support for features may be dropped at any time without notice. The API may change in incompatible ways in a later software release without notice.

* [Version 0.0.1](./0.0.1/schema)

## Examples

* [Version 0.0.1](./0.0.1/examples)

## Known Issues

1. The repository specification is private, so we are not able to use HTTP `$ref` in spec. As a result, we are using a local file reference, e.g. `"file://schema/common/root-fields.json"`. Unfortunately, it needs to be a relative path, so the validator needs to read schemas with a working directory set to [schema root](./0.0.1/schema).
2. The `kind` field is defined in [`root-fields.json`](./0.0.1/schema/common/root-fields.json). As a result it is not restricted to a given OCF manifest name.
3. It is important to note that the schemas listed in an allOf, anyOf or oneOf array know nothing of one another. While it might be surprising, [allOf](https://json-schema.org/understanding-json-schema/reference/combining.html#allof) can not be used to “extend” a schema to add more details to it in the sense of object-oriented inheritance. 
   In that way, we cannot restrict other fields. The possible solution is to use a generator that instead of using allOf is inline the common fields directly in the specification object. For example, we can use [this generator](https://github.com/mokkabonna/json-schema-merge-allof). 

   _Source: https://json-schema.org/understanding-json-schema/reference/combining.html_
