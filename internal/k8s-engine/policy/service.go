package policy

import (
	"context"

	"capact.io/capact/pkg/engine/k8s/policy"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const policyConfigMapKey = "cluster-policy.yaml"

// Service provides functionality to manage Capact Policy configuration.
type Service struct {
	log          *zap.Logger
	k8sCli       client.Client
	policyObjKey client.ObjectKey
}

// NewService returns a new Service instance.
func NewService(log *zap.Logger, actionCli client.Client, cfg Config) *Service {
	return &Service{
		log:    log.With(zap.String("module", "policyService")),
		k8sCli: actionCli,
		policyObjKey: client.ObjectKey{
			Namespace: cfg.Namespace,
			Name:      cfg.Name,
		},
	}
}

// Get returns current Capact Policy configuration.
func (s *Service) Get(ctx context.Context) (policy.Policy, error) {
	cfgMap, err := s.getConfigMap(ctx)
	if err != nil {
		return policy.Policy{}, err
	}

	p, err := policy.FromYAMLString(cfgMap.Data[policyConfigMapKey])
	if err != nil {
		return policy.Policy{},
			errors.Wrapf(err, "while unmarshaling policy from ConfigMap '%s/%s' from %q key",
				s.policyObjKey.Namespace,
				s.policyObjKey.Name,
				policyConfigMapKey,
			)
	}

	return p, nil
}

// Update updates current Capact Policy configuration with a given input.
func (s *Service) Update(ctx context.Context, in policy.Policy) (policy.Policy, error) {
	cfgMap, err := s.getConfigMap(ctx)
	if err != nil {
		return policy.Policy{}, err
	}

	policyStr, err := in.ToYAMLString()
	if err != nil {
		return policy.Policy{}, errors.Wrap(err, "while marshaling Policy")
	}

	cfgMap.Data[policyConfigMapKey] = policyStr

	s.log.Info("Updating Policy")
	err = s.k8sCli.Update(ctx, cfgMap)
	if err != nil {
		errContext := "while updating Policy ConfigMap"
		s.log.Error(errContext, zap.Error(err))
		return policy.Policy{}, errors.Wrap(err, errContext)
	}

	return in, nil
}

func (s *Service) getConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	s.log.Info("Getting Policy")

	policyCfgMap := &corev1.ConfigMap{}

	err := s.k8sCli.Get(ctx, s.policyObjKey, policyCfgMap)
	if err != nil {
		errContext := "while getting ConfigMap from K8s"
		switch {
		case apierrors.IsNotFound(err):
			s.log.Debug(errContext, zap.Error(ErrPolicyConfigMapNotFound))
			return nil, errors.Wrap(ErrPolicyConfigMapNotFound, errContext)
		default:
			s.log.Error(errContext, zap.Error(err))
			return nil, errors.Wrap(err, errContext)
		}
	}

	return policyCfgMap, nil
}
