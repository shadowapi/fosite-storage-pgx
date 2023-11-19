package storage

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/ory/fosite"
)

// GetAuthorizeCodeSession stores the authorization request for a given authorization code.
func (p *PgStorage) CreateAuthorizeCodeSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	wr := &Request{
		Signature: signature,
		Active:    true,
	}
	wr.CastFromFosite(request)
	return p.dbCreateRequest(ctx, wr)
}

// GetAuthorizeCodeSession hydrates the session based on the given code and returns the authorization request.
// If the authorization code has been invalidated with `InvalidateAuthorizeCodeSession`, this
// method should return the ErrInvalidatedAuthorizeCode error.
//
// Make sure to also return the fosite.Requester value when returning the fosite.ErrInvalidatedAuthorizeCode error!
func (p *PgStorage) GetAuthorizeCodeSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	request, err = p.dbFindRequestBySignature(ctx, signature)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fosite.ErrNotFound
	}
	if err != nil {
		slog.Error("GetAuthorizeCodeSession", "dbFindRequestBySignature", err)
		return nil, err
	}

	if !request.(*Request).Active {
		return request, fosite.ErrInactiveToken
	}
	return
}

// InvalidateAuthorizeCodeSession is called when an authorize code is being used. The state of the authorization
// code should be set to invalid and consecutive requests to GetAuthorizeCodeSession should return the
// ErrInvalidatedAuthorizeCode error.
func (p *PgStorage) InvalidateAuthorizeCodeSession(ctx context.Context, signature string) (err error) {
	_, err = p.db.Conn().Exec(ctx, `UPDATE `+p.tablesPrefix+`request SET active = false WHERE signature = $1`, signature)
	return
}

// CreateAccessTokenSession stores an access token
func (p *PgStorage) CreateAccessTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	wr := &Request{
		Signature: p.hashAccessTokenSignature(signature),
		Active:    true,
	}
	wr.CastFromFosite(request)
	return p.dbCreateRequest(ctx, wr)
}

// GetAccessTokenSession returns an access token for a given signature
func (p *PgStorage) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	request, err = p.dbFindRequestBySignature(ctx, p.hashAccessTokenSignature(signature))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fosite.ErrNotFound
	}
	if err != nil {
		slog.Error("GetAccessTokenSession", "dbFindRequestBySignature", err)
		return nil, err
	}

	if !request.(*Request).Active {
		return request, fosite.ErrInactiveToken
	}
	return
}

// DeleteAccessTokenSession deletes an access token
func (p *PgStorage) DeleteAccessTokenSession(ctx context.Context, signature string) (err error) {
	_, err = p.db.Conn().Exec(ctx, `DELETE FROM `+p.tablesPrefix+`request WHERE signature = $1`, p.hashAccessTokenSignature(signature))
	return
}

func (p *PgStorage) CreateRefreshTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	wr := &Request{
		Signature: signature,
		Active:    true,
	}
	wr.CastFromFosite(request)
	return p.dbCreateRequest(ctx, wr)
}

func (p *PgStorage) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	request, err = p.dbFindRequestBySignature(ctx, signature)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fosite.ErrNotFound
	}
	if err != nil {
		slog.Error("GetRefreshTokenSession", "dbFindRequestBySignature", err)
		return nil, err
	}

	if !request.(*Request).Active {
		return request, fosite.ErrInactiveToken
	}
	return
}

func (p *PgStorage) DeleteRefreshTokenSession(ctx context.Context, signature string) (err error) {
	_, err = p.db.Conn().Exec(ctx, `DELETE FROM `+p.tablesPrefix+`request WHERE signature = $1`, signature)
	return
}

// hashAccessTokenSignature to limit it size
func (p *PgStorage) hashAccessTokenSignature(signature string) string {
	hash := sha512.Sum384([]byte(signature))
	return hex.EncodeToString(hash[:])
}
