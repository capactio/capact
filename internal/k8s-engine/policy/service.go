package policy

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"capact.io/capact/pkg/engine/k8s/clusterpolicy"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const clusterPolicyConfigMapKey = "cluster-policy.yaml"

type Service struct {
	log                 *zap.Logger
	k8sCli              client.Client
	clusterPolicyObjKey client.ObjectKey
}

func NewService(log *zap.Logger, actionCli client.Client, cfg Config) *Service {
	return &Service{
		log:    log.With(zap.String("module", "policyService")),
		k8sCli: actionCli,
		clusterPolicyObjKey: client.ObjectKey{
			Namespace: cfg.Namespace,
			Name:      cfg.Name,
		},
	}
}

func (s *Service) Get(ctx context.Context) (clusterpolicy.ClusterPolicy, error) {
	cfgMap, err := s.getConfigMap(ctx)
	if err != nil {
		return clusterpolicy.ClusterPolicy{}, err
	}

	policy, err := clusterpolicy.FromYAMLString(cfgMap.Data[clusterPolicyConfigMapKey])
	if err != nil {
		return clusterpolicy.ClusterPolicy{},
			errors.Wrapf(err, "while unmarshaling policy from ConfigMap '%s/%s' from %q key",
				s.clusterPolicyObjKey.Namespace,
				s.clusterPolicyObjKey.Name,
				clusterPolicyConfigMapKey,
			)
	}

	return policy, nil
}

func (s *Service) Update(ctx context.Context, in clusterpolicy.ClusterPolicy) (clusterpolicy.ClusterPolicy, error) {
	cfgMap, err := s.getConfigMap(ctx)
	if err != nil {
		return clusterpolicy.ClusterPolicy{}, err
	}

	policyStr, err := in.ToYAMLString()
	if err != nil {
		return clusterpolicy.ClusterPolicy{}, errors.Wrap(err, "while marshaling Policy")
	}

	cfgMap.Data[clusterPolicyConfigMapKey] = policyStr

	s.log.Info("Updating Policy")
	err = s.k8sCli.Update(ctx, cfgMap)
	if err != nil {
		errContext := "while updating Policy ConfigMap"
		s.log.Error(errContext, zap.Error(err))
		return clusterpolicy.ClusterPolicy{}, errors.Wrap(err, errContext)
	}

	return in, nil
}

func (s *Service) getConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	s.log.Info("Getting Policy")

	policyCfgMap := &corev1.ConfigMap{}

	err := s.k8sCli.Get(ctx, s.clusterPolicyObjKey, policyCfgMap)
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
