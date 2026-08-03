---
title: Thoughts on debugging
slug: thoughts-on-debugging
date: 2026-05-17
excerpt: Step-through debugging remains useful, but routine dependence on it can reveal deeper problems with test coverage, modularity, or system boundaries.
reading_time: 3 min
category: engineering
---

As LLMs take on a larger role in software development, one recurring concern is that engineers will be asked to maintain generated code they do not fully understand. Step-through debugging is often presented as the way to recover that understanding.

LLM-generated code can certainly increase complexity in large systems. Mature codebases rely on conventions that help engineers navigate the system and infer intent. Without sufficient developer oversight, generated code may disregard those conventions and increase the comprehension burden.

What I question is the assumption that step-through debugging should be the primary way to understand business logic. I rarely rely on it now. When I do, it is usually to confirm a specific mismatch between a call’s output and a unit test’s expectation, not to discover how the surrounding system works.

That change happened gradually. One factor was a development environment that was slow and difficult to run consistently. Starting the required services took time, failures were common, and avoidance of the local environment contributed to configuration drift. Those constraints pushed me toward fast, focused tests that covered expected behavior and edge cases without requiring the entire system to run.

The architecture reinforced that shift. Once a workflow crosses several services, stepping through one local process provides only a narrow view. Logs, traces, contract tests, integration tests, and production-like validation become more useful than following one thread of execution line by line.

Unit tests do not cover every failure mode in a distributed system, but neither does a local debugger. They are one part of a broader verification strategy.

Step-through debugging remains particularly useful when:

- You are working on a complex algorithm and debugging it in the context of a focused test.

- You are working with low-level memory behavior, pointers, or concurrency where runtime state is difficult to infer.

- You are investigating unfamiliar legacy code with limited tests and unclear boundaries.

Routine dependence on a debugger to understand ordinary business logic can point to broader structural problems: insufficient test coverage, weak modularity, hidden side effects, or unclear ownership of state. LLM-generated code can amplify those problems when it is accepted without careful review.

The useful opportunity is that AI-assisted development can reduce the cost of adding test coverage and exploring refactorings. That does not remove the need for engineering judgment or independent review, but it can make incremental improvement more practical. The goal is not to eliminate debugging; it is to build systems whose behavior can usually be understood and verified without depending on it.
