package secretstoragebackend

import tellercore "github.com/spectralops/teller/pkg/core"

func (h *Handler) GetProviderFromContext(contextBytes []byte) (tellercore.Provider, []byte, error) {
	return h.getProviderFromContext(contextBytes)
}
