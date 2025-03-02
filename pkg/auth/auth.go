package auth

// UserInfo stores information about an authenticated user.
type UserInfo struct {
	Username        string
	Name            string
	Company         string
	Blog            string
	Location        string
	Email           string
	Bio             string
	TwitterUsername string
	PublicKeys      []string
}

// Authenticator defines the interface for SSH key authentication.
type Authenticator interface {
	// ValidKey checks if an SSH key is valid and returns the associated username.
	ValidKey(key string) *UserInfo

	// UpdateLoop starts the background update process.
	UpdateLoop()

	// Update refreshes the authentication data.
	Update() error

	// Close stops background processes and frees resources.
	Close()
}
