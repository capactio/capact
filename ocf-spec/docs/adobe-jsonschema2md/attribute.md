# Untitled object in undefined Schema

```txt
undefined
```

Attribute is a primitive, which is used to categorize Implementations. You can use Attributes to find and filter Implementations.

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                 |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Allowed               | none                | [attribute.json](../../0.0.1/schema/attribute.json "open original schema") |

# Untitled object in undefined Properties

| Property                  | Type     | Required | Nullable       | Defined by                                                                                                                             |
| :------------------------ | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------- |
| [ocfVersion](#ocfversion) | `string` | Required | cannot be null | [Untitled schema](attribute-properties-ocfversion.md "#/properties/ocfVersion#/properties/ocfVersion")                                 |
| [kind](#kind)             | `string` | Required | cannot be null | [Untitled schema](attribute-properties-kind.md "#/properties/kind#/properties/kind")                                                   |
| [revision](#revision)     | `string` | Required | cannot be null | [Untitled schema](attribute-properties-revision.md "#/properties/revision#/properties/revision")                                       |
| [signature](#signature)   | `object` | Required | cannot be null | [Untitled schema](attribute-properties-signature.md "#/properties/signature#/properties/signature")                                    |
| [metadata](#metadata)     | `object` | Required | cannot be null | [Untitled schema](attribute-properties-ocf-metadata.md "https://projectvoltron.dev/schemas/common/metadata.json#/properties/metadata") |
| [spec](#spec)             | `object` | Optional | cannot be null | [Untitled schema](attribute-properties-spec.md "#/properties/spec#/properties/spec")                                                   |

## ocfVersion



`ocfVersion`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](attribute-properties-ocfversion.md "#/properties/ocfVersion#/properties/ocfVersion")

### ocfVersion Constraints

**constant**: the value of this property must be equal to:

```json
"0.0.1"
```

## kind



`kind`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](attribute-properties-kind.md "#/properties/kind#/properties/kind")

### kind Constraints

**enum**: the value of this property must be equal to one of the following values:

| Value         | Explanation |
| :------------ | :---------- |
| `"Attribute"` |             |

## revision

Version of the manifest content in the SemVer format.

`revision`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](attribute-properties-revision.md "#/properties/revision#/properties/revision")

### revision Constraints

**minimum length**: the minimum number of characters for this string is: `5`

**pattern**: the string must match the following regular expression: 

```regexp
^(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$
```

[try pattern](https://regexr.com/?expression=%5E\(%3F%3A0%7C%5B1-9%5D%5Cd\*\)%5C.\(%3F%3A0%7C%5B1-9%5D%5Cd\*\)%5C.\(%3F%3A0%7C%5B1-9%5D%5Cd\*\)%24 "try regular expression with regexr.com")

## signature

Ensures the authenticity and integrity of a given manifest.

`signature`

*   is required

*   Type: `object` ([Details](attribute-properties-signature.md))

*   cannot be null

*   defined in: [Untitled schema](attribute-properties-signature.md "#/properties/signature#/properties/signature")

## metadata

A container for the OCF metadata definitions.

`metadata`

*   is required

*   Type: `object` ([OCF Metadata](attribute-properties-ocf-metadata.md))

*   cannot be null

*   defined in: [Untitled schema](attribute-properties-ocf-metadata.md "https://projectvoltron.dev/schemas/common/metadata.json#/properties/metadata")

## spec

A container for the Attribute specification definition.

`spec`

*   is optional

*   Type: `object` ([Details](attribute-properties-spec.md))

*   cannot be null

*   defined in: [Untitled schema](attribute-properties-spec.md "#/properties/spec#/properties/spec")
