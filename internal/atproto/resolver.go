package atproto

import (
	"context"
	"fmt"

	"arabica/internal/models"

	"github.com/bluesky-social/indigo/atproto/syntax"
)

// ATURIComponents holds the parsed components of an AT-URI
type ATURIComponents struct {
	DID        string
	Collection string
	RKey       string
}

// ResolveATURI parses an AT-URI and returns its components
// AT-URI format: at://did:plc:abc123/social.arabica.brew/3jxyabc
func ResolveATURI(uri string) (*ATURIComponents, error) {
	atURI, err := syntax.ParseATURI(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid AT-URI: %w", err)
	}

	return &ATURIComponents{
		DID:        atURI.Authority().String(),
		Collection: atURI.Collection().String(),
		RKey:       atURI.RecordKey().String(),
	}, nil
}

// resolveRef is a generic helper that fetches and converts a record from an AT-URI
func resolveRef[T any](
	ctx context.Context,
	client *Client,
	atURI string,
	sessionID string,
	expectedCollection string,
	convert func(map[string]interface{}, string) (*T, error),
) (*T, error) {
	if atURI == "" {
		return nil, nil
	}

	components, err := ResolveATURI(atURI)
	if err != nil {
		return nil, err
	}

	if components.Collection != expectedCollection {
		return nil, fmt.Errorf("expected %s collection, got %s", expectedCollection, components.Collection)
	}

	didObj, err := syntax.ParseDID(components.DID)
	if err != nil {
		return nil, fmt.Errorf("invalid DID: %w", err)
	}

	output, err := client.GetRecord(ctx, didObj, sessionID, &GetRecordInput{
		Collection: components.Collection,
		RKey:       components.RKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s record: %w", expectedCollection, err)
	}

	result, err := convert(output.Value, atURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %s record: %w", expectedCollection, err)
	}

	return result, nil
}

// ResolveBeanRef fetches a bean record from an AT-URI
func ResolveBeanRef(ctx context.Context, client *Client, atURI string, sessionID string) (*models.Bean, error) {
	return resolveRef(ctx, client, atURI, sessionID, NSIDBean, RecordToBean)
}

// ResolveRoasterRef fetches a roaster record from an AT-URI
func ResolveRoasterRef(ctx context.Context, client *Client, atURI string, sessionID string) (*models.Roaster, error) {
	return resolveRef(ctx, client, atURI, sessionID, NSIDRoaster, RecordToRoaster)
}

// ResolveGrinderRef fetches a grinder record from an AT-URI
func ResolveGrinderRef(ctx context.Context, client *Client, atURI string, sessionID string) (*models.Grinder, error) {
	return resolveRef(ctx, client, atURI, sessionID, NSIDGrinder, RecordToGrinder)
}

// ResolveBrewerRef fetches a brewer record from an AT-URI
func ResolveBrewerRef(ctx context.Context, client *Client, atURI string, sessionID string) (*models.Brewer, error) {
	return resolveRef(ctx, client, atURI, sessionID, NSIDBrewer, RecordToBrewer)
}

// ResolveBrewRefs resolves all references within a brew record
// This is a convenience function that resolves bean, grinder, and brewer refs in one call
func ResolveBrewRefs(ctx context.Context, client *Client, brew *models.Brew, beanRef, grinderRef, brewerRef, sessionID string) error {
	var err error

	// Resolve bean reference (required)
	if beanRef != "" {
		brew.Bean, err = ResolveBeanRef(ctx, client, beanRef, sessionID)
		if err != nil {
			return fmt.Errorf("failed to resolve bean reference: %w", err)
		}

		// If bean has a roaster reference, resolve it too
		if brew.Bean != nil && brew.Bean.RoasterRKey != "" {
			// Note: We need to get the roasterRef from the bean record
			// This requires storing the raw record data or fetching it again
			// For now, we'll skip nested resolution and handle it in store.go
		}
	}

	// Resolve grinder reference (optional)
	if grinderRef != "" {
		brew.GrinderObj, err = ResolveGrinderRef(ctx, client, grinderRef, sessionID)
		if err != nil {
			return fmt.Errorf("failed to resolve grinder reference: %w", err)
		}
	}

	// Resolve brewer reference (optional)
	if brewerRef != "" {
		brew.BrewerObj, err = ResolveBrewerRef(ctx, client, brewerRef, sessionID)
		if err != nil {
			return fmt.Errorf("failed to resolve brewer reference: %w", err)
		}
	}

	return nil
}
