## fosite PostgreSQL storage

This library implements [fosite](https://github.com/ory/fosite) Storage
interface for PostgreSQL database.

## Tests

To test this library you need to create test database for it, or grant test user
create table permissions. After that you can run this command:

```bash
FS_STORE_PGX_URI='postgres://<USER>:<PASSWORD>@<HOST>:<PORT>/<NAME>' FS_PG_TEST_USER=oauth2_test go test ./...
```

Example of run test with PostgreSQL in the Docker:

```bash
docker run --name test-db -e POSTGRES_PASSWORD=fosite -e POSTGRES_DB=fosite_test -p 5432:5432 -d postgres
FS_STORE_PGX_URI='postgres://postgres:fosite@localhost:5432/fosite_test' go test ./...
```
