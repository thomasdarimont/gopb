package microvault

import (
	"context"

	"encoding/json"
	"net/http"

	"errors"

	"github.com/go-kit/kit/endpoint"
	"golang.org/x/crypto/bcrypt"
)

// Service provides password hashing capabilities.
type Service interface {
	Hash(ctx context.Context, password string) (string, error)

	Validate(ctx context.Context, password, hash string) (bool, error)
}

// NewService makes a new Service
func NewService() Service {
	return microVaultService{}
}

type microVaultService struct{}

func (microVaultService) Hash(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (microVaultService) Validate(ctx context.Context, password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		return false, nil
	}

	return true, nil
}

type hashRequest struct {
	Password string `json:"password"`
}

type hashResponse struct {
	Hash string `json:"hash"`
	Err  string `json:"err,omitempty"`
}

func decodeHashRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req hashRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

type validateRequest struct {
	Password string `json:"password"`
	Hash     string `json:"hash"`
}

type validateResponse struct {
	Valid bool   `json:"valid"`
	Err   string `json:"err,omitempty"`
}

func decodeValidateRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req validateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func MakeHashEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(hashRequest)
		h, err := svc.Hash(ctx, req.Password)
		if err != nil {
			return hashResponse{h, err.Error()}, nil
		}

		return hashResponse{h, ""}, nil
	}
}

func MakeValidateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(validateRequest)
		v, err := svc.Validate(ctx, req.Password, req.Hash)
		if err != nil {
			return validateResponse{false, err.Error()}, nil
		}

		return validateResponse{v, ""}, nil
	}
}

type Endpoints struct {
	HashEndpoint     endpoint.Endpoint
	ValidateEndpoint endpoint.Endpoint
}

func (e Endpoints) Hash(ctx context.Context, password string) (string, error) {
	req := hashRequest{Password: password}
	resp, err := e.HashEndpoint(ctx, req)
	if err != nil {
		return "", err
	}
	hashResp := resp.(hashResponse)
	if hashResp.Err != "" {
		return "", errors.New(hashResp.Err)
	}
	return hashResp.Hash, nil
}

func (e Endpoints) Validate(ctx context.Context, password, hash string) (bool, error) {
	req := validateRequest{Password: password, Hash: hash}

	resp, err := e.ValidateEndpoint(ctx, req)
	if err != nil {
		return false, err
	}
	validResp := resp.(validateResponse)
	if validResp.Err != "" {
		return false, errors.New(validResp.Err)
	}
	return validResp.Valid, nil
}
