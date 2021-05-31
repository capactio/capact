# JSON Schema

*The description of an action and its prerequisites (dependencies). An implementation implements at least one interface.*

## Properties

- **`ocfVersion`** *(string)*
- **`kind`** *(string)*: Must be one of: `['Implementation']`.
- **`revision`** *(string)*: Version of the manifest content in the SemVer format.
- **`signature`** *(object)*: Ensures the authenticity and integrity of a given manifest.
  - **`och`** *(string)*: The signature signed with the HUB key.
- **`metadata`** *(object)*
- **`spec`** *(object)*: A container for the Implementation specification definition. Cannot contain additional properties.
  - **`appVersion`** *(string)*: The supported application versions in SemVer2 format. Cannot contain additional properties.
  - **`additionalInput`** *(object)*: Specifies additional input for a given Implementation. Cannot contain additional properties.
    - **`typeInstances`**: Refer to *https://capact.io/schemas/common/input-type-instances.json*.
  - **`additionalOutput`** *(object)*: Specifies additional output for a given Implementation. Cannot contain additional properties.
    - **`typeInstances`**: Refer to *https://capact.io/schemas/common/output-type-instances.json*.
  - **`outputTypeInstanceRelations`** *(object)*: Defines all output TypeInstances to upload with relations between them. It relates to both optional and required TypeInstances. No TypeInstance name specified here means it won't be uploaded to OCH after workflow run. Can contain additional properties.
  - **`implements`** *(array)*: Defines what kind of interfaces this implementation fulfills.
    - **Items** *(object)*: Cannot contain additional properties.
      - **`path`** *(string)*: The Interface path, for example cap.interfaces.db.mysql.install.
      - **`revision`** *(string)*: The exact Interface revision. Cannot contain additional properties.
  - **`requires`** *(object)*: List of the system prerequisites that need to be present on the cluster. There has to be an Instance for every concrete type. Can contain additional properties.
  - **`imports`** *(array)*: List of external Interfaces that this Implementation requires to be able to execute the action.
    - **Items** *(object)*: Cannot contain additional properties.
      - **`interfaceGroupPath`** *(string)*: The name of the Interface Group that contains specific actions that you want to import, for example cap.interfaces.db.mysql.
      - **`alias`** *(string)*: The alias for the full name of the imported group name. It can be used later in the workflow definition instead of using full name.
      - **`appVersion`** *(string)*: The supported application versions in SemVer2 format.
      - **`methods`** *(array)*: The list of all required actions’ names that must be imported.
        - **Items** *(object)*: Cannot contain additional properties.
          - **`name`** *(string)*: The name of the action for a given Interface group, e.g. install.
          - **`revision`** *(string)*: Revision of the Interface for a given action. If not specified, the latest revision is used.
  - **`action`** *(object)*: An explanation about the purpose of this instance. Cannot contain additional properties.
    - **`args`** *(object)*: Holds all parameters that should be passed to the selected runner, for example repoUrl, or chartName for the Helm3 runner.
    - **`runnerInterface`** *(string)*: The Interface of a Runner, which handles the execution, for example, cap.interface.runner.helm3.run.
## Definitions

- **`requireEntity`** *(object)*: Cannot contain additional properties.
  - **`valueConstraints`** *(object)*: Holds the configuration constraints for the given entry. It needs to be valid against the Type JSONSchema.
  - **`name`** *(string)*: The name of the Type. Root prefix can be skipped if it’s a core Type. If it is a custom Type then it MUST be defined as full path to that Type. Custom Type MUST extend the abstract node which is defined as a root prefix for that entry.
  - **`alias`** *(string)*: If provided, the TypeInstance of the Type, configured in policy, is injected to the workflow under the alias.
  - **`revision`** *(string)*: The exact revision of the given Type.
