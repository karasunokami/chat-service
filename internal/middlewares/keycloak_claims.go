package middlewares

import (
	"errors"

	keycloakclient "github.com/karasunokami/chat-service/internal/clients/keycloak"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang-jwt/jwt"
)

var (
	ErrNoAllowedResources = errors.New("no allowed resources")
	ErrSubjectNotDefined  = errors.New(`"sub" is not defined`)
)

type claims struct {
	jwt.StandardClaims
	Audience        keycloakclient.StringsSliceFromStringOrSlice `json:"aud,omitempty"`
	Subject         types.UserID                                 `json:"sub,omitempty"`
	ResourcesAccess resourceAccess                               `json:"resource_access"`
	// Exp field is copy of claims ExpiresAt int64 field
	// it must be copied after parsing jwt to be accessible from handlers
	// Adding json tag to this field is unavailable because it is
	// overriding default claims exp field used by default jwt exp validation
	Exp int64
}

// Valid returns errors:
// - from StandardClaims validation;
// - ErrNoAllowedResources, if claims doesn't contain `resource_access` map, or it's empty;
// - ErrSubjectNotDefined, if claims doesn't contain `sub` field or subject is zero UUID.
func (c claims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return err
	}

	if len(c.ResourcesAccess) == 0 {
		return ErrNoAllowedResources
	}

	if c.Subject.IsZero() {
		return ErrSubjectNotDefined
	}

	return nil
}

func (c claims) UserID() types.UserID {
	return c.Subject
}

func (c claims) ExpiresAtUnix() int64 {
	return c.Exp
}

type resourceAccess map[string]struct {
	Roles []string `json:"roles"`
}

func (ra resourceAccess) HasResourceRole(resource, role string) bool {
	access, ok := ra[resource]
	if !ok {
		return false
	}

	for _, r := range access.Roles {
		if r == role {
			return true
		}
	}
	return false
}
