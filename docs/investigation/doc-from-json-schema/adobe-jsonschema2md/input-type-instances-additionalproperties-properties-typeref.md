# Untitled object in undefined Schema

```txt
https://projectvoltron.dev/schemas/common/type-ref.json#/additionalProperties/properties/typeRef
```

The full path to the Type from which the TypeInstance is created.

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                                               |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [input-type-instances.json*](../../../../ocf-spec/0.0.1/schema/common/input-type-instances.json "open original schema") |

# typeRef Properties

| Property              | Type     | Required | Nullable       | Defined by                                                                                      |
| :-------------------- | :------- | :------- | :------------- | :---------------------------------------------------------------------------------------------- |
| [path](#path)         | `string` | Required | cannot be null | [Untitled schema](type-ref-properties-path.md "#/properties/path#/properties/path")             |
| [revision](#revision) | `string` | Optional | cannot be null | [Untitled schema](type-ref-properties-revision.md "#/properties/revision#/properties/revision") |

## path

Path of a given Type

`path`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](type-ref-properties-path.md "#/properties/path#/properties/path")

## revision

Version of the manifest content in the SemVer format.

`revision`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](type-ref-properties-revision.md "#/properties/revision#/properties/revision")

### revision Constraints

**minimum length**: the minimum number of characters for this string is: `5`

**pattern**: the string must match the following regular expression: 

```regexp
^(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$
```

[try pattern](https://regexr.com/?expression=%5E\(%3F%3A0%7C%5B1-9%5D%5Cd\*\)%5C.\(%3F%3A0%7C%5B1-9%5D%5Cd\*\)%5C.\(%3F%3A0%7C%5B1-9%5D%5Cd\*\)%24 "try regular expression with regexr.com")
