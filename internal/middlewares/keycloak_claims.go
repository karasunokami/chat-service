package middlewares

import (
	"errors"
	"fmt"

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

	Subject        types.UserID                                 `json:"sub,omitempty"`
	ResourceAccess map[string]map[string][]string               `json:"resource_access,omitempty"`
	Audience       keycloakclient.StringsSliceFromStringOrSlice `json:"aud,omitempty"`
}

// Valid returns errors:
// - from StandardClaims validation;
// - ErrNoAllowedResources, if claims doesn't contain `resource_access` map or it's empty;
// - ErrSubjectNotDefined, if claims doesn't contain `sub` field or subject is zero UUID.
func (c claims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return fmt.Errorf("start claims validation, err=%w", err)
	}

	if c.ResourceAccess == nil || len(c.ResourceAccess) == 0 {
		return ErrNoAllowedResources
	}

	if c.Subject.String() == types.UserIDNil.String() {
		return ErrSubjectNotDefined
	}

	return nil
}

func (c claims) UserID() types.UserID {
	return c.Subject
}

func (c claims) hasRoleForResource(role string, resource string) bool {
	for res, data := range c.ResourceAccess {
		if res == resource {
			for key, list := range data {
				if key == "roles" {
					for _, rl := range list {
						if rl == role {
							return true
						}
					}
				}
			}
		}
	}

	return false
}
