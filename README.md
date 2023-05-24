# Bytebase Forward Auth

A simple service that create [bytebase](https://www.bytebase.com/) login session (HTTP cookies) based on request header.

## Headers

- `X-User-Email`: the email of user. Should be unique.
- `X-User-Name`: the username of user. Should be unique.
- `X-User-Role`: the role assigned to the user. If multiple (comma separated) values are given, than it will use only the first one. If `BYTEBASE_FORWARDAUTH_GROUP_PREFIX` environment variable is specified, then only prefixed value are choosen and the actual role name are extracted by removing the prefix, e.g. when prefix is `bytebase/`, then `bytebase/owner` become `owner`.

## Environment Variables

- `BYTEBASE_FORWARDAUTH_PG_URL`: the url to external postgres database. This should be the same as [`PG_URL`](https://www.bytebase.com/docs/get-started/install/external-postgres/#:~:text=pg%20or%20pass-,PG_URL,-environment%20variable%20to) used by bytebase.
- `BYTEBASE_FORWARDAUTH_LISTEN_ADDRESS`: the listen address, default `:8080`.
- `BYTEBASE_FORWARDAUTH_CREATOR_ID`:  the existing user ID used for creating new user, default `101`.
- `BYTEBASE_FORWARDAUTH_GROUP_PREFIX`: as noted above.

## Example

See [docker-compose.yml](./docker-compose.yml) to integrate with [caddy security plugin](https://authp.github.io/) for enabling OAuth2 based SSO.
