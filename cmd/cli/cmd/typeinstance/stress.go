package typeinstance

import (
	"capact.io/capact/pkg/httputil"
	"fmt"
	"math/rand"
	"os"
	"time"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"

	"github.com/spf13/cobra"
)

const (
	megabyte = 1024 * 1000
	creator  = "stressor"
)

// NewStress returns a cobra.Command for creating a TypeInstance on a Local Hub.
func NewStress() *cobra.Command {
	var (
		number      int
		payloadSize float32
	)

	cmd := &cobra.Command{
		Use:   "stress",
		Short: "Creates N new TypeInstances",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			server := config.GetDefaultContext()

			hubCli, err := client.NewHub(server, httputil.WithTimeout(90*time.Second))
			if err != nil {
				return err
			}

			switch args[0] {
			case "create":
				for _, ti := range fixedTypeInstances(number, int(payloadSize*megabyte)) {
					// we need to send one per req, to do not be timed out
					_, err = hubCli.CreateTypeInstance(cmd.Context(), ti)
					if err != nil {
						return err
					}
				}

			case "clean-up":
				out, err := hubCli.ListTypeInstances(cmd.Context(), &gqllocalapi.TypeInstanceFilter{
					CreatedBy: ptr.String(creator),
				})
				if err != nil {
					return err
				}
				var gotIDs []string
				for _, ti := range out {
					gotIDs = append(gotIDs, ti.ID)
				}

				return deleteTI(cmd.Context(), gotIDs, os.Stdout)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&number, "repeat", "r", 5, "repeat create request")
	flags.Float32VarP(&payloadSize, "payload-size", "s", 0.5, "payload size for `value`")

	return cmd
}

func fixedTypeInstances(n, sizeBytes int) []*gqllocalapi.CreateTypeInstanceInput {
	var typeInstances []*gqllocalapi.CreateTypeInstanceInput
	for i := 1; i <= n; i++ {
		ti := gqllocalapi.CreateTypeInstanceInput{
			Alias:     ptr.String(fmt.Sprintf("id_%d", i)),
			CreatedBy: ptr.String(creator),
			TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
				Path:     "cap.type.terraform.state",
				Revision: "0.1.0",
			},
			Value: body{Data: randStringBytes(sizeBytes)},
		}
		typeInstances = append(typeInstances, &ti)
	}
	return typeInstances
}

type body struct {
	Data string `json:"data"`
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		//nolint:gosec
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
