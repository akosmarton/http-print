package protocol

// AuthReq is the authorization request
type AuthReq struct {
	APIKey string
}

// AuthResp is the authorization response
type AuthResp struct {
	AccessToken string
}
