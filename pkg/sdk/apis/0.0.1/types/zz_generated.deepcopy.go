// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package types

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManifestRef) DeepCopyInto(out *ManifestRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManifestRef.
func (in *ManifestRef) DeepCopy() *ManifestRef {
	if in == nil {
		return nil
	}
	out := new(ManifestRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManifestRefWithOptRevision) DeepCopyInto(out *ManifestRefWithOptRevision) {
	*out = *in
	if in.Revision != nil {
		in, out := &in.Revision, &out.Revision
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManifestRefWithOptRevision.
func (in *ManifestRefWithOptRevision) DeepCopy() *ManifestRefWithOptRevision {
	if in == nil {
		return nil
	}
	out := new(ManifestRefWithOptRevision)
	in.DeepCopyInto(out)
	return out
}
