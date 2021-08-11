package manifest

import "capact.io/capact/pkg/sdk/apis/0.0.1/types"

// ValidatorOption is used to provide additional configuration options for the validation.
type ValidatorOption func(validator *FSValidator)

// WithRemoteChecks enables validation checks for manifests against Capact Hub.
func WithRemoteChecks(hubCli Hub) ValidatorOption {
	return func(r *FSValidator) {
		r.kindValidators[types.TypeManifestKind] = append(r.kindValidators[types.TypeManifestKind], NewRemoteTypeValidator(hubCli))
		r.kindValidators[types.InterfaceManifestKind] = append(r.kindValidators[types.InterfaceManifestKind], NewRemoteInterfaceValidator(hubCli))
		r.kindValidators[types.ImplementationManifestKind] = append(r.kindValidators[types.ImplementationManifestKind], NewRemoteImplementationValidator(hubCli))
	}
}
