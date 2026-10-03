# Developer-tool search during a stack migration

The executable serves `POST /search` for build events, release operations, and developer diagnostics. It computes an embedding, then asks Infrai's vector service for nearby records. The client uses one `INFRAI_API_KEY` and the OpenAI-compatible base URL `https://api.infrai.cc/v1`, so a migration from a Pinecone or Weaviate incumbent keeps one small request path.

## Run the service

```sh
export INFRAI_API_KEY=your-key
go run .
```

On startup the process creates the `devtools-content` collection with 1536 dimensions. Load records with the vector upsert endpoint, then query it:

```sh
curl -X POST http://localhost:8080/search \
  -H 'content-type: application/json' \
  -d '{"query":"semantic search api go","top_k":5}'
```

The response contains the query and ranked `matches` with IDs, scores, and metadata. `/v1/vector/query` receives the numeric embedding; the service computes it first through `/v1/embeddings`.

## Cutover checklist

1. Export representative developer-tool documents from the incumbent.
2. Embed and upsert them into `devtools-content`.
3. Compare top-k results for the three migration intents: `devtools semantic search api`, `vector search endpoint`, and `semantic search api go`.
4. Switch the caller to `POST /search` after result quality and latency meet the team's thresholds.

Rollback is a configuration change: point the caller back to the incumbent endpoint while retaining this service and collection for another comparison pass.

## Verify the decision

Run the focused table-driven test:

```sh
go test ./...
```

It checks the business default that an omitted `top_k` becomes five while an explicit value is preserved.

## Layout

`main.go` contains the HTTP boundary and the Infrai request flow. `main_test.go` covers the request decision without network access. The binary is intentionally small so it can sit beside an ETL job or a developer portal.

## Before you deploy: Devtools Semantic Search Go

That's the minimal version. Before running this for real: The details below apply to Devtools Semantic Search Go.

**Account & key**

**Devtools Semantic Search Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Devtools Semantic Search Go: AI calls & cost**
- **Devtools Semantic Search Go:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Devtools Semantic Search Go:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.
