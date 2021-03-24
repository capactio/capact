# Untitled object in undefined Schema

```txt
undefined
```

Primitive, that holds the JSONSchema which describes that Type. It’s also used for validation. There are core and custom Types. Type can be also a composition of other Types.

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                       |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :--------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [type.json](../../../../ocf-spec/0.0.1/schema/type.json "open original schema") |

# Untitled object in undefined Properties

| Property                  | Type     | Required | Nullable       | Defined by                                                                                        |
| :------------------------ | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------ |
| [ocfVersion](#ocfversion) | `string` | Required | cannot be null | [Untitled schema](type-properties-ocfversion.md "#/properties/ocfVersion#/properties/ocfVersion") |
| [kind](#kind)             | `string` | Required | cannot be null | [Untitled schema](type-properties-kind.md "#/properties/kind#/properties/kind")                   |
| [revision](#revision)     | `string` | Required | cannot be null | [Untitled schema](type-properties-revision.md "#/properties/revision#/properties/revision")       |
| [signature](#signature)   | `object` | Required | cannot be null | [Untitled schema](type-properties-signature.md "#/properties/signature#/properties/signature")    |
| [metadata](#metadata)     | Merged   | Required | cannot be null | [Untitled schema](type-properties-metadata.md "#/properties/metadata#/properties/metadata")       |
| [spec](#spec)             | `object` | Required | cannot be null | [Untitled schema](type-properties-spec.md "#/properties/spec#/properties/spec")                   |

## ocfVersion



`ocfVersion`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](type-properties-ocfversion.md "#/properties/ocfVersion#/properties/ocfVersion")

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

*   defined in: [Untitled schema](type-properties-kind.md "#/properties/kind#/properties/kind")

### kind Constraints

**enum**: the value of this property must be equal to one of the following values:

| Value    | Explanation |
| :------- | :---------- |
| `"Type"` |             |

## revision

Version of the manifest content in the SemVer format.

`revision`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](type-properties-revision.md "#/properties/revision#/properties/revision")

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

*   Type: `object` ([Details](type-properties-signature.md))

*   cannot be null

*   defined in: [Untitled schema](type-properties-signature.md "#/properties/signature#/properties/signature")

## metadata



`metadata`

*   is required

*   Type: `object` ([Details](type-properties-metadata.md))

*   cannot be null

*   defined in: [Untitled schema](type-properties-metadata.md "#/properties/metadata#/properties/metadata")

## spec

A container for the Type specification definition.

`spec`

*   is required

*   Type: `object` ([Details](type-properties-spec.md))

*   cannot be null

*   defined in: [Untitled schema](type-properties-spec.md "#/properties/spec#/properties/spec")
