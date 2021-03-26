# JSON Schema

*Interface defines an action signature. It describes the action name, input, and output parameters.*

## Properties

- **`ocfVersion`** *(string)*
- **`kind`** *(string)*: Must be one of: `['Interface']`.
- **`revision`** *(string)*: Version of the manifest content in the SemVer format.
- **`signature`** *(object)*: Ensures the authenticity and integrity of a given manifest.
  - **`och`** *(string)*: The signature signed with the HUB key.
- **`metadata`**: Refer to *https://projectvoltron.dev/schemas/common/metadata.json*.
- **`spec`** *(object)*: A container for the Interface specification definition. Cannot contain additional properties.
  - **`input`** *(object)*: The input schema for Interface action.
    - **`parameters`**: Can contain additional properties.
    - **`typeInstances`**: Refer to *https://projectvoltron.dev/schemas/common/input-type-instances.json*.
  - **`output`** *(object)*: The output schema for Interface action. Cannot contain additional properties.
    - **`typeInstances`**: Refer to *https://projectvoltron.dev/schemas/common/output-type-instances.json*.
  - **`abstract`** *(boolean)*: If true, the Interface cannot be implemented. Default: `False`.
