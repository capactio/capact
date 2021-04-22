package dbpopulator

import "capact.io/capact/internal/logger"

// Config holds application related configuration.
type Config struct {
	// Neo4jAddr is the TCP address the GraphQL endpoint binds to.
	Neo4jAddr string `envconfig:"default=neo4j://localhost:7687"`

	// Neo4jUser is the Neo4j admin user name.
	Neo4jUser string `envconfig:"default=neo4j"`

	// Neo4jUser is the Neo4j admin password.
	Neo4jPassword string `envconfig:"default=okon"`

	// JSONPublishAddr is the address on which populator will serve
	// converted YAML files. It can be k8s service or for example
	// local IP address
	JSONPublishAddr string

	// JSONPublishPort is the port number on which populator will
	// serve converted YAML files. Defaults to 8080
	JSONPublishPort int `envconfig:"default=8080"`

	// ManifestsPath is a path to a directory in a repository where
	// manifests are stored
	ManifestsPath string `envconfig:"default=och-content"`

	// UpdateOnGitCommit makes populator to populate a new data
	// only when a git commit chaned in source repository
	UpdateOnGitCommit bool `envconfig:"default=false"`

	Logger logger.Config
}
