package atproto

import (
	"testing"
)

func TestResolveATURI(t *testing.T) {
	tests := []struct {
		name           string
		uri            string
		wantDID        string
		wantCollection string
		wantRKey       string
		wantErr        bool
	}{
		{
			name:           "valid plc DID URI",
			uri:            "at://did:plc:abc123/social.arabica.alpha.bean/3jxyabc",
			wantDID:        "did:plc:abc123",
			wantCollection: "social.arabica.alpha.bean",
			wantRKey:       "3jxyabc",
			wantErr:        false,
		},
		{
			name:           "valid web DID URI",
			uri:            "at://did:web:example.com/social.arabica.alpha.brew/xyz789",
			wantDID:        "did:web:example.com",
			wantCollection: "social.arabica.alpha.brew",
			wantRKey:       "xyz789",
			wantErr:        false,
		},
		{
			name:           "long TID rkey",
			uri:            "at://did:plc:longtestdid123/social.arabica.alpha.grinder/3kfk4slgu6s2h",
			wantDID:        "did:plc:longtestdid123",
			wantCollection: "social.arabica.alpha.grinder",
			wantRKey:       "3kfk4slgu6s2h",
			wantErr:        false,
		},
		{
			name:           "bsky app collection",
			uri:            "at://did:plc:user123/app.bsky.feed.post/abc123",
			wantDID:        "did:plc:user123",
			wantCollection: "app.bsky.feed.post",
			wantRKey:       "abc123",
			wantErr:        false,
		},
		{
			name:    "invalid scheme",
			uri:     "http://did:plc:abc123/social.arabica.alpha.bean/3jxyabc",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			uri:     "did:plc:abc123/social.arabica.alpha.bean/3jxyabc",
			wantErr: true,
		},
		{
			name:    "empty URI",
			uri:     "",
			wantErr: true,
		},
		{
			name:           "URI without collection/rkey (valid DID reference)",
			uri:            "at://did:plc:abc123",
			wantDID:        "did:plc:abc123",
			wantCollection: "", // No collection is allowed
			wantRKey:       "",
		},
		{
			name:    "garbage input",
			uri:     "not a valid uri at all",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveATURI(tt.uri)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveATURI(%q) expected error, got nil", tt.uri)
				}
				return
			}

			if err != nil {
				t.Fatalf("ResolveATURI(%q) unexpected error = %v", tt.uri, err)
			}

			if got.DID != tt.wantDID {
				t.Errorf("DID = %q, want %q", got.DID, tt.wantDID)
			}
			if got.Collection != tt.wantCollection {
				t.Errorf("Collection = %q, want %q", got.Collection, tt.wantCollection)
			}
			if got.RKey != tt.wantRKey {
				t.Errorf("RKey = %q, want %q", got.RKey, tt.wantRKey)
			}
		})
	}
}

func TestATURIComponents(t *testing.T) {
	// Test that the struct correctly holds parsed components
	uri := "at://did:plc:testuser/social.arabica.alpha.roaster/roaster123"
	components, err := ResolveATURI(uri)
	if err != nil {
		t.Fatalf("ResolveATURI() error = %v", err)
	}

	// Verify struct fields are accessible
	if components.DID == "" {
		t.Error("DID should not be empty")
	}
	if components.Collection == "" {
		t.Error("Collection should not be empty")
	}
	if components.RKey == "" {
		t.Error("RKey should not be empty")
	}
}

func TestResolveATURI_AllCollections(t *testing.T) {
	// Test with each Arabica collection type
	did := "did:plc:testuser"
	rkey := "abc123"

	collections := []string{
		NSIDBean,
		NSIDBrew,
		NSIDBrewer,
		NSIDGrinder,
		NSIDRoaster,
	}

	for _, collection := range collections {
		t.Run(collection, func(t *testing.T) {
			uri := BuildATURI(did, collection, rkey)
			components, err := ResolveATURI(uri)
			if err != nil {
				t.Fatalf("ResolveATURI() error = %v", err)
			}

			if components.DID != did {
				t.Errorf("DID = %q, want %q", components.DID, did)
			}
			if components.Collection != collection {
				t.Errorf("Collection = %q, want %q", components.Collection, collection)
			}
			if components.RKey != rkey {
				t.Errorf("RKey = %q, want %q", components.RKey, rkey)
			}
		})
	}
}

func TestBuildAndResolveRoundTrip(t *testing.T) {
	// Test that BuildATURI and ResolveATURI are inverses
	tests := []struct {
		did        string
		collection string
		rkey       string
	}{
		{"did:plc:abc123", NSIDBean, "bean123"},
		{"did:plc:xyz789", NSIDBrew, "3kfk4slgu6s2h"},
		{"did:web:example.com", NSIDRoaster, "roaster456"},
		{"did:plc:longdidvalue123456789", NSIDGrinder, "g1"},
	}

	for _, tt := range tests {
		t.Run(tt.did+"/"+tt.collection+"/"+tt.rkey, func(t *testing.T) {
			// Build the URI
			uri := BuildATURI(tt.did, tt.collection, tt.rkey)

			// Resolve it back
			components, err := ResolveATURI(uri)
			if err != nil {
				t.Fatalf("ResolveATURI() error = %v", err)
			}

			// Verify round-trip
			if components.DID != tt.did {
				t.Errorf("DID round-trip: got %q, want %q", components.DID, tt.did)
			}
			if components.Collection != tt.collection {
				t.Errorf("Collection round-trip: got %q, want %q", components.Collection, tt.collection)
			}
			if components.RKey != tt.rkey {
				t.Errorf("RKey round-trip: got %q, want %q", components.RKey, tt.rkey)
			}
		})
	}
}
