package action

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"
)

const (
	waitPollInterval = time.Second
	forSeparator     = "="
)

type conditionFunc func(*gqlengine.Action) error

// WaitOptions holds wait related configuration.
type WaitOptions struct {
	ActionName string
	Namespace  string
	Timeout    time.Duration
	For        string
}

// Wait waits for a given Action's condition.
func Wait(ctx context.Context, opts WaitOptions, w io.Writer) (err error) {
	status := printer.NewStatus(w, "")
	defer func() {
		status.End(err == nil)
	}()

	server := config.GetDefaultContext()
	actionCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	conditionFn, kvPair, err := conditionFuncFor(opts.For)
	if err != nil {
		return err
	}

	status.Step("Waiting ≤ %s for %q to equal %q", opts.Timeout, kvPair[0], kvPair[1])
	return waitUntilSatisfied(ctx, actionCli, conditionFn, opts)
}

func waitUntilSatisfied(ctx context.Context, actCli client.ClusterClient, conditionFn conditionFunc, opts WaitOptions) error {
	var (
		lastErr   error
		ctxWithNs = namespace.NewContext(ctx, opts.Namespace)
	)

	err := wait.PollImmediate(waitPollInterval, opts.Timeout, func() (done bool, err error) {
		if ctxWithNs.Err() != nil {
			return false, ctxWithNs.Err()
		}

		act, err := actCli.GetAction(ctxWithNs, opts.ActionName)
		if err != nil { // may be network issue, ignoring
			lastErr = err
			return false, nil
		}

		if act == nil {
			return true, fmt.Errorf("Action %q not found in Namespace %q", opts.ActionName, opts.Namespace)
		}

		if err := conditionFn(act); err != nil {
			lastErr = err // condition not met yet
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return lastErr
		}
		return err
	}

	return nil
}

func conditionFuncFor(condition string) (conditionFunc, []string, error) {
	items := strings.SplitN(condition, forSeparator, 2)
	if len(items) != 2 {
		return nil, nil, fmt.Errorf("invalid format, require 'condition=condition-value'")
	}

	conditionName := items[0]
	conditionValue := items[1]

	switch conditionName {
	case "phase":
		return func(act *gqlengine.Action) error {
			if act.Status.Phase != gqlengine.ActionStatusPhase(conditionValue) {
				return fmt.Errorf("the Action still doesn't have phase=%s", conditionValue)
			}
			return nil
		}, items, nil
	default:
		return nil, nil, fmt.Errorf("unrecognized condition: %q", condition)
	}
}
