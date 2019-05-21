package ccv3_test

import (
	"code.cloudfoundry.org/cli/types"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Spaces", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetSpaces", func() {
		var (
			query Query

			spaces     []Space
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			spaces, warnings, executeErr = client.GetSpaces(query)
		})

		When("spaces exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/spaces?names=some-space-name&page=2&per_page=2"
		}
	},
  "resources": [
    {
      "name": "space-name-1",
      "guid": "space-guid-1",
      "relationships": {
        "organization": {
          "data": { "guid": "org-guid-1" }
        }
      }
    },
    {
      "name": "space-name-2",
      "guid": "space-guid-2",
      "relationships": {
        "organization": {
          "data": { "guid": "org-guid-2" }
        }
      }
    }
  ]
}`, server.URL())
				response2 := `{
  "pagination": {
    "next": null
  },
  "resources": [
    {
      "name": "space-name-3",
      "guid": "space-guid-3",
      "relationships": {
        "organization": {
          "data": { "guid": "org-guid-3" }
        }
      }
    }
  ]
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/spaces", "names=some-space-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/spaces", "names=some-space-name&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)

				query = Query{
					Key:    NameFilter,
					Values: []string{"some-space-name"},
				}
			})

			It("returns the queried spaces and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(spaces).To(ConsistOf(
					Space{Name: "space-name-1", GUID: "space-guid-1", Relationships: Relationships{
						constant.RelationshipTypeOrganization: Relationship{GUID: "org-guid-1"},
					}},
					Space{Name: "space-name-2", GUID: "space-guid-2", Relationships: Relationships{
						constant.RelationshipTypeOrganization: Relationship{GUID: "org-guid-2"},
					}},
					Space{Name: "space-name-3", GUID: "space-guid-3", Relationships: Relationships{
						constant.RelationshipTypeOrganization: Relationship{GUID: "org-guid-3"},
					}},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "Space not found",
      "title": "CF-SpaceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/spaces"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Space not found",
							Title:  "CF-SpaceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateSpace", func() {
		var (
			spaceToUpdate Space
			updatedSpace  Space
			warnings      Warnings
			executeErr    error
		)

		JustBeforeEach(func() {
			updatedSpace, warnings, executeErr = client.UpdateSpace(spaceToUpdate)
		})

		When("the organization is updated successfully", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-space-guid",
					"name": "some-space-name",
					"metadata": {
						"labels": {
							"k1":"v1",
							"k2":"v2"
						}
					}
				}`

				expectedBody := map[string]interface{}{
					"name": "some-space-name",
					"metadata": map[string]interface{}{
						"labels": map[string]string{
							"k1": "v1",
							"k2": "v2",
						},
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/spaces/some-guid"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				spaceToUpdate = Space{
					Name: "some-space-name",
					GUID: "some-guid",
					Metadata: &Metadata{
						Labels: map[string]types.NullString{
							"k1": types.NewNullString("v1"),
							"k2": types.NewNullString("v2"),
						},
					},
				}
			})

			It("should include the labels in the JSON", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(3))
				Expect(len(warnings)).To(Equal(1))
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(updatedSpace.Metadata.Labels).To(BeEquivalentTo(
					map[string]types.NullString{
						"k1": types.NewNullString("v1"),
						"k2": types.NewNullString("v2"),
					}))
			})

		})

	})
})
