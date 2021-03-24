# Untitled object in undefined Schema

```txt
https://projectvoltron.dev/schemas/implementation.json#/definitions/requireEntity
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                            |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [implementation.json*](../../../../ocf-spec/0.0.1/schema/implementation.json "open original schema") |

# requireEntity Properties

| Property                              | Type     | Required | Nullable       | Defined by                                                                                                                                                                             |
| :------------------------------------ | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [valueConstraints](#valueconstraints) | `object` | Optional | cannot be null | [Untitled schema](implementation-definitions-requireentity-properties-valueconstraints.md "#/properties/spec/properties/value#/definitions/requireEntity/properties/valueConstraints") |
| [name](#name)                         | `string` | Required | cannot be null | [Untitled schema](implementation-definitions-requireentity-properties-name.md "#/properties/spec/properties/name#/definitions/requireEntity/properties/name")                          |
| [alias](#alias)                       | `string` | Optional | cannot be null | [Untitled schema](implementation-definitions-requireentity-properties-alias.md "#/properties/spec/properties/alias#/definitions/requireEntity/properties/alias")                       |
| [revision](#revision)                 | `string` | Required | cannot be null | [Untitled schema](implementation-definitions-requireentity-properties-revision.md "#/properties/spec/properties/revision#/definitions/requireEntity/properties/revision")              |

## valueConstraints

Holds the configuration constraints for the given entry. It needs to be valid against the Type JSONSchema.

`valueConstraints`

*   is optional

*   Type: `object` ([Details](implementation-definitions-requireentity-properties-valueconstraints.md))

*   cannot be null

*   defined in: [Untitled schema](implementation-definitions-requireentity-properties-valueconstraints.md "#/properties/spec/properties/value#/definitions/requireEntity/properties/valueConstraints")

## name

The name of the Type. Root prefix can be skipped if it’s a core Type. If it is a custom Type then it MUST be defined as full path to that Type. Custom Type MUST extend the abstract node which is defined as a root prefix for that entry.

`name`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](implementation-definitions-requireentity-properties-name.md "#/properties/spec/properties/name#/definitions/requireEntity/properties/name")

## alias

If provided, the TypeInstance of the Type, configured in policy, is injected to the workflow under the alias.

`alias`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](implementation-definitions-requireentity-properties-alias.md "#/properties/spec/properties/alias#/definitions/requireEntity/properties/alias")

## revision

The exact revision of the given Type.

`revision`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](implementation-definitions-requireentity-properties-revision.md "#/properties/spec/properties/revision#/definitions/requireEntity/properties/revision")
