// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/testhelpers"
)

var (
	iterationBucketSlug       = fmt.Sprintf("alpine-acc-itertest-%s", time.Now().Format("200601021504"))
	iterationUbuntuBucketSlug = fmt.Sprintf("ubuntu-acc-itertest-%s", time.Now().Format("200601021504"))
	iterationChannelSlug      = "production-iter-test"
)

var (
	iterationConfigAlpineProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "alpine" {
		bucket_name  = %q
		channel = %q
	}
    # we make sure that this won't fail even when revoke_at is not set
	output "revoke_at" {
  		value = data.hcp_packer_iteration.alpine.revoke_at
	}
`, iterationBucketSlug, iterationChannelSlug)

	iterationConfigUbuntuProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "ubuntu" {
		bucket_name  = %q
		channel = %q
	}
`, iterationUbuntuBucketSlug, iterationChannelSlug)
)

func TestAcc_dataSourcePackerIteration(t *testing.T) {
	resourceName := "data.hcp_packer_iteration.alpine"
	fingerprint := "43"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, iterationBucketSlug, iterationChannelSlug, false)
			deleteIteration(t, iterationBucketSlug, fingerprint, false)
			deleteBucket(t, iterationBucketSlug, false)
			return nil
		},

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					upsertBucket(t, iterationBucketSlug)
					upsertIteration(t, iterationBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, iterationBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, iterationBucketSlug, fingerprint, itID)
					upsertChannel(t, iterationBucketSlug, iterationChannelSlug, itID)
				},
				Config: testhelpers.TestConfig(iterationConfigAlpineProduction),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
				),
			},
		},
	})
}

func TestAcc_dataSourcePackerIteration_revokedIteration(t *testing.T) {
	resourceName := "data.hcp_packer_iteration.ubuntu"
	fingerprint := "43"
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(5 * time.Minute))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, iterationUbuntuBucketSlug, iterationChannelSlug, false)
			deleteIteration(t, iterationUbuntuBucketSlug, fingerprint, false)
			deleteBucket(t, iterationUbuntuBucketSlug, false)
			return nil
		},

		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, iterationUbuntuBucketSlug)
					upsertIteration(t, iterationUbuntuBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, iterationUbuntuBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, iterationUbuntuBucketSlug, fingerprint, itID)
					upsertChannel(t, iterationUbuntuBucketSlug, iterationChannelSlug, itID)
					// Schedule revocation to the future, otherwise we won't be able to revoke an iteration that
					// it's assigned to a channel
					revokeIteration(t, itID, iterationUbuntuBucketSlug, revokeAt)
					// Sleep to make sure the iteration is revoked when we test
					time.Sleep(5 * time.Second)
				},
				Config: testhelpers.TestConfig(iterationConfigUbuntuProduction),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "revoke_at", revokeAt.String()),
				),
			},
		},
	})
}
