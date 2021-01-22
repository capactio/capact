MATCH (n) DETACH DELETE n

WITH collect(n) as deleted

MERGE (rp:RepoMetadata {path: "com.org.test"})
MERGE (rp)-[:CONTAINS]->(r1:RepoMetadataRevision {revision:"0.1.0"})
MERGE (r1)-[:DESCRIBED_BY]->(r1m:GenericMetadata {path:"com.org.test1"})

MERGE (rp)-[:CONTAINS]->(r2:RepoMetadataRevision {revision:"0.2.0"})
MERGE (r2)-[:DESCRIBED_BY]->(r3m:GenericMetadata {path:"com.org.test2"})

MERGE (ifg:InterfaceGroup {path:"com.org.group"})
MERGE (ifg)-[:DESCRIBED_BY]->(ifgm:GenericMetadata {path:"com.org.group"})
MERGE (ifg)-[:CONTAINS]->(if1:Interface {path:"com.org.group.install"})

MERGE (ifg2:InterfaceGroup {path:"com.org2.group"})
MERGE (ifg2)-[:DESCRIBED_BY]->(ifgm2:GenericMetadata {path:"com.org2.group"})
MERGE (ifg2)-[:CONTAINS]->(ifx:Interface {path:"com.org2.group.install"})

MERGE (if1)-[:CONTAINS]->(ir11:InterfaceRevision {revision:"0.1.0"})
MERGE (ir11)-[:SPECIFIED_BY]->(ifs11:InterfaceSpec)

MERGE (if1)-[:CONTAINS]->(ir12:InterfaceRevision {revision:"0.2.0"})
MERGE (ir12)-[:SPECIFIED_BY]->(ifs12:InterfaceSpec)

MERGE (impl1:Implementation {path:"com.org.concrete.install"})
MERGE (impl1)-[:CONTAINS]->(impl1Rev1:ImplementationRevision {revision:"0.1.0"})
MERGE (impl1Rev1)-[:DESCRIBED_BY]->(impl1Rev1Meta:ImplementationMetadata{path:"org.concrete.install", prefix:"org.concrete"})
MERGE (impl1Rev1)-[:SPECIFIED_BY]->(impl1Rev1Spec:ImplementationSpec {appVersion:"0.1.0-0.2.0"})
MERGE (impl1Rev1)-[:IMPLEMENTS]->(ir11) 

MERGE (impl1Rev1Spec)-[:REQUIRES]->(req111:ImplementationRequirement)
MERGE (req111)-[:ONE_OF]->(:ImplementationRequirementItem)-[:REFERENCES_TYPE]->(:TypeReference {path: "com.ti1", revision: "0.1.0"})
CREATE (req111)-[:ONE_OF]->(:ImplementationRequirementItem)-[:REFERENCES_TYPE]->(:TypeReference {path: "com.ti2", revision: "0.1.0"})

MERGE (req111)-[:ALL_OF]->(:ImplementationRequirementItem)-[:REFERENCES_TYPE]->(:TypeReference {path: "com.ti3", revision: "0.1.0"})
MERGE (req111)-[:ALL_OF]->(:ImplementationRequirementItem)-[:REFERENCES_TYPE]->(:TypeReference {path: "com.ti4", revision: "0.1.0"})

//CREATE (impl2Rev1Spec)-[:REQUIRES]->(req112:ImplementationRequirement)
//MERGE (req112)-[:ONE_OF]->(:ImplementationRequirementItem)-[:REFERENCES_TYPE]->(:TypeReference {path: "com.ti3", revision: "0.1.0"})

//MERGE (req112)-[:ANY_OF]->(:ImplementationRequirementItem)-[:REFERENCES_TYPE]->(:TypeReference {path: "com.ti6", revision: "0.1.0"})
//MERGE (req112)-[:ANY_OF]->(:ImplementationRequirementItem)-[:REFERENCES_TYPE]->(:TypeReference {path: "com.ti7", revision: "0.1.0"})

MERGE (impl2:Implementation {path:"com.org.concrete.delete"})
MERGE (impl2)-[:CONTAINS]->(impl2Rev1:ImplementationRevision {revision:"0.1.0"})
MERGE (impl2Rev1)-[:DESCRIBED_BY]->(impl2Rev1Meta:ImplementationMetadata{path:"com.concrete.delete", prefix:"com.concrete"})
MERGE (impl2Rev1)-[:SPECIFIED_BY]->(impl2Rev1Spec:ImplementationSpec {appVersion:"0.1.0-0.2.0"})
MERGE (impl2Rev1)-[:IMPLEMENTS]->(ir11) 

MERGE (impl2Rev1Meta)-[:CHARACTERIZED_BY]->(:AttributeRevision {revision: "0.1.0"})-[:DESCRIBED_BY]->(:GenericMetadata {path: "com.attr1"})

query {
  repoMetadata {
    name
    prefix
    path
    latestRevision {
      revision
      metadata {
        path
        name
        prefix
      }
    }
    revisions {
      revision
      metadata {
        path
      }
    }
    revision(revision: "0.2.0") {
      revision
      metadata {
        path
      }
    }
  }
  
  interfaceGroups {
    interfaces {
      path
      revision(revision: "0.1.0") {
        revision
      }
    }
  }
  
  interface(
    path: "com.org.group.install"
  ) {
    revision(revision:"0.1.0") {
      revision
      implementationRevisions(filter: {}) {
        revision
        metadata {
          path
        }
      }
    }
  }
}