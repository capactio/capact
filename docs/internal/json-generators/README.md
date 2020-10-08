# JSON Schema generators

This document describes the available libraries to generated Go Struct from JSON Schema, their pros and cons. 

we are usign the const, one, ... patternProperties

- get rid of the title, if present then it is sued for 

I've reviewed paweł pull request, and we also have a discussion about the rendering flow
I've finished creating JSON Schemas for all OCF manifests
Currently, I'm doing research. 
Currently, we cannot find any library to support the `patternProperties`, `const`, `allOf`, `anyOf` and `oneOf` features. The possible workaround it to:
- convert JSON Schema to protobufs and then to Go struct
- convert JSON Schema to Open API 3.0 and then to Go struct
- get rid of more complex statements (probably not possible with PatternProperties)
- fork a given lib and add adjust for our needs
- generate types as they are and just adjust them manually, 



### Libraries JSON Schema -> Go struct

This section describes the open source libraries that are able to convert JSON Schema to Go struct.

## Schematic

Github: https://github.com/interagent/schematic/
 
>**NOTE:** Does not support references. This library focuses on client generation and not the Go structs.

## go-jsonschema

Github: https://github.com/atombender/go-jsonschema       
Last release: v0.7.0 on Mar 12, 2019
Stars: 148
Last update: Mar 29, 2020

#### Pros 
    1. The generated boilerplate struct is quschema-generateite small.  
### Cons 
    1. Fail fast with unmarshaling (if first element is not provided then rest are not checked)
    2. `additionalItems` field not supported
       ```
       json: cannot unmarshal bool into Go struct field Type.properties.allOf.properties.properties.additionalItems of type schemas.Type
       ```
    3. Does not support the `allOf`, `anyOf` and `oneOf` features
    4. Does not support additionalProperties object

## generate 

Github: https://github.com/a-h/generate
Last release: none
Stars: 221
Last update: Feb 4, 2019

### Pros
    1. Checks for all validation problems instead of failing fast with unmarshaling (if first element is not provided then rest are still checked)
### Cons 
    1. Generates a lof of boilerplate
    2. Struct are named by the description  
    3. Does not support the `allOf`, `anyOf` and `oneOf` features
    4. Does not support enum and const
    5. Uses title for naming types

## quicktype

### Pros
    1. Checks for all validation problems instead of failing fast with unmarshaling (if first element is not provided then rest are still checked)
### Cons 
    1. Do not generate the validation rules and unmarshaling
    2. Uses title for naming types
    3. Does not support `const`
    4. OneOf generates the struct with both fields as a pointer. no validator that both cannot be set, e.g.
    ```
    "license": {
                  "$id": "#/properties/metadata/properties/license",
                  "type": "object",
                  "description": "This entry allows you to specify a license, so people know how they are permitted to use it, and what kind of restrictions you are placing on it.",
                  "oneOf": [
                    {
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "name": {
                          "$id": "#/properties/metadata/properties/license/name",
                          "type": "string",
                          "description": "If you are using a common license such as BSD-2-Clause or MIT, add a current SPDX license identifier for the license you’re using e.g. BSD-3-Clause. If your package is licensed under multiple common licenses, use an SPDX license expression syntax version 2.0 string, e.g. (ISC OR GPL-3.0)"
                        }
                      }
                    },
                    {
                      "required": [
                        "ref"
                      ],
                      "properties": {
                        "ref": {
                          "$id": "#/properties/metadata/properties/license/ref",
                          "type": "string",
                          "description": "If you are using a license that hasn’t been assigned an SPDX identifier, or if you are using a custom license, use the direct link to the license file e.g. https://raw.githubusercontent.com/project/v1/license.md. The resource under given link MUST be immutable and publicly accessible."
                        }
                      }
                    }
                  ]
                }
    ```
    ```
    type License struct {
    	Name *string `json:"name,omitempty"`// If you are using a common license such as BSD-2-Clause or MIT, add a current SPDX license; identifier for the license you’re using e.g. BSD-3-Clause. If your package is licensed; under multiple common licenses, use an SPDX license expression syntax version 2.0 string,; e.g. (ISC OR GPL-3.0)
    	Ref  *string `json:"ref,omitempty"` // If you are using a license that hasn’t been assigned an SPDX identifier, or if you are; using a custom license, use the direct link to the license file e.g.; https://raw.githubusercontent.com/project/v1/license.md. The resource under given link; MUST be immutable and publicly accessible.
    }
    ```

----------

Open API 3.0

https://github.com/openapi-contrib/json-schema-to-openapi-schema
https://github.com/OpenAPITools/openapi-generator
