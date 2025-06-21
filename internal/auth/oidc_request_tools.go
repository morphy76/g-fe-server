package auth

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/cmd/options"
)

// GetUserInfoFn retrieves user information from the session.
type GetUserInfoFn func(r *http.Request) (*UserInfo, error)

// GetAccessTokenFn retrieves the access token from the session.
type GetAccessTokenFn func(r *http.Request) (string, error)

// GetRefreshTokenFn retrieves the refresh token from the session.
type GetRefreshTokenFn func(r *http.Request) (string, error)

// GetExpiresInFn retrieves the expiration time of the access token from the session.
type GetExpiresInFn func(r *http.Request) (int64, error)

// BuildOIDCTools constructs functions to retrieve user info, access token, and refresh token from the session store.
func BuildOIDCTools(
	httpOptions *options.HTTPOptions,
	sessionStore sessions.Store,
) (
	getUserInfoFn GetUserInfoFn,
	getAccessTokenFn GetAccessTokenFn,
	getRefreshTokenFn GetRefreshTokenFn,
	getExpiresInFn GetExpiresInFn,
) {
	getUserInfoFn = func(r *http.Request) (*UserInfo, error) {
		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			return nil, err
		}
		rv, found := session.Values["user_info"]
		if !found {
			return nil, fmt.Errorf("user info not found in session")
		}
		userInfo, ok := rv.(*UserInfo)
		if !ok {
			return nil, fmt.Errorf("user info in session is not of correct type")
		}
		return userInfo, nil
	}

	getAccessTokenFn = func(r *http.Request) (string, error) {
		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			return "", err
		}
		rv, found := session.Values["access_token"]
		if !found {
			return "", fmt.Errorf("access token not found in session")
		}
		accessToken, ok := rv.(string)
		if !ok {
			return "", fmt.Errorf("access token in session is not of correct type")
		}
		return accessToken, nil
	}

	getRefreshTokenFn = func(r *http.Request) (string, error) {
		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			return "", err
		}
		rv, found := session.Values["refresh_token"]
		if !found {
			return "", fmt.Errorf("refresh token not found in session")
		}
		refreshToken, ok := rv.(string)
		if !ok {
			return "", fmt.Errorf("refresh token in session is not of correct type")
		}
		return refreshToken, nil
	}

	getExpiresInFn = func(r *http.Request) (int64, error) {
		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			return 0, err
		}
		rv, found := session.Values["expires_in"]
		if !found {
			return 0, fmt.Errorf("expires_in not found in session")
		}
		expiresIn, ok := rv.(int64)
		if !ok {
			return 0, fmt.Errorf("expires_in in session is not of correct type")
		}
		return expiresIn, nil
	}

	return getUserInfoFn, getAccessTokenFn, getRefreshTokenFn, getExpiresInFn
}
