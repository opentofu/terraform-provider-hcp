// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/testhelpers"
)

func TestAccPackerChannel(t *testing.T) {
	resourceName := "hcp_packer_channel.production"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, alpineBucketSlug, false)
			return nil
		},

		Steps: []resource.TestStep{
			{
				PreConfig: func() { upsertBucket(t, alpineBucketSlug) },
				Config:    testhelpers.TestConfig(channelConfigBasic(alpineBucketSlug, productionChannelSlug)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", alpineBucketSlug),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", productionChannelSlug),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// Testing that we can import bucket channel created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					bucketName := rs.Primary.Attributes["bucket_name"]
					channelName := rs.Primary.Attributes["name"]
					return fmt.Sprintf("%s:%s", bucketName, channelName), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPackerChannel_AssignedIteration(t *testing.T) {
	resourceName := "hcp_packer_channel.production"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, alpineBucketSlug, false)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					fingerprint := "channel-assigned-iteration"
					upsertBucket(t, alpineBucketSlug)
					upsertIteration(t, alpineBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, alpineBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, alpineBucketSlug, fingerprint, itID)
				},
				Config: testhelpers.TestConfig(channelConfigAssignedLatestIteration(alpineBucketSlug, productionChannelSlug)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", alpineBucketSlug),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.incremental_version"),
					resource.TestCheckResourceAttr(resourceName, "iteration.0.fingerprint", "channel-assigned-iteration"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// Testing that we can import bucket channel created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					bucketName := rs.Primary.Attributes["bucket_name"]
					channelName := rs.Primary.Attributes["name"]
					return fmt.Sprintf("%s:%s", bucketName, channelName), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPackerChannel_UpdateAssignedIteration(t *testing.T) {
	resourceName := "hcp_packer_channel.production"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, alpineBucketSlug, false)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					fingerprint := "channel-update-it1"
					upsertBucket(t, alpineBucketSlug)
					upsertIteration(t, alpineBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, alpineBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, alpineBucketSlug, fingerprint, itID)
				},
				Config: testhelpers.TestConfig(channelConfigAssignedLatestIteration(alpineBucketSlug, productionChannelSlug)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", alpineBucketSlug),
					resource.TestCheckResourceAttr(resourceName, "name", productionChannelSlug),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.id"),
					resource.TestCheckResourceAttr(resourceName, "iteration.0.fingerprint", "channel-update-it1"),
				),
			},
			{
				PreConfig: func() {
					fingerprint := "channel-update-it2"
					upsertIteration(t, alpineBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, alpineBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, alpineBucketSlug, fingerprint, itID)
				},
				Config: testhelpers.TestConfig(channelConfigAssignedLatestIteration(alpineBucketSlug, productionChannelSlug)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", alpineBucketSlug),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.incremental_version"),
					resource.TestCheckResourceAttr(resourceName, "iteration.0.fingerprint", "channel-update-it2"),
					resource.TestCheckResourceAttr(resourceName, "name", productionChannelSlug),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccPackerChannel_UpdateAssignedIterationWithFingerprint(t *testing.T) {
	resourceName := "hcp_packer_channel.production"

	fingerprint := "channel-update-it1"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, alpineBucketSlug, false)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					upsertBucket(t, alpineBucketSlug)
					upsertIteration(t, alpineBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, alpineBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, alpineBucketSlug, fingerprint, itID)
				},
				Config: testhelpers.TestConfig(channelConfigIterationFingerprint(alpineBucketSlug, productionChannelSlug, fingerprint)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", alpineBucketSlug),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.fingerprint"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.incremental_version"),
					resource.TestCheckResourceAttr(resourceName, "name", productionChannelSlug),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

var channelConfigBasic = func(bucketName, channelName string) string {
	return fmt.Sprintf(`
	resource "hcp_packer_channel" "production" {
		bucket_name  = %q
		name = %q
	}`, bucketName, channelName)
}

var channelConfigAssignedLatestIteration = func(bucketName, channelName string) string {
	return fmt.Sprintf(`
	data "hcp_packer_image_iteration" "test" {
		bucket_name = %[2]q
		channel     = "latest"
	}
	resource "hcp_packer_channel" "production" {
		name = %[1]q
		bucket_name  = %[2]q
		iteration {
		  id = data.hcp_packer_image_iteration.test.id
		}
	}`, channelName, bucketName)
}

var channelConfigIterationFingerprint = func(bucketName, channelName, fingerprint string) string {
	return fmt.Sprintf(`
	resource "hcp_packer_channel" "production" {
		bucket_name  = %q
		name = %q
		iteration {
		  fingerprint = %q
		}
	}`, bucketName, channelName, fingerprint)
}
