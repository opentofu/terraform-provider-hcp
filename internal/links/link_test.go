// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package links_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/links"
	"github.com/stretchr/testify/require"
)

func Test_linkURL(t *testing.T) {
	baseLink := &sharedmodels.HashicorpCloudLocationLink{
		Type: "hashicorp.network.hvn",
		ID:   "test-hvn",
		Location: &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: uuid.New().String(),
			ProjectID:      uuid.New().String(),
		},
	}

	t.Run("valid ID", func(t *testing.T) {
		l := *baseLink

		urn, err := links.LinkURL(&l)
		require.NoError(t, err)

		expected := fmt.Sprintf("/project/%s/%s/%s",
			l.Location.ProjectID,
			l.Type,
			l.ID)
		require.Equal(t, expected, urn)
	})

	t.Run("missing organization ID", func(t *testing.T) {
		l := *baseLink
		l.Location.OrganizationID = ""

		_, err := links.LinkURL(&l)
		require.NoError(t, err)
	})

	t.Run("missing project ID", func(t *testing.T) {
		l := *baseLink
		l.Location.ProjectID = ""

		_, err := links.LinkURL(&l)
		require.Error(t, err)
	})

	t.Run("missing resource type", func(t *testing.T) {
		l := *baseLink
		l.Type = ""

		_, err := links.LinkURL(&l)
		require.Error(t, err)
	})

	t.Run("missing resource ID", func(t *testing.T) {
		l := *baseLink
		l.ID = ""

		_, err := links.LinkURL(&l)
		require.Error(t, err)
	})

	t.Run("missing Location", func(t *testing.T) {
		l := *baseLink
		l.Location = nil

		_, err := links.LinkURL(&l)
		require.Error(t, err)
	})
}

func Test_parseLinkURL(t *testing.T) {
	svcType := "hashicorp.network.hvn"
	id := "test-hvn"
	projID := uuid.New().String()

	t.Run("valid URL", func(t *testing.T) {
		urn := fmt.Sprintf("/project/%s/%s/%s",
			projID,
			svcType,
			id)

		l, err := links.ParseLinkURL(urn, svcType)
		require.NoError(t, err)

		require.Equal(t, projID, l.Location.ProjectID)
		require.Equal(t, svcType, l.Type)
		require.Equal(t, id, l.ID)
	})

	t.Run("extracting type from the URL", func(t *testing.T) {
		urn := fmt.Sprintf("/project/%s/%s/%s",
			projID,
			svcType,
			id)

		l, err := links.ParseLinkURL(urn, "")
		require.NoError(t, err)

		require.Equal(t, projID, l.Location.ProjectID)
		require.Equal(t, svcType, l.Type)
		require.Equal(t, id, l.ID)
	})

	t.Run("missing project ID", func(t *testing.T) {
		urn := fmt.Sprintf("/project/%s/%s/%s",
			"",
			svcType,
			id)

		_, err := links.ParseLinkURL(urn, svcType)
		require.Error(t, err)
	})

	t.Run("missing resource type", func(t *testing.T) {
		urn := fmt.Sprintf("/project/%s/%s/%s",
			projID,
			"",
			id)

		_, err := links.ParseLinkURL(urn, svcType)
		require.Error(t, err)
	})

	t.Run("mismatched resource type", func(t *testing.T) {
		urn := fmt.Sprintf("/project/%s/%s/%s",
			projID,
			"other.hvn",
			id)

		_, err := links.ParseLinkURL(urn, svcType)
		require.Error(t, err)
	})

	t.Run("missing resource id", func(t *testing.T) {
		urn := fmt.Sprintf("/project/%s/%s/%s",
			projID,
			svcType,
			"")

		_, err := links.ParseLinkURL(urn, svcType)
		require.Error(t, err)
	})

	t.Run("missing a field", func(t *testing.T) {
		urn := fmt.Sprintf("/%s/%s",
			svcType,
			id)

		_, err := links.ParseLinkURL(urn, svcType)
		require.Error(t, err)
	})

	t.Run("too many fields before", func(t *testing.T) {
		urn := fmt.Sprintf("/extra/value/project/%s/%s/%s",
			projID,
			svcType,
			id)

		_, err := links.ParseLinkURL(urn, svcType)
		require.Error(t, err)
	})

	t.Run("too many fields after", func(t *testing.T) {
		urn := fmt.Sprintf("/project/%s/%s/%s/extra/value",
			projID,
			svcType,
			id)

		_, err := links.ParseLinkURL(urn, svcType)
		require.Error(t, err)
	})
}
