// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package collector

import "errors"

type config struct {
	UseTypes     bool `config:"use_types" yaml:"use_types,omitempty"`
	RateCounters bool `config:"rate_counters" yaml:"rate_counters,omitempty"`
}

func (c *config) Validate() error {
	if c.RateCounters && !c.UseTypes {
		return errors.New("'rate_counters' can only be enabled when `use_types` is also enabled")
	}

	return nil
}
