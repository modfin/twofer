package frejam

import "fmt"

func AuthErrorCodes() map[int]string {
	return map[int]string{
		1001: "Invalid or missing userInfoType.",
		1002: "Invalid or missing userInfo.",
		1003: "Invalid restrict.",
		1004: "You are not allowed to call this method.",
		1005: "User has disabled your service.",
		1007: "Invalid min registration level.",
		1008: "Unknown Relying Party.",
		1009: "You are not allowed to request integratorSpecificUserId parameter.",
		1010: "JSON request cannot be parsed.",
		1012: "User with the specified userInfo does not exist in Freja eID database.",
		1100: "Invalid reference (for example, nonexistent or expired).",
		1200: "Invalid or missing includePrevious parameter.",
		2000: "Authentication request failed. Previous authentication request was rejected due to security reasons.",
		2002: "Invalid attributesToReturn parameter.",
		2003: "Custom identifier has to exist when it is requested.",

		3000: "Invalid or missing dataToSignType.",
		3001: "Invalid or missing dataToSign.",
		3002: "Invalid or missing signatureType.",
		3003: "Invalid expiry time.",
		3004: "Invalid push notification.",
		3005: "Invalid attributesToReturn parameter.",
		3006: "Custom identifier has to exist when it is requested.",
		3007: "Invalid title.",
	}
}

type FrejaError struct {
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
}

func (a FrejaError) Error() string {

	d, ok := a.Code.(float64)
	if ok {
		a.Code = int(d)
	}

	return fmt.Sprintf("got code %v, %s", a.Code, a.Message)
}
