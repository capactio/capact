# Untitled array in OCF Metadata Schema

```txt
#/properties/metadata/properties/maintainers#/properties/maintainers
```

The list of maintainers with contact information.

| Abstract            | Extensible | Status         | Identifiable            | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                                       |
| :------------------ | :--------- | :------------- | :---------------------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | Unknown identifiability | Forbidden         | Allowed               | none                | [metadata.json*](../../../../ocf-spec/0.0.1/schema/common/metadata.json "open original schema") |

## maintainers Constraints

**minimum number of items**: the minimum number of items for this array is: `1`

## maintainers Examples

```yaml
- email: foo@example.com
  name: Foo Bar
  url: https://foo.bar
- email: foo@example.com
  name: Foo Bar
  url: https://foo.bar

```
