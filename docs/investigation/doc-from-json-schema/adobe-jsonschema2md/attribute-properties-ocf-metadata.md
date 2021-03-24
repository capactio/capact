# OCF Metadata Schema

```txt
https://projectvoltron.dev/schemas/common/metadata.json#/properties/metadata
```

A container for the OCF metadata definitions.

| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                  |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :-------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Allowed               | none                | [attribute.json*](../../../../ocf-spec/0.0.1/schema/attribute.json "open original schema") |

# metadata Properties

| Property                              | Type     | Required | Nullable       | Defined by                                                                                                                               |
| :------------------------------------ | :------- | :------- | :------------- | :--------------------------------------------------------------------------------------------------------------------------------------- |
| [name](#name)                         | `string` | Required | cannot be null | [OCF Metadata](metadata-properties-name.md "#/properties/metadata/properties/name#/properties/name")                                     |
| [prefix](#prefix)                     | `string` | Optional | cannot be null | [OCF Metadata](metadata-properties-prefix.md "#/properties/metadata/properties/prefix#/properties/prefix")                               |
| [displayName](#displayname)           | `string` | Optional | cannot be null | [OCF Metadata](metadata-properties-displayname.md "#/properties/metadata/properties/displayName#/properties/displayName")                |
| [description](#description)           | `string` | Required | cannot be null | [OCF Metadata](metadata-properties-description.md "#/properties/metadata/properties/description#/properties/description")                |
| [maintainers](#maintainers)           | `array`  | Required | cannot be null | [OCF Metadata](metadata-properties-maintainers.md "#/properties/metadata/properties/maintainers#/properties/maintainers")                |
| [documentationURL](#documentationurl) | `string` | Optional | cannot be null | [OCF Metadata](metadata-properties-documentationurl.md "#/properties/metadata/properties/documentationURL#/properties/documentationURL") |
| [supportURL](#supporturl)             | `string` | Optional | cannot be null | [OCF Metadata](metadata-properties-supporturl.md "#/properties/metadata/properties/supportURL#/properties/supportURL")                   |
| [iconURL](#iconurl)                   | `string` | Optional | cannot be null | [OCF Metadata](metadata-properties-iconurl.md "#/properties/metadata/properties/iconURL#/properties/iconURL")                            |

## name

The name of OCF manifest that uniquely identifies this object within the entity sub-tree. Must be a non-empty string. We recommend using a CLI-friendly name.

`name`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-name.md "#/properties/metadata/properties/name#/properties/name")

### name Examples

```yaml
config

```

## prefix

The prefix value is automatically computed and set when storing manifest in OCH.

> Value set by user is ignored and this field is always managed by OCH

`prefix`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-prefix.md "#/properties/metadata/properties/prefix#/properties/prefix")

### prefix Examples

```yaml
cap.type.database.mysql

```

## displayName

The name of the OCF manifest to be displayed in graphical clients.

`displayName`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-displayname.md "#/properties/metadata/properties/displayName#/properties/displayName")

### displayName Examples

```yaml
MySQL Config

```

## description

A short description of the OCF manifest. Must be a non-empty string.

`description`

*   is required

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-description.md "#/properties/metadata/properties/description#/properties/description")

## maintainers

The list of maintainers with contact information.

`maintainers`

*   is required

*   Type: `object[]` ([Details](metadata-properties-maintainers-items.md))

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-maintainers.md "#/properties/metadata/properties/maintainers#/properties/maintainers")

### maintainers Constraints

**minimum number of items**: the minimum number of items for this array is: `1`

### maintainers Examples

```yaml
- email: foo@example.com
  name: Foo Bar
  url: https://foo.bar
- email: foo@example.com
  name: Foo Bar
  url: https://foo.bar

```

## documentationURL

Link to documentation page for the OCF manifest.

`documentationURL`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-documentationurl.md "#/properties/metadata/properties/documentationURL#/properties/documentationURL")

### documentationURL Constraints

**URI**: the string must be a URI, according to [RFC 3986](https://tools.ietf.org/html/rfc3986 "check the specification")

### documentationURL Examples

```yaml
https://example.com/docs

```

## supportURL

Link to support page for the OCF manifest.

`supportURL`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-supporturl.md "#/properties/metadata/properties/supportURL#/properties/supportURL")

### supportURL Constraints

**URI**: the string must be a URI, according to [RFC 3986](https://tools.ietf.org/html/rfc3986 "check the specification")

### supportURL Examples

```yaml
https://example.com/online-support

```

## iconURL

The URL to an icon or a data URL containing an icon.

`iconURL`

*   is optional

*   Type: `string`

*   cannot be null

*   defined in: [OCF Metadata](metadata-properties-iconurl.md "#/properties/metadata/properties/iconURL#/properties/iconURL")

### iconURL Constraints

**URI**: the string must be a URI, according to [RFC 3986](https://tools.ietf.org/html/rfc3986 "check the specification")

### iconURL Examples

```yaml
https://example.com/favicon.ico

```
