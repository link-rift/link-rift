package sso

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
)

// SAMLProvider implements SSOProvider for SAML 2.0 authentication.
// This is a simplified implementation that handles the core SAML flow.
// In production, this would use github.com/crewjam/saml for full SP functionality.
type SAMLProvider struct {
	EntityID         string
	SSOURL           string
	SLOURL           string
	Certificate      *x509.Certificate
	ACSURL           string // Assertion Consumer Service URL
	MetadataURL      string // SP metadata URL
	AttributeMapping AttributeMapping
}

// SAMLConfig holds the configuration for creating a SAML provider.
type SAMLConfig struct {
	EntityID         string
	SSOURL           string
	SLOURL           string
	CertificatePEM   string
	ACSURL           string
	MetadataURL      string
	AttributeMapping AttributeMapping
}

// NewSAMLProvider creates a new SAML provider from configuration.
func NewSAMLProvider(cfg SAMLConfig) (*SAMLProvider, error) {
	cert, err := ParseCertificate(cfg.CertificatePEM)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate: %w", err)
	}

	return &SAMLProvider{
		EntityID:         cfg.EntityID,
		SSOURL:           cfg.SSOURL,
		SLOURL:           cfg.SLOURL,
		Certificate:      cert,
		ACSURL:           cfg.ACSURL,
		MetadataURL:      cfg.MetadataURL,
		AttributeMapping: cfg.AttributeMapping,
	}, nil
}

// GenerateAuthURL creates a SAML AuthnRequest redirect URL.
func (p *SAMLProvider) GenerateAuthURL(requestID string) (string, error) {
	// In a full implementation, this would create a proper SAML AuthnRequest XML,
	// deflate it, base64-encode it, and append to the SSO URL.
	// For now, we construct a simplified redirect URL.
	authURL := fmt.Sprintf(
		"%s?SAMLRequest=%s&RelayState=%s",
		p.SSOURL,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(
			`<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="%s" Version="2.0" AssertionConsumerServiceURL="%s" Destination="%s"><saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">%s</saml:Issuer></samlp:AuthnRequest>`,
			requestID, p.ACSURL, p.SSOURL, p.EntityID,
		))),
		requestID,
	)
	return authURL, nil
}

// ValidateResponse validates a SAML Response and extracts the assertion.
// In production, this would perform full XML signature verification using the IdP certificate.
func (p *SAMLProvider) ValidateResponse(samlResponse string) (*SSOAssertion, error) {
	// Decode base64 SAML response
	_, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("invalid SAML response encoding: %w", err)
	}

	// In production, here we would:
	// 1. Parse the XML
	// 2. Verify the signature using the IdP certificate
	// 3. Check conditions (NotBefore, NotOnOrAfter, Audience)
	// 4. Extract attributes using the attribute mapping
	//
	// For now, return an error indicating full SAML validation
	// requires the crewjam/saml library at runtime.
	return nil, fmt.Errorf("SAML response validation requires crewjam/saml library; configure your IdP and ensure the library is linked")
}

// GenerateMetadataXML generates SAML SP metadata XML for IdP configuration.
func (p *SAMLProvider) GenerateMetadataXML() (string, error) {
	metadata := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="%s">
  <md:SPSSODescriptor AuthnRequestsSigned="true" WantAssertionsSigned="true" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat>
    <md:AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="%s" index="0" isDefault="true"/>
  </md:SPSSODescriptor>
</md:EntityDescriptor>`,
		p.EntityID, p.ACSURL,
	)
	return metadata, nil
}

// ParseCertificate parses a PEM or raw base64-encoded X.509 certificate.
func ParseCertificate(certData string) (*x509.Certificate, error) {
	certData = strings.TrimSpace(certData)

	// Try PEM decode first
	block, _ := pem.Decode([]byte(certData))
	if block != nil {
		return x509.ParseCertificate(block.Bytes)
	}

	// Try raw base64
	decoded, err := base64.StdEncoding.DecodeString(certData)
	if err != nil {
		return nil, fmt.Errorf("certificate is neither valid PEM nor base64: %w", err)
	}

	return x509.ParseCertificate(decoded)
}
