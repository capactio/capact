package dbpopulator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// SourceInfo defines a single source with Hub manifests.
type SourceInfo struct {
	RootDir string
	Files   []string
	GitHash []byte
}

const (
	commitEncodeSep  = ","
	maxStoredCommits = 1000
)

var attributeQuery = `
MERGE (attribute:Attribute:unpublished{
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name})
CREATE (metadata:GenericMetadata:unpublished {
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.iconURL})

CREATE (attributeRevision: AttributeRevision:unpublished {revision: value.revision})

CREATE (attributeRevision)-[:DESCRIBED_BY]->(metadata)
CREATE (attribute)-[:CONTAINS]->(attributeRevision)

WITH value, metadata 
UNWIND value.metadata.maintainers as m
MERGE (maintainer:Maintainer:unpublished {
  email: m.email,
  name: m.name,
  url: m.url})
CREATE (metadata)-[:MAINTAINED_BY]->(maintainer)
`

var typeQuery = `
MERGE (type:Type:unpublished{
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name})
CREATE (metadata:TypeMetadata:unpublished {
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.iconURL})

CREATE (typeSpec:TypeSpec:unpublished {jsonSchema: value.spec.jsonSchema.value})
CREATE (typeRevision:TypeRevision:unpublished {revision: value.revision})

// VirtualType allows us to find types which define spec.additionalRefs
CREATE (vType:VirtualType:unpublished {path: "<PREFIX>"})
CREATE (vType)-[:CONTAINS]->(type)

CREATE (typeRevision)-[:SPECIFIED_BY]->(typeSpec)
CREATE (typeRevision)-[:DESCRIBED_BY]->(metadata)
CREATE (type)-[:CONTAINS]->(typeRevision)

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

var interfaceGroupQuery = `
CREATE (metadata:GenericMetadata:unpublished {
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.iconURL})
CREATE (interfaceGroup:InterfaceGroup:unpublished{
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name})

CREATE (interfaceGroup)-[:DESCRIBED_BY]->(metadata)

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

MERGE (interface:Interface:unpublished{
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name})

CREATE (input:InterfaceInput:unpublished)

WITH *, value.spec.input.parameters as parameters
CALL {
  WITH input, parameters
  UNWIND keys(parameters) as name
    CREATE (parameter: InputParameter:unpublished {
      name: name,
      jsonSchema: parameters[name].jsonSchema.value
    })
    CREATE (input)-[:HAS]->(parameter)

    WITH *
    CALL apoc.do.when(
      parameters[name].typeRef IS NOT NULL,
      "MERGE (typeReference: TypeReference:unpublished{path: typeRef.path,revision: typeRef.revision}) CREATE (parameter)-[:OF_TYPE]->(typeReference)",
      "",
      {parameter: parameter, typeRef: parameters[name].typeRef}
    ) YIELD value

  RETURN count([]) as _tmp0
}

CREATE (output:InterfaceOutput:unpublished)

CREATE (spec:InterfaceSpec:unpublished {abstract: value.spec.abstract})
CREATE (spec)-[:HAS_INPUT]->(input)
CREATE (spec)-[:OUTPUTS]->(output)


CREATE (metadata:GenericMetadata:unpublished {
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.iconURL})
CREATE (interfaceRevision:InterfaceRevision:unpublished {revision: value.revision})
CREATE (interfaceRevision)-[:DESCRIBED_BY]->(metadata)
CREATE (interfaceRevision)-[:SPECIFIED_BY]->(spec)
CREATE (interface)-[:CONTAINS]->(interfaceRevision)

MERGE (interfaceGroup)-[:CONTAINS]->(interface)
MERGE (interface)-[:CONTAINED_BY]->(interfaceGroup)

WITH *, value.spec.input.typeInstances as typeInstances
CALL {
 WITH typeInstances, input, interfaceRevision
 UNWIND keys(typeInstances) as name
  CREATE (inputTypeInstance: InputTypeInstance:unpublished{
    name: name,
    verbs: typeInstances[name].verbs})
  MERGE (typeReference: TypeReference:unpublished{
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
  MERGE (typeReference: TypeReference:unpublished{
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
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name})
CREATE (implementationRevision:ImplementationRevision:unpublished {revision: value.revision})

CREATE (implementation)-[:CONTAINS]->(implementationRevision)

MERGE (license: License:unpublished{name: value.metadata.license.name})

CREATE (metadata:ImplementationMetadata:unpublished {
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.iconURL})
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
    MERGE (typeReference:TypeReference:unpublished{
      path: apoc.text.join([r, listItem.name], "."),
      revision: listItem.revision})
    CREATE (item:ImplementationRequirementItem:unpublished {valueConstraints: listItem.valueConstraints, alias: listItem.alias})
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

WITH *, value.spec.additionalInput.typeInstances as typeInstances, value.spec.additionalInput.parameters as parameters
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
  MERGE (typeReference: TypeReference:unpublished{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  CREATE (inputTypeInstance)-[:OF_TYPE]->(typeReference)
  RETURN count([]) as _tmp4
}
CALL {
 WITH parameters, implementationAdditionalInput
 UNWIND keys(parameters) as name
 CREATE (additionalParameter: ImplementationAdditionalInputParameter:unpublished {
    name: name})
 CREATE (implementationAdditionalInput)-[:CONTAINS]->(additionalParameter)
 MERGE (typeReference: TypeReference:unpublished{
   path: parameters[name].typeRef.path,
   revision: parameters[name].typeRef.revision})
 CREATE (additionalParameter)-[:OF_TYPE]->(typeReference)
 RETURN count([]) as _tmpAdditionalParameters
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
  MERGE (typeReference: TypeReference:unpublished{
    path: typeInstances[name].typeRef.path,
    revision: typeInstances[name].typeRef.revision})
  CREATE (outputTypeInstance)-[:OF_TYPE]->(typeReference)
 RETURN count([]) as _tmp5
}

WITH *, value.spec.outputTypeInstanceRelations as outputTypeInstanceRelations
CALL {
 WITH outputTypeInstanceRelations, spec
 UNWIND keys(outputTypeInstanceRelations) as name
  CREATE (typeInstanceRelationItem:TypeInstanceRelationItem:unpublished{
    typeInstanceName: name,
    uses: outputTypeInstanceRelations[name].uses})
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
MERGE (repo:RepoMetadata:unpublished{path: apoc.text.join(["<PREFIX", value.metadata.name], "."), prefix: "<PREFIX>", name: value.metadata.name})
CREATE (repoRevision:RepoMetadataRevision:unpublished {revision: value.revision})
CREATE (repo)-[:CONTAINS]->(repoRevision)

CREATE (metadata:GenericMetadata:unpublished {
  path: apoc.text.join(["<PREFIX>", value.metadata.name], "."),
  prefix: "<PREFIX>",
  name: value.metadata.name,
  displayName: value.metadata.displayName,
  description: value.metadata.description,
  documentationURL: value.metadata.documentationURL,
  supportURL: value.metadata.supportURL,
  iconURL: value.metadata.supportURL})
CREATE (repoRevision)-[:DESCRIBED_BY]->(metadata)

CREATE (spec:RepoMetadataSpec:unpublished{hubVersion: value.spec.hubVersion})
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

// Populate imports Public Hub manifests into a Neo4j database.
func Populate(ctx context.Context, log *zap.Logger, session neo4j.Session, sources []SourceInfo, publishPath string) (bool, error) {
	for _, source := range sources {
		err := populate(ctx, log, session, source.Files, source.RootDir, publishPath)
		if err != nil {
			return false, errors.Wrap(err, "while adding new manifests")
		}
	}

	err := swap(session)
	if err != nil {
		return false, errors.Wrap(err, "while swapping manifests")
	}

	err = cleanOld(session)
	if err != nil {
		return true, errors.Wrap(err, "while cleaning old manifests")
	}

	err = warmup(session)
	return true, err
}

// IsDataInDB checks whether the commits have already existed in the DB
func IsDataInDB(session neo4j.Session, log *zap.Logger, commits []string) (bool, error) {
	currentCommits, err := currentCommits(session)
	if err != nil {
		return false, errors.Wrap(err, "while getting commit hash of populated data")
	}

	storedCommits := decodeCommits(currentCommits)
	if len(storedCommits) != len(commits) {
		log.Info("new sources were added or removed", zap.String("current commits", currentCommits), zap.String("given commits", encodeCommits(commits)))
		return false, nil
	}

	unknownCommits, found := detectUnknownCommit(commits, storedCommits)
	if found {
		log.Info("detected unknown commits", zap.String("current commits", currentCommits), zap.String("unknown commits", encodeCommits(unknownCommits)))
		return false, nil
	}

	log.Info("git commits did not change. Finishing")
	return true, nil
}

// SaveCommitsMetadata saves the commits from the repositories in the DB
// TODO: gather commits per repository, now only repositories from the last run are cached
func SaveCommitsMetadata(session neo4j.Session, commits []string) error {
	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		contentMetadata := "CREATE (n:ContentMetadata:published { commits: '%s', timestamp: '%s'}) RETURN *"
		q := fmt.Sprintf(contentMetadata, encodeCommits(commits), time.Now())
		result, err := transaction.Run(q, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "while running %s", q)
		}
		err = result.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "when checking results %s", q)
		}
		return nil, nil
	})
	return errors.Wrap(err, "while executing neo4j transaction")
}

func populate(ctx context.Context, log *zap.Logger, session neo4j.Session, paths []string, rootDir string, publishPath string) error {
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
				prefix := getPrefix(manifestPath, rootDir)
				q := renderQuery(query, publishPath, manifestPath, prefix)

				select {
				case <-ctx.Done():
					// returning error to not commit transaction
					return nil, errors.New("canceled")
				default:
					log.Info("Processing manifest", zap.String("manifest", manifestPath))
					log.Debug("Executing query", zap.String("query", q))
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

func currentCommits(session neo4j.Session) (string, error) {
	result, err := session.Run("MATCH (c:ContentMetadata:published) RETURN c.commits", map[string]interface{}{})
	if err != nil {
		return "", errors.Wrap(err, "while querying ContentMetadata")
	}

	var record *neo4j.Record
	result.NextRecord(&record)

	if record == nil || len(record.Values) == 0 {
		return "", nil
	}
	commit, ok := record.Values[0].(string)
	if !ok {
		return "", fmt.Errorf("failed to convert database response: %v", record.Values[0])
	}

	return commit, errors.Wrap(result.Err(), "while executing neo4j transaction")
}

func detectUnknownCommit(newCommits []string, storedCommits []string) ([]string, bool) {
	var out []string
	indexed := indexStringSlice(storedCommits)
	for _, commit := range newCommits {
		if _, found := indexed[commit]; found {
			continue
		}
		out = append(out, commit)
	}
	return out, len(out) > 0
}

func decodeCommits(in string) []string {
	return strings.SplitN(in, commitEncodeSep, maxStoredCommits)
}

func encodeCommits(in []string) string {
	return strings.Join(in, commitEncodeSep)
}

func indexStringSlice(slice []string) map[string]struct{} {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	return set
}

func warmup(session neo4j.Session) error {
	_, err := session.Run("CALL apoc.warmup.run(true, true, true)", map[string]interface{}{})
	return errors.Wrap(err, "while warming up the data")
}

func getPrefix(manifestPath string, rootDir string) string {
	path := strings.TrimPrefix(manifestPath, rootDir)
	parts := strings.Split(path, "/")
	prefix := strings.Join(parts[:len(parts)-1], ".")
	prefix = "cap" + prefix
	return prefix
}

func renderQuery(query, publishPath, manifestPath, prefix string) string {
	json := fmt.Sprintf("call apoc.load.json(\"%s/%s\") yield value", publishPath, manifestPath)
	renderedQuery := strings.ReplaceAll(query, "<PREFIX>", prefix)
	return fmt.Sprintf("%s\n%s", json, renderedQuery)
}
