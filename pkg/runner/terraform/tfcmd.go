package terraform

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// TODO: Refactor and use Terraform Go client instead of the binary

type tfCmd struct {
	workDir string
	log     *zap.Logger
}

func newTFCmd(log *zap.Logger, workDir string) *tfCmd {
	return &tfCmd{
		log:     log,
		workDir: workDir,
	}
}

func (t *tfCmd) Init(ctx context.Context) error {
	return t.executeAndStreamOutput(ctx, "init")
}

func (t *tfCmd) Plan(ctx context.Context) error {
	return t.executeAndStreamOutput(ctx, "plan")
}

func (t *tfCmd) Apply(ctx context.Context) error {
	return t.executeAndStreamOutput(ctx, "apply", "-auto-approve")
}

func (t *tfCmd) Destroy(ctx context.Context) error {
	return t.executeAndStreamOutput(ctx, "destroy", "-auto-approve")
}

func (t *tfCmd) Output() ([]byte, error) {
	cmd := t.cmd("output", "-json")
	t.log.Info("Running command", zap.Strings("args", cmd.Args))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "while waiting for command to exit")
	}

	return out, nil
}

func (t *tfCmd) executeAndStreamOutput(ctx context.Context, command string, args ...string) error {
	cmd := t.cmd(command, args...)
	t.log.Info("Running command", zap.Strings("args", cmd.Args))

	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "while getting stdout pipe")
	}

	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "while getting stderr pipe")
	}

	err = cmd.Start()
	if err != nil {
		return errors.Wrap(err, "while starting command")
	}

	t.log.Info("Piping the Terraform output to stdout and stderr")

	cmdOutLogger := t.log.
		Named("tf").
		WithOptions(zap.WithCaller(false), zap.AddStacktrace(zap.PanicLevel))

	t.readAndPrintConcurrently(ctx, stdOut, func(s string) { cmdOutLogger.Info(s) })
	t.readAndPrintConcurrently(ctx, stdErr, func(s string) { cmdOutLogger.Error(s) })

	err = cmd.Wait()
	if err != nil {
		return errors.Wrap(err, "while waiting for command to exit")
	}

	return nil
}

func (t *tfCmd) cmd(command string, args ...string) *exec.Cmd {
	allArgs := []string{command, "-no-color"}
	allArgs = append(allArgs, args...)

	// #nosec
	cmd := exec.Command("terraform", allArgs...)
	cmd.Dir = t.workDir
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=true")

	return cmd
}

func (t *tfCmd) readAndPrintConcurrently(ctx context.Context, pipe io.ReadCloser, printFn func(string)) {
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanLines)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for scanner.Scan() {
					text := scanner.Text()
					printFn(text)
				}
			}
		}
	}(ctx)
}
