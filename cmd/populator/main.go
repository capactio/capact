package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
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

	// ManifestsPath is a path to a directory in a repository where
	// manifests are stored
	ManifestsPath string `envconfig:"default=och-content"`

	// UpdateOnGitCommit makes populator to populate a new data
	// only when a git commit chaned in source repository
	UpdateOnGitCommit bool `envconfig:"default=false"`

	// LoggerDevMode sets the logger to use (or not use) development mode (more human-readable output, extra stack traces
	// and logging information, etc).
	LoggerDevMode bool `envconfig:"default=false"`
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

	// setup logger
	var logCfg zap.Config
	if cfg.LoggerDevMode {
		logCfg = zap.NewDevelopmentConfig()
	} else {
		logCfg = zap.NewProductionConfig()
	}

	logger, err := logCfg.Build()
	exitOnError(err, "while creating zap logger")

	parent, err := ioutil.TempDir("/tmp", "*-och-parent")
	exitOnError(err, "while creating temporary directory")
	dstDir := path.Join(parent, "och")
	defer os.RemoveAll(parent)

	err = dbpopulator.Download(ctx, src, dstDir)
	exitOnError(err, "while downloading och content")

	rootDir := path.Join(dstDir, cfg.ManifestsPath)
	files, err := dbpopulator.List(rootDir)
	exitOnError(err, "while loading manifests")

	go dbpopulator.MustServeJSON(ctx, cfg.JSONPublishPort, files)

	driver, err := neo4j.NewDriver(cfg.Neo4jAddr, neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	exitOnError(err, "while connecting to Neo4j db")
	defer driver.Close()

	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	gitHash := []byte{}
	if cfg.UpdateOnGitCommit {
		logger.Info("APP_UPDATE_ON_GIT_COMMIT set. Updating manifests only if git commit changed.")
		gitHash, err = getGitHash(rootDir)
		exitOnError(err, "while getting `git rev-parse HEAD`")
	} else {
		logger.Info("APP_UPDATE_ON_GIT_COMMIT not set. Ignoring git commit, always updating manifests.")
	}

	start := time.Now()
	err = retry.Do(func() error {
		hash := strings.TrimSpace(string(gitHash))
		populated, err := dbpopulator.Populate(
			ctx, logger, session, files, rootDir, fmt.Sprintf("%s:%d", cfg.JSONPublishAddr, cfg.JSONPublishPort), hash)
		if err != nil {
			return err
		}
		if populated {
			end := time.Now()
			logger.Info("Populated new data", zap.Duration("duration (seconds)", end.Sub(start)))
		}
		return nil
	}, retry.Attempts(3), retry.Delay(1*time.Minute))
	exitOnError(err, "while populating manifests")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("Error %s: %v", context, err)
	}
}

// git is used directly because it's already required by go-getter
// When go-getter starts using go-git we can also move to using a library instead of binary
func getGitHash(rootDir string) ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = rootDir
	return cmd.CombinedOutput()
}
