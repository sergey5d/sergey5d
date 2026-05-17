---
title: A small theory of on-call
slug: a-small-theory-of-on-call
date: 2026-03-29
excerpt: Pages should teach you something. If they don't, the runbook is doing the wrong job.
reading_time: 5 min
category: operations
---

Good on-call is not just about waking the right person up. It is about teaching the system and the team something every time a page happens.

## A page should have a lesson inside it

If an alert fires and nobody learns anything from it, then it is not really doing the job. It is just background pain with a timestamp attached.

## Runbooks are not legal documents

The best runbooks make the first five minutes easier:

- what failed
- what usually causes it
- what to check first
- what is dangerous to do in a hurry

Everything else is secondary.

## The real goal

The point of an on-call system is not to prove that engineers are resilient. It is to reduce the number of surprising things that can happen in production, and shorten the time it takes to understand the ones that still do.
