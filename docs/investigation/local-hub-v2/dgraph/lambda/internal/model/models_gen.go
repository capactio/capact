// @generated - This was created as a part of investigation. We mark it as generate to exlude it from goreportcard to do not have missleading issues.:golint
package model

type TypeInstanceBackendReference struct {
	Id       string `json:"id" dql:"uid"`
	Abstract bool   `json:"abstract" dql:"TypeInstanceBackendReference.abstract"`
}
type TypeInstanceResourceVersionSpec struct {
	Id        string                        `json:"id" dql:"uid"`
	Backend   *TypeInstanceBackendReference `json:"backend" dql:"TypeInstanceResourceVersionSpec.backend"`
	BackendID string                        `json:"backendID" dql:"TypeInstanceResourceVersionSpec.backendID"`
	Value     string                        `json:"value" dql:"TypeInstanceResourceVersionSpec.value"`
}
