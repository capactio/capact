import { JSONSchemaType } from "ajv";

export interface StorageTypeInstanceSpec {
  url: string;
  acceptValue: boolean;
  contextSchema: object | null;
}

export const StorageTypeInstanceSpecSchema: JSONSchemaType<StorageTypeInstanceSpec> =
  {
    $schema: "http://json-schema.org/draft-07/schema",
    title: "The Storage TypeInstance's spec field schema",
    type: "object",
    properties: {
      url: {
        $id: "#/properties/url",
        type: "string",
        format: "uri",
      },
      acceptValue: { type: "boolean" },
      contextSchema: { type: "object", nullable: true },
    },
    required: ["url", "acceptValue"],
    additionalProperties: true,
  };
