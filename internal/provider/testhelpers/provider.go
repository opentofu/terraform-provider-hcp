// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testhelpers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	//"github.com/hashicorp/terraform-provider-hcp/internal/provider"
)

var (
	providerState = providerConfigurator{
		envVarsAtConfigTime: make(map[string]string),
		requiredCredVarKeys: map[string][]string{
			"hcp": {
				"HCP_CLIENT_ID",
				"HCP_CLIENT_SECRET",
			},
			"aws": {
				"AWS_ACCESS_KEY_ID",
				"AWS_SECRET_ACCESS_KEY",
				"AWS_SESSION_TOKEN",
			},
			"azure": {
				"ARM_TENANT_ID",
				"ARM_SUBSCRIPTION_ID",
			},
		},
	}

	providerFactories = map[string]func() (*schema.Provider, error){
		"dummy": func() (*schema.Provider, error) {
			return newDummyProvider(), nil
		},
	}
)

// Must be called once in the provider package's `init()`.
// Injection is being used to avoid a circular dependency
func InitializeHCPProviderFactory(newHCPProvider func() *schema.Provider) {
	if _, ok := providerFactories["hcp"]; ok {
		panic(fmt.Errorf("hcp provider factory was already initialized"))
	}
	providerFactories["hcp"] = func() (*schema.Provider, error) {
		return newHCPProvider(), nil
	}
}

// DefaultProvider is "main" provider instance during testing. It's
// configuration contains the default values provided via environment variables.
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// PreCheck() must be called before using this provider instance.
func DefaultProvider() *schema.Provider {
	if providerState.provider == nil {
		panic("provider not configured before calling")
	}
	return providerState.provider
}

// Provider Factories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
func ProviderFactories(requestedProviders ...string) map[string]func() (*schema.Provider, error) {
	outFactories := map[string]func() (*schema.Provider, error){}

	// If no arguments passed, return all providers
	if len(requestedProviders) == 0 {
		// Copy to new map so the main map can't be modified
		for k, v := range providerFactories {
			outFactories[k] = v
		}
		return outFactories
	}

	invalidProviders := []string{}
	for _, requestedProvider := range requestedProviders {
		factory, ok := providerFactories[requestedProvider]
		if !ok {
			invalidProviders = append(invalidProviders, requestedProvider)
			continue
		}
		outFactories[requestedProvider] = factory
	}

	if len(invalidProviders) > 0 {
		validProviders := []string{}
		for provider := range providerFactories {
			validProviders = append(validProviders, provider)
		}
		panic(fmt.Errorf("requested provider factories that are not defined: %v. Valid options are: %v", invalidProviders, validProviders))
	}

	return outFactories
}

// PreCheck verifies and sets required provider testing configuration
//
// This PreCheck function should be present in every acceptance test. It ensures
// testing functions that attempt to call HCP APIs are previously configured.
//
// These verifications and configuration are preferred at this level to prevent
// provider developers from experiencing less clear errors for every test.
func PreCheck(t *testing.T, requiredCreds map[string]bool) {
	t.Helper()
	providerState.configure()

	if _, ok := requiredCreds["hcp"]; ok {
		t.Logf("credental type \"hcp\" should not be present in requiredCreds, as it is always required")
	}

	requiredCreds["hcp"] = true
	if err := providerState.checkCredVarsSet(requiredCreds); err != nil {
		t.Fatal(err)
	}
}

func TestConfig(res ...string) string {
	provider := `provider "hcp" {}`

	c := []string{provider}
	c = append(c, res...)
	return strings.Join(c, "\n")
}

type providerConfigurator struct {
	// The PreCheck(t) function is invoked for every test and this prevents
	// extraneous reconfiguration to the same values each time. However, this does
	// not prevent reconfiguration that may happen should the address of
	// testhelpers.DefaultProvider() be errantly reused in ProviderFactories.
	configureOnce sync.Once

	provider *schema.Provider

	// Contains the values of all environment variables defined in
	// `requiredCredVarKeys` at the time `configure()` is called
	envVarsAtConfigTime map[string]string

	// A map from types of required credentials to the list of environment
	// variable keys that are needed for that type of credential
	requiredCredVarKeys map[string][]string
}

func (c *providerConfigurator) configure() {
	c.configureOnce.Do(func() {
		for _, varKeys := range c.requiredCredVarKeys {
			for _, varKey := range varKeys {
				c.envVarsAtConfigTime[varKey] = os.Getenv(varKey)
			}
		}

		provider, err := ProviderFactories("hcp")["hcp"]()
		if err != nil {
			panic(fmt.Errorf("failed to create provider, recieved error: %v", err))
		}
		c.provider = provider
		if diags := c.provider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil)); diags.HasError() {
			panic(fmt.Errorf("failed to configure provider, recieved error diagnostic: %#v", diags))
		}
	})
}

func (c *providerConfigurator) checkCredVarsSet(requiredCreds map[string]bool) error {
	var errors *multierror.Error
	for credType, credRequired := range requiredCreds {
		credTypeVarKeys, ok := c.requiredCredVarKeys[credType]
		if !ok {
			errors = multierror.Append(errors, fmt.Errorf("credential type %q is not valid", credType))
			continue
		}

		if credRequired {
			for _, key := range credTypeVarKeys {
				if val := c.envVarsAtConfigTime[key]; val == "" {
					errors = multierror.Append(errors,
						fmt.Errorf("credential type %q requires environment variable %q to be a non-empty string, but it was empty", credType, key),
					)
				}
			}
		}
	}

	if len(errors.Errors) > 0 {
		return errors
	}

	return nil
}

// Used to create a dummy non-empty state so that `CheckDestroy` can be used to
// clean up resources created in `PreCheck` for tests that don't generate a
// non-empty state on their own.
func DummyNonemptyStateConfig() string {
	return `resource "dummy_state" "dummy" {}`
}

// Contains a dummy resource that is used in `DummyNonemptyStateConfig`
func newDummyProvider() *schema.Provider {
	setDummyID := func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
		d.SetId("dummy_id")
		return nil
	}
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"dummy_state": {
				CreateContext: setDummyID,
				ReadContext:   setDummyID,
				DeleteContext: func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics { return nil },
			},
		},
	}
}
