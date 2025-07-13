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

// ResourceAccessType defines the structure for resource access claims
type ResourceAccessType map[string]map[string][]string

// UserInfo represents the user information returned by the OIDC provider
type UserInfo struct {
	Subject             string  // sub
	Name                string  // name
	GivenName           string  // given_name
	FamilyName          string  // family_name
	MiddleName          string  // middle_name
	Nickname            string  // nickname
	PreferredUsername   string  // preferred_username
	Profile             string  // profile
	Picture             string  // picture
	Website             string  // website
	Email               string  // email
	EmailVerified       bool    // email_verified
	Gender              string  // gender
	Birthdate           string  // birthdate
	Zoneinfo            string  // zoneinfo
	Locale              string  // locale
	PhoneNumber         string  // phone_number
	PhoneNumberVerified bool    // phone_number_verified
	Address             Address // address (OIDC address claim is a JSON object)
	UpdatedAt           int64   // updated_at (Unix timestamp)

	ResourceAccess ResourceAccessType // resource_access (map of resource names to scopes)

	TenantID       string // Tenant ID, if applicable
	SubscriptionID string // Subscription ID, if applicable
}

// Address represents the address structure in OIDC UserInfo
type Address struct {
	Formatted     string
	StreetAddress string
	Locality      string
	Region        string
	PostalCode    string
	Country       string
}

// Convert converts oidc.UserInfo to our UserInfo struct
func Convert(userInfo *oidc.UserInfo, resourceAccess map[string]interface{}) *UserInfo {
	if userInfo == nil {
		return nil
	}

	ui := &UserInfo{}
	populateBasicFields(ui, userInfo)
	populateAddress(ui, userInfo)
	populateClaims(ui, userInfo)
	populateResourceAccess(ui, resourceAccess)

	return ui
}

func populateBasicFields(ui *UserInfo, userInfo *oidc.UserInfo) {
	ui.Subject = userInfo.Subject
	ui.Name = userInfo.Name
	ui.GivenName = userInfo.GivenName
	ui.FamilyName = userInfo.FamilyName
	ui.MiddleName = userInfo.MiddleName
	ui.Nickname = userInfo.Nickname
	ui.PreferredUsername = userInfo.PreferredUsername
	ui.Profile = userInfo.Profile
	ui.Picture = userInfo.Picture
	ui.Website = userInfo.Website
	ui.Email = userInfo.Email
	ui.EmailVerified = bool(userInfo.EmailVerified)
	ui.Gender = string(userInfo.Gender)
	ui.Birthdate = userInfo.Birthdate
	ui.Zoneinfo = userInfo.Zoneinfo
	ui.Locale = userInfo.Locale.String()
	ui.PhoneNumber = userInfo.PhoneNumber
	ui.PhoneNumberVerified = userInfo.PhoneNumberVerified
	ui.UpdatedAt = int64(userInfo.UpdatedAt)
}

func populateAddress(ui *UserInfo, userInfo *oidc.UserInfo) {
	if userInfo.Address != nil {
		ui.Address = Address{
			Formatted:     userInfo.Address.Formatted,
			StreetAddress: userInfo.Address.StreetAddress,
			Locality:      userInfo.Address.Locality,
			Region:        userInfo.Address.Region,
			PostalCode:    userInfo.Address.PostalCode,
			Country:       userInfo.Address.Country,
		}
	} else {
		ui.Address = Address{}
	}
}

func populateClaims(ui *UserInfo, userInfo *oidc.UserInfo) {
	if userInfo.Claims != nil {
		ui.TenantID, _ = userInfo.Claims["tenantId"].(string)
		ui.SubscriptionID, _ = userInfo.Claims["subscriptionId"].(string)
	}
}

func populateResourceAccess(ui *UserInfo, resourceAccess map[string]interface{}) {
	if resourceAccess != nil {
		ui.ResourceAccess = make(ResourceAccessType)
		for resource, scopes := range resourceAccess {
			if scopeMap, ok := scopes.(map[string]interface{}); ok {
				ui.ResourceAccess[resource] = make(map[string][]string)
				for scope, values := range scopeMap {
					if valueArray, ok := values.([]interface{}); ok {
						stringArray := make([]string, len(valueArray))
						for i, v := range valueArray {
							if str, ok := v.(string); ok {
								stringArray[i] = str
							}
						}
						ui.ResourceAccess[resource][scope] = stringArray
					}
				}
			}
		}
	} else {
		ui.ResourceAccess = nil
	}
}
