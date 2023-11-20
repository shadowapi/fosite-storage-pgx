package storage

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizeCodeSession(t *testing.T) {
	_, err := conn.Exec(context.TODO(), `BEGIN;`)
	assert.NoError(t, err)
	defer conn.Exec(context.TODO(), `ROLLBACK;`)

	s := New(
		&connAsDB{conn},
		&mockClient{})
	err = s.MigrateUp(context.Background())
	assert.NoError(t, err)

	request := &fosite.Request{
		ID:             "id",
		RequestedAt:    time.Now().Round(time.Minute).UTC(),
		Client:         &fosite.DefaultClient{},
		RequestedScope: fosite.Arguments{"scope"},
		GrantedScope:   fosite.Arguments{"scope"},
		RequestedAudience: fosite.Arguments{
			"audience1",
			"audience2",
		},
		GrantedAudience: fosite.Arguments{
			"audience1",
			"audience2",
		},
		Form: url.Values{
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
	}

	t.Run("create authorize code session", func(t *testing.T) {
		err = s.CreateAuthorizeCodeSession(context.TODO(), "signCreateAuthorizeCodeSession", request)
		assert.NoError(t, err)
	})

	t.Run("get authorize code session", func(t *testing.T) {
		raw, err := s.GetAuthorizeCodeSession(context.TODO(), "signCreateAuthorizeCodeSession", &fosite.DefaultSession{})
		assert.NoError(t, err)

		schema, ok := raw.(*Request)
		assert.True(t, ok)

		assert.Equal(t, request.ID, schema.ID)
		assert.Equal(t, request.RequestedAt, schema.RequestedAt)
		assert.Equal(t, request.Client.GetID(), schema.Client.GetID())
		assert.Equal(t, request.RequestedScope, schema.RequestedScope)
		assert.Equal(t, request.GrantedScope, schema.GrantedScope)
		assert.Equal(t, request.RequestedAudience, schema.RequestedAudience)
		assert.Equal(t, request.GrantedAudience, schema.GrantedAudience)
		assert.Equal(t, request.Form, schema.Form)
		assert.Equal(t, request.Session, schema.Session)
		assert.Equal(t, true, schema.Active)
	})

	t.Run("invalidate authorize code session", func(t *testing.T) {
		err = s.InvalidateAuthorizeCodeSession(context.TODO(), "signCreateAuthorizeCodeSession")
		assert.NoError(t, err)

		db, err := s.dbFindRequestBySignature(context.Background(), "signCreateAuthorizeCodeSession")
		assert.NoError(t, err)

		assert.False(t, db.Active)
	})
}
