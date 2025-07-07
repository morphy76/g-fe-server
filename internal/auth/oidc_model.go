package auth

import (
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

const (
	// SessionKeyAuthenticated is the key used in the session to indicate if the user is authenticated
	SessionKeyAuthenticated = "authenticated"
	// SessionKeyIssuer is the key used in the session to store the OIDC issuer
	SessionKeyIssuer = "issuer"
	// SessionKeySubject is the key used in the session to store the OIDC subject
	SessionKeySubject = "subject"
	// SessionKeySessionID is the key used in the session to store the OIDC session ID
	SessionKeySessionID = "session_id"
	// SessionKeyIDToken is the key used in the session to store the OIDC ID token
	SessionKeyIDToken = "id_token"
	// SessionKeyUserInfo is the key used in the session to store the user information
	SessionKeyUserInfo = "user_info"
	// SessionKeyAccessToken is the key used in the session to store the OIDC access token
	SessionKeyAccessToken = "access_token"
	// SessionKeyRefreshToken is the key used in the session to store the OIDC refresh token
	SessionKeyRefreshToken = "refresh_token"
	// SessionKeyExpiresIn is the key used in the session to store the OIDC token expiration time
	SessionKeyExpiresIn = "expires_in"
	// SessionKeyJTI is the key used in the session to store the JWT ID (JTI)
	SessionKeyJTI = "jti"
	// SessionKeyExpiresAt is the key used in the session to store the token expiration time
	SessionKeyExpiresAt = "expires_at"
	// SessionKeySessionState is the key used in the session to store the OIDC session state
	SessionKeySessionState = "session_state"
	// SessionKeyTokenIssuedAt is the key used in the session to store the token issued at time
	SessionKeyTokenIssuedAt = "token_issued_at"
)

// OIDCOptions holds the configuration for the OIDC client
type OIDCOptions struct {
	Disabled            bool
	Issuer              string
	ClientID            string
	ClientSecret        string
	Scopes              []string
	ExtraAuthArgs       map[string]string
	ResourceAccessClaim string
}

// UserInfo represents the user information returned by the OIDC provider
type UserInfo struct {
	Subject             string                 // sub
	Name                string                 // name
	GivenName           string                 // given_name
	FamilyName          string                 // family_name
	MiddleName          string                 // middle_name
	Nickname            string                 // nickname
	PreferredUsername   string                 // preferred_username
	Profile             string                 // profile
	Picture             string                 // picture
	Website             string                 // website
	Email               string                 // email
	EmailVerified       bool                   // email_verified
	Gender              string                 // gender
	Birthdate           string                 // birthdate
	Zoneinfo            string                 // zoneinfo
	Locale              string                 // locale
	PhoneNumber         string                 // phone_number
	PhoneNumberVerified bool                   // phone_number_verified
	Address             map[string]interface{} // address (OIDC address claim is a JSON object)
	UpdatedAt           int64                  // updated_at (Unix timestamp)
	RawClaims           map[string]interface{} // any additional claims
	ResourceAccess      map[string]interface{} // resource_access (map of resource names to scopes)
}

// Convert converts oidc.UserInfo to our UserInfo struct
func Convert(userInfo *oidc.UserInfo, resourceAccess map[string]interface{}) *UserInfo {
	if userInfo == nil {
		return nil
	}

	// Copy standard claims
	ui := &UserInfo{
		Subject:             userInfo.Subject,
		Name:                userInfo.Name,
		GivenName:           userInfo.GivenName,
		FamilyName:          userInfo.FamilyName,
		MiddleName:          userInfo.MiddleName,
		Nickname:            userInfo.Nickname,
		PreferredUsername:   userInfo.PreferredUsername,
		Profile:             userInfo.Profile,
		Picture:             userInfo.Picture,
		Website:             userInfo.Website,
		Email:               userInfo.Email,
		EmailVerified:       bool(userInfo.EmailVerified),
		Gender:              string(userInfo.Gender),
		Birthdate:           userInfo.Birthdate,
		Zoneinfo:            userInfo.Zoneinfo,
		Locale:              userInfo.Locale.String(),
		PhoneNumber:         userInfo.PhoneNumber,
		PhoneNumberVerified: userInfo.PhoneNumberVerified,
		UpdatedAt:           int64(userInfo.UpdatedAt),
	}

	if userInfo.Address != nil {
		addr := make(map[string]interface{})
		if userInfo.Address.Formatted != "" {
			addr["formatted"] = userInfo.Address.Formatted
		}
		if userInfo.Address.StreetAddress != "" {
			addr["street_address"] = userInfo.Address.StreetAddress
		}
		if userInfo.Address.Locality != "" {
			addr["locality"] = userInfo.Address.Locality
		}
		if userInfo.Address.Region != "" {
			addr["region"] = userInfo.Address.Region
		}
		if userInfo.Address.PostalCode != "" {
			addr["postal_code"] = userInfo.Address.PostalCode
		}
		if userInfo.Address.Country != "" {
			addr["country"] = userInfo.Address.Country
		}
		ui.Address = addr
	}

	if userInfo.Claims != nil {
		ui.RawClaims = make(map[string]interface{})
		for k, v := range userInfo.Claims {
			ui.RawClaims[k] = v
		}
	}

	if resourceAccess != nil {
		ui.ResourceAccess = make(map[string]interface{})
		for resource, scopes := range resourceAccess {
			if scopesMap, ok := scopes.(map[string]interface{}); ok {
				ui.ResourceAccess[resource] = scopesMap
			} else {
				ui.ResourceAccess[resource] = scopes
			}
		}
	}

	return ui
}
