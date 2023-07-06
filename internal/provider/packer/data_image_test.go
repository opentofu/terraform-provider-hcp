// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packer_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/testhelpers"
)

var (
	imageBucketSlug       = fmt.Sprintf("alpine-acc-imagetest-%s", time.Now().Format("200601021504"))
	ubuntuImageBucketSlug = fmt.Sprintf("ubuntu-acc-imagetest-%s", time.Now().Format("200601021504"))
	archImageBucketSlug   = fmt.Sprintf("arch-acc-imagetest-%s", time.Now().Format("200601021504"))
	imageChannelSlug      = "production-image-test"
	componentType         = "amazon-ebs.example"
)

var (
	imageConfigAlpineProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "alpine-imagetest" {
		bucket_name  = %q
		channel = %q
	}

	data "hcp_packer_image" "foo" {
		bucket_name    = %q
		cloud_provider = "aws"
		iteration_id   = data.hcp_packer_iteration.alpine-imagetest.id
		region         = "us-east-1"
		component_type = %q
	}

	# we make sure that this won't fail even when revoke_at is not set
	output "revoke_at" {
  		value = data.hcp_packer_iteration.alpine-imagetest.revoke_at
	}
`, imageBucketSlug, imageChannelSlug, imageBucketSlug, componentType)

	imageConfigAlpineProductionError = fmt.Sprintf(`
	data "hcp_packer_iteration" "alpine-imagetest" {
		bucket_name  = %q
		channel = %q
	}

	data "hcp_packer_image" "foo" {
		bucket_name    = %q
		cloud_provider = "aws"
		iteration_id   = data.hcp_packer_iteration.alpine-imagetest.id
		region         = "us-east-1"
		component_type = "amazon-ebs.do-not-exist"
	}

	# we make sure that this won't fail even when revoke_at is not set
	output "revoke_at" {
  		value = data.hcp_packer_iteration.alpine-imagetest.revoke_at
	}
`, imageBucketSlug, imageChannelSlug, imageBucketSlug)

	imageConfigUbuntuProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "ubuntu-imagetest" {
		bucket_name  = %q
		channel = %q
	}

	data "hcp_packer_image" "ubuntu-foo" {
		bucket_name    = %q
		cloud_provider = "aws"
		iteration_id   = data.hcp_packer_iteration.ubuntu-imagetest.id
		region         = "us-east-1"
	}
`, ubuntuImageBucketSlug, imageChannelSlug, ubuntuImageBucketSlug)

	imageConfigBothChanAndIter = fmt.Sprintf(`
	data "hcp_packer_image" "arch-btw" {
		bucket_name = %q
		cloud_provider = "aws"
		iteration_id = "234567"
		channel = "chanSlug"
		region = "us-east-1"
	}
`, archImageBucketSlug)

	imageConfigBothChanAndIterRef = fmt.Sprintf(`
	data "hcp_packer_iteration" "arch-imagetest" {
		bucket_name = %q
		channel = %q
	}

	data "hcp_packer_image" "arch-btw" {
		bucket_name = %q
		cloud_provider = "aws"
		iteration_id = data.hcp_packer_iteration.arch-imagetest.id
		channel = %q
		region = "us-east-1"
	}
`, archImageBucketSlug, imageChannelSlug, archImageBucketSlug, imageChannelSlug)

	imageConfigArchProduction = fmt.Sprintf(`
	data "hcp_packer_image" "arch-btw" {
		bucket_name = %q
		cloud_provider = "aws"
		channel = %q
		region = "us-east-1"
	}
`, archImageBucketSlug, imageChannelSlug)
)

func TestAcc_dataSourcePackerImage(t *testing.T) {
	resourceName := "data.hcp_packer_image.foo"
	fingerprint := "44"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, imageBucketSlug, imageChannelSlug, false)
			deleteIteration(t, imageBucketSlug, fingerprint, false)
			deleteBucket(t, imageBucketSlug, false)
			return nil
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			// Testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					upsertBucket(t, imageBucketSlug)
					upsertIteration(t, imageBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, imageBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, imageBucketSlug, fingerprint, itID)
					upsertChannel(t, imageBucketSlug, imageChannelSlug, itID)
				},
				Config: imageConfigAlpineProduction,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr("data.hcp_packer_image.foo", "labels.test-key", "test-value"),
				),
			},
			// Testing that filtering non-existent image fails properly
			{
				PlanOnly:    true,
				Config:      imageConfigAlpineProductionError,
				ExpectError: regexp.MustCompile("Error: Unable to load image"),
			},
		},
	})
}

func TestAcc_dataSourcePackerImage_revokedIteration(t *testing.T) {
	fingerprint := fmt.Sprintf("%d", rand.Int())
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(5 * time.Minute))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, ubuntuImageBucketSlug, imageChannelSlug, true)
			deleteIteration(t, ubuntuImageBucketSlug, fingerprint, true)
			deleteBucket(t, ubuntuImageBucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, ubuntuImageBucketSlug)
					upsertIteration(t, ubuntuImageBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, ubuntuImageBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, ubuntuImageBucketSlug, fingerprint, itID)
					upsertChannel(t, ubuntuImageBucketSlug, imageChannelSlug, itID)
					// Schedule revocation to the future, otherwise we won't be able to revoke an iteration that
					// it's assigned to a channel
					revokeIteration(t, itID, ubuntuImageBucketSlug, revokeAt)
				},
				Config: testhelpers.TestConfig(imageConfigUbuntuProduction),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hcp_packer_image.ubuntu-foo", "revoke_at", revokeAt.String()),
					resource.TestCheckResourceAttr("data.hcp_packer_image.ubuntu-foo", "cloud_image_id", "ami-42"),
				),
			},
		},
	})
}

func TestAcc_dataSourcePackerImage_emptyChannel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, archImageBucketSlug, imageChannelSlug, true)
			deleteBucket(t, archImageBucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, archImageBucketSlug)
					upsertChannel(t, archImageBucketSlug, imageChannelSlug, "")
				},
				Config:      testhelpers.TestConfig(imageConfigArchProduction),
				ExpectError: regexp.MustCompile(`.*Channel does not have an assigned iteration.*`),
			},
		},
	})
}

func TestAcc_dataSourcePackerImage_channelAndIterationIDReject(t *testing.T) {
	fingerprint := "rejectIterationAndChannel"
	configs := []string{
		imageConfigBothChanAndIter,
		imageConfigBothChanAndIterRef,
	}

	for _, cfg := range configs {
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
			ProviderFactories: testhelpers.ProviderFactories(),
			Steps: []resource.TestStep{
				// basically just testing that we don't pass validation here
				{
					PlanOnly: true,
					PreConfig: func() {
						deleteChannel(t, archImageBucketSlug, imageChannelSlug, false)
						deleteIteration(t, archImageBucketSlug, fingerprint, false)
						deleteBucket(t, archImageBucketSlug, false)

						upsertRegistry(t)
						upsertBucket(t, archImageBucketSlug)
						upsertIteration(t, archImageBucketSlug, fingerprint)
						itID, err := getIterationIDFromFingerPrint(t, archImageBucketSlug, fingerprint)
						if err != nil {
							t.Fatal(err.Error())
						}
						upsertBuild(t, archImageBucketSlug, fingerprint, itID)
						upsertChannel(t, archImageBucketSlug, imageChannelSlug, itID)
					},
					Config:      testhelpers.TestConfig(cfg),
					ExpectError: regexp.MustCompile("Error: Invalid combination of arguments"),
				},
			},
		})
	}
}

func TestAcc_dataSourcePackerImage_channelAccept(t *testing.T) {
	fingerprint := "acceptChannel"
	resourceName := "data.hcp_packer_image.arch-btw"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testhelpers.PreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: testhelpers.ProviderFactories(),
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, archImageBucketSlug, imageChannelSlug, false)
			deleteIteration(t, archImageBucketSlug, fingerprint, false)
			deleteBucket(t, archImageBucketSlug, false)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, archImageBucketSlug)
					upsertIteration(t, archImageBucketSlug, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, archImageBucketSlug, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, archImageBucketSlug, fingerprint, itID)
					upsertChannel(t, archImageBucketSlug, imageChannelSlug, itID)
				},
				Config: testhelpers.TestConfig(imageConfigArchProduction),
				Check: resource.ComposeTestCheckFunc(
					// build_id is only known at runtime
					// and the test works on a reset value,
					// therefore we can only check it's set
					resource.TestCheckResourceAttrSet(resourceName, "build_id"),
				),
			},
		},
	})
}
