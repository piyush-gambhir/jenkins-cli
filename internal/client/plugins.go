package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Plugin represents a Jenkins plugin.
type Plugin struct {
	ShortName       string           `json:"shortName"`
	LongName        string           `json:"longName"`
	Version         string           `json:"version"`
	URL             string           `json:"url"`
	Active          bool             `json:"active"`
	Enabled         bool             `json:"enabled"`
	HasUpdate       bool             `json:"hasUpdate"`
	Pinned          bool             `json:"pinned"`
	Dependencies    []PluginDep      `json:"dependencies"`
	BackupVersion   string           `json:"backupVersion"`
}

// PluginDep represents a plugin dependency.
type PluginDep struct {
	ShortName string `json:"shortName"`
	Version   string `json:"version"`
	Optional  bool   `json:"optional"`
}

// PluginListResponse wraps a list of plugins.
type PluginListResponse struct {
	Plugins []Plugin `json:"plugins"`
}

// PluginInstallStatus represents the status of a plugin installation.
type PluginInstallStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

// UpdateCenter represents update center info.
type UpdateCenter struct {
	AvailableUpdates []Plugin `json:"updates"`
}

// ListPlugins lists all installed plugins.
func (c *Client) ListPlugins() ([]Plugin, error) {
	data, err := c.Get("/pluginManager", DepthParam(1))
	if err != nil {
		return nil, fmt.Errorf("listing plugins: %w", err)
	}

	var resp PluginListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing plugins: %w", err)
	}

	return resp.Plugins, nil
}

// GetPlugin gets details about a specific plugin.
func (c *Client) GetPlugin(shortName string) (*Plugin, error) {
	plugins, err := c.ListPlugins()
	if err != nil {
		return nil, err
	}

	for _, p := range plugins {
		if p.ShortName == shortName {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("plugin %q not found", shortName)
}

// InstallPlugin installs a plugin by name.
func (c *Client) InstallPlugin(name, version string) error {
	pluginID := name
	if version != "" {
		pluginID = name + "@" + version
	}

	xmlBody := fmt.Sprintf(`<jenkins><install plugin="%s" /></jenkins>`, pluginID)

	_, _, err := c.PostXML("/pluginManager/installNecessaryPlugins", nil, xmlBody)
	if err != nil {
		return fmt.Errorf("installing plugin: %w", err)
	}

	return nil
}

// UninstallPlugin uninstalls a plugin.
func (c *Client) UninstallPlugin(name string) error {
	path := fmt.Sprintf("/pluginManager/plugin/%s/doUninstall", url.PathEscape(name))

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("uninstalling plugin: %w", err)
	}

	return nil
}

// CheckPluginUpdates checks for available plugin updates.
func (c *Client) CheckPluginUpdates() ([]Plugin, error) {
	// First trigger an update check
	_, _ = c.Post("/pluginManager/checkUpdatesServer", nil)

	// Then get available updates
	data, err := c.Get("/updateCenter", TreeParam("updates[name,version]"))
	if err != nil {
		// Fallback: list plugins and filter those with updates
		plugins, err := c.ListPlugins()
		if err != nil {
			return nil, err
		}
		var updates []Plugin
		for _, p := range plugins {
			if p.HasUpdate {
				updates = append(updates, p)
			}
		}
		return updates, nil
	}

	var uc UpdateCenter
	if err := json.Unmarshal(data, &uc); err != nil {
		return nil, fmt.Errorf("parsing update center: %w", err)
	}

	return uc.AvailableUpdates, nil
}
