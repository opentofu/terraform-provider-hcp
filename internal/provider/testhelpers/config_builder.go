// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testhelpers

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

const maxSlugLength = 36
const minCharsFromTestName = 4

func CreateTestSlug(t *testing.T, midfix ...string) string {
	if len(midfix) > 1 {
		panic(fmt.Errorf("only 1 midfix is allowed, recieved %d", len(midfix)))
	}

	suffix := fmt.Sprintf("-%s", time.Now().Format("0601021504"))

	if len(midfix) == 1 {
		midfix := midfix[0]
		if minCharsFromTestName+len(midfix)+len(suffix) > maxSlugLength {
			panic(fmt.Errorf(
				"length of midfix and suffix too long (midfix length: %d) (suffix length: %d) (total length: %d), expected max total length of %d",
				len(midfix), len(suffix), len(midfix)+len(suffix), maxSlugLength-minCharsFromTestName,
			))
		}
		suffix = fmt.Sprintf("%s-%s", midfix, suffix)
	}

	charsRemaining := maxSlugLength - len(suffix)
	testName := strings.ReplaceAll(t.Name(), "_", "")
	if charsRemaining < len(testName) {
		testName = testName[len(testName)-charsRemaining:]
	}

	return fmt.Sprintf("%s%s", testName, suffix)
}

// TODO: Add support for blocks
type ConfigBuilder interface {
	IsData() bool
	ResourceType() string
	UniqueName() string
	ResourceName() string
	AttributeRef(string) string
	Attributes() map[string]string
}

func BuildTestConfig(builders ...ConfigBuilder) string {
	configs := []string{}
	for _, builder := range builders {
		configs = append(configs, ConfigBuilderToString(builder))
	}
	return TestConfig(configs...)
}

func ConfigBuilderToString(builder ConfigBuilder) string {
	rOrD := "resource"
	if builder.IsData() {
		rOrD = "data"
	}

	attributeStrings := []string{}
	for key, value := range builder.Attributes() {
		if key != "" && value != "" {
			attributeStrings = append(attributeStrings, fmt.Sprintf("    %s = %s", key, value))
		}
	}

	return fmt.Sprintf(`
%s %q %q {
%s
}
`, rOrD, builder.ResourceType(), builder.UniqueName(), strings.Join(attributeStrings, "\n"))
}

type genericConfigBuilder struct {
	isData       bool
	resourceType string
	uniqueName   string
	// Attribute values must be as they would be in the config file.
	// Ex: "value" can be represented in Go with `"value"` or fmt.Sprintf("%q", "value")
	// An empty string is equivalent to the attribute not being present in the map.
	attributes map[string]string
}

var _ ConfigBuilder = genericConfigBuilder{}

func NewResourceConfigBuilder(resourceType string, uniqueName string, attributes map[string]string) ConfigBuilder {
	return &genericConfigBuilder{
		isData:       false,
		resourceType: resourceType,
		uniqueName:   uniqueName,
		attributes:   attributes,
	}
}

func NewDataConfigBuilder(dataType string, uniqueName string, attributes map[string]string) ConfigBuilder {
	return &genericConfigBuilder{
		isData:       true,
		resourceType: dataType,
		uniqueName:   uniqueName,
		attributes:   attributes,
	}
}

func (b genericConfigBuilder) IsData() bool {
	return b.isData
}

func (b genericConfigBuilder) ResourceType() string {
	return b.resourceType
}

func (b genericConfigBuilder) UniqueName() string {
	return b.uniqueName
}

func (b genericConfigBuilder) ResourceName() string {
	if b.isData {
		return fmt.Sprintf("data.%s.%s", b.ResourceType(), b.UniqueName())
	}

	return fmt.Sprintf("%s.%s", b.ResourceType(), b.UniqueName())
}

func (b genericConfigBuilder) Attributes() map[string]string {
	return b.attributes
}

func (b genericConfigBuilder) AttributeRef(path string) string {
	return fmt.Sprintf("%s.%s", b.ResourceName(), path)
}
