const convert = require('@openapi-contrib/json-schema-to-openapi-schema');

const schema = {
        "$schema": "http://json-schema.org/draft-07/schema",
        "$id": "https://capact.io/schemas/implementation.json",
        "type": "object",
        "title": "The OCF Type manifest schema",
        "description": "Primitive, that holds the JSONSchema which describes that Type. It’s also used for validation. There are core and custom Types. Type can be also a composition of other Types.",
        "definitions": {
            "requireEntity": {
                "type": "object",
                "required": [
                    "name",
                    "revision"
                ],
                "properties": {
                    "value": {
                        "$id": "#/properties/spec/properties/requires/properties/cap.core.type.platform/properties/oneOf/items/anyOf/0/properties/constraints",
                        "type": "object",
                        "title": "The value schema",
                        "description": "Holds the configuration constraints for the given entry. It needs to be valid against the Type JSONSchema."
                    },
                    "name": {
                        "$id": "#/properties/spec/properties/requires/properties/cap.core.type.platform/properties/oneOf/items/anyOf/0/properties/name",
                        "type": "string",
                        "title": "The name schema",
                        "description": "The name of the Type. Root prefix can be skipped if it’s a core Type. If it is a custom Type then it MUST be defined as full path to that Type. Custom Type MUST extend the abstract node which is defined as a root prefix for that entry."
                    },
                    "revision": {
                        "$id": "#/properties/spec/properties/requires/properties/cap.core.type.platform/properties/oneOf/items/anyOf/0/properties/revision",
                        "type": "string",
                        "title": "The revision schema",
                        "description": "The revision version of the given Type."
                    }
                },
                "additionalProperties": false
            }
        },
        "required": [
            "metadata",
            "spec",
            "revision",
            "signature",
            "ocfVersion",
            "kind"
        ],
        "properties": {
            "ocfVersion": {
                "$id": "#/properties/ocfVersion",
                "type": "string",
                "const": "0.0.1"
            },
            "kind": {
                "$comment": "TODO: How to restrict kind to a given name",
                "$id": "#/properties/kind",
                "type": "string",
                "title": "The OCF manifest kind",
                "enum": [
                    "Implementation"
                ]
            },
            "revision": {
                "$id": "#/properties/revision",
                "type": "string",
                "minLength": 5,
                "pattern": "^(?:0|[1-9]\\d*)\\.(?:0|[1-9]\\d*)\\.(?:0|[1-9]\\d*)$",
                "title": "Version of the manifest content in the SemVer format.",
                "description": "Version of the manifest content in the SemVer format."
            },
            "signature": {
                "$id": "#/properties/signature",
                "type": "object",
                "title": "Ensures the authenticity and integrity of a given manifest.",
                "description": "Ensures the authenticity and integrity of a given manifest.",
                "required": [
                    "och"
                ],
                "properties": {
                    "och": {
                        "type": "string"
                    }
                }
            },
            "metadata": {
                "$id": "#/properties/metadata",
                "type": "object",
                "title": "The metadata schema",
                "required": [
                    "license",
                    "name",
                    "description",
                    "maintainers"
                ],
                "properties": {
                    "name": {
                        "$id": "#/properties/metadata/properties/name",
                        "type": "string",
                        "title": "The name of OCF manifest.",
                        "description": "The name of OCF manifest that uniquely identifies this object within the entity sub-tree. Must be a non-empty string. We recommend using a CLI-friendly name.",
                        "examples": [
                            "config"
                        ]
                    },
                    "prefix": {
                        "$comment": "TODO: How to restrict prefix to be set only by controller?",
                        "$id": "#/properties/metadata/properties/prefix",
                        "type": "string",
                        "title": "The prefix value is automatically set when storing manifest in OCH.",
                        "description": "The prefix value is automatically computed and set when storing manifest in OCH.",
                        "examples": [
                            "cap.type.database.mysql"
                        ]
                    },
                    "displayName": {
                        "$id": "#/properties/metadata/properties/displayName",
                        "type": "string",
                        "title": "The display name of the OCF manifest.",
                        "description": "The name of the OCF manifest to be displayed in graphical clients.",
                        "examples": [
                            "MySQL Config"
                        ]
                    },
                    "description": {
                        "$id": "#/properties/metadata/properties/description",
                        "type": "string",
                        "title": "A short description.",
                        "description": "A short description of the OCF manifest. Must be a non-empty string."
                    },
                    "maintainers": {
                        "$id": "#/properties/metadata/properties/maintainers",
                        "type": "array",
                        "title": "The maintainers schema",
                        "description": "The list of maintainers with contact information.",
                        "examples": [
                            [
                                {
                                    "email": "foo@example.com",
                                    "name": "Foo Bar",
                                    "url": "https://foo.bar"
                                },
                                {
                                    "email": "foo@example.com",
                                    "name": "Foo Bar",
                                    "url": "https://foo.bar"
                                }
                            ]
                        ],
                        "minItems": 1,
                        "items": {
                            "$id": "#/properties/metadata/properties/maintainers/items",
                            "type": "object",
                            "title": "Holds contact information.",
                            "examples": [
                                {
                                    "email": "foo@example.com",
                                    "name": "Foo Bar",
                                    "url": "https://example.com"
                                }
                            ],
                            "required": [
                                "email"
                            ],
                            "properties": {
                                "email": {
                                    "$id": "#/properties/metadata/properties/maintainers/items/anyOf/0/properties/email",
                                    "format": "email",
                                    "type": "string",
                                    "title": "Email address of the person."
                                },
                                "name": {
                                    "$id": "#/properties/metadata/properties/maintainers/items/anyOf/0/properties/name",
                                    "type": "string",
                                    "title": "Name of the person."
                                },
                                "url": {
                                    "$id": "#/properties/metadata/properties/maintainers/items/anyOf/0/properties/url",
                                    "format": "iri",
                                    "type": "string",
                                    "title": "URL of the person’s site."
                                }
                            }
                        }
                    },
                    "documentationURL": {
                        "$id": "#/properties/metadata/properties/documentationURL",
                        "format": "uri",
                        "type": "string",
                        "title": "Link to documentation page for the OCF manifest.",
                        "description": "Link to documentation page for the OCF manifest.",
                        "examples": [
                            "https://example.com/docs"
                        ]
                    },
                    "supportURL": {
                        "$id": "#/properties/metadata/properties/supportURL",
                        "format": "uri",
                        "type": "string",
                        "title": "Link to support page for the OCF manifest.",
                        "description": "Link to support page for the OCF manifest.",
                        "examples": [
                            "https://example.com/online-support"
                        ]
                    },
                    "iconURL": {
                        "$id": "#/properties/metadata/properties/iconURL",
                        "format": "uri",
                        "type": "string",
                        "title": "The URL to an icon or a data URL containing an icon.",
                        "description": "The URL to an icon or a data URL containing an icon.",
                        "examples": [
                            "https://example.com/favicon.ico"
                        ]
                    },
                    "tags": {
                        "$id": "#/properties/metadata/properties/tags",
                        "type": "object",
                        "title": "The tags schema",
                        "description": "The tags is a list of key value, OCF Tags. Describes the OCF Implementation (provides generic categorization) and are used to filter out a specific Implementation.",
                        "patternProperties": {
                            "^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$": {
                                "type": "object",
                                "required": [
                                    "revision"
                                ],
                                "properties": {
                                    "revision": {
                                        "type": "string"
                                    }
                                }
                            }
                        },
                        "additionalProperties": false
                    },
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
                }
            },
            "spec": {
                "$id": "#/properties/spec",
                "type": "object",
                "title": "The spec schema",
                "description": "A container for the Implementation specification definition.",
                "required": [
                    "appVersion",
                    "implements",
                    "action"
                ],
                "properties": {
                    "appVersion": {
                        "$id": "#/properties/spec/properties/appVersion",
                        "type": "string",
                        "title": "The appVersion schema",
                        "description": "The supported application versions in SemVer2 format.",
                        "additionalProperties": false
                    },
                    "implements": {
                        "$id": "#/properties/spec/properties/implements",
                        "type": "array",
                        "title": "The implements schema",
                        "description": "Defines what kind of interfaces this implementation fulfills.",
                        "items": {
                            "$id": "#/properties/spec/properties/implements/items",
                            "type": "object",
                            "required": [
                                "name"
                            ],
                            "properties": {
                                "name": {
                                    "$id": "#/properties/spec/properties/implements/items/anyOf/0/properties/name",
                                    "type": "string",
                                    "title": "The name schema",
                                    "description": "The Interface name, for example cap.interfaces.db.mysql.install"
                                },
                                "revision": {
                                    "$id": "#/properties/spec/properties/implements/items/anyOf/0/properties/revision",
                                    "type": "string",
                                    "title": "The revision schema",
                                    "description": "The Interface revision.",
                                    "additionalProperties": false
                                }
                            },
                            "additionalProperties": false
                        }
                    },
                    "requires": {
                        "$id": "#/properties/spec/properties/requires",
                        "type": "object",
                        "title": "The requires schema",
                        "description": "List of the system prerequisites that need to be present on the cluster. There has to be an Instance for every concrete type.",
                        "patternProperties": {
                            "^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$": {
                                "$id": "#/properties/spec/properties/requires/properties/cap.core.type.platform",
                                "type": "object",
                                "title": "The prefix schema",
                                "description": "Prefix MUST be an abstract node and represents a core abstract Type e.g. cap.core.type.platform. Custom Types are not allowed.",
                                "properties": {
                                    "oneOf": {
                                        "$id": "#/properties/spec/properties/requires/properties/oneOf",
                                        "type": "array",
                                        "title": "The oneOf schema",
                                        "description": "Exactly one of the given types MUST have an Instance on the cluster. Element on the list MUST resolves to concrete Type.",
                                        "items": {
                                            "$ref": "#/definitions/requireEntity"
                                        }
                                    },
                                    "allOf": {
                                        "$id": "#/properties/spec/properties/requires/properties/allOf",
                                        "type": "array",
                                        "title": "The allOf schema",
                                        "description": "All of the given types MUST have an Instance on the cluster. Element on the list MUST resolves to concrete Type.",
                                        "items": {
                                            "$ref": "#/definitions/requireEntity"
                                        }
                                    },
                                    "anyOf": {
                                        "$id": "#/properties/spec/properties/requires/properties/anyOf",
                                        "type": "array",
                                        "title": "The anyOf schema",
                                        "description": "Any (one or more) of the given types MUST have an Instance on the cluster. Element on the list MUST resolves to concrete Type.",
                                        "items": {
                                            "$ref": "#/definitions/requireEntity"
                                        }
                                    }
                                },
                                "additionalProperties": false
                            }
                        },
                        "additionalProperties": false
                    },
                    "imports": {
                        "$id": "#/properties/spec/properties/imports",
                        "type": "array",
                        "title": "The imports schema",
                        "description": "List of external Interfaces that this Implementation requires to be able to execute the action.",
                        "items": {
                            "$id": "#/properties/imports/items",
                            "type": "object",
                            "required": [
                                "name",
                                "methods"
                            ],
                            "properties": {
                                "name": {
                                    "$id": "#/properties/imports/items/0/properties/name",
                                    "type": "string",
                                    "title": "The name schema",
                                    "description": "The name of the group that holds specific actions that you want to import, for example cap.interfaces.db.mysql"
                                },
                                "alias": {
                                    "$id": "#/properties/imports/items/0/properties/alias",
                                    "type": "string",
                                    "title": "The alias schema",
                                    "description": "The alias for the full name of the imported group name. It can be used later in the workflow definition instead of using full name."
                                },
                                "appVersion": {
                                    "$id": "#/properties/imports/items/0/properties/appVersion",
                                    "type": "string",
                                    "title": "The appVersion schema",
                                    "description": "The supported application versions in SemVer2 format.",
                                    "examples": [
                                        "5.6.x, 5.7.x"
                                    ]
                                },
                                "methods": {
                                    "$id": "#/properties/imports/items/0/properties/methods",
                                    "type": "array",
                                    "title": "The methods schema",
                                    "description": "The list of all required actions’ names that must be imported.",
                                    "items": {
                                        "type": "string"
                                    }
                                }
                            },
                            "additionalProperties": false
                        }
                    },
                    "action": {
                        "$id": "#/properties/spec/properties/action",
                        "type": "object",
                        "title": "The action schema",
                        "description": "An explanation about the purpose of this instance.",
                        "required": [
                            "args",
                            "type"
                        ],
                        "properties": {
                            "args": {
                                "$id": "#/properties/spec/properties/action/properties/args",
                                "type": "object",
                                "title": "The args schema",
                                "description": "Holds all parameters that should be passed to the selected runner, for example repoUrl, or chartName for the Helm3 runner."
                            },
                            "type": {
                                "$id": "#/properties/spec/properties/action/properties/type",
                                "type": "string",
                                "title": "The type schema",
                                "description": "The Interface or Implementation of a runner, which handles the execution, for example, cap.interface.runner.helm3.run"
                            }
                        },
                        "additionalProperties": false
                    }
                }
            }
        }
    }
;

(async () => {
    const convertedSchema = await convert(schema);
    console.log(JSON.stringify(convertedSchema, null, 2));
})();
