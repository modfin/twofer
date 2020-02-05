package mfreja

type Verifiable interface {
	JWSToken() string
}

type SSN struct {
	Country string `json:"country"`
	SSN     string `json:"ssn"`
}
type BasicUserInfo struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
}


type UserInfoType string
const (
	UIT_PHONE    UserInfoType = "PHONE"
	UIT_EMAIL    UserInfoType = "EMAIL"
	UIT_SSN      UserInfoType = "SSN"
	UIT_INFERRED UserInfoType = "INFERRED" // QR code auth
)


type Status string
const (
	STATUS_STARTED             = "STARTED"             //(the transaction has been started but not yet delivered to Freja eID application associated with the end user),
	STATUS_DELIVERED_TO_MOBILE = "DELIVERED_TO_MOBILE" //(the Freja eID app has downloaded the transaction),
	STATUS_CANCELED            = "CANCELED"            //(the end user declined the authentication request),
	STATUS_RP_CANCELED         = "RP_CANCELED"         //(the authentication request was sent to the user but cancelled by the RP before the user could respond),
	STATUS_EXPIRED             = "EXPIRED"             //(the authentication request was not approved by the end user within the authentication validity limit of two minutes),
	STATUS_APPROVED            = "APPROVED"            //(the authentication was successful) or
	STATUS_REJECTED            = "REJECTED"            //(e.g. if you try to run more than one authentication transaction for the same user at the same time).
)

type RegistrationLevel string
const (
	RL_BASIC    RegistrationLevel = "BASIC"
	RL_EXTENDED RegistrationLevel = "EXTENDED"
	RL_PLUS     RegistrationLevel = "PLUS"
)

type AttributeName string
const (
	// (name and surname),
	ATTR_BASIC_USER_INFO AttributeName = "BASIC_USER_INFO"

	// (user's email address),
	ATTR_EMAIL_ADDRESS AttributeName = "EMAIL_ADDRESS"

	// (date of birth),
	ATTR_DATE_OF_BIRTH AttributeName = "DATE_OF_BIRTH"

	// (social security number and country ),
	// country: string, mandatory. Contains the ISO-3166 two-alphanumeric country code of the country where the SSN is
	// issued. In the current version of Freja eID, one of: ''SE'' (Sweden), ''NO'' (Norway), ''FI'' (Finland), ''DK'' (Danmark).
	// ssn: string, mandatory. Expected SSN of the end user as per pre-registration.
	// If country equal to "SE", the value must be the 12-digit format of the Swedish "personnummer" without spaces or
	//   hyphens. Example: 195210131234.
	// If country equal to ''NO'', the value must be the 11-digit format of the Norwegian "personnummer" without spaces
	//   or hyphens. Example: 13105212345.
	// If country equal to ''FI'', the value must be the 10-characters format of the Finish ''koodi'', with the hyphen
	//   before the last four control characters. Hyphen can be replaced with the letter A. Example format: 131052-308T or 131052A308T.
	// If country equal to ''DK'', the value must be the 10-digit format of the Danish "personnummer" without spaces or
	//   hyphens. Example: 1310521234.
	ATTR_SSN AttributeName = "SSN"

	// (a unique, user-specific value that allows the Relying Party to identify the same user across multiple sessions),
	ATTR_RELYING_PARTY_USER_ID AttributeName = "RELYING_PARTY_USER_ID"

	// a unique, user-specific value that allows the Integrator to identify the same user across multiple sessions
	// regardless of the Integrated Relying Party service that the user is using. For more info, see Integrator
	// Relying Party Management
	ATTR_INTEGRATOR_SPECIFIC_USER_ID AttributeName = "INTEGRATOR_SPECIFIC_USER_ID"

	// a unique, Relying Party-specific, user identifier, set by the Relying Party through the Custom Identifier Management.
	ATTR_CUSTOM_IDENTIFIER AttributeName = "CUSTOM_IDENTIFIER"
)




