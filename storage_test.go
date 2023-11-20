package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
)

var conn *pgx.Conn

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set up database connection
	dbURL := os.Getenv("FS_STORE_PGX_URI") // Ensure this points to your test database

	cfg, err := pgx.ParseConfig(dbURL)
	assert.NoError(nil, err)

	conn, err = pgx.ConnectConfig(ctx, cfg)
	assert.NoError(nil, err)

	// Run tests
	code := m.Run()

	// Teardown
	err = conn.Close(ctx)
	assert.NoError(nil, err)

	os.Exit(code)
}

type connAsDB struct {
	conn *pgx.Conn
}

func (m *connAsDB) Conn(_ context.Context) (*pgx.Conn, error) {
	return m.conn, nil
}

type mockClient struct{}

func (m *mockClient) GetClientByID(ctx context.Context, id string) (fosite.Client, error) {
	return &fosite.DefaultClient{}, nil
}

func TestCreateStorageWithOptions(t *testing.T) {
	storage := New(
		&connAsDB{conn},
		&mockClient{},
		WithMigrationTableName("auth_test_fosite_migrations"),
		WithTablesPrefix("auth_test_fosite_"),
	)
	assert.Equal(t, "auth_test_fosite_migrations", storage.migrationTableName)
	assert.Equal(t, "auth_test_fosite_", storage.tablesPrefix)
}

func TestRequestToSchema(t *testing.T) {
	_, err := conn.Exec(context.TODO(), `BEGIN;`)
	assert.NoError(t, err)
	defer conn.Exec(context.TODO(), `ROLLBACK;`)

	request := &Request{
		Request: fosite.Request{
			ID:          "id",
			RequestedAt: time.Now(),
			Client:      &fosite.DefaultClient{},
			RequestedScope: []string{
				"scope1",
				"scope2",
			},
			GrantedScope: []string{
				"scope1",
				"scope2",
			},
			RequestedAudience: []string{
				"audience1",
				"audience2",
			},
			GrantedAudience: []string{
				"audience1",
				"audience2",
			},
			Form: map[string][]string{
				"form1": {
					"form1",
					"form2",
				},
				"form2": {
					"form1",
					"form2",
				},
			},
			Session: &fosite.DefaultSession{},
		},
		Signature: "signature",
		Active:    true,
	}
	s := New(&connAsDB{conn}, &mockClient{})
	err = s.MigrateUp(context.TODO())
	assert.NoError(t, err)
	schema, err := s.requestToSchema(context.TODO(), request)
	assert.NoError(t, err)
	assert.Equal(t, request.ID, schema["id"])
	assert.Equal(t, request.Signature, schema["signature"])
	assert.Equal(t, request.RequestedAt, schema["requested_at"])
	assert.Equal(t, request.Client.GetID(), schema["client_id"])
	assert.Equal(t, request.RequestedScope, schema["requested_scope"])
	assert.Equal(t, request.GrantedScope, schema["granted_scope"])
	assert.Equal(t, request.RequestedAudience, schema["requested_audience"])
	assert.Equal(t, request.GrantedAudience, schema["granted_audience"])
	assert.Equal(t, sql.NullString{String: request.Form.Encode(), Valid: true}, schema["form"])
	session, err := json.Marshal(request.Session)
	assert.NoError(t, err)
	assert.Equal(t, session, schema["session"])
	assert.Equal(t, request.Active, schema["active"])
}

func TestSchemaToRequest(t *testing.T) {
	_, err := conn.Exec(context.TODO(), `BEGIN;`)
	assert.NoError(t, err)
	defer conn.Exec(context.TODO(), `ROLLBACK;`)

	request := &Request{
		Request: fosite.Request{
			ID:          "id",
			RequestedAt: time.Now().UTC().Round(time.Minute),
			Client:      &fosite.DefaultClient{},
			RequestedScope: []string{
				"scope1",
				"scope2",
			},
			GrantedScope: []string{
				"scope1",
				"scope2",
			},
			RequestedAudience: []string{
				"audience1",
				"audience2",
			},
			GrantedAudience: []string{
				"audience1",
				"audience2",
			},
			Form: map[string][]string{
				"form1": {
					"form1",
					"form2",
				},
				"form2": {
					"form1",
					"form2",
				},
			},
			Session: &fosite.DefaultSession{},
		},
		Signature: "signature",
		Active:    true,
	}
	s := New(&connAsDB{conn}, &mockClient{})
	err = s.MigrateUp(context.TODO())
	assert.NoError(t, err)

	err = s.dbCreateRequest(context.TODO(), request)
	assert.NoError(t, err)

	schema, err := s.dbFindRequestBySignature(context.TODO(), request.Signature)
	assert.NoError(t, err)

	assert.Equal(t, request.ID, schema.ID)
	assert.Equal(t, request.Signature, schema.Signature)
	assert.Equal(t, request.RequestedAt, schema.RequestedAt)
	assert.Equal(t, request.Client.GetID(), schema.Client.GetID())
	assert.Equal(t, request.RequestedScope, schema.RequestedScope)
	assert.Equal(t, request.GrantedScope, schema.GrantedScope)
	assert.Equal(t, request.RequestedAudience, schema.RequestedAudience)
	assert.Equal(t, request.GrantedAudience, schema.GrantedAudience)
	assert.Equal(t, request.Form.Encode(), schema.Form.Encode())
	assert.Equal(t, request.Session, schema.Session)
	assert.Equal(t, request.Active, schema.Active)
}
