# Open Capability Format

The Open Capability Format (OCF) is a standard way of representing cloud-native application capabilities and prerequisites. By design, cloud-agnostic.

## The specification

The current version should be treated as unstable. Support for features may be dropped at any time without notice. The API may change in incompatible ways in a later specification release without notice.

* [Version 0.0.1](./0.0.1/schema)

## Examples

* [Version 0.0.1](./0.0.1/examples)

## Known Issues

1. The repository specification is private, so we are not able to expose HTTPS `$ref` in spec. As a result, we always need to use a tool that supports passing the `$ref` resolvers.
2. We are not using the `allOf` on the root of the JSON Schemas because generators do not support that. As a result, we need to redefine the root fields for each manifest.  
3. It is important to note that the schemas listed in an allOf, anyOf or oneOf array know nothing of one another. While it might be surprising, [allOf](https://json-schema.org/understanding-json-schema/reference/combining.html#allof) can not be used to “extend” a schema to add more details to it in the sense of object-oriented inheritance. 
   In that way, we cannot restrict other fields. The possible solution is to use a generator that instead of using allOf is inline the common fields directly in the specification object. For example, we can use [this generator](https://github.com/mokkabonna/json-schema-merge-allof). 

   _Source: https://json-schema.org/understanding-json-schema/reference/combining.html_
