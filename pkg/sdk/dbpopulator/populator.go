package dbpopulator

import (
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
)

type Populator struct {
	Session neo4j.Session
}

// TODO: add AtributeSpec
var attributeQuery = `
MERGE (signature:Signature{och: value.signature.och})

MERGE (attribute:Attribute{path: "<PATH>", name: value.metadata.name})
CREATE (metadata:GenericMetadata {
  path: "<PATH>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL})

CREATE (attributeRevision: AttributeRevision {revision: value.revision})

CREATE (attributeRevision)-[:DESCRIBED_BY]->(metadata)
CREATE (attribute)-[:CONTAINS]->(attributeRevision)
CREATE (attributeRevision)-[:SIGNED_WITH]->(signature)

WITH value, metadata 
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer {
  email: m.email,
  name: m.name,
  url: m.url})
CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

// TODO handle optional fields
var typeQuery = `
MERGE (signature:Signature{och: value.signature.och})
MERGE (type:Type{path: "<PATH>", name: value.metadata.name})
CREATE (metadata:TypeMetadata {
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL})

CREATE (typeSpec:TypeSpec {jsonSchema: value.spec.jsonSchema.value})
CREATE (typeRevision:TypeRevision {revision: value.revision})
CREATE (typeRevision)-[:SPECIFIED_BY]->(typeSpec)
CREATE (typeRevision)-[:DESCRIBED_BY]->(metadata)
CREATE (type)-[:CONTAINS]->(typeRevision)
CREATE (typeRevision)-[:SIGNED_WITH]->(signature)

WITH value, metadata 
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer {
  email: m.email,
  name: m.name,
  url: m.url})
CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var interfaceGroupQuery = `
MERGE (signature:Signature{och: value.signature.och})
MERGE (metadata:GenericMetadata {
  path: "<PATH>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL})
MERGE (interfaceGroup:InterfaceGroup{path: "<PATH>"})

MERGE (interfaceGroup)-[:DESCRIBED_BY]->(metadata)
MERGE (interfaceGroup)-[:SIGNED_WITH]->(signature)

WITH value, metadata 
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer {
  email: m.email,
  name: m.name,
  url: m.url})
MERGE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var interfaceQuery = `
MATCH (interfaceGroup:InterfaceGroup{path: "<PREFIX>"})
CREATE (input:InterfaceInput)
CREATE (inputParameters:InputParameters {
  jsonSchema: value.spec.input.parameters.jsonSchema.value})
MERGE (input)-[:HAS]->(inputParameters)

CREATE (output:InterfaceOutput)

CREATE (spec:InterfaceSpec)
MERGE (spec)-[:HAS_INPUT]->(input)
MERGE (spec)-[:OUTPUTS]->(output)

MERGE (signature:Signature{och: value.signature.och})

MERGE (metadata:GenericMetadata {
  path: "<PATH>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL})
CREATE (interfaceRevision:InterfaceRevision {revision: value.revision})
MERGE (interfaceRevision)-[:DESCRIBED_BY]->(metadata)
MERGE (interfaceRevision)-[:SIGNED_WITH]->(signature)
MERGE (interfaceRevision)-[:SPECIFIED_BY]->(spec)
MERGE (interface:Interface {path: "<PATH>"})
MERGE (interface)-[:CONTAINS]->(interfaceRevision)

MERGE (interfaceGroup)-[:CONTAINS]->(interface)

WITH output, input, value, metadata, value.spec.input.typeInstances as typeInstances
UNWIND (CASE keys(typeInstances) WHEN null then [null] else keys(typeInstances) end) as name
CREATE (inputTypeInstance: InputTypeInstance{
  name: name,
  verbs: typeInstances[name].verbs})
CREATE (typeReference: TypeReference{
  path: typeInstances[name].typeRef.path,
  revision: typeInstances[name].typeRef.revision})
MERGE (inputTypeInstance)-[:OF_TYPE]->(typeReference)
MERGE (input)-[:HAS]->(inputTypeInstance)

WITH distinct output, value, metadata, value.spec.output.typeInstances as typeInstances
UNWIND (CASE keys(typeInstances) WHEN null then [null] else keys(typeInstances) end) as name
CREATE (outputTypeInstance: OutputTypeInstance{name: name})
CREATE (typeReference: TypeReference{
  path: typeInstances[name].typeRef.path,
  revision: typeInstances[name].typeRef.revision})
MERGE (outputTypeInstance)-[:OF_TYPE]->(typeReference)
MERGE (output)-[:HAS]->(outputTypeInstance)

WITH value, metadata
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer {
  email: m.email,
  name: m.name,
  url: m.url})
MERGE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var implementationQuery = `
MERGE (implementation:Implementation{path: "<PATH>", prefix: "<PREFIX>"})
CREATE (implementationRevision:ImplementationRevision {revision: value.Revision})

CREATE (implementation)-[:CONTAINS]->(implementationRevision)

MERGE (signature:Signature{och: value.signature.och})
CREATE (implementationRevision)-[:SIGNED_WITH]->(signature)

CREATE (metadata:ImplementationMetadata {
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.supportURL})
CREATE (implementationRevision)-[:DESCRIBED_BY]->(metadata)

CREATE (spec:ImplementationSpec{appVersion: value.spec.appVersion})
CREATE (implementationRevision)-[:SPECIFIED_BY]->(spec)

CREATE (action:ImplementationAction {
  runnerInterface: value.spec.action.runnerInterface,
  args: apoc.convert.toJson(value.spec.action.args)})
CREATE (spec)-[:DOES]->(action)

WITH *
UNWIND value.spec.implements as interface
 MATCH (interfaceRevision: InterfaceRevision {revision: interface.revision})-[:DESCRIBED_BY]->(m:GenericMetadata{path: interface.path})
 CREATE (interfaceReference: InterfaceReference{path: interface.path, revision: interface.revision})
 MERGE (spec)-[:IMPLEMENTS]->(interfaceReference)
 MERGE (implementationRevision)-[:IMPLEMENTS]->(interfaceRevision)
 MERGE (interfaceRevision)-[:IMPLEMENTED_BY]->(implementationRevision)

WITH distinct value, spec, metadata, value.spec.requires as requires
UNWIND (CASE keys(requires) WHEN null then [null] else keys(requires) end) as r
 CREATE (implementationRequirement:ImplementationRequirement{prefix: r})
 CREATE (spec)-[:REQUIRES]->(implementationRequirement)
 WITH *
 UNWIND (CASE keys(requires[r]) WHEN null then [null] else keys(requires[r]) end) as of
  UNWIND requires[r][of] as listItem
   CREATE (item:ImplementationRequirementItem)
   CREATE (type:TypeReference{path:listItem.name, revision:listItem.revision})
   CREATE (item)-[:REFERENCES_TYPE]->(type)
   WITH *, {oneOf: "ONE_OF", anyOf: "ANY_OF", allOf: "ALL_OF"} as requirementTypes
   CALL apoc.create.relationship(implementationRequirement, requirementTypes[of], {}, item) YIELD rel

WITH distinct value, spec, metadata, value.spec.imports as imports
UNWIND (CASE imports WHEN null then [null] else imports end) as import
 CREATE (implementationImport:ImplementationImport {
   interfaceGroupPath: import.interfaceGrupPath,
   alias: import.alias,
   appVersion: import.appVersion})
 CREATE (spec)-[:IMPORTS]->(implementationImport)
 WITH *
 UNWIND (CASE import.methods WHEN null then [null] else import.methods end) as method
  CREATE (implementationImportMethod:ImplementationImportMethod {
    name: method.name,
	revision: method.revision})
  CREATE (implementationImport)-[:HAS]->(implementationImportMethod)

WITH distinct value, spec, metadata, value.spec.additionalInput.typeInstances as typeInstances
CREATE (implementationAdditionalInput:ImplementationAdditionalInput)
CREATE (spec)-[:USES]->(implementationAdditionalInput)
WITH *
UNWIND (CASE keys(typeInstances) WHEN null then [null] else keys(typeInstances) end) as name
 CREATE (inputTypeInstance: InputTypeInstance {
   name: name,
   verbs: typeInstances[name].verbs})
 CREATE (implementationAdditionalInput)-[:CONTAINS]->(inputTypeInstance)
 CREATE (typeReference: TypeReference{
   path: typeInstances[name].typeRef.path,
   revision: typeInstances[name].typeRef.revision})
 CREATE (inputTypeInstance)-[:OF_TYPE]->(typeReference)

WITH distinct value, spec, metadata, value.spec.additionalOutput.typeInstances as typeInstances
CREATE (implementationAdditionalOutput:ImplementationAdditionalOutput)
CREATE (spec)-[:OUTPUTS]->(implementationAdditionalOutput)
WITH *
UNWIND (CASE keys(typeInstances) WHEN null then [null] else keys(typeInstances) end) as name
 CREATE (outputTypeInstance: OutputTypeInstance {
   name: name,
   verbs: typeInstances[name].verbs})
 CREATE (implementationAdditionalOutput)-[:CONTAINS]->(outputTypeInstance)
 CREATE (typeReference: TypeReference{
   path: typeInstances[name].typeRef.path,
   revision: typeInstances[name].typeRef.revision})
 MERGE (outputTypeInstance)-[:OF_TYPE]->(typeReference)

WITH distinct value, spec, metadata, value.spec.additionalOutput.typeInstanceRelations as typeInstanceRelations
UNWIND (CASE keys(typeInstanceRelations) WHEN null then [null] else keys(typeInstanceRelations) end) as name
 CREATE (typeInstanceRelationItem:TypeInstanceRelationItem{
   typeInstanceName: name,
   uses: typeInstanceRelations[name].uses})
 CREATE (spec)-[:RELATIONS]->(typeInstanceRelationItem)


WITH value, metadata
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer {
  email: m.email,
  name: m.name,
  url: m.url})
MERGE (metadata)-[:MAINTAINED_BY]->(maintainer)

//TODO: Attributes
`

//TODO performance: switch merge with create wherever possible

func Populate(session neo4j.Session, paths []string, prefixPath string, publishPath string) error {
	var queries = map[string]string{
		"Attribute":      attributeQuery,
		"Type":           typeQuery,
		"InterfaceGroup": interfaceGroupQuery,
		"Interface":      interfaceQuery,
		"Implementation": implementationQuery,
	}
	grouped, err := Group(paths)
	if err != nil {
		return err
	}

	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		for _, kind := range ordered {
			manifests := grouped[kind]
			query := queries[kind]
			for _, manifest := range manifests {
				// TODO move it to a function
				path := strings.TrimPrefix(manifest, prefixPath)
				path = strings.TrimSuffix(path, ".yaml")
				path = strings.ReplaceAll(path, "/", ".")
				path = strings.TrimPrefix(path, ".")
				path = "cap." + path
				parts := strings.Split(path, ".")
				prefix := strings.Join(parts[:len(parts)-1], ".")

				json := fmt.Sprintf("call apoc.load.json(\"%s/%s\") yield value", publishPath, manifest)
				renderedQuery := strings.ReplaceAll(query, "<PATH>", path)
				renderedQuery = strings.ReplaceAll(renderedQuery, "<PREFIX>", prefix)
				q := fmt.Sprintf("%s\n%s", json, renderedQuery)

				result, err := transaction.Run(q, nil)
				if err != nil {
					return nil, errors.Wrapf(err, "when adding manifest %s", manifest)
				}
				err = result.Err()
				if err != nil {
					return nil, errors.Wrapf(err, "when adding manifest %s", manifest)
				}
			}
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
