package api

type AuthResponse struct {
	StatusCode int    `json:"statusCode"`
	Time       string `json:"time"`
	Data       struct {
		Message string `json:"message"`
	} `json:"data"`
}

type SignInRequest struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Type      string `json:"type"`
}

type SignInResponse struct {
	StatusCode int    `json:"statusCode"`
	Time       string `json:"time"`
	Data       struct {
		AccessToken string `json:"accessToken"`
		ExpiresAt   string `json:"expiresAt"`
	} `json:"data"`
}

type Country struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	LogoURL string `json:"logoURL"`
}

type VoteOrder struct {
	ID     string `json:"_id"`
	Status string `json:"status"`
}

type VoteOrderResponse struct {
	StatusCode int    `json:"statusCode"`
	Time       string `json:"time"`
	Data       struct {
		ID     string `json:"_id"`
		Status string `json:"status"`
	} `json:"data"`
}

type ConfirmResponse struct {
	StatusCode int    `json:"statusCode"`
	Time       string `json:"time"`
	Data       struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	} `json:"data"`
}
