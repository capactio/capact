# Untitled object in OCF Metadata Schema

```txt
#/properties/metadata/properties/maintainers/items#/properties/maintainers/items
```

Holds contact information.

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                       |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Allowed               | none                | [metadata.json*](../../0.0.1/schema/common/metadata.json "open original schema") |

## items Examples

```yaml
email: foo@example.com
name: Foo Bar
url: https://example.com

```

# items Properties

| Property        | Type     | Required | Nullable       | Defined by                                                                                                                                                                                             |
| :-------------- | :------- | :------- | :------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [email](#email) | `string` | Required | cannot be null | [OCF Metadata](metadata-properties-maintainers-items-properties-email.md "#/properties/metadata/properties/maintainers/items/anyOf/0/properties/email#/properties/maintainers/items/properties/email") |
| [name](#name)   | `string` | Optional | cannot be null | [OCF Metadata](metadata-properties-maintainers-items-properties-name.md "#/properties/metadata/properties/maintainers/items/anyOf/0/properties/name#/properties/maintainers/items/properties/name")    |
| [url](#url)     | `string` | Optional | cannot be null | [OCF Metadata](metadata-properties-maintainers-items-properties-url.md "#/properties/metadata/properties/maintainers/items/anyOf/0/properties/url#/properties/maintainers/items/properties/url")       |

## email

Email address of the person.

`email`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-maintainers-items-properties-email.md "#/properties/metadata/properties/maintainers/items/anyOf/0/properties/email#/properties/maintainers/items/properties/email")

### email Constraints

**email**: the string must be an email address, according to [RFC 5322, section 3.4.1](https://tools.ietf.org/html/rfc5322 "check the specification")

## name

Name of the person.

`name`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-maintainers-items-properties-name.md "#/properties/metadata/properties/maintainers/items/anyOf/0/properties/name#/properties/maintainers/items/properties/name")

## url

URL of the person’s site.

`url`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-maintainers-items-properties-url.md "#/properties/metadata/properties/maintainers/items/anyOf/0/properties/url#/properties/maintainers/items/properties/url")

### url Constraints

**IRI**: the string must be a IRI, according to [RFC 3987](https://tools.ietf.org/html/rfc3987 "check the specification")
