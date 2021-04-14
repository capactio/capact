# Untitled object in undefined Schema

```txt
https://capact.io/schemas/implementation.json
```

The description of an action and its prerequisites (dependencies). An implementation implements at least one interface.

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                           |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :----------------------------------------------------------------------------------- |
| Can be instantiated | Yes        | Unknown status | No           | Forbidden         | Allowed               | none                | [implementation.json](../../../../ocf-spec/0.0.1/schema/implementation.json "open original schema") |

# Untitled object in undefined Properties

| Property                  | Type     | Required | Nullable       | Defined by                                                                                                  |
| :------------------------ | :------- | :------- | :------------- | :---------------------------------------------------------------------------------------------------------- |
| [ocfVersion](#ocfversion) | `string` | Required | cannot be null | [Untitled schema](implementation-properties-ocfversion.md "#/properties/ocfVersion#/properties/ocfVersion") |
| [kind](#kind)             | `string` | Required | cannot be null | [Untitled schema](implementation-properties-kind.md "#/properties/kind#/properties/kind")                   |
| [revision](#revision)     | `string` | Required | cannot be null | [Untitled schema](implementation-properties-revision.md "#/properties/revision#/properties/revision")       |
| [signature](#signature)   | `object` | Required | cannot be null | [Untitled schema](implementation-properties-signature.md "#/properties/signature#/properties/signature")    |
| [metadata](#metadata)     | Merged   | Required | cannot be null | [Untitled schema](implementation-properties-metadata.md "#/properties/metadata#/properties/metadata")       |
| [spec](#spec)             | `object` | Required | cannot be null | [Untitled schema](implementation-properties-spec.md "#/properties/spec#/properties/spec")                   |

## ocfVersion



`ocfVersion`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](implementation-properties-ocfversion.md "#/properties/ocfVersion#/properties/ocfVersion")

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

*   defined in: [Untitled schema](implementation-properties-kind.md "#/properties/kind#/properties/kind")

### kind Constraints

**enum**: the value of this property must be equal to one of the following values:

| Value              | Explanation |
| :----------------- | :---------- |
| `"Implementation"` |             |

## revision

Version of the manifest content in the SemVer format.

`revision`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](implementation-properties-revision.md "#/properties/revision#/properties/revision")

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

*   Type: `object` ([Details](implementation-properties-signature.md))

*   cannot be null

*   defined in: [Untitled schema](implementation-properties-signature.md "#/properties/signature#/properties/signature")

## metadata



`metadata`

*   is required

*   Type: `object` ([Details](implementation-properties-metadata.md))

*   cannot be null

*   defined in: [Untitled schema](implementation-properties-metadata.md "#/properties/metadata#/properties/metadata")

## spec

A container for the Implementation specification definition.

`spec`

*   is required

*   Type: `object` ([Details](implementation-properties-spec.md))

*   cannot be null

*   defined in: [Untitled schema](implementation-properties-spec.md "#/properties/spec#/properties/spec")

# Untitled object in undefined Definitions

## Definitions group requireEntity

Reference this group by using

```json
{"$ref":"https://capact.io/schemas/implementation.json#/definitions/requireEntity"}
```

| Property                              | Type     | Required | Nullable       | Defined by                                                                                                                                                                             |
| :------------------------------------ | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [valueConstraints](#valueconstraints) | `object` | Optional | cannot be null | [Untitled schema](implementation-definitions-requireentity-properties-valueconstraints.md "#/properties/spec/properties/value#/definitions/requireEntity/properties/valueConstraints") |
| [name](#name)                         | `string` | Required | cannot be null | [Untitled schema](implementation-definitions-requireentity-properties-name.md "#/properties/spec/properties/name#/definitions/requireEntity/properties/name")                          |
| [alias](#alias)                       | `string` | Optional | cannot be null | [Untitled schema](implementation-definitions-requireentity-properties-alias.md "#/properties/spec/properties/alias#/definitions/requireEntity/properties/alias")                       |
| [revision](#revision-1)               | `string` | Required | cannot be null | [Untitled schema](implementation-definitions-requireentity-properties-revision.md "#/properties/spec/properties/revision#/definitions/requireEntity/properties/revision")              |

### valueConstraints

Holds the configuration constraints for the given entry. It needs to be valid against the Type JSONSchema.

`valueConstraints`

*   is optional

*   Type: `object` ([Details](implementation-definitions-requireentity-properties-valueconstraints.md))

*   cannot be null

*   defined in: [Untitled schema](implementation-definitions-requireentity-properties-valueconstraints.md "#/properties/spec/properties/value#/definitions/requireEntity/properties/valueConstraints")

### name

The name of the Type. Root prefix can be skipped if it’s a core Type. If it is a custom Type then it MUST be defined as full path to that Type. Custom Type MUST extend the abstract node which is defined as a root prefix for that entry.

`name`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](implementation-definitions-requireentity-properties-name.md "#/properties/spec/properties/name#/definitions/requireEntity/properties/name")

### alias

If provided, the TypeInstance of the Type, configured in policy, is injected to the workflow under the alias.

`alias`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](implementation-definitions-requireentity-properties-alias.md "#/properties/spec/properties/alias#/definitions/requireEntity/properties/alias")

### revision

The exact revision of the given Type.

`revision`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](implementation-definitions-requireentity-properties-revision.md "#/properties/spec/properties/revision#/definitions/requireEntity/properties/revision")
