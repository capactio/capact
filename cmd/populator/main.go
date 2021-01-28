package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/vrischmann/envconfig"
	"projectvoltron.dev/voltron/pkg/sdk/dbpopulator"
)

func help() string {
	return `
	populator <path>
`
}

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
}

func main() {
	if len(os.Args) < 2 {
		print(help())
		os.Exit(1)
	}
	src := os.Args[1]

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-c
		signal.Reset(os.Interrupt)
		cancel()
	}()

	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	parent, err := ioutil.TempDir("/tmp", "*-och-parent")
	exitOnError(err, "while creating temporary directory")
	dstDir := path.Join(parent, "och")
	defer os.RemoveAll(parent)

	err = dbpopulator.Download(ctx, src, dstDir)
	exitOnError(err, "while downloading och content")

	rootDir := path.Join(dstDir, "och-content/")
	files, err := dbpopulator.List(rootDir)
	exitOnError(err, "while loading manifests")

	go dbpopulator.MustServeJSON(ctx, cfg.JSONPublishPort, files)

	driver, err := neo4j.NewDriver(cfg.Neo4jAddr, neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	exitOnError(err, "while connecting to Neo4j db")
	defer driver.Close()

	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	err = dbpopulator.Populate(ctx, session, files, rootDir, fmt.Sprintf("%s:%d", cfg.JSONPublishAddr, cfg.JSONPublishPort))
	exitOnError(err, "while populating manifests")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("Error %s: %v", context, err)
	}
}
