package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"sync"

	"github.com/hashicorp/go-getter"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
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

	// Neo4jUser is the TCP address the GraphQL endpoint binds to.
	Neo4jUser string `envconfig:"default=neo4j"`

	// Neo4jPassword is the TCP address the GraphQL endpoint binds to.
	Neo4jPassword string `envconfig:"default=okon"`

	// JSONPublishAddr is the address on which populator will serve
	// convert YAML files. It can be k8s service or for example
	// local IP address
	JSONPublishAddr string `envconfig:""`
}

func main() {
	if len(os.Args) < 2 {
		print(help())
		os.Exit(1)
	}
	src := os.Args[1]

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	parent, err := ioutil.TempDir("/tmp", "*-och-parent")
	exitOnError(err, "while creating temporary directory")
	dstDir := path.Join(parent, "och")
	defer os.RemoveAll(dstDir)

	err = download(src, dstDir, c)
	exitOnError(err, "while downloading och content")

	prefixPath := path.Join(dstDir, "och-content/")
	files, err := dbpopulator.List(prefixPath)
	exitOnError(err, "when loading manifests")

	dbpopulator.ServeJson(files)

	driver, err := neo4j.NewDriver(cfg.Neo4jAddr, neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	exitOnError(err, "when connecting to Neo4j db")
	defer driver.Close()

	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	err = dbpopulator.Populate(session, files, prefixPath, cfg.JSONPublishAddr)
	exitOnError(err, "when populating manifests")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}

func download(src string, dst string, c chan os.Signal) error {
	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Error getting pwd")
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Build the client
	client := &getter.Client{
		Ctx:  ctx,
		Src:  src,
		Dst:  dst,
		Pwd:  pwd,
		Mode: getter.ClientModeDir,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.Get(); err != nil {
			errChan <- err
		}
	}()

	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		log.Printf("signal %v", sig)
	case <-ctx.Done():
		wg.Wait()
	case err := <-errChan:
		wg.Wait()
		return err
	}
	return nil
}
