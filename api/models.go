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

type OrderResponse struct {
	StatusCode int    `json:"statusCode"`
	Time       string `json:"time"`
	Data       struct {
		Order struct {
			ID         string `json:"_id"`
			Status     string `json:"status"`
			FeedAmount int    `json:"feedAmount"`
		} `json:"order"`
		Payment struct {
			ContractAddress string `json:"contractAddress"`
			ABI             []struct {
				Inputs []struct {
					Name         string `json:"name"`
					Type         string `json:"type"`
					InternalType string `json:"internalType"`
				} `json:"inputs"`
				Name            string     `json:"name"`
				Outputs         []struct{} `json:"outputs"`
				StateMutability string     `json:"stateMutability"`
				Type            string     `json:"type"`
			} `json:"abi"`
			FunctionName string `json:"functionName"`
			Params       struct {
				CandidateID        string `json:"candidateID"`
				FeedAmount         int    `json:"feedAmount"`
				RequestID          string `json:"requestID"`
				RequestData        string `json:"requestData"`
				UserHashedMessage  string `json:"userHashedMessage"`
				IntegritySignature string `json:"integritySignature"`
			} `json:"params"`
		} `json:"payment"`
	} `json:"data"`
}
