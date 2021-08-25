package policy

import "capact.io/capact/internal/maps"

// DeepCopyInto writes a deep copy of AdditionalParametersToInject into out.
// controller-gen doesn't support interface{} so writing it manually
func (in *AdditionalParametersToInject) DeepCopyInto(out *AdditionalParametersToInject) {
	*out = *in
	out.Value = maps.Merge(out.Value, in.Value)
}

// DeepCopy returns a new deep copy of AdditionalParametersToInject.
// controller-gen doesn't support interface{} so writing it manually
func (in *AdditionalParametersToInject) DeepCopy() *AdditionalParametersToInject {
	if in == nil {
		return nil
	}
	out := new(AdditionalParametersToInject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopy returns a new deep copy of InjectData.
// controller-gen doesn't support interface{} so writing it manually
func (in *InjectData) DeepCopy() *InjectData {
	if in == nil {
		return nil
	}
	out := new(InjectData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto writes a deep copy of InjectData into out.
// controller-gen doesn't support interface{} so writing it manually
func (in *InjectData) DeepCopyInto(out *InjectData) {
	*out = *in
	if in.RequiredTypeInstances != nil {
		in, out := &in.RequiredTypeInstances, &out.RequiredTypeInstances
		*out = make([]RequiredTypeInstanceToInject, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AdditionalParameters != nil {
		in, out := &in.AdditionalParameters, &out.AdditionalParameters
		*out = make([]AdditionalParametersToInject, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}
