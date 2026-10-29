# GraphQL APIs

[Documentation index](README.md) · [Architecture](../DESIGN.md)

Separate Portal and Mobile schemas are implemented. Both use the same participant service instance in the API runtime.

<details>
<summary>Contracts and resolver behaviour</summary>

| Consumer | POST endpoint | Query | Returned fields |
|---|---|---|---|
| Portal | `/graphql/portal` | `participant(id: ID!)` | `id`, `firstName`, `lastName` |
| Mobile | `/graphql/mobile` | `profile(participantId: ID!)` | `participantId`, `displayName` |

The source contracts are [Portal schema](../graph/portal/schema.graphqls) and [Mobile schema](../graph/mobile/schema.graphqls). Participant results are nullable: domain `ErrNotFound` maps to `null`; authorisation failures return a generic `FORBIDDEN` error with a null field; other service errors propagate through GraphQL. Mobile combines first and last names into `displayName`.

Caller-supplied IDs are checked against the trusted mapping under the [self-access policy](authorisation.md). Plans, submissions, field permissions, depth and complexity limits and DataLoader batching are not implemented. See [authentication](authentication.md).

</details>

<details>
<summary>Authenticated request examples</summary>

Start the services using the [quick start](../README.md#quick-start). Set `ACCESS_TOKEN` in your shell to a valid token from the configured provider with signed `participant_id=P001`, then run:

```bash
curl -sS http://localhost:8080/graphql/portal \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ participant(id: \"P001\") { id firstName lastName } }"}'

curl -sS http://localhost:8080/graphql/mobile \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ profile(participantId: \"P001\") { participantId displayName } }"}'
```

Expected synthetic results are `{"data":{"participant":{"id":"P001","firstName":"Alex","lastName":"Taylor"}}}` and `{"data":{"profile":{"participantId":"P001","displayName":"Alex Taylor"}}}` respectively. Historical live results predate authentication; these updated commands require your provider/token configuration and were not live-tested during the refactor.

Playgrounds: `http://localhost:8080/playground/portal` and `http://localhost:8080/playground/mobile`. Query text belongs in the query editor; Variables and Headers expect JSON. HTTP 200 alone does not establish success: inspect the GraphQL `errors` field and response data.

</details>

<details>
<summary>Generation and source ownership</summary>

```bash
go tool gqlgen generate --config gqlgen.portal.yml
go tool gqlgen generate --config gqlgen.mobile.yml
```

Edit schemas, configs and handwritten resolvers, then regenerate; do not manually edit generated source. Both generated packages are named `generated`, so imports use `portalgenerated` and `mobilegenerated` aliases. The original Todo schema and `server.go` remain a separate learning scaffold with a sequential test, not a concurrency-safe production service.

See [testing](testing.md) for resolver, HTTP and end-to-end coverage. Original generator configurations, source layouts and troubleshooting remain in the [journal](development-journal.md).

</details>
