// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import "testing"

func TestIfconfigIPFinder_Impl(t *testing.T) {
	var raw interface{}
	raw = &IfconfigIPFinder{}
	if _, ok := raw.(HostIPFinder); !ok {
		t.Fatalf("IfconfigIPFinder is not a host IP finder")
	}
}
