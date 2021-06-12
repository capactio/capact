# Untitled object in undefined Schema

```txt
https://capact.io/schemas/common/metadata-attributes.json#/properties/metadata/allOf/1/properties/attributes
```

Object that holds Attributes references.

| Abstract            | Extensible | Status         | Identifiable            | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                            |
| :------------------ | :--------- | :------------- | :---------------------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | Unknown identifiability | Forbidden         | Allowed               | none                | [implementation.json*](../../../../ocf-spec/0.0.1/schema/implementation.json "open original schema") |

# attributes Properties

| Property              | Type     | Required | Nullable       | Defined by                                                                                                                                                |
| :-------------------- | :------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Additional Properties | `object` | Optional | cannot be null | [Untitled schema](metadata-attributes-additionalproperties.md "https://capact.io/schemas/common/metadata-attributes.json#/additionalProperties") |

## Additional Properties

Additional properties are allowed, as long as they follow this schema:

The attribute object contains OCF Attributes references. It provides generic categorization for Implementations, Types and TypeInstances. Attributes are used to filter out a specific Implementation.

*   is optional

*   Type: `object` ([Details](metadata-attributes-additionalproperties.md))

*   cannot be null

*   defined in: [Untitled schema](metadata-attributes-additionalproperties.md "https://capact.io/schemas/common/metadata-attributes.json#/additionalProperties")
