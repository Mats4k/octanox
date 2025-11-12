package octanox

import (
	"context"
	"strings"
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
	pkces                StringStateMap
	// Optional OIDC ID token validation
	validateIDToken bool
	oidcIssuer      string
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
	// Generate a state and PKCE pair
	state := a.states.Generate(300)
	verifier, challenge := generatePKCE()
	a.pkces.Store(state, verifier, 600)

	// Request authorization code with PKCE (S256)
	url := a.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		// Ensure scopes are sent as a space-delimited string
		oauth2.SetAuthURLParam("scope", strings.Join(a.config.Scopes, " ")),
	)

	c.Redirect(302, url)
}

func (a *OAuth2BearerAuthenticator) callback(c *gin.Context) {
	state := c.Query("state")
	if !a.states.ValidateOnce(state) {
		c.String(400, "invalid state")
		return
	}

	code := c.Query("code")

	// Retrieve PKCE verifier for this state
	verifier := a.pkces.Pop(state)
	if verifier == "" {
		c.String(400, "missing PKCE verifier")
		return
	}

	token, err := a.config.Exchange(context.Background(), code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		c.String(400, "Token Exchange Failed")
		return
	}

	// Optionally validate ID token using OIDC discovery + JWKS
	if a.validateIDToken {
		if raw := token.Extra("id_token"); raw != nil {
			idToken, _ := raw.(string)
			if err := validateIDTokenWithIssuer(idToken, a.oidcIssuer, a.config.ClientID); err != nil {
				c.String(400, "Invalid ID Token")
				return
			}
		} else {
			c.String(400, "Missing ID Token")
			return
		}
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

// EnableOIDCValidation enforces validation of ID token against the given issuer using JWKS.
func (a *OAuth2BearerAuthenticator) EnableOIDCValidation(issuer string) {
	a.oidcIssuer = issuer
	a.validateIDToken = true
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
