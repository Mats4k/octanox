package octanox

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserProvider is an interface that allows the authentication module to access the user data.
type UserProvider interface {
	// ProvideByUserPass provides the user data for the given username and password. If the user data cannot be provided, it should return an error.
	ProvideByUserPass(username, password string) (User, error)
	// ProvideByID provides the user data for the given user ID. If the user data cannot be provided, it should return an error.
	// This should be used to provide the user data when the authentication is called, like providing it by the user ID in the token.
	ProvideByID(id uuid.UUID) (User, error)
	// ProvideByApiKey provides the user data for the given API key. If the user data cannot be provided, it should return an error.
	ProvideByApiKey(apiKey string) (User, error)
}

// Authenticator is an struct that defines the authentication module.
type Authenticator interface {
	// Authenticate authenticates the client request. Gets the client request context and returns the authenticated user.
	// If the authentication fails, it should return nil.
	Authenticate(c *gin.Context) (User, error)
}

// AuthenticatorBuilder is a struct that helps build the Authenticator.
type AuthenticatorBuilder struct {
	instance *Instance
	provider UserProvider
}

// Plugs in the authentication module into Octanox.
func (i *Instance) Authenticate(provider UserProvider) *AuthenticatorBuilder {
	if i.Authenticator != nil {
		panic("octanox: authenticator already exists")
	}

	return &AuthenticatorBuilder{i, provider}
}

// Bearer creates a new BearerAuthenticator with the given secret and plugs it into the Authenticator.
// The basePath is the base path for the authentication routes.
// The secret is the secret key used to sign the JWT token.
// Defaults to 1 day for the token expiration time.
func (b *AuthenticatorBuilder) Bearer(secret, basePath string) *BearerAuthenticator {
	bearer := &BearerAuthenticator{
		provider: b.provider,
		secret:   []byte(secret),
		exp:      86400,
	}

	bearer.registerRoutes(b.instance.Gin.Group(basePath))

	b.instance.Authenticator = bearer

	return bearer
}

// Basic creates a new BasicAuthenticator and plugs it into the Authenticator.
func (b *AuthenticatorBuilder) Basic() *BasicAuthenticator {
	basic := &BasicAuthenticator{
		provider: b.provider,
	}

	b.instance.Authenticator = basic

	return basic
}

// ApiKey creates a new ApiKeyAuthenticator and plugs it into the Authenticator.
func (b *AuthenticatorBuilder) ApiKey() *ApiKeyAuthenticator {
	apiKey := &ApiKeyAuthenticator{
		provider: b.provider,
	}

	b.instance.Authenticator = apiKey

	return apiKey
}
