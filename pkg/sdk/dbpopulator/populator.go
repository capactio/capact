package dbpopulator

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
)

// TODO: add AtributeSpec
var attributeQuery = `
MERGE (signature:Signature{och: value.signature.och})

MERGE (attribute:Attribute{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})
CREATE (metadata:GenericMetadata {
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.supportURL})

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

var typeQuery = `
MERGE (signature:Signature{och: value.signature.och})
MERGE (type:Type{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})
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
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL})
MERGE (interfaceGroup:InterfaceGroup{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})

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

MERGE (interface:Interface{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})

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
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL})
CREATE (interfaceRevision:InterfaceRevision {revision: value.revision})
MERGE (interfaceRevision)-[:DESCRIBED_BY]->(metadata)
MERGE (interfaceRevision)-[:SIGNED_WITH]->(signature)
MERGE (interfaceRevision)-[:SPECIFIED_BY]->(spec)
MERGE (interface)-[:CONTAINS]->(interfaceRevision)

MERGE (interfaceGroup)-[:CONTAINS]->(interface)
MERGE (interface)-[:CONTAINED_BY]->(interfaceGroup)

WITH *, value.spec.input.typeInstances as typeInstances
CALL {
 WITH typeInstances, input, interfaceRevision
 UNWIND keys(typeInstances) as name
  CREATE (inputTypeInstance: InputTypeInstance{
    name: name,
    verbs: typeInstances[name].verbs})
  CREATE (typeReference: TypeReference{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  MERGE (inputTypeInstance)-[:OF_TYPE]->(typeReference)
  MERGE (input)-[:HAS]->(inputTypeInstance)
  WITH *
  MATCH (:Type{
    path: typeInstances[name].typeRef.path})-[:CONTAINS]->(typeRevision:TypeRevision{revision:typeInstances[name].typeRef.revision})
  MERGE (input)-[:HAS]->(typeRevision)
  MERGE (typeRevision)-[:USED_BY]->(interfaceRevision)
 RETURN count([]) as _tmp1
}

WITH *, value.spec.output.typeInstances as typeInstances
CALL {
 WITH typeInstances, output, interfaceRevision
 UNWIND keys(typeInstances) as name
  CREATE (outputTypeInstance: OutputTypeInstance{name: name})
  CREATE (typeReference: TypeReference{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  MERGE (outputTypeInstance)-[:OF_TYPE]->(typeReference)
  MERGE (output)-[:OUTPUTS]->(outputTypeInstance)
  WITH *
  MATCH (:Type{
    path: typeInstances[name].typeRef.path})-[:CONTAINS]->(typeRevision:TypeRevision{revision:typeInstances[name].typeRef.revision})
  MERGE (output)-[:OUTPUTS]->(typeRevision)
  MERGE (typeRevision)-[:USED_BY]->(interfaceRevision)
 RETURN count([]) as _tmp2
}

WITH value, metadata
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer {
  email: m.email,
  name: m.name,
  url: m.url})
MERGE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var implementationQuery = `
MERGE (implementation:Implementation{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})
CREATE (implementationRevision:ImplementationRevision {revision: value.revision})

CREATE (implementation)-[:CONTAINS]->(implementationRevision)

MERGE (signature:Signature{och: value.signature.och})
CREATE (implementationRevision)-[:SIGNED_WITH]->(signature)

MERGE (license: License{name: value.metadata.license.name})

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
CREATE (metadata)-[:LICENSED_WITH]->(license)

CREATE (spec:ImplementationSpec{appVersion: value.spec.appVersion})
CREATE (implementationRevision)-[:SPECIFIED_BY]->(spec)

CREATE (action:ImplementationAction {
  runnerInterface: value.spec.action.runnerInterface,
  args: apoc.convert.toJson(value.spec.action.args)})
CREATE (spec)-[:DOES]->(action)

WITH *
CALL {
 WITH value, implementationRevision, spec
 UNWIND value.spec.implements as interface
  MATCH (interfaceRevision: InterfaceRevision {revision: interface.revision})-[:DESCRIBED_BY]->(m:GenericMetadata{path: interface.path})
  CREATE (interfaceReference: InterfaceReference{path: interface.path, revision: interface.revision})
  MERGE (spec)-[:IMPLEMENTS]->(interfaceReference)
  MERGE (implementationRevision)-[:IMPLEMENTS]->(interfaceRevision)
  MERGE (interfaceRevision)-[:IMPLEMENTED_BY]->(implementationRevision)
 return count([]) as _tmp1
}

WITH *, value.spec.requires as requires
CALL {
 WITH value, spec, requires
 UNWIND keys(requires) as r
  CREATE (implementationRequirement:ImplementationRequirement{prefix: r})
  CREATE (spec)-[:REQUIRES]->(implementationRequirement)
  WITH *
  UNWIND keys(requires[r]) as of
   UNWIND requires[r][of] as listItem
    MATCH (type:Type{path: apoc.text.join([r, listItem.name], ".")})-[:CONTAINS]->(typeRevision:TypeRevision {revision: listItem.revision})
    CREATE (typeReference:TypeReference{
      path: apoc.text.join([r, listItem.name], "."),
      revision: listItem.revision})
    CREATE (item:ImplementationRequirementItem {valueConstraints: listItem.valueConstraints})
    CREATE (item)-[:REFERENCES_TYPE]->(typeReference)
    WITH *, {oneOf: "ONE_OF", anyOf: "ANY_OF", allOf: "ALL_OF"} as requirementTypes
    CALL apoc.create.relationship(implementationRequirement, requirementTypes[of], {}, item) YIELD rel as t1
    CALL apoc.create.relationship(implementationRequirement, requirementTypes[of], {}, typeRevision) YIELD rel as t2
 RETURN count([]) as _tmp2
}

WITH *, value.spec.imports as imports
CALL {
 WITH value, imports, spec
 UNWIND imports as import
  CREATE (implementationImport:ImplementationImport {
    interfaceGroupPath: import.interfaceGroupPath,
    alias: import.alias,
    appVersion: import.appVersion})
  CREATE (spec)-[:IMPORTS]->(implementationImport)
  WITH *
  UNWIND import.methods as method
   CREATE (implementationImportMethod:ImplementationImportMethod {
     name: method.name,
     revision: method.revision})
   CREATE (implementationImport)-[:HAS]->(implementationImportMethod)
 RETURN count([]) as _tmp3
}

WITH *, value.spec.additionalInput.typeInstances as typeInstances
CREATE (implementationAdditionalInput:ImplementationAdditionalInput)
CREATE (spec)-[:USES]->(implementationAdditionalInput)
WITH *
CALL {
 WITH typeInstances, implementationAdditionalInput
 UNWIND keys(typeInstances) as name
  CREATE (inputTypeInstance: InputTypeInstance {
    name: name,
    verbs: typeInstances[name].verbs})
  CREATE (implementationAdditionalInput)-[:CONTAINS]->(inputTypeInstance)
  CREATE (typeReference: TypeReference{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  CREATE (inputTypeInstance)-[:OF_TYPE]->(typeReference)
  RETURN count([]) as _tmp4
}

WITH *, value.spec.additionalOutput.typeInstances as typeInstances
CALL {
 WITH typeInstances, spec
 CREATE (implementationAdditionalOutput:ImplementationAdditionalOutput)
 CREATE (spec)-[:OUTPUTS]->(implementationAdditionalOutput)
 WITH *
 UNWIND keys(typeInstances) as name
  CREATE (outputTypeInstance: OutputTypeInstance {
    name: name,
    verbs: typeInstances[name].verbs})
  CREATE (implementationAdditionalOutput)-[:CONTAINS]->(outputTypeInstance)
  CREATE (typeReference: TypeReference{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  MERGE (outputTypeInstance)-[:OF_TYPE]->(typeReference)
 RETURN count([]) as _tmp5
}

WITH *, value.spec.additionalOutput.typeInstanceRelations as typeInstanceRelations
CALL {
 WITH typeInstanceRelations, spec
 UNWIND keys(typeInstanceRelations) as name
  CREATE (typeInstanceRelationItem:TypeInstanceRelationItem{
    typeInstanceName: name,
    uses: typeInstanceRelations[name].uses})
  CREATE (spec)-[:RELATIONS]->(typeInstanceRelationItem)
 RETURN count([]) as _tmp6
}

WITH *
CALL {
 WITH value, metadata
 UNWIND value.metadata.maintainers as m
  MERGE (maintainer:Maintainer {
    email: m.email,
    name: m.name,
    url: m.url})
  MERGE (metadata)-[:MAINTAINED_BY]->(maintainer)
 RETURN count([]) as _tmp7
}

WITH *, value.metadata.attributes as attributes
CALL {
 WITH attributes, metadata
 UNWIND keys(attributes) as path
  MATCH (attribute:Attribute{path: path})-[:CONTAINS]->(revision:AttributeRevision {revision: attributes[path].revision})
  CREATE (metadata)-[:CHARACTERIZED_BY]->(revision)
 RETURN count([]) as _tmp8
}

RETURN []
`

var repoMetadataQuery = `
MERGE (repo:RepoMetadata{path: "<PATH>", prefix: "<PREFIX>", name: value.metadata.name})
CREATE (repoRevision:RepoMetadataRevision {revision: value.revision})
CREATE (repo)-[:CONTAINS]->(repoRevision)

CREATE (metadata:GenericMetadata {
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.supportURL})
CREATE (repoRevision)-[:DESCRIBED_BY]->(metadata)

MERGE (signature:Signature{och: value.signature.och})
CREATE (repoRevision)-[:SIGNED_WITH]->(signature)

CREATE (spec:RepoMetadataSpec{ochVersion: value.spec.ochVersion})
CREATE (repoRevision)-[:SPECIFIED_BY]->(spec)

CREATE (ocfVersion:RepoOCFVersion {
  supported: value.spec.ocfVersion.supported,
  default: value.spec.ocfVersion.default})
CREATE (spec)-[:SUPPORTS]->(ocfVersion)

CREATE (implementation: RepoImplementationConfig)
CREATE (spec)-[:CONFIGURED]->(implementation)

CREATE (appVersion: RepoImplementationAppVersionConfig)
CREATE (implementation)-[:APP_VERSION]->(appVersion)

CREATE (semVerTaggingStrategy:SemVerTaggingStrategy)
CREATE (appVersion)-[:TAGGING_STRATEGY]->(semVerTaggingStrategy)

CREATE (latest:LatestSemVerTaggingStrategy{
  pointsTo: toUpper(value.spec.implementation.appVersion.semVerTaggingStrategy.latest.pointsTo)})
CREATE (semVerTaggingStrategy)-[:LATEST]->(latest)

WITH value, metadata
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer {
  email: m.email,
  name: m.name,
  url: m.url})
CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

func Populate(ctx context.Context, session neo4j.Session, paths []string, rootDir string, publishPath string) error {
	var queries = map[string]string{
		"RepoMetadata":   repoMetadataQuery,
		"Attribute":      attributeQuery,
		"Type":           typeQuery,
		"InterfaceGroup": interfaceGroupQuery,
		"Interface":      interfaceQuery,
		"Implementation": implementationQuery,
	}
	grouped, err := Group(paths)
	if err != nil {
		return errors.Wrap(err, "while grouping manifests")
	}

	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		for _, kind := range ordered {
			paths = grouped[kind]
			query := queries[kind]
			for _, manifestPath := range paths {
				path, prefix := getPathPrefix(manifestPath, rootDir)
				q := renderQuery(query, publishPath, manifestPath, path, prefix)

				select {
				case <-ctx.Done():
					// returning error to not commit transaction
					return nil, errors.New("canceled")
				default:
					result, err := transaction.Run(q, nil)
					if err != nil {
						return nil, errors.Wrapf(err, "when adding manifest %s", manifestPath)
					}
					err = result.Err()
					if err != nil {
						return nil, errors.Wrapf(err, "when adding manifest %s", manifestPath)
					}
				}
			}
		}
		return nil, nil
	})
	return errors.Wrap(err, "while executing neo4j transaction")
}

func getPathPrefix(manifestPath string, rootDir string) (string, string) {
	path := strings.TrimPrefix(manifestPath, rootDir)
	path = strings.TrimSuffix(path, ".yaml")
	path = strings.ReplaceAll(path, "/", ".")
	path = "cap" + path
	parts := strings.Split(path, ".")
	prefix := strings.Join(parts[:len(parts)-1], ".")
	return path, prefix
}

func renderQuery(query, publishPath, manifestPath, path, prefix string) string {
	json := fmt.Sprintf("call apoc.load.json(\"%s/%s\") yield value", publishPath, manifestPath)
	renderedQuery := strings.ReplaceAll(query, "<PATH>", path)
	renderedQuery = strings.ReplaceAll(renderedQuery, "<PREFIX>", prefix)
	return fmt.Sprintf("%s\n%s", json, renderedQuery)
}
