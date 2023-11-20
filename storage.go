package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/ory/fosite"
)

type (
	// option to configure storage
	option func(*PgStorage)

	// DB interface
	DB interface {
		Conn(context.Context) (*pgx.Conn, error)
	}

	// Client interface
	Client interface {
		GetClientByID(ctx context.Context, id string) (fosite.Client, error)
	}

	// PgStorage is a fosite storage implementation for PostgreSQL
	PgStorage struct {
		db                 DB
		client             Client
		tablesPrefix       string
		migrationTableName string
	}
)

// New storage instance constructor
//
// The storage instance can be configured with options:
// Example:
//
//	storage := storage.New(db, storage.WithTablesPrefix("auth_fosite_"))
func New(db DB, client Client, options ...option) *PgStorage {
	p := &PgStorage{
		db:                 db,
		client:             client,
		tablesPrefix:       "auth_fosite_",
		migrationTableName: "public.auth_fosite_migrations",
	}
	for _, option := range options {
		option(p)
	}
	return p
}

func (p *PgStorage) schemaToRequest(ctx context.Context, row pgx.Row) (request *Request, err error) {
	request = &Request{}
	var (
		clientID  string
		session   fosite.DefaultSession
		urlValues sql.NullString
	)
	err = row.Scan(
		&request.ID,
		&request.Type,
		&request.Signature,
		&clientID,
		&request.RequestedAt,
		&request.RequestedScope,
		&request.GrantedScope,
		&request.RequestedAudience,
		&request.GrantedAudience,
		&urlValues,
		&session,
		&request.Active,
	)
	if err != nil {
		return nil, err
	}
	request.Session = &session

	if urlValues.Valid {
		request.Form, err = url.ParseQuery(urlValues.String)
		if err != nil {
			return nil, err
		}
	}

	request.Client, err = p.client.GetClientByID(ctx, clientID)
	return
}

func (p *PgStorage) requestToSchema(ctx context.Context, request *Request) (args pgx.NamedArgs, err error) {
	form := sql.NullString{}
	if v := request.GetRequestForm(); v != nil {
		form.String = v.Encode()
		form.Valid = true
	}

	session, err := json.Marshal(request.GetSession())
	if err != nil {
		return nil, err
	}

	args = pgx.NamedArgs{
		"id":                 request.GetID(),
		"type":               request.Type,
		"signature":          request.Signature,
		"client_id":          request.GetClient().GetID(),
		"requested_at":       request.GetRequestedAt(),
		"requested_scope":    request.GetRequestedScopes(),
		"granted_scope":      request.GetGrantedScopes(),
		"requested_audience": request.GetRequestedAudience(),
		"granted_audience":   request.GetGrantedAudience(),
		"form":               form,
		"session":            session,
		"active":             request.Active,
	}
	return
}

func (p PgStorage) dbCreateRequest(ctx context.Context, request *Request) (err error) {
	args, err := p.requestToSchema(ctx, request)
	if err != nil {
		return err
	}

	conn, err := p.db.Conn(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, `
		INSERT INTO `+p.tablesPrefix+`request (
			id,
			"type",
			signature,
			client_id,
			requested_at,
			requested_scope,
			granted_scope,
			requested_audience,
			granted_audience,
			form,
			session,
			active
		) VALUES (
			@id,
			@type,
			@signature,
			@client_id,
			@requested_at,
			@requested_scope,
			@granted_scope,
			@requested_audience,
			@granted_audience,
			@form,
			@session,
			@active
		)
	`, args)
	return err
}

func (p PgStorage) dbFindRequestBySignature(ctx context.Context, signature string) (request *Request, err error) {
	conn, err := p.db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	row := conn.QueryRow(ctx, `
		SELECT
			id,
			"type",
			signature,
			client_id,
			requested_at,
			requested_scope,
			granted_scope,
			requested_audience,
			granted_audience,
			form,
			session,
			active
		FROM `+p.tablesPrefix+`request
		WHERE signature = @signature
		`, pgx.NamedArgs{"signature": signature})
	return p.schemaToRequest(ctx, row)
}
