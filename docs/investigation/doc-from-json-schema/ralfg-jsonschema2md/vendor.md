# JSON Schema

*Remote OCH repositories can be mounted under the vendor sub-tree in the local repository. OCF Vendor manifest stores connection details of the external OCH, such as URI of the repository (base path) or federation strategy.*

## Properties

- **`ocfVersion`** *(string)*
- **`kind`** *(string)*: Must be one of: `['Vendor']`.
- **`revision`** *(string)*: Version of the manifest content in the SemVer format.
- **`signature`** *(object)*: Ensures the authenticity and integrity of a given manifest.
  - **`och`** *(string)*: The signature signed with the HUB key.
- **`metadata`**: Refer to *https://capact.io/schemas/common/metadata.json*.
- **`spec`** *(object)*: A container for the Vendor specification definition. Cannot contain additional properties.
  - **`federation`** *(object)*: Holds configuration for vendor federation. Cannot contain additional properties.
    - **`uri`** *(string)*: The URI of the external OCH.
