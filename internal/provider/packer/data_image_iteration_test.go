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
	alpineBucketSlug      = fmt.Sprintf("alpine-acc-%s", time.Now().Format("200601021504"))
	ubuntuBucketSlug      = fmt.Sprintf("ubuntu-acc-%s", time.Now().Format("200601021504"))
	productionChannelSlug = fmt.Sprintf("packer-acc-channel-%s", time.Now().Format("200601021504"))
)

var (
	alpineImageConfigProduction = fmt.Sprintf(`
	data "hcp_packer_image_iteration" "alpine" {
		bucket_name  = %q
		channel = %q
	}`, alpineBucketSlug, productionChannelSlug)
	ubuntuImageConfigProduction = fmt.Sprintf(`
	data "hcp_packer_image_iteration" "ubuntu" {
		bucket_name  = %q
		channel = %q
	}`, ubuntuBucketSlug, productionChannelSlug)
)

func TestAcc_dataSourcePacker(t *testing.T) {
	resourceName := "data.hcp_packer_image_iteration.alpine"
	fingerprint := "42"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, alpineBucketSlug, productionChannelSlug, false)
			deleteIteration(t, alpineBucketSlug, fingerprint, false)
			deleteBucket(t, alpineBucketSlug, false)
			return nil
		},

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, alpineBucketSlug)
					upsertIteration(t, alpineBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, alpineBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, alpineBucketSlug, fingerprint, itID)
					upsertChannel(t, alpineBucketSlug, productionChannelSlug, itID)
				},
				Config: testhelpers.TestConfig(alpineImageConfigProduction),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
				),
			},
		},
	})
}

func TestAcc_dataSourcePacker_revokedIteration(t *testing.T) {
	resourceName := "data.hcp_packer_image_iteration.ubuntu"
	fingerprint := "42"
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(5 * time.Minute))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, ubuntuBucketSlug, productionChannelSlug, false)
			deleteIteration(t, ubuntuBucketSlug, fingerprint, false)
			deleteBucket(t, ubuntuBucketSlug, false)
			return nil
		},

		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, ubuntuBucketSlug)
					upsertIteration(t, ubuntuBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, ubuntuBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, ubuntuBucketSlug, fingerprint, itID)
					upsertChannel(t, ubuntuBucketSlug, productionChannelSlug, itID)
					// Schedule revocation to the future, otherwise we won't be able to revoke an iteration that
					// it's assigned to a channel
					revokeIteration(t, itID, ubuntuBucketSlug, revokeAt)
					// Sleep to make sure the iteration is revoked when we test
					time.Sleep(5 * time.Second)
				},
				Config: testhelpers.TestConfig(ubuntuImageConfigProduction),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "revoke_at", revokeAt.String()),
				),
			},
		},
	})
}
