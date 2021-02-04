// +build integration

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"projectvoltron.dev/voltron/internal/ptr"
	graphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
)

var _ = Describe("GraphQL API", func() {
	cli := getGraphQLClient()
	ctx := context.Background()

	Context("Public OCH", func() {
		It("lists interfaces", func() {
			interfaces, err := cli.ListInterfacesMetadata(ctx)

			Expect(err).ToNot(HaveOccurred())
			Expect(interfaces).To(HaveLen(2))
		})
	})

	Context("Local OCH", func() {
		It("creates and deletes TypeInstance", func() {
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

			typeInstance, err := cli.GetTypeInstance(ctx, createdTypeInstance.Metadata.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(typeInstance).To(Equal(&graphql.TypeInstance{
				ResourceVersion: 1,
				Metadata: &graphql.TypeInstanceMetadata{
					ID: createdTypeInstance.Metadata.ID,
					Attributes: []*graphql.AttributeReference{
						{
							Path:     "com.voltron.attribute1",
							Revision: "0.1.0",
						},
					},
				},
				Spec: &graphql.TypeInstanceSpec{
					TypeRef: &graphql.TypeReference{
						Path:     "com.voltron.ti",
						Revision: "0.1.0",
					},
					Value: map[string]interface{}{
						"foo": "bar",
					},
				},
			}))

			err = cli.DeleteTypeInstance(ctx, typeInstance.Metadata.ID)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
