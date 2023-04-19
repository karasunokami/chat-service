package keycloakclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type IntrospectTokenResult struct {
	Exp    int      `json:"exp"`
	Iat    int      `json:"iat"`
	Aud    []string `json:"aud"`
	Active bool     `json:"active"`
}

func (r *IntrospectTokenResult) UnmarshalJSON(data []byte) error {
	tr := struct {
		Exp    int             `json:"exp"`
		Iat    int             `json:"iat"`
		Aud    json.RawMessage `json:"aud"`
		Active bool            `json:"active"`
	}{}

	err := json.Unmarshal(data, &tr)
	if err != nil {
		return fmt.Errorf("unmarshal data to tmp introspect token result, err=%v", err)
	}

	r.Exp = tr.Exp
	r.Iat = tr.Iat
	r.Active = tr.Active

	r.Aud, err = unmarshalStringOrStringSliceToSlice(tr.Aud)
	if err != nil {
		return fmt.Errorf("unamrshal string or string slice to slice, err=%w", err)
	}

	return nil
}

// IntrospectToken implements
// https://www.keycloak.org/docs/latest/authorization_services/index.html#obtaining-information-about-an-rpt
func (c *Client) IntrospectToken(ctx context.Context, token string) (*IntrospectTokenResult, error) {
	url := fmt.Sprintf("realms/%s/protocol/openid-connect/token/introspect", c.realm)

	var result IntrospectTokenResult

	resp, err := c.auth(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"token_type_hint": "requesting_party_token",
			"token":           token,
		}).
		SetResult(&result).
		Post(url)
	if err != nil {
		return nil, fmt.Errorf("send request to keycloak: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("errored keycloak response: %v", resp.Status())
	}

	return &result, nil
}

func (c *Client) auth(ctx context.Context) *resty.Request {
	c.cli.SetBasicAuth(c.clientID, c.clientSecret)

	return c.cli.R().SetContext(ctx)
}

func unmarshalStringOrStringSliceToSlice(data []byte) ([]string, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var firstChar = string(data[0])

	if firstChar == "\"" {
		var str string

		err := json.Unmarshal(data, &str)
		if err != nil {
			return nil, fmt.Errorf("unmarshal aud to string, err=%v", err)
		}

		return []string{str}, nil
	}

	if firstChar == "[" {
		var list []string

		err := json.Unmarshal(data, &list)
		if err != nil {
			return nil, fmt.Errorf("unmarshal aud to []string, err=%v", err)
		}

		return list, nil
	}

	return nil, fmt.Errorf("unsupported data type, data=%s", data)
}
