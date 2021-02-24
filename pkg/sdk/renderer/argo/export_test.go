package argo

func (r *TypeInstanceHandler) SetGenUUID(genUUID func() string) {
	r.genUUID = genUUID
}
