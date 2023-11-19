package storage

import (
	"github.com/ory/fosite"
)

const (
	// AuthTypeAccessToken is a type of access token
	AuthTypeAccessToken AuthType = "access_token"
)

type (
	AuthType string
	// Request is a fosite request with an additional fields
	Request struct {
		fosite.Request
		Type      AuthType `json:"type"`
		Signature string   `json:"signature"`
		Active    bool     `json:"active"`
	}
)

func (r *Request) CastFromFosite(request fosite.Requester) {
	if v, ok := request.(*fosite.Request); ok {
		r.Request = *v
		return
	}
	r.Request.ID = request.GetID()
	r.Request.RequestedAt = request.GetRequestedAt()
	r.Request.Client = request.GetClient()
	r.Request.Form = request.GetRequestForm()
	r.Request.Session = request.GetSession()
	r.Request.GrantedScope = request.GetGrantedScopes()
	r.Request.RequestedScope = request.GetRequestedScopes()
	r.Request.RequestedAudience = request.GetRequestedAudience()
	r.Request.GrantedAudience = request.GetGrantedAudience()
}
