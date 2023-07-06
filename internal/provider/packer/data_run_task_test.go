// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packer

import (
	"context"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/testhelpers"
)

func TestAcc_dataSourcePackerRunTask(t *testing.T) {
	runTask := testAccPackerDataRunTaskBuilder("runTask")
	config := testhelpers.BuildTestConfig(runTask)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
		},
		ProviderFactories: testhelpers.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  testAccCheckPackerRunTaskStateMatchesAPI(runTask.ResourceName()),
			},
			{ // Change HMAC and check that it updates in state
				PreConfig: func() {
					client := testhelpers.DefaultProvider().Meta().(*clients.Client)
					loc := &sharedmodels.HashicorpCloudLocationLocation{
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}
					_, err := clients.RegenerateHMAC(context.Background(), client, loc)
					if err != nil {
						t.Errorf("error while regenerating HMAC key: %v", err)
						return
					}
				},
				Config: config,
				Check:  testAccCheckPackerRunTaskStateMatchesAPI(runTask.ResourceName()),
			},
		},
	})
}

func testAccPackerDataRunTaskBuilder(uniqueName string) testhelpers.ConfigBuilder {
	return testhelpers.NewDataConfigBuilder(
		"hcp_packer_run_task",
		uniqueName,
		nil,
	)
}
