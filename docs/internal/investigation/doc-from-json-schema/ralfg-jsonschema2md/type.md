# JSON Schema

*Primitive, that holds the JSONSchema which describes that Type. It’s also used for validation. There are core and custom Types. Type can be also a composition of other Types.*

## Properties

- **`ocfVersion`** *(string)*
- **`kind`** *(string)*: Must be one of: `['Type']`.
- **`revision`** *(string)*: Version of the manifest content in the SemVer format.
- **`signature`** *(object)*: Ensures the authenticity and integrity of a given manifest.
  - **`och`** *(string)*: The signature signed with the HUB key.
- **`metadata`** *(object)*
- **`spec`** *(object)*: A container for the Type specification definition. Cannot contain additional properties.
  - **`jsonSchema`**: Refer to *https://capact.io/schemas/common/json-schema-type.json*.
  - **`additionalRefs`** *(array)*: List of the full path of additional parent nodes the Type is attached to. The parent nodes MUST reside under “cap.core.type” or “cap.type” subtree. The connection means that the Type becomes a child of the referenced parent nodes. In a result, the Type has multiple parents. Cannot contain additional properties.
    - **Items** *(string)*: Full path of additional parent nodes the Attribute is attached to.
