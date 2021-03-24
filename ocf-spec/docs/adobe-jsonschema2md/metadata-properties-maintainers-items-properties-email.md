# Untitled string in OCF Metadata Schema

```txt
#/properties/metadata/properties/maintainers/items/anyOf/0/properties/email#/properties/maintainers/items/properties/email
```

Email address of the person.

| Abstract            | Extensible | Status         | Identifiable            | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                       |
| :------------------ | :--------- | :------------- | :---------------------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | Unknown identifiability | Forbidden         | Allowed               | none                | [metadata.json*](../../0.0.1/schema/common/metadata.json "open original schema") |

## email Constraints

**email**: the string must be an email address, according to [RFC 5322, section 3.4.1](https://tools.ietf.org/html/rfc5322 "check the specification")
