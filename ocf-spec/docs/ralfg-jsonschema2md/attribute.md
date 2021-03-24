# JSON Schema

*Attribute is a primitive, which is used to categorize Implementations. You can use Attributes to find and filter Implementations.*

## Properties

- **`ocfVersion`** *(string)*
- **`kind`** *(string)*: Must be one of: `['Attribute']`.
- **`revision`** *(string)*: Version of the manifest content in the SemVer format.
- **`signature`** *(object)*: Ensures the authenticity and integrity of a given manifest.
  - **`och`** *(string)*: The signature signed with the HUB key.
- **`metadata`**: Refer to *https://projectvoltron.dev/schemas/common/metadata.json*.
- **`spec`** *(object)*: A container for the Attribute specification definition. Cannot contain additional properties.
  - **`additionalRefs`** *(array)*: List of the full path of additional parent nodes the Attribute is attached to. The parent nodes MUST reside under “cap.core.attribute” or “cap.attribute” subtree. The connection means that the Attribute becomes a child of the referenced parent nodes. In a result, the Attribute has multiple parents. Cannot contain additional properties.
    - **Items** *(string)*: Full path of additional parent nodes the Attribute is attached to.
