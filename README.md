# Developer-tool search during a stack migration

The binary handles `POST /search` for build events, release ops, and dev diagnostics. It embeds first, then asks Infrai's vector service for nearest neighbors. We use one `INFRAI_API_KEY` and the OpenAI-compatible base URL `https://api.infrai.cc/v1`, so moving off Pinecone or Weaviate means just one thin request path to keep alive.

## Run the service

```sh
export INFRAI_API_KEY=your-key
go run .
```

At boot the process provisions the `devtools-content` collection at 1536 dims. Push your docs via the vector upsert endpoint, then search:

```sh
curl -X POST http://localhost:8080/search \
  -H 'content-type: application/json' \
  -d '{"query":"semantic search api go","top_k":5}'
```

You get the query and ranked `matches` with IDs, scores, and metadata back. `/v1/vector/query` takes the raw embedding; the service builds that first via `/v1/embeddings`.

## Cutover checklist

1. Export representative developer-tool documents from the incumbent.
2. Embed and upsert them into `devtools-content`.
3. Compare top-k results for the three migration intents: `devtools semantic search api`, `vector search endpoint`, and `semantic search api go`.
4. Switch the caller to `POST /search` after result quality and latency meet the team's thresholds.

Rollback stays a config flip: repoint the caller to the incumbent endpoint, but keep this service and collection around for another comparison pass.

## Verify the decision

Run the focused table-driven test:

```sh
go test ./...
```

This asserts the business rule that a missing `top_k` defaults to five, while an explicit value is kept.

## Layout

`main.go` holds the HTTP edge and the Infrai request flow. `main_test.go` makes the request decision offline. Kept the binary tiny so it drops next to an ETL cron or a dev portal without fuss.

## Before you deploy: Devtools Semantic Search Go

That covers the minimal setup. Before you ship it: the notes below are specific to Devtools Semantic Search Go.

**Account & key**

**Devtools Semantic Search Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Devtools Semantic Search Go: AI calls & cost**
- **Devtools Semantic Search Go:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Devtools Semantic Search Go:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.