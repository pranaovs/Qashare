package models

type Jwtoken struct {
	AToken string `json:"access_token"`
	RToken string `json:"refresh_token"`
	Type   string `json:"token_type"`
	Expiry int64  `json:"expires_in"`
}
