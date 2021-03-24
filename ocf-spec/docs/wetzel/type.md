> WETZEL_WARNING: Only JSON Schema 3 or 4 is supported. Treating as Schema 3.

## Objects
* [`OCF Metadata`](#reference-common/metadata-json)
* [`undefined`](#reference-common/json-schema-type-json)
   * [`OCF Metadata`](#reference-common/metadata-json)


---------------------------------------
<a name="reference-common/metadata-json"></a>
### OCF Metadata

A container for the OCF metadata definitions.

**`OCF Metadata` Properties**

|   |Type|Description|Required|
|---|---|---|---|
|**name**|`string`|The name of OCF manifest that uniquely identifies this object within the entity sub-tree. Must be a non-empty string. We recommend using a CLI-friendly name.|No|
|**prefix**|`string`|The prefix value is automatically computed and set when storing manifest in OCH.|No|
|**displayName**|`string`|The name of the OCF manifest to be displayed in graphical clients.|No|
|**description**|`string`|A short description of the OCF manifest. Must be a non-empty string.|No|
|**maintainers**|`object` `[1-*]`|The list of maintainers with contact information.|No|
|**documentationURL**|`string`|Link to documentation page for the OCF manifest.|No|
|**supportURL**|`string`|Link to support page for the OCF manifest.|No|
|**iconURL**|`string`|The URL to an icon or a data URL containing an icon.|No|

Additional properties are allowed.

#### common/metadata.json.name

The name of OCF manifest that uniquely identifies this object within the entity sub-tree. Must be a non-empty string. We recommend using a CLI-friendly name.

* **Type**: `string`
* **Required**: No
* **Examples**:
   * `"config"`

#### common/metadata.json.prefix

The prefix value is automatically computed and set when storing manifest in OCH.

* **Type**: `string`
* **Required**: No
* **Examples**:
   * `"cap.type.database.mysql"`

#### common/metadata.json.displayName

The name of the OCF manifest to be displayed in graphical clients.

* **Type**: `string`
* **Required**: No
* **Examples**:
   * `"MySQL Config"`

#### common/metadata.json.description

A short description of the OCF manifest. Must be a non-empty string.

* **Type**: `string`
* **Required**: No

#### common/metadata.json.maintainers

The list of maintainers with contact information.

* **Type**: `object` `[1-*]`
* **Required**: No
* **Examples**:
   * `[object Object],[object Object]`

#### common/metadata.json.documentationURL

Link to documentation page for the OCF manifest.

* **Type**: `string`
* **Required**: No
* **Format**: uri
* **Examples**:
   * `"https://example.com/docs"`

#### common/metadata.json.supportURL

Link to support page for the OCF manifest.

* **Type**: `string`
* **Required**: No
* **Format**: uri
* **Examples**:
   * `"https://example.com/online-support"`

#### common/metadata.json.iconURL

The URL to an icon or a data URL containing an icon.

* **Type**: `string`
* **Required**: No
* **Format**: uri
* **Examples**:
   * `"https://example.com/favicon.ico"`




---------------------------------------
<a name="reference-common/json-schema-type-json"></a>
### WETZEL_WARNING: title not defined

The JSONSchema definition.

** Properties**

|   |Type|Description|Required|
|---|---|---|---|
|**value**|`string`|Inline JSON Schema definition for the parameters.|No|

Additional properties are not allowed.

#### common/json-schema-type.json.value

Inline JSON Schema definition for the parameters.

* **Type**: `string`
* **Required**: No


