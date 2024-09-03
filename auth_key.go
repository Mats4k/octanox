package octanox

import "github.com/gin-gonic/gin"

type ApiKeyAuthenticator struct {
	provider UserProvider
}

func (a *ApiKeyAuthenticator) Authenticate(c *gin.Context) (User, error) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return nil, nil
	}

	user, err := a.provider.ProvideByApiKey(apiKey)
	if err != nil {
		return nil, err
	}

	return user, nil
}
