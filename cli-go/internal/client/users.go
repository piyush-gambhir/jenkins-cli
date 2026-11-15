package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type User struct {
	ID                     string   `json:"id"`
	FullName               string   `json:"fullName"`
	Description            string   `json:"description"`
	AbsoluteURL            string   `json:"absoluteUrl"`
	LastGrantedAuthorities []string `json:"lastGrantedAuthorities,omitempty"`
}

type UserInfo struct {
	ID          string `json:"id"`
	FullName    string `json:"fullName"`
	Description string `json:"description"`
	AbsoluteURL string `json:"absoluteUrl"`
}

type UserListItem struct {
	User       User        `json:"user"`
	LastChange int64       `json:"lastChange"`
	Project    UserProject `json:"project"`
}

type UserProject struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type UserListResponse struct {
	Users []UserListItem `json:"users"`
}

func (c *Client) ListUsers() ([]UserListItem, error) {
	data, err := c.Get("/asynchPeople", nil)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	var resp UserListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing users: %w", err)
	}

	return resp.Users, nil
}

func (c *Client) GetUser(id string) (*User, error) {
	path := fmt.Sprintf("/user/%s", url.PathEscape(id))

	data, err := c.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("parsing user: %w", err)
	}

	return &user, nil
}

func (c *Client) WhoAmI() (*UserInfo, error) {
	data, err := c.Get("/me", nil)
	if err != nil {
		return nil, fmt.Errorf("getting current user: %w", err)
	}

	var user UserInfo
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("parsing current user: %w", err)
	}

	return &user, nil
}
