package terraform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

const (
	variablesFile = "terraform.tfvars"
	stateFile     = "terraform.tfstate"
)

type terraform struct {
	log       *zap.Logger
	workdir   string
	args      Arguments
	_waitCh   chan error
	runOutput []byte
}

func newTerraform(log *zap.Logger, workdir string, args Arguments) *terraform {
	return &terraform{
		log:     log,
		workdir: workdir,
		args:    args,
	}
}

// Start starts the Terraform operation.
func (t *terraform) Start(dryRun bool) error {
	t._waitCh = make(chan error)

	go func() {
		err := t.init()
		if err != nil {
			t._waitCh <- errors.Wrap(err, "while initializing terraform")
			return
		}

		err = t.run(dryRun)
		if err != nil {
			t._waitCh <- errors.Wrap(err, "while running terraform")
			return
		}
		// TODO returning error here is misleading as resources were deployed
		out, err := t.output()
		if err != nil {
			t._waitCh <- errors.Wrap(err, "while running terraform")
			return
		}
		t.runOutput = out
		close(t._waitCh)
	}()
	return nil
}

// Wait blocks until the Terraform operation is not finished.
func (t *terraform) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err, open := <-t._waitCh:
		if !open {
			break
		}
		return err
	}

	return nil
}

func (t *terraform) releaseInfo() ([]byte, error) {
	release := Release{Name: t.args.Module.Name, Source: t.args.Module.Source}
	bytes, err := yaml.Marshal(&release)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling yaml")
	}
	return bytes, nil
}

func (t *terraform) _execute(command string, arg ...string) ([]byte, error) {
	allArgs := []string{command, "-no-color"}
	allArgs = append(allArgs, arg...)

	// #nosec
	cmd := exec.Command("terraform", allArgs...)
	cmd.Dir = t.workdir
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=true")

	t.log.Info("Running terraform command", zap.Strings("args", allArgs))
	out, err := cmd.CombinedOutput()
	t.log.Debug("Terraform output", zap.ByteString("output", out))
	return out, err
}

func (t *terraform) init() error {
	_, err := t._execute("init")
	return err
}

func (t *terraform) output() ([]byte, error) {
	out, err := t._execute("output", "-json")
	return out, err
}

func (t *terraform) run(dryRun bool) error {
	if dryRun {
		return t._plan()
	} else if t.args.Command == ApplyCommand {
		if err := t._plan(); err != nil {
			return err
		}

		return t._apply()
	} else if t.args.Command == DestroyCommand {
		return t._destroy()
	}
	return fmt.Errorf("command `%s` is not supported", t.args.Command)
}

func (t *terraform) _plan() error {
	_, err := t._execute("plan")
	return err
}

func (t *terraform) _apply() error {
	_, err := t._execute("apply", "-auto-approve")
	return err
}

func (t *terraform) _destroy() error {
	_, err := t._execute("destroy", "-auto-approve")
	return err
}

func (t *terraform) tfstate() ([]byte, error) {
	return ioutil.ReadFile(path.Join(t.workdir, "terraform.tfstate"))
}

func (t *terraform) variables() ([]byte, error) {
	return ioutil.ReadFile(path.Join(t.workdir, variablesFile))
}

func (t *terraform) renderOutput() ([]byte, error) {
	if t.args.Output.GoTemplate == "" {
		return []byte{}, nil
	}
	if len(t.runOutput) == 0 {
		return []byte{}, nil
	}

	tmpl, err := template.New("output").Parse(t.args.Output.GoTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load template")
	}

	data, err := outputToMap(t.runOutput)
	if err != nil {
		return nil, errors.Wrap(err, "while converting json output")
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, data); err != nil {
		return nil, errors.Wrap(err, "while rendering output")
	}
	return buff.Bytes(), nil
}

// Converts, terraform output -json
// {
//  "instance_ip_addr": {
//    "value": "35.223.194.84",
//    "type": "string"
// }
// to
// {"instance_ip_addr": "35.223.194.84"}
func outputToMap(jsonOutput []byte) (map[string]interface{}, error) {
	tOutput := map[string]interface{}{}
	err := json.Unmarshal(jsonOutput, &tOutput)
	if err != nil {
		return nil, err
	}

	output := map[string]interface{}{}
	for k, v := range tOutput {
		s, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert terraform output, %+v", v)
		}
		output[k] = s["value"]
	}
	return output, nil
}
