// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// Code generated by beats/dev-tools/cmd/asset/asset.go - DO NOT EDIT.

package aws

import (
	"github.com/elastic/beats/v7/libbeat/asset"
)

func init() {
	if err := asset.SetFields("metricbeat", "aws", asset.ModuleFieldsPri, AssetAws); err != nil {
		panic(err)
	}
}

// AssetAws returns asset data.
// This is the base64 encoded gzipped contents of module/aws.
func AssetAws() string {
	return "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"
}
