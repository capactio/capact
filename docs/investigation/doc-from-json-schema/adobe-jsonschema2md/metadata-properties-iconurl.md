# Untitled string in OCF Metadata Schema

```txt
#/properties/metadata/properties/iconURL#/properties/iconURL
```

The URL to an icon or a data URL containing an icon.

| Abstract            | Extensible | Status         | Identifiable            | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                       |
| :------------------ | :--------- | :------------- | :---------------------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | Unknown identifiability | Forbidden         | Allowed               | none                | [metadata.json*](../../../../ocf-spec/0.0.1/schema/common/metadata.json "open original schema") |

## iconURL Constraints

**URI**: the string must be a URI, according to [RFC 3986](https://tools.ietf.org/html/rfc3986 "check the specification")

## iconURL Examples

```yaml
https://example.com/favicon.ico

```
