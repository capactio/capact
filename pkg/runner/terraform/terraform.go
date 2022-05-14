package terraform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

const (
	variablesFile = "terraform.tfvars"
	stateFile     = "terraform.tfstate"
)

type terraform struct {
	tfCmd     *tfCmd
	log       *zap.Logger
	args      Arguments
	_waitCh   chan error
	runOutput []byte
}

func newTerraform(log *zap.Logger, args Arguments, tfCmd *tfCmd) *terraform {
	return &terraform{
		log:   log,
		args:  args,
		tfCmd: tfCmd,
	}
}

// Start starts the Terraform operation.
func (t *terraform) Start(ctx context.Context, dryRun bool) error {
	t._waitCh = make(chan error)

	go func() {
		err := t.tfCmd.Init(ctx)
		if err != nil {
			t._waitCh <- errors.Wrap(err, "while initializing terraform")
			return
		}

		err = t.run(ctx, dryRun)
		if err != nil {
			t._waitCh <- errors.Wrap(err, "while running terraform")
			return
		}
		// TODO returning error here is misleading as resources were deployed
		out, err := t.tfCmd.Output()
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

func (t *terraform) run(ctx context.Context, dryRun bool) error {
	if dryRun {
		return t.tfCmd.Plan(ctx)
	}

	switch t.args.Command {
	case ApplyCommand:
		if err := t.tfCmd.Plan(ctx); err != nil {
			return err
		}

		return t.tfCmd.Apply(ctx)
	case DestroyCommand:
		return t.tfCmd.Destroy(ctx)
	}

	return fmt.Errorf("command `%s` is not supported", t.args.Command)
}

func (t *terraform) ReadTFStateFile(dir string) ([]byte, error) {
	return t.readFile(path.Join(dir, stateFile))
}

func (t *terraform) ReadVariablesFile(dir string) ([]byte, error) {
	return t.readFile(path.Join(dir, variablesFile))
}

func (t *terraform) readFile(filePath string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, errors.Wrapf(err, "while reading file from path %q", filePath)
	}
	return bytes, nil
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
