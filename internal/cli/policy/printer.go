package policy

import (
	"fmt"
	"io"

	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	"github.com/hokaccha/go-prettyjson"
	"sigs.k8s.io/yaml"
)

// TODO: Migrate to pkg/printer once the PR https://github.com/Project-Voltron/go-voltron/pull/296 is merged

type PrintFormat string

const (
	YAMLFormat = "yaml"
	JSONFormat = "json"
)

type printer func(in *gqlengine.Policy, w io.Writer) error

func selectPrinter(format string) (printer, error) {
	switch format {
	case JSONFormat:
		return func(in *gqlengine.Policy, w io.Writer) error {
			return printJSON(in, w)
		}, nil
	case YAMLFormat:
		return func(in *gqlengine.Policy, w io.Writer) error {
			return printYAML(in, w)
		}, nil
	}

	return nil, fmt.Errorf("Unknown output PrintFormat %q", format)
}

func printJSON(in interface{}, w io.Writer) error {
	out, err := prettyjson.Marshal(in)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}

func printYAML(in interface{}, w io.Writer) error {
	out, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}
