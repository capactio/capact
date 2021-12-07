package validate

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"capact.io/capact/internal/cli"
	"capact.io/capact/pkg/sdk/validation/manifest"

	"github.com/fatih/color"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/schema"
	"github.com/pkg/errors"
)

// Options struct defines validation options for OCF manifest validation.
type Options struct {
	SchemaLocation  string
	ServerSide      bool
	RecursiveSearch bool
	MaxConcurrency  int
}

// Validate validates the Options struct fields.
func (o *Options) Validate() error {
	if o.MaxConcurrency < 1 {
		return errors.New("concurrency parameter cannot be less than 1")
	}

	return nil
}

// ValidationResult defines a validation error.
type ValidationResult struct {
	Path   string
	Errors []error
}

// IsSuccess returns if there were any validation errors.
func (r *ValidationResult) IsSuccess() bool {
	return len(r.Errors) == 0
}

// Error returns error message based on the ValidationResult data.
func (r *ValidationResult) Error() string {
	if r == nil || len(r.Errors) == 0 {
		return ""
	}

	var errMsgs []string
	for _, err := range r.Errors {
		errMsgs = append(errMsgs, err.Error())
	}

	return fmt.Sprintf("%q:\n    * %s\n", r.Path, strings.Join(errMsgs, "\n    * "))
}

// respectedManifestsExt defines valid extensions for OCF manifest files.
// Manifests with different extensions are ignored during validation process.
var respectedManifestsExt = map[string]struct{}{
	".yaml": {},
	".yml":  {},
}

// Validation provides functionality to validate OCF manifests.
type Validation struct {
	hubCli          client.Hub
	writer          io.Writer
	maxWorkers      int
	recursiveSearch bool
	validatorFn     func() manifest.FileSystemValidator
}

// New creates new Validation.
func New(writer io.Writer, opts Options) (*Validation, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	server := config.GetDefaultContext()
	fs, ocfSchemaRootPath := schema.NewProvider(opts.SchemaLocation).FileSystem()

	var (
		hubCli        client.Hub
		err           error
		validatorOpts []manifest.ValidatorOption
	)

	if opts.ServerSide {
		hubCli, err = client.NewHub(server)
		if err != nil {
			return nil, errors.Wrap(err, "while creating Hub client")
		}

		validatorOpts = append(validatorOpts, manifest.WithRemoteChecks(hubCli))
	}

	return &Validation{
		// TODO: To improve: Share a single validator for all workers.
		//		Current implementation makes OCF JSON schemas caching separated per validationWorker.
		//		That enforces thread-safe JSON validator implementations. OCF Schema validator is not thread-safe.
		validatorFn: func() manifest.FileSystemValidator {
			return manifest.NewDefaultFilesystemValidator(fs, ocfSchemaRootPath, validatorOpts...)
		},
		hubCli:          hubCli,
		writer:          writer,
		recursiveSearch: opts.RecursiveSearch,
		maxWorkers:      opts.MaxConcurrency,
	}, nil
}

// Run runs validation across all JSON validators.
func (v *Validation) Run(ctx context.Context, paths []string) error {
	filePaths, err := v.getFilesToParse(paths)
	if err != nil {
		return errors.Wrap(err, "while collecting files for validation")
	}

	var workersCount = v.maxWorkers
	if len(filePaths) < workersCount {
		workersCount = len(filePaths)
	}

	v.printIntroMessage(filePaths, workersCount)

	jobsCh := make(chan string, len(filePaths))
	resultsCh := make(chan ValidationResult, len(filePaths))

	var wg sync.WaitGroup
	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		worker := newValidationWorker(&wg, v.validatorFn())
		go worker.Do(ctx, jobsCh, resultsCh)
	}

	for _, filepath := range filePaths {
		jobsCh <- filepath
	}
	close(jobsCh)

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var processedFilesCount, errsCount int
	for res := range resultsCh {
		processedFilesCount++
		errsCount += len(res.Errors)
		v.printPartialResult(res)
	}

	return v.outputResultSummary(processedFilesCount, errsCount)
}

func (v *Validation) printIntroMessage(filePaths []string, workersCount int) {
	fileNoun := properNounFor("file", len(filePaths))
	fmt.Fprintf(v.writer, "Validating %s in %d concurrent %s...\n", fileNoun, workersCount, properNounFor("job", workersCount))
}

func (v *Validation) outputResultSummary(processedFilesCount int, errsCount int) error {
	fileNoun := properNounFor("file", processedFilesCount)
	fmt.Fprintf(v.writer, "\nValidated %d %s in total.\n", processedFilesCount, fileNoun)

	if errsCount > 0 {
		errNoun := properNounFor("error", errsCount)
		return fmt.Errorf("detected %d validation %s", errsCount, errNoun)
	}

	fmt.Fprintf(v.writer, "🚀 No errors detected.\n")
	return nil
}

func (v *Validation) printPartialResult(res ValidationResult) {
	if !res.IsSuccess() {
		fmt.Fprintf(v.writer, "- %s %s\n", color.RedString("✗"), res.Error())
		return
	}

	// Print successes only in verbose mode
	if !cli.VerboseMode.IsEnabled() {
		return
	}
	fmt.Fprintf(v.writer, "- %s %q\n", color.GreenString("✓"), res.Path)
}

func (v *Validation) getFilesToParse(paths []string) ([]string, error) {
	var files []string

	for _, path := range paths {
		if !isDir(path) {
			files = append(files, path)
			continue
		}

		var fileList []string
		var err error
		if v.recursiveSearch {
			fileList, err = getListOfFilesRecursive(path, isCorrectExtension)
		} else {
			fileList, err = getListOfFilesFromSingleDir(path, isCorrectExtension)
		}
		if err != nil {
			return nil, err
		}
		files = append(files, fileList...)
	}
	return removeDuplicatePaths(files), nil
}

type filterFun func(string) bool

func getListOfFilesRecursive(root string, filterFn filterFun) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filterFn(path) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func getListOfFilesFromSingleDir(path string, filterFn filterFun) ([]string, error) {
	var files []string
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return files, err
	}
	defer file.Close()

	fileNames, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	for _, name := range fileNames {
		fullPath := filepath.Join(path, name)
		if isDir(fullPath) || !filterFn(fullPath) {
			continue
		}
		files = append(files, fullPath)
	}
	return files, nil
}

func removeDuplicatePaths(paths []string) []string {
	var result []string
	allPaths := make(map[string]struct{})
	for _, path := range paths {
		if _, ok := allPaths[path]; ok {
			continue
		}
		allPaths[path] = struct{}{}
		result = append(result, path)
	}
	return result
}

func isDir(in string) bool {
	f, err := os.Stat(in)
	return err == nil && f.IsDir()
}

func isCorrectExtension(path string) bool {
	_, found := respectedManifestsExt[filepath.Ext(path)]
	return found
}

type validationWorker struct {
	wg        *sync.WaitGroup
	validator manifest.FileSystemValidator
}

func newValidationWorker(wg *sync.WaitGroup, validator manifest.FileSystemValidator) *validationWorker {
	return &validationWorker{wg: wg, validator: validator}
}

// Do executes the validationWorker logic.
func (w *validationWorker) Do(ctx context.Context, jobCh <-chan string, resultCh chan<- ValidationResult) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case filePath, ok := <-jobCh:
			if !ok {
				return
			}

			var resultErrs []error
			res, err := w.validator.Do(ctx, filePath)
			if err != nil {
				resultErrs = append(resultErrs, errors.Wrap(err, "internal:"))
			} else {
				resultErrs = append(resultErrs, res.Errors...)
			}

			resultCh <- ValidationResult{
				Path:   filePath,
				Errors: resultErrs,
			}
		}
	}
}

func properNounFor(str string, numberOfItems int) string {
	if numberOfItems == 1 {
		return str
	}

	return str + "s"
}
