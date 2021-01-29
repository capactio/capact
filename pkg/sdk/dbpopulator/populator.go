package dbpopulator

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
)

var attributeQuery = `
MERGE (signature:Signature:unpublished{och: value.signature.och})

MERGE (attribute:Attribute:unpublished{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})
CREATE (metadata:GenericMetadata:unpublished {
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.supportURL})

CREATE (attributeRevision: AttributeRevision:unpublished {revision: value.revision})

CREATE (attributeRevision)-[:DESCRIBED_BY]->(metadata)
CREATE (attribute)-[:CONTAINS]->(attributeRevision)
CREATE (attributeRevision)-[:SIGNED_WITH]->(signature)

WITH value, metadata 
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer:unpublished {
  email: m.email,
  name: m.name,
  url: m.url})
CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var typeQuery = `
MERGE (signature:Signature:unpublished{och: value.signature.och})
MERGE (type:Type:unpublished{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})
CREATE (metadata:TypeMetadata:unpublished {
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL})

CREATE (typeSpec:TypeSpec:unpublished {jsonSchema: value.spec.jsonSchema.value})
CREATE (typeRevision:TypeRevision:unpublished {revision: value.revision})

CREATE (vType:VirtualType:unpublished {path: "<PREFIX>"})
CREATE (vType)-[:CONTAINS]->(type)

CREATE (typeRevision)-[:SPECIFIED_BY]->(typeSpec)
CREATE (typeRevision)-[:DESCRIBED_BY]->(metadata)
CREATE (type)-[:CONTAINS]->(typeRevision)
CREATE (typeRevision)-[:SIGNED_WITH]->(signature)

WITH *, value.spec.additionalRefs as refs
CALL {
 WITH *
 UNWIND refs AS path
  MATCH (baseType:VirtualType:unpublished {path:path})
  CREATE (baseType)-[:CONTAINS]->(type)
 RETURN count([]) as _tmp1
}

WITH value, metadata 
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer:unpublished {
  email: m.email,
  name: m.name,
  url: m.url})
CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var interfaceGroupQuery = `
MERGE (signature:Signature:unpublished{och: value.signature.och})
CREATE (metadata:GenericMetadata:unpublished {
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL})
CREATE (interfaceGroup:InterfaceGroup:unpublished{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})

CREATE (interfaceGroup)-[:DESCRIBED_BY]->(metadata)
CREATE (interfaceGroup)-[:SIGNED_WITH]->(signature)

WITH value, metadata 
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer:unpublished {
  email: m.email,
  name: m.name,
  url: m.url})
CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var interfaceQuery = `
MATCH (interfaceGroup:InterfaceGroup:unpublished{path: "<PREFIX>"})

CREATE (interface:Interface:unpublished{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})

CREATE (input:InterfaceInput:unpublished)
CREATE (inputParameters:InputParameters:unpublished {
  jsonSchema: value.spec.input.parameters.jsonSchema.value})
CREATE (input)-[:HAS]->(inputParameters)

CREATE (output:InterfaceOutput:unpublished)

CREATE (spec:InterfaceSpec:unpublished {abstract: value.spec.abstract})
CREATE (spec)-[:HAS_INPUT]->(input)
CREATE (spec)-[:OUTPUTS]->(output)

MERGE (signature:Signature:unpublished{och: value.signature.och})

CREATE (metadata:GenericMetadata:unpublished {
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL})
CREATE (interfaceRevision:InterfaceRevision:unpublished {revision: value.revision})
CREATE (interfaceRevision)-[:DESCRIBED_BY]->(metadata)
CREATE (interfaceRevision)-[:SIGNED_WITH]->(signature)
CREATE (interfaceRevision)-[:SPECIFIED_BY]->(spec)
CREATE (interface)-[:CONTAINS]->(interfaceRevision)

CREATE (interfaceGroup)-[:CONTAINS]->(interface)
CREATE (interface)-[:CONTAINED_BY]->(interfaceGroup)

WITH *, value.spec.input.typeInstances as typeInstances
CALL {
 WITH typeInstances, input, interfaceRevision
 UNWIND keys(typeInstances) as name
  CREATE (inputTypeInstance: InputTypeInstance:unpublished{
    name: name,
    verbs: typeInstances[name].verbs})
  CREATE (typeReference: TypeReference:unpublished{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  CREATE (inputTypeInstance)-[:OF_TYPE]->(typeReference)
  CREATE (input)-[:HAS]->(inputTypeInstance)
  WITH *
  MATCH (:Type:unpublished{
    path: typeInstances[name].typeRef.path})-[:CONTAINS]->(typeRevision:TypeRevision{revision:typeInstances[name].typeRef.revision})
  CREATE (input)-[:HAS]->(typeRevision)
  CREATE (typeRevision)-[:USED_BY]->(interfaceRevision)
 RETURN count([]) as _tmp1
}

WITH *, value.spec.output.typeInstances as typeInstances
CALL {
 WITH typeInstances, output, interfaceRevision
 UNWIND keys(typeInstances) as name
  CREATE (outputTypeInstance: OutputTypeInstance:unpublished{name: name})
  CREATE (typeReference: TypeReference:unpublished{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  CREATE (outputTypeInstance)-[:OF_TYPE]->(typeReference)
  CREATE (output)-[:OUTPUTS]->(outputTypeInstance)
  WITH *
  MATCH (:Type:unpublished{
    path: typeInstances[name].typeRef.path})-[:CONTAINS]->(typeRevision:TypeRevision{revision:typeInstances[name].typeRef.revision})
  CREATE (output)-[:OUTPUTS]->(typeRevision)
  CREATE (typeRevision)-[:USED_BY]->(interfaceRevision)
 RETURN count([]) as _tmp2
}

WITH value, metadata
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer:unpublished {
  email: m.email,
  name: m.name,
  url: m.url})
CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var implementationQuery = `
MERGE (implementation:Implementation:unpublished{
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name})
CREATE (implementationRevision:ImplementationRevision:unpublished {revision: value.revision})

CREATE (implementation)-[:CONTAINS]->(implementationRevision)

MERGE (signature:Signature:unpublished{och: value.signature.och})
CREATE (implementationRevision)-[:SIGNED_WITH]->(signature)

MERGE (license: License:unpublished{name: value.metadata.license.name})

CREATE (metadata:ImplementationMetadata:unpublished {
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

CREATE (spec:ImplementationSpec:unpublished{appVersion: value.spec.appVersion})
CREATE (implementationRevision)-[:SPECIFIED_BY]->(spec)

CREATE (action:ImplementationAction:unpublished {
  runnerInterface: value.spec.action.runnerInterface,
  args: apoc.convert.toJson(value.spec.action.args)})
CREATE (spec)-[:DOES]->(action)

WITH *
CALL {
 WITH value, implementationRevision, spec
 UNWIND value.spec.implements as interface
  MATCH (interfaceRevision: InterfaceRevision:unpublished {revision: interface.revision})-[:DESCRIBED_BY]->(m:GenericMetadata{path: interface.path})
  CREATE (interfaceReference: InterfaceReference:unpublished{path: interface.path, revision: interface.revision})
  CREATE (spec)-[:IMPLEMENTS]->(interfaceReference)
  CREATE (implementationRevision)-[:IMPLEMENTS]->(interfaceRevision)
  CREATE (interfaceRevision)-[:IMPLEMENTED_BY]->(implementationRevision)
 return count([]) as _tmp1
}

WITH *, value.spec.requires as requires
CALL {
 WITH value, spec, requires
 UNWIND keys(requires) as r
  CREATE (implementationRequirement:ImplementationRequirement:unpublished{prefix: r})
  CREATE (spec)-[:REQUIRES]->(implementationRequirement)
  WITH *
  UNWIND keys(requires[r]) as of
   UNWIND requires[r][of] as listItem
    MATCH (type:Type:unpublished{path: apoc.text.join([r, listItem.name], ".")})-[:CONTAINS]->(typeRevision:TypeRevision {revision: listItem.revision})
    CREATE (typeReference:TypeReference:unpublished{
      path: apoc.text.join([r, listItem.name], "."),
      revision: listItem.revision})
    CREATE (item:ImplementationRequirementItem:unpublished {valueConstraints: listItem.valueConstraints})
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
  CREATE (implementationImport:ImplementationImport:unpublished {
    interfaceGroupPath: import.interfaceGroupPath,
    alias: import.alias,
    appVersion: import.appVersion})
  CREATE (spec)-[:IMPORTS]->(implementationImport)
  WITH *
  UNWIND import.methods as method
   CREATE (implementationImportMethod:ImplementationImportMethod:unpublished {
     name: method.name,
     revision: method.revision})
   CREATE (implementationImport)-[:HAS]->(implementationImportMethod)
 RETURN count([]) as _tmp3
}

WITH *, value.spec.additionalInput.typeInstances as typeInstances
CREATE (implementationAdditionalInput:ImplementationAdditionalInput:unpublished)
CREATE (spec)-[:USES]->(implementationAdditionalInput)
WITH *
CALL {
 WITH typeInstances, implementationAdditionalInput
 UNWIND keys(typeInstances) as name
  CREATE (inputTypeInstance: InputTypeInstance:unpublished {
    name: name,
    verbs: typeInstances[name].verbs})
  CREATE (implementationAdditionalInput)-[:CONTAINS]->(inputTypeInstance)
  CREATE (typeReference: TypeReference:unpublished{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  CREATE (inputTypeInstance)-[:OF_TYPE]->(typeReference)
  RETURN count([]) as _tmp4
}

WITH *, value.spec.additionalOutput.typeInstances as typeInstances
CALL {
 WITH typeInstances, spec
 CREATE (implementationAdditionalOutput:ImplementationAdditionalOutput:unpublished)
 CREATE (spec)-[:OUTPUTS]->(implementationAdditionalOutput)
 WITH *
 UNWIND keys(typeInstances) as name
  CREATE (outputTypeInstance: OutputTypeInstance:unpublished {
    name: name,
    verbs: typeInstances[name].verbs})
  CREATE (implementationAdditionalOutput)-[:CONTAINS]->(outputTypeInstance)
  CREATE (typeReference: TypeReference:unpublished{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  CREATE (outputTypeInstance)-[:OF_TYPE]->(typeReference)
 RETURN count([]) as _tmp5
}

WITH *, value.spec.additionalOutput.typeInstanceRelations as typeInstanceRelations
CALL {
 WITH typeInstanceRelations, spec
 UNWIND keys(typeInstanceRelations) as name
  CREATE (typeInstanceRelationItem:TypeInstanceRelationItem:unpublished{
    typeInstanceName: name,
    uses: typeInstanceRelations[name].uses})
  CREATE (spec)-[:RELATIONS]->(typeInstanceRelationItem)
 RETURN count([]) as _tmp6
}

WITH *
CALL {
 WITH value, metadata
 UNWIND value.metadata.maintainers as m
  MERGE (maintainer:Maintainer:unpublished {
    email: m.email,
    name: m.name,
    url: m.url})
  CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
 RETURN count([]) as _tmp7
}

WITH *, value.metadata.attributes as attributes
CALL {
 WITH attributes, metadata
 UNWIND keys(attributes) as path
  MATCH (attribute:Attribute:unpublished{path: path})-[:CONTAINS]->(revision:AttributeRevision {revision: attributes[path].revision})
  CREATE (metadata)-[:CHARACTERIZED_BY]->(revision)
  CREATE (revision)-[:CHARACTERIZES]->(metadata)
 RETURN count([]) as _tmp8
}

RETURN []
`

var repoMetadataQuery = `
MERGE (repo:RepoMetadata:unpublished{path: "<PATH>", prefix: "<PREFIX>", name: value.metadata.name})
CREATE (repoRevision:RepoMetadataRevision:unpublished {revision: value.revision})
CREATE (repo)-[:CONTAINS]->(repoRevision)

CREATE (metadata:GenericMetadata:unpublished {
  path: "<PATH>",
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.supportURL})
CREATE (repoRevision)-[:DESCRIBED_BY]->(metadata)

MERGE (signature:Signature:unpublished{och: value.signature.och})
CREATE (repoRevision)-[:SIGNED_WITH]->(signature)

CREATE (spec:RepoMetadataSpec:unpublished{ochVersion: value.spec.ochVersion})
CREATE (repoRevision)-[:SPECIFIED_BY]->(spec)

CREATE (ocfVersion:RepoOCFVersion:unpublished {
  supported: value.spec.ocfVersion.supported,
  default: value.spec.ocfVersion.default})
CREATE (spec)-[:SUPPORTS]->(ocfVersion)

CREATE (implementation: RepoImplementationConfig:unpublished)
CREATE (spec)-[:CONFIGURED]->(implementation)

CREATE (appVersion: RepoImplementationAppVersionConfig:unpublished)
CREATE (implementation)-[:APP_VERSION]->(appVersion)

CREATE (semVerTaggingStrategy:SemVerTaggingStrategy:unpublished)
CREATE (appVersion)-[:TAGGING_STRATEGY]->(semVerTaggingStrategy)

CREATE (latest:LatestSemVerTaggingStrategy:unpublished{
  pointsTo: toUpper(value.spec.implementation.appVersion.semVerTaggingStrategy.latest.pointsTo)})
CREATE (semVerTaggingStrategy)-[:LATEST]->(latest)

WITH value, metadata
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer:unpublished {
  email: m.email,
  name: m.name,
  url: m.url})
CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var swapQuery = `
CALL {
 MATCH (n:published)
 CALL apoc.create.addLabels( n, [ "to_remove" ] ) YIELD node
 CALL apoc.create.removeLabels( n, [ "published" ] ) YIELD node as node1
 return count(n)
}

MATCH (n:unpublished)
CALL apoc.create.addLabels( n, [ "published" ] ) YIELD node
CALL apoc.create.removeLabels( n, [ "unpublished" ] ) YIELD node as node1
RETURN count(n)
`

var cleanQuery = `
call apoc.periodic.iterate("MATCH (n:to_remove) return n", "DETACH DELETE n", {batchSize:1000})
yield batches, total return batches, total
`

func Populate(ctx context.Context, session neo4j.Session, paths []string, rootDir string, publishPath string, commit string) error {
	currentCommit, err := currentCommit(session)
	if err != nil {
		return errors.Wrap(err, "while adding new manifests")
	}

	if currentCommit == commit {
		return nil
	}

	err = populate(ctx, session, paths, rootDir, publishPath, commit)
	if err != nil {
		return errors.Wrap(err, "while adding new manifests")
	}

	err = swap(session)
	if err != nil {
		return errors.Wrap(err, "while swapping manifests")
	}

	err = cleanOld(session)
	if err != nil {
		return errors.Wrap(err, "while cleaning old manifests")
	}
	return nil
}

func populate(ctx context.Context, session neo4j.Session, paths []string, rootDir string, publishPath string, commit string) error {
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
		q := fmt.Sprintf("CREATE (n:ContentMetadata:unpublished { commit: '%s' }) RETURN *", commit)
		result, err := transaction.Run(q, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "when running query %s", q)
		}
		err = result.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "when running query %s", q)
		}
		return nil, nil
	})
	return errors.Wrap(err, "while executing neo4j transaction")
}

func swap(session neo4j.Session) error {
	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(swapQuery, nil)
		return nil, err
	})
	return errors.Wrap(err, "while executing neo4j transaction")
}

func cleanOld(session neo4j.Session) error {
	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(cleanQuery, nil)
		return nil, err
	})
	return errors.Wrap(err, "while executing neo4j transaction")
}

func currentCommit(session neo4j.Session) (string, error) {
	result, err := session.Run("MATCH (c:ContentMetadata:published) RETURN c.commit", map[string]interface{}{})
	if err != nil {
		return "", errors.Wrap(err, "while quering ContextMetadada")
	}

	var record *neo4j.Record
	result.NextRecord(&record)

	if record == nil || len(record.Values) == 0 {
		return "", nil
	}
	return record.Values[0].(string), errors.Wrap(result.Err(), "while executing neo4j transaction")
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
