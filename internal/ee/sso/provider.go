package sso

import (
	"encoding/json"
)

// SSOProvider defines the interface for SSO authentication providers.
type SSOProvider interface {
	// GenerateAuthURL generates a redirect URL to the Identity Provider.
	GenerateAuthURL(requestID string) (string, error)
	// ValidateResponse validates the SSO response from the IdP and returns an assertion.
	ValidateResponse(responseData string) (*SSOAssertion, error)
	// GenerateMetadataXML generates SP metadata XML.
	GenerateMetadataXML() (string, error)
}

// SSOAssertion represents the claims extracted from a successful SSO authentication.
type SSOAssertion struct {
	NameID     string            `json:"name_id"`
	Email      string            `json:"email"`
	FirstName  string            `json:"first_name,omitempty"`
	LastName   string            `json:"last_name,omitempty"`
	Groups     []string          `json:"groups,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// Name returns the full name from the assertion.
func (a *SSOAssertion) Name() string {
	if a.FirstName != "" && a.LastName != "" {
		return a.FirstName + " " + a.LastName
	}
	if a.FirstName != "" {
		return a.FirstName
	}
	if a.LastName != "" {
		return a.LastName
	}
	return ""
}

// AttributeMapping defines how SAML attributes map to Linkrift user fields.
type AttributeMapping struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Groups    string `json:"groups"`
}

// DefaultAttributeMapping returns the default SAML attribute mapping.
func DefaultAttributeMapping() AttributeMapping {
	return AttributeMapping{
		Email:     "email",
		FirstName: "firstName",
		LastName:  "lastName",
		Groups:    "groups",
	}
}

// ParseAttributeMapping parses attribute mapping from JSON.
func ParseAttributeMapping(data json.RawMessage) AttributeMapping {
	if len(data) == 0 || string(data) == "{}" || string(data) == "null" {
		return DefaultAttributeMapping()
	}
	var mapping AttributeMapping
	if err := json.Unmarshal(data, &mapping); err != nil {
		return DefaultAttributeMapping()
	}
	if mapping.Email == "" {
		mapping.Email = "email"
	}
	return mapping
}
