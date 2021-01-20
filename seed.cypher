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
MERGE (impl1Rev1)-[:DESCRIBED_BY]->(impl1Rev1Meta:ImplementationMetadata{path:"com.org.concrete.install"})
MERGE (impl1Rev1)-[:SPECIFIED_BY]->(impl1Rev1Spec:ImplementationSpec {appVersion:"0.1.0-0.2.0"})
MERGE (impl1Rev1)-[:IMPLEMENTS]->(ir11) 


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