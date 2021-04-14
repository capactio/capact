# Untitled schema Schema

```txt
#/properties/metadata#/properties/metadata/allOf/1
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                            |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Allowed               | none                | [implementation.json*](../../../../ocf-spec/0.0.1/schema/implementation.json "open original schema") |

# 1 Properties

| Property                  | Type     | Required | Nullable       | Defined by                                                                                                                                                                                                     |
| :------------------------ | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [attributes](#attributes) | `object` | Optional | cannot be null | [Untitled schema](implementation-properties-metadata-allof-1-properties-attributes.md "https://capact.io/schemas/common/metadata-attributes.json#/properties/metadata/allOf/1/properties/attributes") |
| [license](#license)       | Merged   | Required | cannot be null | [Untitled schema](implementation-properties-metadata-allof-1-properties-license.md "#/properties/metadata/properties/license#/properties/metadata/allOf/1/properties/license")                                 |

## attributes

Object that holds Attributes references.

`attributes`

*   is optional

*   Type: `object` ([Details](implementation-properties-metadata-allof-1-properties-attributes.md))

*   cannot be null

*   defined in: [Untitled schema](implementation-properties-metadata-allof-1-properties-attributes.md "https://capact.io/schemas/common/metadata-attributes.json#/properties/metadata/allOf/1/properties/attributes")

## license

This entry allows you to specify a license, so people know how they are permitted to use it, and what kind of restrictions you are placing on it.

`license`

*   is required

*   Type: `object` ([Details](implementation-properties-metadata-allof-1-properties-license.md))

*   cannot be null

*   defined in: [Untitled schema](implementation-properties-metadata-allof-1-properties-license.md "#/properties/metadata/properties/license#/properties/metadata/allOf/1/properties/license")
