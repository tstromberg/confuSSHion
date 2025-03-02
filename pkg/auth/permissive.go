package auth

import (
	"context"
	"log"
)

// PermissiveAuthenticator is an authenticator that allows access to anyone.
type PermissiveAuthenticator struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewPermissiveAuthenticator creates an authenticator that allows all access.
func NewPermissiveAuthenticator() *PermissiveAuthenticator {
	ctx, cancel := context.WithCancel(context.Background())
	return &PermissiveAuthenticator{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Close stops any background processes.
func (a *PermissiveAuthenticator) Close() {
	a.cancel()
}

// Update is a no-op for PermissiveAuthenticator.
func (a *PermissiveAuthenticator) Update() error {
	// Nothing to update
	return nil
}

// UpdateLoop is a no-op for PermissiveAuthenticator.
func (a *PermissiveAuthenticator) UpdateLoop() {
	// Nothing to update
	log.Println("Permissive authenticator active - all access allowed")
}

// ValidKey always returns "anonymous" for any key.
func (a *PermissiveAuthenticator) ValidKey(key string) *UserInfo {
	// Allow any key, return "anonymous" as the username
	return &UserInfo{Name: "anonymous"}
}
