package atproto

import "fmt"

// NSID (Namespaced Identifier) constants for Arabica lexicons.
// The domain is reversed following ATProto conventions: arabica.social -> social.arabica
// Using "alpha" namespace during development - will migrate to stable namespace later.
const (
	// NSIDBase is the base namespace for all Arabica lexicons
	NSIDBase = "social.arabica.alpha"

	// Collection NSIDs
	NSIDBean    = NSIDBase + ".bean"
	NSIDBrew    = NSIDBase + ".brew"
	NSIDBrewer  = NSIDBase + ".brewer"
	NSIDGrinder = NSIDBase + ".grinder"
	NSIDRoaster = NSIDBase + ".roaster"
)

// BuildATURI constructs an AT-URI from a DID, collection NSID, and record key
func BuildATURI(did, collection, rkey string) string {
	return fmt.Sprintf("at://%s/%s/%s", did, collection, rkey)
}
