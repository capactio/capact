# Untitled array in undefined Schema

```txt
#/properties/spec/properties/additionalRefs#/properties/spec/properties/additionalRefs
```

List of the full path of additional parent nodes the Attribute is attached to. The parent nodes MUST reside under “cap.core.attribute” or “cap.attribute” subtree. The connection means that the Attribute becomes a child of the referenced parent nodes. In a result, the Attribute has multiple parents.

| Abstract            | Extensible | Status         | Identifiable            | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                  |
| :------------------ | :--------- | :------------- | :---------------------- | :---------------- | :-------------------- | :------------------ | :-------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | Unknown identifiability | Forbidden         | Forbidden             | none                | [attribute.json*](../../../../ocf-spec/0.0.1/schema/attribute.json "open original schema") |

## additionalRefs Constraints

**unique items**: all items in this array must be unique. Duplicates are not allowed.
