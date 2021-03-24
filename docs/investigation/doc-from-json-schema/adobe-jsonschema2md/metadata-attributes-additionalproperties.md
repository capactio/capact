# Untitled object in undefined Schema

```txt
https://projectvoltron.dev/schemas/common/metadata-attributes.json#/additionalProperties
```

The attribute object contains OCF Attributes references. It provides generic categorization for Implementations, Types and TypeInstances. Attributes are used to filter out a specific Implementation.

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                                             |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :----------------------------------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Allowed               | none                | [metadata-attributes.json*](../../../../ocf-spec/0.0.1/schema/common/metadata-attributes.json "open original schema") |

# additionalProperties Properties

| Property              | Type     | Required | Nullable       | Defined by                                                                                                                                                                                        |
| :-------------------- | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [revision](#revision) | `string` | Required | cannot be null | [Untitled schema](metadata-attributes-additionalproperties-properties-revision.md "https://projectvoltron.dev/schemas/common/metadata-attributes.json#/additionalProperties/properties/revision") |

## revision

The exact Attribute revision.

`revision`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](metadata-attributes-additionalproperties-properties-revision.md "https://projectvoltron.dev/schemas/common/metadata-attributes.json#/additionalProperties/properties/revision")

### revision Constraints

**minimum length**: the minimum number of characters for this string is: `5`

**pattern**: the string must match the following regular expression: 

```regexp
^(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$
```

[try pattern](https://regexr.com/?expression=%5E\(%3F%3A0%7C%5B1-9%5D%5Cd\*\)%5C.\(%3F%3A0%7C%5B1-9%5D%5Cd\*\)%5C.\(%3F%3A0%7C%5B1-9%5D%5Cd\*\)%24 "try regular expression with regexr.com")
