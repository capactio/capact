# JSON Schema

*InterfaceGroup stores metadata for a group of Interfaces.*

## Properties

- **`ocfVersion`** *(string)*
- **`kind`** *(string)*: Must be one of: `['InterfaceGroup']`.
- **`signature`** *(object)*: Ensures the authenticity and integrity of a given manifest.
  - **`och`** *(string)*: The signature signed with the HUB key.
- **`metadata`**: Refer to *https://projectvoltron.dev/schemas/common/metadata.json*.
