package secretstoragebackend

import tellercore "github.com/spectralops/teller/pkg/core"

func (h *Handler) GetProviderFromContext(contextBytes []byte) (tellercore.Provider, error) {
	return h.getProviderFromContext(contextBytes)
}
