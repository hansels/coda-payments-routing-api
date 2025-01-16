package model

type ServerRegisterRequest struct {
	URL string `json:"url"`
}

type ServerUnregisterRequest struct {
	URL string `json:"url"`
}
