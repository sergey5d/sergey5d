---
title: Notes on idempotency keys that won't haunt you
slug: notes-on-idempotency-keys
date: 2026-05-14
excerpt: Most retry bugs come from idempotency that's almost-but-not-quite right. A short field guide drawn from four years of payments work.
reading_time: 9 min
category: distributed systems
---

Most retry bugs in payments do not come from missing idempotency. They come from idempotency that is *almost* right — the kind that survives the obvious retries and fails the interesting ones. Here are four patterns I have watched teams reinvent, and one that is actually worth keeping.

## 1. The request-body-as-key trap

Hashing the request body and using that as the idempotency key looks elegant for a while. Then a client adds an extra field for internal logging, the hash changes, and suddenly the same business action no longer looks the same. Idempotency keys should be *client-owned*. The server’s job is to honor them, not invent them.

## 2. Storing only the key

The first part of an idempotency layer is asking whether you have seen a key before. The second part is remembering what you told the client the first time. If you store only the key and not the response, your retry path behaves differently from the original path, which defeats most of the value.

```sql
CREATE TABLE idempotency (
  key            text PRIMARY KEY,
  request_hash   bytea NOT NULL,
  response_body  jsonb NOT NULL,
  response_code  smallint NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),
  expires_at     timestamptz NOT NULL
);
```

## 3. Reusing the same key for a different payload

When the same idempotency key arrives with a different request body, the right answer is not to shrug and return the old response. It is to reject the request clearly. If two requests claim to be the same retry but are not actually the same request, the system should refuse to guess.

> If two requests claim to be the same retry but are not, one of them is lying.

## 4. Getting the expiry window wrong

Teams often choose either an expiry window that is too short or one that is absurdly long. The useful number is neither arbitrary nor aesthetic. It should match your longest legitimate retry path, plus some safety margin.

## The one pattern worth keeping

Keep the idempotency layer thin. It should be a small table, a payload check, and a cached response. As soon as it starts accumulating business logic, it becomes the kind of component nobody wants to touch and everybody is afraid to migrate.
