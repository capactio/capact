// +build integration

package e2e

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/httputil"
	graphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	"projectvoltron.dev/voltron/pkg/och/client"
)

var _ = Describe("GraphQL API", func() {
	Context("Public OCH", func() {
		It("lists interfaces", func() {
			httpClient := httputil.NewClient(
				20*time.Second,
				true,
				httputil.WithBasicAuth(cfg.Gateway.Username, cfg.Gateway.Password),
			)
			cli := client.NewClient(cfg.Gateway.Endpoint, httpClient)

			interfaces, err := cli.ListInterfacesMetadata(context.Background())

			Expect(err).ToNot(HaveOccurred())
			Expect(interfaces).To(BeEmpty())
		})
	})

	Context("Local OCH", func() {
		It("creates and deletes TypeInstance", func() {
			httpClient := httputil.NewClient(
				20*time.Second,
				true,
				httputil.WithBasicAuth(cfg.Gateway.Username, cfg.Gateway.Password),
			)
			cli := client.NewClient(cfg.Gateway.Endpoint, httpClient)
			ctx := context.Background()

			createdTypeInstance, err := cli.CreateTypeInstance(ctx, &graphql.CreateTypeInstanceInput{
				TypeRef: &graphql.TypeReferenceInput{
					Path:     "com.voltron.ti",
					Revision: ptr.String("0.1.0"),
				},
				Attributes: []*graphql.AttributeReferenceInput{
					{
						Path:     "com.voltron.attribute1",
						Revision: ptr.String("0.1.0"),
					},
				},
				Value: map[string]interface{}{
					"foo": "bar",
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(createdTypeInstance.Spec.Value).To(Equal(map[string]interface{}{
				"foo": "bar",
			}))

			typeInstance, err := cli.GetTypeInstance(ctx, createdTypeInstance.Metadata.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdTypeInstance.Spec.Value).To(Equal(map[string]interface{}{
				"foo": "bar",
			}))

			err = cli.DeleteTypeInstance(ctx, typeInstance.Metadata.ID)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
