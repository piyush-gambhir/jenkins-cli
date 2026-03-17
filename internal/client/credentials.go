package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Credential represents a Jenkins credential.
type Credential struct {
	ID          string `json:"id"`
	TypeName    string `json:"typeName"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Domain      string `json:"domain,omitempty"`
	Fingerprint *CredentialFingerprint `json:"fingerprint,omitempty"`
}

// CredentialFingerprint represents credential usage info.
type CredentialFingerprint struct {
	Usage []CredentialUsage `json:"usage,omitempty"`
}

// CredentialUsage represents where a credential is used.
type CredentialUsage struct {
	Name   string `json:"name"`
	Ranges json.RawMessage `json:"ranges"`
}

// CredentialStore represents a credential store.
type CredentialStore struct {
	Domains []CredentialDomain `json:"domains"`
}

// CredentialDomain represents a credential domain.
type CredentialDomain struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	URLName     string       `json:"urlName"`
	Credentials []Credential `json:"credentials"`
}

// CredentialListResponse wraps a list of credentials.
type CredentialListResponse struct {
	Credentials []Credential `json:"credentials"`
}

// credentialBasePath builds the base path for credential operations.
func credentialBasePath(store, domain string) string {
	if store == "" {
		store = "system"
	}
	if domain == "" {
		domain = "_"
	}
	return fmt.Sprintf("/credentials/store/%s/domain/%s", url.PathEscape(store), url.PathEscape(domain))
}

// ListCredentials lists all credentials in a store/domain.
func (c *Client) ListCredentials(store, domain string) ([]Credential, error) {
	path := credentialBasePath(store, domain)

	data, err := c.Get(path, DepthParam(1))
	if err != nil {
		return nil, fmt.Errorf("listing credentials: %w", err)
	}

	var resp struct {
		Credentials []Credential `json:"credentials"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	return resp.Credentials, nil
}

// GetCredential gets a specific credential.
func (c *Client) GetCredential(store, domain, id string) (*Credential, error) {
	path := credentialBasePath(store, domain) + "/credential/" + url.PathEscape(id)

	data, err := c.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting credential: %w", err)
	}

	var cred Credential
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, fmt.Errorf("parsing credential: %w", err)
	}

	return &cred, nil
}

// CreateCredential creates a new credential.
func (c *Client) CreateCredential(store, domain, configXML string) error {
	path := credentialBasePath(store, domain) + "/createCredentials"

	_, _, err := c.PostXML(path, nil, configXML)
	if err != nil {
		return fmt.Errorf("creating credential: %w", err)
	}

	return nil
}

// UpdateCredential updates a credential.
func (c *Client) UpdateCredential(store, domain, id, configXML string) error {
	path := credentialBasePath(store, domain) + "/credential/" + url.PathEscape(id) + "/config.xml"

	_, _, err := c.PostXML(path, nil, configXML)
	if err != nil {
		return fmt.Errorf("updating credential: %w", err)
	}

	return nil
}

// DeleteCredential deletes a credential.
func (c *Client) DeleteCredential(store, domain, id string) error {
	path := credentialBasePath(store, domain) + "/credential/" + url.PathEscape(id) + "/doDelete"

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("deleting credential: %w", err)
	}

	return nil
}
