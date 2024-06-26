// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceApp(t *testing.T) {
	testAppName := generateRandomSlug()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "hcp_project" "example" {
						name        = "test-project"
					}
					resource "hcp_vault_secrets_app" "example" {
						app_name = %q
						description = "Acceptance test run"
						project_id = hcp_project.example.resource_id
				  }`, testAppName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_app.example", "app_name", testAppName),
					resource.TestCheckResourceAttr("hcp_vault_secrets_app.example", "description", "Acceptance test run"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app.example", "project_id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_secrets_app" "example" {
						app_name = %q
						description = "Acceptance test run"
				  }`, testAppName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_app.example", "app_name", testAppName),
					resource.TestCheckResourceAttr("hcp_vault_secrets_app.example", "description", "Acceptance test run"),
				),
			},
		},
	})
}
