// Package auth provides authentication interfaces for SSH access.
package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
)

var nonASCIIRe = regexp.MustCompile("[[:^ascii:]]")

// GitHubAuthenticator validates SSH keys against GitHub organization membership.
type GitHubAuthenticator struct {
	orgName        string
	client         *github.Client
	httpClient     *http.Client
	keyToUser      sync.Map // maps SSH key to username
	users          sync.Map // maps username to UserInfo
	updateInterval time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.Mutex
	updating       bool
}

// NewGitHubAuthenticator creates an authenticator for the specified GitHub organization.
func NewGitHubAuthenticator(orgName string, updateInterval time.Duration) *GitHubAuthenticator {
	ctx, cancel := context.WithCancel(context.Background())

	// Set up HTTP client for public endpoints
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Set up GitHub API client with token if available
	var client *github.Client
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
		klog.Infof("Using authenticated GitHub client for org %s", orgName)
	} else {
		client = github.NewClient(nil)
		klog.Infof("Using anonymous GitHub client for org %s", orgName)
	}

	auth := &GitHubAuthenticator{
		orgName:        orgName,
		client:         client,
		httpClient:     httpClient,
		updateInterval: updateInterval,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Initial data fetch
	auth.Update()

	return auth
}

// Close stops the background update loop.
func (a *GitHubAuthenticator) Close() {
	a.cancel()
}

// Update refreshes the cache of users and SSH keys.
func (a *GitHubAuthenticator) Update() error {
	a.mu.Lock()
	if a.updating {
		a.mu.Unlock()
		return nil
	}
	a.updating = true
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.updating = false
		a.mu.Unlock()
	}()

	klog.Infof("Fetching members of GitHub Org %q", a.orgName)

	members, err := a.fetchOrgMembers()
	if err != nil {
		klog.Errorf("org members fetch failed: %w", err)
		return fmt.Errorf("failed to fetch organization members: %w", err)
	}

	// Process members
	keyMap := make(map[string]*UserInfo)
	userMap := make(map[string]UserInfo)

	klog.Infof("fetching SSH keys for %d members ...", len(members))
	for _, member := range members {
		// Get keys using public endpoint instead of API
		klog.Infof("gathering keys for %s", member.Username)
		keys, err := a.fetchUserKeysPublic(member.Username)
		if err != nil {
			klog.Infof("Warning: failed to fetch keys for %s: %v", member.Username, err)
		}
		member.PublicKeys = keys

		// Update maps
		userMap[member.Username] = *member
		for _, key := range keys {
			if key != "" {
				keyMap[key] = member
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	// Update sync.Maps
	for k, v := range keyMap {
		a.keyToUser.Store(k, v)
	}
	for k, v := range userMap {
		a.users.Store(k, v)
	}

	klog.Infof("Updated auth data: %d users, %d keys", len(userMap), len(keyMap))

	return nil
}

// UpdateLoop starts a background goroutine that periodically updates the cache.
func (a *GitHubAuthenticator) UpdateLoop() {
	go func() {
		ticker := time.NewTicker(a.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := a.Update(); err != nil {
					klog.Infof("Error updating auth data: %v", err)
				}
			case <-a.ctx.Done():
				return
			}
		}
	}()
	klog.Infof("Started SSH key update loop with interval %v", a.updateInterval)
}

// ValidKey checks if an SSH key belongs to an organization member.
func (a *GitHubAuthenticator) ValidKey(key string) *UserInfo {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}

	if userInfo, ok := a.keyToUser.Load(key); ok {
		return userInfo.(*UserInfo)
	}
	return nil
}

// fetchOrgMembers retrieves all members of the GitHub organization with detailed user information.
func (a *GitHubAuthenticator) fetchOrgMembers() ([]*UserInfo, error) {
	var members []*UserInfo
	opts := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		users, resp, err := a.client.Organizations.ListMembers(a.ctx, a.orgName, opts)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			klog.Infof("looking up %s ...", user.GetLogin())
			if username := user.GetLogin(); username != "" {
				// Get detailed user information
				details, _, err := a.client.Users.Get(a.ctx, username)
				if err != nil {
					klog.Infof("Warning: couldn't fetch detailed info for %s: %v", username, err)
					members = append(members, &UserInfo{Username: username})
					continue
				}

				members = append(members, &UserInfo{
					Username:        username,
					Name:            cleanChar(details.GetName()),
					Company:         cleanChar(details.GetCompany()),
					Blog:            cleanChar(details.GetBlog()),
					Location:        cleanChar(details.GetLocation()),
					Email:           cleanChar(details.GetEmail()),
					Bio:             cleanChar(details.GetBio()),
					TwitterUsername: cleanChar(details.GetTwitterUsername()),
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return members, nil
}

func cleanChar(s string) string {
	return strings.TrimSpace(nonASCIIRe.ReplaceAllLiteralString(s, ""))
}

// fetchUserKeysPublic retrieves SSH keys using GitHub's public keys endpoint.
func (a *GitHubAuthenticator) fetchUserKeysPublic(username string) ([]string, error) {
	url := fmt.Sprintf("https://github.com/%s.keys", username)

	klog.Infof("fetching %s ...", url)
	req, err := http.NewRequestWithContext(a.ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch keys, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	if content == "" {
		return []string{}, nil
	}

	// Split content by newlines to get individual keys
	keys := strings.Split(strings.TrimSpace(content), "\n")
	return keys, nil
}
