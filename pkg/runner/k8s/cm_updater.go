package k8s

import "sigs.k8s.io/controller-runtime/pkg/client"

type ConfigMapStatusSetter struct {
	cli client.Client
}

func NewConfigMapStatusReporter(cli client.Client) *ConfigMapStatusSetter {
	return &ConfigMapStatusSetter{
		cli: cli,
	}
}


func (c *ConfigMapStatusSetter) Report(status interface{})  error {



	return nil
}
