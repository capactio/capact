package typeinstance

import (
	"capact.io/capact/internal/cli/typeinstance"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
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

	mapBackendContext := func() interface{} {
		if in.LatestResourceVersion.Spec.Backend == nil {
			return nil
		}
		return in.LatestResourceVersion.Spec.Backend.Context
	}

	if in.LatestResourceVersion != nil {
		out.TypeInstance = &gqllocalapi.UpdateTypeInstanceInput{
			Attributes: mapAttrs(),
			Value:      mapSpecValue(),
			Backend: &gqllocalapi.UpdateTypeInstanceBackendInput{
				Context: mapBackendContext(),
			},
		}
	}

	return out
}

// setTypeInstanceDataForMarshaling sets TypeInstance data based on backend data.
func setTypeInstanceDataForMarshaling(backendData *typeinstance.StorageBackendData, in *gqllocalapi.UpdateTypeInstancesInput) {
	if backendData == nil {
		return
	}
	if backendData.ContextSchema == nil {
		in.TypeInstance.Backend = nil
	}
	if !backendData.AcceptValue {
		in.TypeInstance.Value = nil
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
