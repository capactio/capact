package register

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/getter"
	"capact.io/capact/internal/io"
	"capact.io/capact/internal/logger"
	"capact.io/capact/internal/multierror"
	"capact.io/capact/pkg/sdk/dbpopulator"
	"github.com/avast/retry-go"
	"github.com/docker/cli/cli"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
)

type sourceInfo struct {
	dir     string
	files   []string
	gitHash []byte
}

type filterPath func(string) string

// NewOCFManifests returns a cobra.Command for populating manifests into a Neo4j database.
// TODO: support configuration both via flags and environment variables
func NewOCFManifests(cliName string) *cobra.Command {
	var sources []string
	cmd := &cobra.Command{
		Use:   "ocf-manifests [MANIFEST_PATH]",
		Short: "Populates locally available manifests into Neo4j database",
		Example: heredoc.WithCLIName(`
			APP_JSON_PUBLISH_ADDR=http://{HOST_IP} <cli> .
		`, cliName),
		Args: cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBPopulateWithSources(cmd.Context(), sources)
		},
	}
	flags := cmd.Flags()
	flags.StringSliceVar(&sources, "source", []string{}, "Manifests location")
	return cmd
}

func runDBPopulateWithSources(ctx context.Context, sources []string) (err error) {
	var cfg dbpopulator.Config
	err = envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		return errors.Wrap(err, "while loading configuration")
	}

	log, err := logger.New(cfg.Logger)
	if err != nil {
		return errors.Wrap(err, "while creating zap logger")
	}

	parentDir, err := ioutil.TempDir("/tmp", "*-hubs-parent")
	if err != nil {
		return errors.Wrap(err, "while creating parent temporary directory")
	}

	sources = removeDuplicateSources(sources)
	if len(sources) == 0 {
		return fmt.Errorf("no source information provided")
	}

	sourcesInfo, err := getSourcesInfo(ctx, cfg, log, sources, parentDir)
	if err != nil {
		return errors.Wrap(err, "while getting sources info")
	}

	defer func() {
		if rErr := os.RemoveAll(parentDir); rErr != nil {
			err = multierror.Append(err, rErr)
		}
	}()

	// run server with merge file list from various sources
	seenFiles := make(map[string]struct{})
	var fileList []string
	var commits []string
	for _, src := range sourcesInfo {
		err = filesAlreadyExists(seenFiles, src.files)
		if err != nil {
			return errors.Wrap(err, "while validating the source files")
		}
		fileList = append(fileList, src.files...)
		commits = append(commits, strings.TrimSpace(string(src.gitHash)))
	}
	go dbpopulator.MustServeJSON(ctx, cfg.JSONPublishPort, fileList)

	// create neo4j session
	driver, err := neo4j.NewDriver(cfg.Neo4jAddr, neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	if err != nil {
		return errors.Wrap(err, "while connecting to Neo4j db")
	}
	defer func() {
		if cErr := driver.Close(); cErr != nil {
			err = multierror.Append(err, cErr)
		}
	}()
	session := driver.NewSession(neo4j.SessionConfig{})
	defer func() {
		if sErr := session.Close(); sErr != nil {
			err = multierror.Append(err, sErr)
		}
	}()

	if cfg.UpdateOnGitCommit {
		log.Info("APP_UPDATE_ON_GIT_COMMIT set. Updating manifests only if git commit changed.")
		dataInDB, err := dbpopulator.IsDataInDB(session, log, commits)
		if err != nil {
			return errors.Wrap(err, "while verifying commits in db")
		}
		if dataInDB {
			return nil
		}
	} else {
		log.Info("APP_UPDATE_ON_GIT_COMMIT not set. Ignoring git commit, always updating manifests.")
	}

	for _, src := range sourcesInfo {
		err = runDBPopulate(ctx, cfg, session, log, src)
		if err != nil {
			return errors.Wrap(err, "while populating db")
		}
	}

	if cfg.UpdateOnGitCommit {
		err = dbpopulator.SaveCommitsMetadata(session, commits)
		if err != nil {
			return errors.Wrap(err, "while saving metadata into db")
		}
	}

	return nil
}

func removeDuplicateSources(sources []string) []string {
	var result []string
	allSources := make(map[string]struct{})
	for _, source := range sources {
		if _, ok := allSources[source]; ok {
			continue
		}
		allSources[source] = struct{}{}
		result = append(result, source)
	}
	return result
}

func getSourcesInfo(ctx context.Context, cfg dbpopulator.Config, log *zap.Logger, sources []string, parent string) ([]sourceInfo, error) {
	var sourcesInfo []sourceInfo
	for _, source := range sources {
		dstDir, err := getDestDir(parent, source)
		if err != nil {
			return nil, errors.Wrap(err, "while getting a destination directory")
		}

		err = getter.Download(ctx, source, dstDir)
		if err != nil {
			return nil, errors.Wrap(err, "while downloading Hub manifests")
		}

		log.Info("Populating downloaded manifests...", zap.String("source", source), zap.String("path", cfg.ManifestsPath))
		rootDir := path.Join(dstDir, cfg.ManifestsPath)
		files, err := io.ListYAMLs(rootDir)
		if err != nil {
			return nil, errors.Wrap(err, "while loading manifests")
		}

		if len(files) == 0 {
			return nil, fmt.Errorf("empty list of files for source %s", source)
		}

		var gitHash []byte
		gitHash, err = getGitHash(rootDir)
		if err != nil {
			return nil, errors.Wrap(err, "while getting `git rev-parse HEAD`")
		}
		sourcesInfo = append(sourcesInfo, sourceInfo{
			dir:     parent,
			files:   files,
			gitHash: gitHash,
		})
	}
	return sourcesInfo, nil
}

func getDestDir(parent string, source string) (string, error) {
	encodePath := path.Clean(encodePath(source, []filterPath{trimSSHkey}))
	if encodePath == "." {
		tempDir, err := createTempDirName("-local")
		if err != nil {
			return "", errors.Wrap(err, "while creating temporary directory for local source")
		}
		return path.Join(parent, tempDir), nil
	}
	return path.Join(parent, encodePath), nil
}

func encodePath(path string, filterPaths []filterPath) string {
	toEscape := path
	for _, filterPath := range filterPaths {
		toEscape = filterPath(toEscape)
	}
	return url.QueryEscape(toEscape)
}

func createTempDirName(suffix string) (string, error) {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randBytes) + suffix, nil
}

func filesAlreadyExists(container map[string]struct{}, files []string) error {
	for _, file := range files {
		shortPath := getPathWithoutRootDir(file)
		if _, ok := container[shortPath]; ok {
			return fmt.Errorf("duplicate path for: %s", shortPath)
		}
		container[shortPath] = struct{}{}
	}
	return nil
}

func getPathWithoutRootDir(fullPath string) string {
	parts := strings.Split(fullPath, string(os.PathSeparator))
	return path.Join(parts[4:]...)
}

func trimSSHkey(s string) string {
	if idx := strings.LastIndex(s, "&sshkey="); idx != -1 {
		return s[:idx]
	}
	return s
}

func runDBPopulate(ctx context.Context, cfg dbpopulator.Config, session neo4j.Session, log *zap.Logger, source sourceInfo) (err error) {
	start := time.Now()
	err = retry.Do(func() error {
		populated, err := dbpopulator.Populate(
			ctx, log, session, source.files, source.dir, fmt.Sprintf("%s:%d", cfg.JSONPublishAddr, cfg.JSONPublishPort))
		if err != nil {
			log.Error("Cannot populate a new data", zap.String("error", err.Error()))
			return err
		}
		if populated {
			end := time.Now()
			log.Info("Populated new data", zap.Duration("duration (seconds)", end.Sub(start)))
		}
		return nil
	}, retry.Attempts(6), retry.Delay(30*time.Second))
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
