---
title: Migrating a billion rows without anyone noticing
slug: migrating-a-billion-rows-without-anyone-noticing
date: 2026-01-14
excerpt: The expand-contract pattern, but applied to a column type change on a hot table.
reading_time: 11 min
category: data migrations
---

The best large migration is the one that users do not notice and operators barely remember afterwards.

## The shape of the problem

Changing a hot table with a large amount of traffic is rarely about the SQL itself. It is about choreography:

- new code and old code overlapping
- two schemas being valid at the same time
- backfills that can run safely for days

## Expand first

Add what the future state needs before you remove what the past state depended on. That usually means new columns, dual writes, and a period where both representations are valid.

## Contract last

Only after the system is clearly reading from the new shape should you clean up the old one. Teams get into trouble when they treat deletion as the exciting part.

## A practical rule

If the migration plan looks elegant on a whiteboard but gives on-call engineers nothing to observe or pause, it is not done yet.
