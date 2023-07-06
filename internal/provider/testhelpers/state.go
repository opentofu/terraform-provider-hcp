// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testhelpers

import (
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/links"
	"github.com/hashicorp/terraform-provider-hcp/internal/location"
)

// If the resource is not found, the value will be nil and an error is returned.
// If the attribute is not found, the value will be a blank string, but an error will still be returned.
func GetAttributeFromResourceInState(resourceName string, attribute string, state *terraform.State) (*string, error) {
	resources := state.RootModule().Resources

	resource, ok := resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource %q not found in the present state", resourceName)
	}

	value, ok := resource.Primary.Attributes[attribute]
	if !ok {
		return &value, fmt.Errorf("resource %q does not have an attribute named %q in the present state", resourceName, attribute)
	}

	return &value, nil
}

// Returns a best-effort location from the state of a given resource.
// Will return the default location even if the resource isn't found.
func GetLocationFromState(resourceName string, state *terraform.State) (*sharedmodels.HashicorpCloudLocationLocation, error) {
	client := DefaultProvider().Meta().(*clients.Client)

	projectIDFromState, _ := GetAttributeFromResourceInState(resourceName, "project_id", state)
	if projectIDFromState == nil {
		return &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: client.Config.OrganizationID,
			ProjectID:      client.Config.ProjectID,
		}, fmt.Errorf("resource %q not found in the present state", resourceName)
	}

	projectID, _ := location.GetProjectID(*projectIDFromState, client.Config.OrganizationID)

	return &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}, nil
}

// Checks that the atrribute's value is not the same as diffVal
func CheckResourceAttrDifferent(resourceName string, attribute string, diffVal string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		stateValPtr, err := GetAttributeFromResourceInState(resourceName, attribute, state)
		if err != nil {
			return err
		}

		if *stateValPtr == diffVal {
			return fmt.Errorf("%s: Attribute '%s' expected to not be %#v, but it was", resourceName, attribute, diffVal)
		}

		return nil
	}
}

// Same as `testAccCheckResourceAttrDifferent`, but diffVal is a pointer that is read at check-time
func CheckResourceAttrPtrDifferent(resourceName string, attribute string, diffVal *string) resource.TestCheckFunc {
	if diffVal == nil {
		panic("diffVal cannot be nil")
	}
	return func(state *terraform.State) error {
		return CheckResourceAttrDifferent(resourceName, attribute, *diffVal)(state)
	}
}

func CheckLinkFromState(resourceName, fieldName, expectedID, expectedType, projectIDSourceResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		selfLink, err := GetAttributeFromResourceInState(resourceName, fieldName, s)
		if err != nil {
			return err
		}

		projectID, err := GetAttributeFromResourceInState(projectIDSourceResource, "project_id", s)
		if err != nil {
			return err
		}

		link, err := links.LinkURL(&sharedmodels.HashicorpCloudLocationLink{
			ID: expectedID,
			Location: &sharedmodels.HashicorpCloudLocationLocation{
				ProjectID: *projectID},
			Type: expectedType,
		})
		if err != nil {
			return fmt.Errorf("unable to build link: %v", err)
		}

		if link != *selfLink {
			return fmt.Errorf("incorrect self_link, expected: %s, got: %s", link, *selfLink)
		}

		return nil
	}
}
