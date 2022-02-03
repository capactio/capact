package manifest_test

import (
	"testing"
)

func TestValidateTI(t *testing.T) {
	//tests := map[string]struct {
	//	typeInstance gqllocalapi.CreateTypeInstanceInput
	//	typeRevision hubpublicgraphql.TypeRevision
	//	expErrors    []error
	//}{
	//	"": {
	//		typeInstance: gqllocalapi.CreateTypeInstanceInput{
	//			TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
	//				Path:     "test",
	//				Revision: "0.1.0",
	//			},
	//			Value: manifest.TypeJSONSchema{
	//				Properties: map[string]struct {
	//					Id    string "json:\"$id\""
	//					Type  string "json:\"type\""
	//					Title string "json:\"title\""
	//				}{
	//					"test":  {},
	//					"test2": {},
	//					"test3": {},
	//				},
	//			},
	//		},
	//	},
	//}
	//
	//for tn, tc := range tests {
	//	t.Run(tn, func(t *testing.T) {
	//		//given
	//		hubCli := fakeHub{}
	//
	//		// when
	//		validationResults, err := manifest.ValidateTI(context.Background(),&tc.typeInstance,hubCli)
	//
	//		// then
	//		require.NoError(t, err)
	//		assert.Equal(t, tc.expErrors, validationResults.Errors)
	//	})
	//}
}
