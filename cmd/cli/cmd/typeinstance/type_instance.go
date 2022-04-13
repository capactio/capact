package typeinstance

import (
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	storagebackend "capact.io/capact/pkg/hub/storage-backend"
	"github.com/spf13/cobra"
)

const (
	decodeBufferSize = 4096
)

// NewCmd returns a cobra.Command for TypeInstance related operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "typeinstance",
		Aliases: []string{"ti", "TypeInstance", "typeinstances"},
		Short:   "This command consists of multiple subcommands to interact with target TypeInstances",
	}

	root.AddCommand(
		NewCreate(),
		NewDelete(),
		NewGet(),
		NewEdit(),
		NewApply(),
	)
	return root
}

func mapTypeInstanceToUpdateType(in *gqllocalapi.TypeInstance) gqllocalapi.UpdateTypeInstancesInput {
	out := gqllocalapi.UpdateTypeInstancesInput{
		OwnerID: in.LockedBy,
		ID:      in.ID,
	}

	mapAttrs := func() []*gqllocalapi.AttributeReferenceInput {
		if in.LatestResourceVersion.Metadata == nil || in.LatestResourceVersion.Metadata.Attributes == nil {
			return []*gqllocalapi.AttributeReferenceInput{}
		}

		// An empty slice json.Marshal into "[]"
		// whereas a nil slice json.Marshal into "null"
		out := []*gqllocalapi.AttributeReferenceInput{}
		for _, attr := range in.LatestResourceVersion.Metadata.Attributes {
			out = append(out, &gqllocalapi.AttributeReferenceInput{
				Path:     attr.Path,
				Revision: attr.Revision,
			})
		}
		return out
	}

	mapSpecValue := func() interface{} {
		if in.LatestResourceVersion.Spec == nil {
			return nil
		}
		return in.LatestResourceVersion.Spec.Value
	}

	mapBackend := func() *gqllocalapi.UpdateTypeInstanceBackendInput {
		if in.LatestResourceVersion.Spec.Backend == nil {
			return nil
		}
		return &gqllocalapi.UpdateTypeInstanceBackendInput{
			Context: in.LatestResourceVersion.Spec.Backend.Context,
		}
	}

	if in.LatestResourceVersion != nil {
		out.TypeInstance = &gqllocalapi.UpdateTypeInstanceInput{
			Attributes: mapAttrs(),
			Value:      mapSpecValue(),
			Backend:    mapBackend(),
		}
	}

	return out
}

// setTypeInstanceValueForMarshaling sets TypeInstance value based on backend data.
func setTypeInstanceValueForMarshaling(typeInstanceValue *storagebackend.TypeInstanceValue, in *gqllocalapi.UpdateTypeInstancesInput) {
	if typeInstanceValue == nil {
		return
	}
	if typeInstanceValue.ContextSchema == nil {
		in.TypeInstance.Backend = nil
	}
	if !typeInstanceValue.AcceptValue {
		in.TypeInstance.Value = nil
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
