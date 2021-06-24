package register

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/getter"
	"capact.io/capact/internal/logger"
	"capact.io/capact/pkg/sdk/dbpopulator"

	"github.com/avast/retry-go"
	"github.com/docker/cli/cli"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
)

// NewOCFManifests returns a cobra.Command for populating Manifests into a Neo4j database.
// TODO: support configuration both via flags and environment variables
func NewOCFManifests(cliName string) *cobra.Command {
	return &cobra.Command{
		Use:   "ocf-manifests [MANIFEST_PATH]",
		Short: "Populates locally available manifests into Neo4j database",
		Example: heredoc.WithCLIName(`
			APP_JSON_PUBLISH_ADDR=http://{HOST_IP} <cli> .
		`, cliName),
		Args: cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBPopulate(cmd.Context(), args[0])
		},
	}
}

func runDBPopulate(ctx context.Context, src string) error {
	var cfg dbpopulator.Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		return errors.Wrap(err, "while loading configuration")
	}

	// setup logger
	logger, err := logger.New(cfg.Logger)
	if err != nil {
		return errors.Wrap(err, "while creating zap logger")
	}

	parent, err := ioutil.TempDir("/tmp", "*-hub-parent")
	if err != nil {
		return errors.Wrap(err, "while creating temporary directory")
	}
	dstDir := path.Join(parent, "hub")
	defer os.RemoveAll(parent)

	err = getter.Download(ctx, src, dstDir)
	if err != nil {
		return errors.Wrap(err, "while downloading Hub manifests")
	}

	logger.Info("Populating downloaded manifests...", zap.String("path", cfg.ManifestsPath))
	rootDir := path.Join(dstDir, cfg.ManifestsPath)
	files, err := dbpopulator.List(rootDir)
	if err != nil {
		return errors.Wrap(err, "while loading manifests")
	}

	go dbpopulator.MustServeJSON(ctx, cfg.JSONPublishPort, files)

	driver, err := neo4j.NewDriver(cfg.Neo4jAddr, neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	if err != nil {
		return errors.Wrap(err, "while connecting to Neo4j db")
	}
	defer driver.Close()

	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	gitHash := []byte{}
	if cfg.UpdateOnGitCommit {
		logger.Info("APP_UPDATE_ON_GIT_COMMIT set. Updating manifests only if git commit changed.")
		gitHash, err = getGitHash(rootDir)
		if err != nil {
			return errors.Wrap(err, "while getting `git rev-parse HEAD`")
		}
	} else {
		logger.Info("APP_UPDATE_ON_GIT_COMMIT not set. Ignoring git commit, always updating manifests.")
	}

	start := time.Now()
	err = retry.Do(func() error {
		hash := strings.TrimSpace(string(gitHash))
		populated, err := dbpopulator.Populate(
			ctx, logger, session, files, rootDir, fmt.Sprintf("%s:%d", cfg.JSONPublishAddr, cfg.JSONPublishPort), hash)
		if err != nil {
			logger.Error("Cannot populate a new data", zap.String("error", err.Error()))
			return err
		}
		if populated {
			end := time.Now()
			logger.Info("Populated new data", zap.Duration("duration (seconds)", end.Sub(start)))
		}
		return nil
	}, retry.Attempts(3), retry.Delay(1*time.Minute))
	if err != nil {
		return errors.Wrap(err, "while populating manifests")
	}

	return nil
}

// git is used directly because it's already required by go-getter
// When go-getter starts using go-git we can also move to using a library instead of binary
func getGitHash(rootDir string) ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = rootDir
	return cmd.CombinedOutput()
}
