package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Plugin struct {
	ShortName     string      `json:"shortName"`
	LongName      string      `json:"longName"`
	Version       string      `json:"version"`
	URL           string      `json:"url"`
	Active        bool        `json:"active"`
	Enabled       bool        `json:"enabled"`
	HasUpdate     bool        `json:"hasUpdate"`
	Pinned        bool        `json:"pinned"`
	Dependencies  []PluginDep `json:"dependencies"`
	BackupVersion string      `json:"backupVersion"`
}

type PluginDep struct {
	ShortName string `json:"shortName"`
	Version   string `json:"version"`
	Optional  bool   `json:"optional"`
}

type PluginListResponse struct {
	Plugins []Plugin `json:"plugins"`
}

type PluginInstallStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

type UpdateCenter struct {
	AvailableUpdates []UpdateCenterPlugin `json:"updates"`
}

type UpdateCenterPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

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

func (c *Client) UninstallPlugin(name string) error {
	path := fmt.Sprintf("/pluginManager/plugin/%s/doUninstall", url.PathEscape(name))

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("uninstalling plugin: %w", err)
	}

	return nil
}

func (c *Client) CheckPluginUpdates() ([]Plugin, error) {
	_, _ = c.Post("/pluginManager/checkUpdatesServer", nil)

	data, err := c.Get("/updateCenter", TreeParam("updates[name,version]"))
	if err != nil {
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

	updates := make([]Plugin, 0, len(uc.AvailableUpdates))
	for _, u := range uc.AvailableUpdates {
		updates = append(updates, Plugin{
			ShortName: u.Name,
			Version:   u.Version,
			HasUpdate: true,
		})
	}
	return updates, nil
}
