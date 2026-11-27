package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type View struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Jobs        []Job  `json:"jobs"`
}

type ViewListResponse struct {
	Views []View `json:"views"`
}

func (c *Client) ListViews() ([]View, error) {
	tree := "views[name,url,description]"
	query := TreeParam(tree)

	data, err := c.Get("", query)
	if err != nil {
		return nil, fmt.Errorf("listing views: %w", err)
	}

	var resp ViewListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing views: %w", err)
	}

	return resp.Views, nil
}

func (c *Client) GetView(name string) (*View, error) {
	path := fmt.Sprintf("/view/%s", url.PathEscape(name))

	data, err := c.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting view: %w", err)
	}

	var view View
	if err := json.Unmarshal(data, &view); err != nil {
		return nil, fmt.Errorf("parsing view: %w", err)
	}

	return &view, nil
}

func (c *Client) CreateView(name, viewType string) error {
	if viewType != "" && viewType != "hudson.model.ListView" {
		return fmt.Errorf("unsupported view type %q: only hudson.model.ListView is supported", viewType)
	}

	configXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<listView>
  <name>%s</name>
  <filterExecutors>false</filterExecutors>
  <filterQueue>false</filterQueue>
  <properties class="hudson.model.View$PropertyList"/>
  <jobNames>
    <comparator class="hudson.util.CaseInsensitiveComparator"/>
  </jobNames>
  <jobFilters/>
  <columns>
    <hudson.views.StatusColumn/>
    <hudson.views.WeatherColumn/>
    <hudson.views.JobColumn/>
    <hudson.views.LastSuccessColumn/>
    <hudson.views.LastFailureColumn/>
    <hudson.views.LastDurationColumn/>
    <hudson.views.BuildButtonColumn/>
  </columns>
</listView>`, name)

	query := url.Values{"name": {name}}
	_, _, err := c.PostXML("/createView", query, configXML)
	if err != nil {
		return fmt.Errorf("creating view: %w", err)
	}

	return nil
}

func (c *Client) DeleteView(name string) error {
	path := fmt.Sprintf("/view/%s/doDelete", url.PathEscape(name))

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("deleting view: %w", err)
	}

	return nil
}

func (c *Client) AddJobToView(viewName, jobName string) error {
	path := fmt.Sprintf("/view/%s/addJobToView", url.PathEscape(viewName))
	query := url.Values{"name": {jobName}}

	_, err := c.Post(path, query)
	if err != nil {
		return fmt.Errorf("adding job to view: %w", err)
	}

	return nil
}

func (c *Client) RemoveJobFromView(viewName, jobName string) error {
	path := fmt.Sprintf("/view/%s/removeJobFromView", url.PathEscape(viewName))
	query := url.Values{"name": {jobName}}

	_, err := c.Post(path, query)
	if err != nil {
		return fmt.Errorf("removing job from view: %w", err)
	}

	return nil
}
