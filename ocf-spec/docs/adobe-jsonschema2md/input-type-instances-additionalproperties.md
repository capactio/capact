# Untitled object in undefined Schema

```txt
#/additionalParameters#/additionalProperties
```

Prefix is an alias of the TypeInstance, used in the Implementation

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                                               |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [input-type-instances.json*](../../0.0.1/schema/common/input-type-instances.json "open original schema") |

# additionalProperties Properties

| Property            | Type     | Required | Nullable       | Defined by                                                                                                                                                                            |
| :------------------ | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [typeRef](#typeref) | `object` | Required | cannot be null | [Untitled schema](input-type-instances-additionalproperties-properties-typeref.md "https://projectvoltron.dev/schemas/common/type-ref.json#/additionalProperties/properties/typeRef") |
| [verbs](#verbs)     | `array`  | Required | cannot be null | [Untitled schema](input-type-instances-additionalproperties-properties-verbs.md "#/additionalParameters/verbs#/additionalProperties/properties/verbs")                                |

## typeRef

The full path to the Type from which the TypeInstance is created.

`typeRef`

*   is required

*   Type: `object` ([Details](input-type-instances-additionalproperties-properties-typeref.md))

*   cannot be null

*   defined in: [Untitled schema](input-type-instances-additionalproperties-properties-typeref.md "https://projectvoltron.dev/schemas/common/type-ref.json#/additionalProperties/properties/typeRef")

## verbs

The full list of access rights for a given TypeInstance

`verbs`

*   is required

*   Type: `string[]`

*   cannot be null

*   defined in: [Untitled schema](input-type-instances-additionalproperties-properties-verbs.md "#/additionalParameters/verbs#/additionalProperties/properties/verbs")
