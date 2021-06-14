# JSON Schema

*RepoMetadata stores metadata about the Open Capability Hub.*

## Properties

- **`ocfVersion`** *(string)*
- **`kind`** *(string)*: Must be one of: `['RepoMetadata']`.
- **`revision`** *(string)*: Version of the manifest content in the SemVer format.
- **`signature`** *(object)*: Ensures the authenticity and integrity of a given manifest.
  - **`och`** *(string)*: The signature signed with the HUB key.
- **`metadata`**: Refer to *https://capact.io/schemas/common/metadata.json*.
- **`spec`** *(object)*: A container for the RepoMetadata definition. Cannot contain additional properties.
  - **`implementation`** *(object)*: Holds configuration for the OCF Implementation entities. Cannot contain additional properties.
    - **`appVersion`** *(object)*: Defines the configuration for the appVersion field. Cannot contain additional properties.
      - **`semVerTaggingStrategy`** *(object)*: Defines the tagging strategy. Cannot contain additional properties.
        - **`latest`** *(object)*: Defines the strategy for which version the tag Latest should be applied. You configure this while running OCH. Cannot contain additional properties.
          - **`pointsTo`** *(string)*: An explanation about the purpose of this instance. Must be one of: `['Stable', 'Edge']`.
  - **`ocfVersion`** *(object)*: Holds information about supported OCF versions in OCH server. Cannot contain additional properties.
    - **`default`** *(string)*: The default OCF version that is supported by the OCH. It should be the stored version.
    - **`supported`** *(array)*: The supported OCF version that OCH is able to serve. In general, the OCH takes the stored version and converts it to the supported one.
      - **Items** *(string)*
  - **`hubVersion`** *(string)*: Defines the OCH version in SemVer2 format.
