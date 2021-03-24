# Untitled object in undefined Schema

```txt
#/properties/spec#/properties/spec
```

A container for the Attribute specification definition.

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                  |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :-------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [attribute.json*](../../../../ocf-spec/0.0.1/schema/attribute.json "open original schema") |

# spec Properties

| Property                          | Type    | Required | Nullable       | Defined by                                                                                                                                                         |
| :-------------------------------- | :------ | :------- | :------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [additionalRefs](#additionalrefs) | `array` | Optional | cannot be null | [Untitled schema](attribute-properties-spec-properties-additionalrefs.md "#/properties/spec/properties/additionalRefs#/properties/spec/properties/additionalRefs") |

## additionalRefs

List of the full path of additional parent nodes the Attribute is attached to. The parent nodes MUST reside under “cap.core.attribute” or “cap.attribute” subtree. The connection means that the Attribute becomes a child of the referenced parent nodes. In a result, the Attribute has multiple parents.

`additionalRefs`

*   is optional

*   Type: `string[]`

*   cannot be null

*   defined in: [Untitled schema](attribute-properties-spec-properties-additionalrefs.md "#/properties/spec/properties/additionalRefs#/properties/spec/properties/additionalRefs")

### additionalRefs Constraints

**unique items**: all items in this array must be unique. Duplicates are not allowed.
