package scim

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SCIM 2.0 Schema URIs
const (
	SchemaUser                = "urn:ietf:params:scim:schemas:core:2.0:User"
	SchemaGroup               = "urn:ietf:params:scim:schemas:core:2.0:Group"
	SchemaListResponse        = "urn:ietf:params:scim:api:messages:2.0:ListResponse"
	SchemaPatchOp             = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
	SchemaServiceProviderConfig = "urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"
	SchemaResourceType        = "urn:ietf:params:scim:schemas:core:2.0:ResourceType"
	SchemaSchema              = "urn:ietf:params:scim:schemas:core:2.0:Schema"
	SchemaError               = "urn:ietf:params:scim:api:messages:2.0:Error"
)

// SCIMUser represents a SCIM 2.0 User resource.
type SCIMUser struct {
	Schemas    []string    `json:"schemas"`
	ID         string      `json:"id"`
	ExternalID string      `json:"externalId,omitempty"`
	UserName   string      `json:"userName"`
	Name       *SCIMName   `json:"name,omitempty"`
	Emails     []SCIMEmail `json:"emails,omitempty"`
	Active     bool        `json:"active"`
	Meta       *SCIMMeta   `json:"meta,omitempty"`
}

// SCIMName represents the name component of a SCIM user.
type SCIMName struct {
	Formatted  string `json:"formatted,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	GivenName  string `json:"givenName,omitempty"`
}

// SCIMEmail represents an email in a SCIM user.
type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMGroup represents a SCIM 2.0 Group resource.
type SCIMGroup struct {
	Schemas     []string       `json:"schemas"`
	ID          string         `json:"id"`
	DisplayName string         `json:"displayName"`
	Members     []SCIMMember   `json:"members,omitempty"`
	Meta        *SCIMMeta      `json:"meta,omitempty"`
}

// SCIMMember represents a member reference in a SCIM group.
type SCIMMember struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
	Ref     string `json:"$ref,omitempty"`
}

// SCIMMeta holds resource metadata.
type SCIMMeta struct {
	ResourceType string `json:"resourceType"`
	Created      string `json:"created,omitempty"`
	LastModified string `json:"lastModified,omitempty"`
	Location     string `json:"location,omitempty"`
}

// SCIMListResponse is the standard SCIM 2.0 list response.
type SCIMListResponse struct {
	Schemas      []string        `json:"schemas"`
	TotalResults int             `json:"totalResults"`
	ItemsPerPage int             `json:"itemsPerPage"`
	StartIndex   int             `json:"startIndex"`
	Resources    json.RawMessage `json:"Resources"`
}

// SCIMPatchOp represents a SCIM PATCH operation.
type SCIMPatchOp struct {
	Schemas    []string          `json:"schemas"`
	Operations []SCIMPatchOperation `json:"Operations"`
}

// SCIMPatchOperation is a single operation within a PATCH request.
type SCIMPatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// SCIMError represents a SCIM 2.0 error response.
type SCIMError struct {
	Schemas  []string `json:"schemas"`
	Detail   string   `json:"detail"`
	Status   string   `json:"status"`
	ScimType string   `json:"scimType,omitempty"`
}

// NewSCIMError creates a SCIM error response.
func NewSCIMError(status int, detail string) SCIMError {
	return SCIMError{
		Schemas: []string{SchemaError},
		Detail:  detail,
		Status:  fmt.Sprintf("%d", status),
	}
}

// ServiceProviderConfig returns the SCIM service provider configuration.
func ServiceProviderConfig(baseURL string) map[string]interface{} {
	return map[string]interface{}{
		"schemas": []string{SchemaServiceProviderConfig},
		"documentationUri": "https://linkrift.com/docs/scim",
		"patch": map[string]bool{
			"supported": true,
		},
		"bulk": map[string]interface{}{
			"supported":      false,
			"maxOperations":  0,
			"maxPayloadSize": 0,
		},
		"filter": map[string]interface{}{
			"supported":  true,
			"maxResults": 100,
		},
		"changePassword": map[string]bool{
			"supported": false,
		},
		"sort": map[string]bool{
			"supported": false,
		},
		"etag": map[string]bool{
			"supported": false,
		},
		"authenticationSchemes": []map[string]string{
			{
				"type":        "oauthbearertoken",
				"name":        "OAuth Bearer Token",
				"description": "Authentication scheme using SCIM bearer tokens",
			},
		},
		"meta": map[string]string{
			"resourceType": "ServiceProviderConfig",
			"location":     baseURL + "/ServiceProviderConfig",
		},
	}
}

// ResourceTypes returns the SCIM resource types supported.
func ResourceTypes(baseURL string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"schemas":     []string{SchemaResourceType},
			"id":          "User",
			"name":        "User",
			"endpoint":    "/Users",
			"description": "User Account",
			"schema":      SchemaUser,
			"meta": map[string]string{
				"resourceType": "ResourceType",
				"location":     baseURL + "/ResourceTypes/User",
			},
		},
		{
			"schemas":     []string{SchemaResourceType},
			"id":          "Group",
			"name":        "Group",
			"endpoint":    "/Groups",
			"description": "Group",
			"schema":      SchemaGroup,
			"meta": map[string]string{
				"resourceType": "ResourceType",
				"location":     baseURL + "/ResourceTypes/Group",
			},
		},
	}
}

// FormatTime formats a time for SCIM responses.
func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// ParseFilter is a simplified SCIM filter parser.
// Supports: userName eq "value"
func ParseFilter(filter string) (attr, op, value string, err error) {
	parts := strings.SplitN(filter, " ", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid filter: %s", filter)
	}
	attr = parts[0]
	op = strings.ToLower(parts[1])
	value = strings.Trim(parts[2], "\"")
	return attr, op, value, nil
}
