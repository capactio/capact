package runner

import (
	"encoding/json"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Config holds whole configuration for Manager.
type Config struct {
	InputPath     string
	LoggerDevMode bool `envconfig:"default=false"`
}

type InputData struct {
	Context ExecutionContext `json:"context"`
	Args    json.RawMessage  `json:"args"`
}

// ExecutionContext holds configuration directly connected with specific runner.
type ExecutionContext struct {
	Name     string                   `json:"name"`
	DryRun   bool                     `json:"dryRun"`
	Timeout  Duration                 `json:"timeout"`
	Platform KubernetesPlatformConfig `json:"platform"`
}

// KubernetesPlatformConfig holds Kubernetes specific configuration that can be utilized by K8s runners.
type KubernetesPlatformConfig struct {
	Namespace          string            `json:"namespace"`
	ServiceAccountName string            `json:"serviceAccountName"`
	OwnerRef           v1.OwnerReference `json:"ownerRef"`
}

// Duration implements own unmarshal function to solve problem with:
// `json: cannot unmarshal string into Go struct field of type time.Duration`
type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var value string
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}
	tmp, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	*d = Duration(tmp)
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}
