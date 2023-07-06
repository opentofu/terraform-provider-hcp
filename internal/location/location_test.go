// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/location"
	"github.com/stretchr/testify/assert"
)

// Project ID is read from the resource config or provider
func Test_GetProjectID(t *testing.T) {
	tests := []struct {
		name        string
		resProjID   string
		clientProj  string
		expectedID  string
		expectedErr string
	}{
		{"resource only project defined", "proj1", "", "proj1", ""},
		{"provider only project defined", "", "proj2", "proj2", ""},
		{"resource and provider project defined", "proj1", "proj2", "proj1", ""},
		{"project not defined", "", "", "", "project ID not defined. Verify that project ID is set either in the provider or in the resource config"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			projID, err := location.GetProjectID(testCase.resProjID, testCase.clientProj)
			assert.Equal(t, testCase.expectedID, projID)

			if testCase.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, testCase.expectedErr)
			}
		})
	}
}

// If project ID is not defined on the provider or resource config, the provider
// project ID becomes the organization's oldest existing project
func TestDetermineOldestProject(t *testing.T) {

	testCases := []struct {
		name           string
		projArray      []*models.ResourcemanagerProject
		expectedProjID string
	}{
		{
			name: "One Project",
			projArray: []*models.ResourcemanagerProject{
				{
					CreatedAt: strfmt.DateTime(time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					ID:        "proj1",
				},
			},
			expectedProjID: "proj1",
		},
		{
			name: "Two Projects",
			projArray: []*models.ResourcemanagerProject{
				{
					ID:        "proj1",
					CreatedAt: strfmt.DateTime(time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				{
					ID:        "proj2",
					CreatedAt: strfmt.DateTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			expectedProjID: "proj2",
		},
		{
			name: "Three Projects",
			projArray: []*models.ResourcemanagerProject{
				{
					ID:        "proj1",
					CreatedAt: strfmt.DateTime(time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				{
					ID:        "proj2",
					CreatedAt: strfmt.DateTime(time.Date(2007, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				{
					ID:        "proj3",
					CreatedAt: strfmt.DateTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			expectedProjID: "proj2",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			oldestProject := location.GetOldestProject(testCase.projArray)
			assert.Equal(t, testCase.expectedProjID, oldestProject.ID)

		})

	}
}
