package atproto

import (
	"testing"
)

func TestNSIDConstants(t *testing.T) {
	// Verify base namespace
	if NSIDBase != "social.arabica.alpha" {
		t.Errorf("NSIDBase = %q, want %q", NSIDBase, "social.arabica.alpha")
	}

	// Verify all collection NSIDs are properly prefixed
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"NSIDBean", NSIDBean, "social.arabica.alpha.bean"},
		{"NSIDBrew", NSIDBrew, "social.arabica.alpha.brew"},
		{"NSIDBrewer", NSIDBrewer, "social.arabica.alpha.brewer"},
		{"NSIDGrinder", NSIDGrinder, "social.arabica.alpha.grinder"},
		{"NSIDRoaster", NSIDRoaster, "social.arabica.alpha.roaster"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestBuildATURI(t *testing.T) {
	tests := []struct {
		name       string
		did        string
		collection string
		rkey       string
		expected   string
	}{
		{
			name:       "basic URI",
			did:        "did:plc:abc123",
			collection: "social.arabica.alpha.bean",
			rkey:       "3jxyabc",
			expected:   "at://did:plc:abc123/social.arabica.alpha.bean/3jxyabc",
		},
		{
			name:       "web DID",
			did:        "did:web:example.com",
			collection: "social.arabica.alpha.brew",
			rkey:       "xyz789",
			expected:   "at://did:web:example.com/social.arabica.alpha.brew/xyz789",
		},
		{
			name:       "with grinder collection",
			did:        "did:plc:test456",
			collection: NSIDGrinder,
			rkey:       "rkey123",
			expected:   "at://did:plc:test456/social.arabica.alpha.grinder/rkey123",
		},
		{
			name:       "empty rkey",
			did:        "did:plc:abc",
			collection: NSIDBean,
			rkey:       "",
			expected:   "at://did:plc:abc/social.arabica.alpha.bean/",
		},
		{
			name:       "long TID rkey",
			did:        "did:plc:abc123xyz789",
			collection: NSIDBrew,
			rkey:       "3kfk4slgu6s2h",
			expected:   "at://did:plc:abc123xyz789/social.arabica.alpha.brew/3kfk4slgu6s2h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildATURI(tt.did, tt.collection, tt.rkey)
			if got != tt.expected {
				t.Errorf("BuildATURI(%q, %q, %q) = %q, want %q",
					tt.did, tt.collection, tt.rkey, got, tt.expected)
			}
		})
	}
}

func TestBuildATURI_WithNSIDConstants(t *testing.T) {
	did := "did:plc:testuser"
	rkey := "abc123"

	// Test with each NSID constant
	tests := []struct {
		collection string
		expected   string
	}{
		{NSIDBean, "at://did:plc:testuser/social.arabica.alpha.bean/abc123"},
		{NSIDBrew, "at://did:plc:testuser/social.arabica.alpha.brew/abc123"},
		{NSIDBrewer, "at://did:plc:testuser/social.arabica.alpha.brewer/abc123"},
		{NSIDGrinder, "at://did:plc:testuser/social.arabica.alpha.grinder/abc123"},
		{NSIDRoaster, "at://did:plc:testuser/social.arabica.alpha.roaster/abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.collection, func(t *testing.T) {
			got := BuildATURI(did, tt.collection, rkey)
			if got != tt.expected {
				t.Errorf("BuildATURI with %s = %q, want %q", tt.collection, got, tt.expected)
			}
		})
	}
}
