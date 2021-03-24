# Untitled object in undefined Schema

```txt
undefined
```

InterfaceGroup stores metadata for a group of Interfaces.

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                             |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Allowed               | none                | [interface-group.json](../../0.0.1/schema/interface-group.json "open original schema") |

# Untitled object in undefined Properties

| Property                  | Type     | Required | Nullable       | Defined by                                                                                                                             |
| :------------------------ | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------- |
| [ocfVersion](#ocfversion) | `string` | Required | cannot be null | [Untitled schema](interface-group-properties-ocfversion.md "#/properties/ocfVersion#/properties/ocfVersion")                           |
| [kind](#kind)             | `string` | Required | cannot be null | [Untitled schema](interface-group-properties-kind.md "#/properties/kind#/properties/kind")                                             |
| [signature](#signature)   | `object` | Required | cannot be null | [Untitled schema](interface-group-properties-signature.md "#/properties/signature#/properties/signature")                              |
| [metadata](#metadata)     | `object` | Required | cannot be null | [Untitled schema](attribute-properties-ocf-metadata.md "https://projectvoltron.dev/schemas/common/metadata.json#/properties/metadata") |

## ocfVersion



`ocfVersion`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [Untitled schema](interface-group-properties-ocfversion.md "#/properties/ocfVersion#/properties/ocfVersion")

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

*   defined in: [Untitled schema](interface-group-properties-kind.md "#/properties/kind#/properties/kind")

### kind Constraints

**enum**: the value of this property must be equal to one of the following values:

| Value              | Explanation |
| :----------------- | :---------- |
| `"InterfaceGroup"` |             |

## signature

Ensures the authenticity and integrity of a given manifest.

`signature`

*   is required

*   Type: `object` ([Details](interface-group-properties-signature.md))

*   cannot be null

*   defined in: [Untitled schema](interface-group-properties-signature.md "#/properties/signature#/properties/signature")

## metadata

A container for the OCF metadata definitions.

`metadata`

*   is required

*   Type: `object` ([OCF Metadata](attribute-properties-ocf-metadata.md))

*   cannot be null

*   defined in: [Untitled schema](attribute-properties-ocf-metadata.md "https://projectvoltron.dev/schemas/common/metadata.json#/properties/metadata")
