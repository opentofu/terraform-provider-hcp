// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/testhelpers"
)

func TestProvider(t *testing.T) {
	if err := provider.New().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

var projectID = "prov-project-id-invalid"

func TestAccMultiProject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(t *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{

			{
				PlanOnly: true,
				Config: fmt.Sprintf(`
				provider "hcp" {
					project_id     = "%s"
				}
				resource "hcp_hvn" "test" {
					hvn_id         = hvn-id-ex
					cloud_provider = "aws"
					region         = "us-west-2"
				}
				`, projectID),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected "project_id" to be a valid UUID, got %s`, projectID)),
			},
		},
	})

}

func TestAccMultiProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(t *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{

			{
				PlanOnly: true,
				Config: fmt.Sprintf(`
				provider "hcp" {}
				resource "hcp_hvn" "test" {
					hvn_id         = "hvn-id-ex"
					project_id     = "%s"
					cloud_provider = "aws"
					region         = "us-west-2"
				}
				`, projectID),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected "project_id" to be a valid UUID, got %s`, projectID)),
			},
		},
	})

}
