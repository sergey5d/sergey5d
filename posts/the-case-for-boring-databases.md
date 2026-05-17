---
title: The case for boring databases
slug: the-case-for-boring-databases
date: 2026-04-18
excerpt: Why "just use Postgres" is rarely the wrong answer, and the three specific times it actually is.
reading_time: 7 min
category: databases
---

Most teams do not need a clever database choice. They need a database their engineers understand, their tooling supports, and their on-call rotation can survive at 3am. For a large category of systems, that answer is still Postgres.

## What boring buys you

- A query model people already know.
- Mature backup, migration, and monitoring tooling.
- Fewer architectural debates dressed up as innovation.

That does not make it glamorous, but it does make it reliable.

## Where people talk themselves into complexity

The usual pattern is familiar: the team sees scale in the future, overfits to that imagined future, and pays the complexity bill immediately. Many systems end up spending years carrying distributed-systems complexity they never actually needed.

## The times boring really is wrong

- When you need write patterns the system genuinely cannot support.
- When your consistency model is fundamentally different from what a relational store wants to provide.
- When the data model is not relational in any honest sense, and every query becomes an act of translation.

Those cases exist. They are just rarer than people like to believe.

## My bias

I like systems that are easy to reason about and cheap to operate. “Boring” is often shorthand for both.
