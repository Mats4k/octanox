package octanox

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type OAuth2BearerAuthenticator struct {
	provider             OAuth2UserProvider
	config               oauth2.Config
	loginSuccessRedirect string
	secret               []byte
	exp                  int64
	states               StateMap
}

// SetExp sets the expiration time for the token.
func (a *OAuth2BearerAuthenticator) SetExp(exp int64) {
	a.exp = exp
}

func (a *OAuth2BearerAuthenticator) Method() AuthenticationMethod {
	return AuthenticationMethodBearerOAuth2
}

func (a *OAuth2BearerAuthenticator) Authenticate(c *gin.Context) (User, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		return nil, nil
	}

	userID := a.extractToken(token[7:])
	if userID == nil {
		return nil, nil
	}

	user, err := a.provider.ProvideByID(*userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *OAuth2BearerAuthenticator) login(c *gin.Context) {
	url := a.config.AuthCodeURL(a.states.Generate(300))

	c.Redirect(302, url)
}

func (a *OAuth2BearerAuthenticator) callback(c *gin.Context) {
	state := c.Query("state")
	if !a.states.ValidateOnce(state) {
		c.String(400, "invalid state")
		return
	}

	code := c.Query("code")

	token, err := a.config.Exchange(context.Background(), code)
	if err != nil {
		c.String(400, "Token Exchange Failed")
		return
	}

	user, err := a.provider.ProvideForLogin(token.AccessToken)
	if err != nil {
		panic(err)
	}

	if user == nil {
		c.String(400, "User not found")
		return
	}

	jwt, err := a.createToken(user)
	if err != nil {
		panic("octanox: failed to create token")
	}

	c.Redirect(302, a.loginSuccessRedirect+"?token="+jwt)
}

func (a *OAuth2BearerAuthenticator) registerRoutes(r *gin.RouterGroup) {
	r.GET("/login", a.login)
	r.GET("/oauth2/callback", a.callback)
}

func (a *OAuth2BearerAuthenticator) createToken(user User) (string, error) {
	currTime := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "Octanox Auth",
		"aud": "octanox",
		"sub": user.ID(),
		"exp": time.Now().Add(time.Second * time.Duration(a.exp)).Unix(),
		"iat": currTime,
		"nbf": currTime,
		"jti": uuid.New().String(),
	})

	return token.SignedString(a.secret)
}

func (a *OAuth2BearerAuthenticator) extractToken(tokenString string) *uuid.UUID {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return a.secret, nil
	})
	if err != nil {
		return nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		subClaim, ok := claims["sub"]
		if !ok {
			return nil
		}

		subject, err := uuid.Parse(subClaim.(string))
		if err != nil {
			return nil
		}

		return &subject
	}

	return nil
}
