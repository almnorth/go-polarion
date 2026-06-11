// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Polarion Client Contributors

package polarion

import "testing"

func TestFieldsAllIncludesExternalLinks(t *testing.T) {
	t.Parallel()

	if FieldsAll.ExternallyLinkedWorkItems != "@all" {
		t.Fatalf("FieldsAll.ExternallyLinkedWorkItems = %q, want %q", FieldsAll.ExternallyLinkedWorkItems, "@all")
	}
}

func TestFieldsDefaultIncludesExternalLinks(t *testing.T) {
	t.Parallel()

	if FieldsDefault.ExternallyLinkedWorkItems != "id,role,workItemURI" {
		t.Fatalf("FieldsDefault.ExternallyLinkedWorkItems = %q, want %q", FieldsDefault.ExternallyLinkedWorkItems, "id,role,workItemURI")
	}
}
