# OCF Metadata

_A container for the OCF metadata definitions._

Type: `object`

<i id="#https://capact.io/schemas/common/metadata.json">path: #https://capact.io/schemas/common/metadata.json</i>

&#36;schema: [http://json-schema.org/draft-07/schema#](http://json-schema.org/draft-07/schema#)

<b id="httpscapact.ioschemascommonmetadata.json">&#36;id: https://capact.io/schemas/common/metadata.json</b>

**_Properties_**

 - <b id="#https://capact.io/schemas/common/metadata.json/properties/name">name</b> `required`
	 - _The name of OCF manifest that uniquely identifies this object within the entity sub-tree. Must be a non-empty string. We recommend using a CLI-friendly name._
	 - Type: `string`
	 - <i id="##/properties/metadata/properties/name">path: ##/properties/metadata/properties/name</i>
	 - <b id="propertiesmetadatapropertiesname">&#36;id: #/properties/metadata/properties/name</b>
	 - Example values: 
		 1. _"config"_
 - <b id="#https://capact.io/schemas/common/metadata.json/properties/prefix">prefix</b>
	 - _The prefix value is automatically computed and set when storing manifest in OCH._
	 - Type: `string`
	 - <i id="##/properties/metadata/properties/prefix">path: ##/properties/metadata/properties/prefix</i>
	 - <b id="propertiesmetadatapropertiesprefix">&#36;id: #/properties/metadata/properties/prefix</b>
	 - **Comment**<br/>_Value set by user is ignored and this field is always managed by OCH_
	 - Example values: 
		 1. _"cap.type.database.mysql"_
 - <b id="#https://capact.io/schemas/common/metadata.json/properties/displayName">displayName</b>
	 - _The name of the OCF manifest to be displayed in graphical clients._
	 - Type: `string`
	 - <i id="##/properties/metadata/properties/displayName">path: ##/properties/metadata/properties/displayName</i>
	 - <b id="propertiesmetadatapropertiesdisplayname">&#36;id: #/properties/metadata/properties/displayName</b>
	 - Example values: 
		 1. _"MySQL Config"_
 - <b id="#https://capact.io/schemas/common/metadata.json/properties/description">description</b> `required`
	 - _A short description of the OCF manifest. Must be a non-empty string._
	 - Type: `string`
	 - <i id="##/properties/metadata/properties/description">path: ##/properties/metadata/properties/description</i>
	 - <b id="propertiesmetadatapropertiesdescription">&#36;id: #/properties/metadata/properties/description</b>
 - <b id="#https://capact.io/schemas/common/metadata.json/properties/maintainers">maintainers</b> `required`
	 - _The list of maintainers with contact information._
	 - Type: `array`
	 - <i id="##/properties/metadata/properties/maintainers">path: ##/properties/metadata/properties/maintainers</i>
	 - <b id="propertiesmetadatapropertiesmaintainers">&#36;id: #/properties/metadata/properties/maintainers</b>
	 - Example values: 
		 1. `[object Object],[object Object]`
 - This schema accepts additional items.
	 - Item Count:  &ge; 1
		 - **_Items_**
		 - _Holds contact information._
		 - Type: `object`
		 - <i id="##/properties/metadata/properties/maintainers/items">path: ##/properties/metadata/properties/maintainers/items</i>
		 - <b id="propertiesmetadatapropertiesmaintainersitems">&#36;id: #/properties/metadata/properties/maintainers/items</b>
		 - Example values: 
			 1. `[object Object]`
		 - **_Properties_**
			 - <b id="##/properties/metadata/properties/maintainers/items/properties/email">email</b> `required`
				 - _Email address of the person._
				 - Type: `string`
				 - <i id="##/properties/metadata/properties/maintainers/items/anyOf/0/properties/email">path: ##/properties/metadata/properties/maintainers/items/anyOf/0/properties/email</i>
				 - <b id="propertiesmetadatapropertiesmaintainersitemsanyof0propertiesemail">&#36;id: #/properties/metadata/properties/maintainers/items/anyOf/0/properties/email</b>
				 - String format must be a "email"
			 - <b id="##/properties/metadata/properties/maintainers/items/properties/name">name</b>
				 - _Name of the person._
				 - Type: `string`
				 - <i id="##/properties/metadata/properties/maintainers/items/anyOf/0/properties/name">path: ##/properties/metadata/properties/maintainers/items/anyOf/0/properties/name</i>
				 - <b id="propertiesmetadatapropertiesmaintainersitemsanyof0propertiesname">&#36;id: #/properties/metadata/properties/maintainers/items/anyOf/0/properties/name</b>
			 - <b id="##/properties/metadata/properties/maintainers/items/properties/url">url</b>
				 - _URL of the person’s site._
				 - Type: `string`
				 - <i id="##/properties/metadata/properties/maintainers/items/anyOf/0/properties/url">path: ##/properties/metadata/properties/maintainers/items/anyOf/0/properties/url</i>
				 - <b id="propertiesmetadatapropertiesmaintainersitemsanyof0propertiesurl">&#36;id: #/properties/metadata/properties/maintainers/items/anyOf/0/properties/url</b>
				 - String format must be a "iri"
 - <b id="#https://capact.io/schemas/common/metadata.json/properties/documentationURL">documentationURL</b>
	 - _Link to documentation page for the OCF manifest._
	 - Type: `string`
	 - <i id="##/properties/metadata/properties/documentationURL">path: ##/properties/metadata/properties/documentationURL</i>
	 - <b id="propertiesmetadatapropertiesdocumentationurl">&#36;id: #/properties/metadata/properties/documentationURL</b>
	 - Example values: 
		 1. _"https://example.com/docs"_
	 - String format must be a "uri"
 - <b id="#https://capact.io/schemas/common/metadata.json/properties/supportURL">supportURL</b>
	 - _Link to support page for the OCF manifest._
	 - Type: `string`
	 - <i id="##/properties/metadata/properties/supportURL">path: ##/properties/metadata/properties/supportURL</i>
	 - <b id="propertiesmetadatapropertiessupporturl">&#36;id: #/properties/metadata/properties/supportURL</b>
	 - Example values: 
		 1. _"https://example.com/online-support"_
	 - String format must be a "uri"
 - <b id="#https://capact.io/schemas/common/metadata.json/properties/iconURL">iconURL</b>
	 - _The URL to an icon or a data URL containing an icon._
	 - Type: `string`
	 - <i id="##/properties/metadata/properties/iconURL">path: ##/properties/metadata/properties/iconURL</i>
	 - <b id="propertiesmetadatapropertiesiconurl">&#36;id: #/properties/metadata/properties/iconURL</b>
	 - Example values: 
		 1. _"https://example.com/favicon.ico"_
	 - String format must be a "uri"

_Generated with [json-schema-md-doc](https://brianwendt.github.io/json-schema-md-doc/)_
