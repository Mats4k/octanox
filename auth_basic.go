package octanox

import "github.com/gin-gonic/gin"

type BasicAuthenticator struct {
	provider UserProvider
}

func (a *BasicAuthenticator) Authenticate(c *gin.Context) (User, error) {
	username, password, ok := c.Request.BasicAuth()
	if !ok {
		return nil, nil
	}

	user, err := a.provider.ProvideByUserPass(username, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}
