---
title: Thoughts on debugging
slug: thoughts-on-debugging
date: 2026-05-17
excerpt: Unless you are writing low-level code that works directly with memory, there is usually very little reason to "debug" a system in the old-school sense.
reading_time: 3 min
category: engineering
---

Unless you are writing a low-level language, or working in a part of the system that deals directly with memory, there is usually very little chance that you actually need to debug it in the old-school sense.

Most of the time the problem is not that the machine is doing something mysterious. The problem is that we do not understand the system well enough yet.

In ordinary backend work, most bugs give themselves away if you do the boring things properly:

- read the code carefully
- check the logs
- verify inputs and outputs
- trace the data flow
- confirm your assumptions against reality

That is not really debugging in the classic sense. It is closer to reading, observation, and refusing to guess too early.

Of course there are exceptions. If you are dealing with memory corruption, undefined behavior, concurrency at a low level, or something close to the metal, then yes, real debugging tools matter a lot.

But for most application code, reaching for a debugger too early is often just a way to avoid thinking clearly about what the code is supposed to do.

## Working draft

so I'm seeing a lot of posts on X that oh AI vibeslopped tons of code and now you have to debug it. I'm generally suprised that people still debug the code. in my past 4 years I did it super rarealy, am I outlier or this is just indicates that their systems are not unit test covered enough?

Thought for a couple of seconds

You are probably not an outlier, but you may be using a narrower meaning of “debug.”

In mature backend systems, especially ones with decent unit/integration tests, good logs, reproducible local/dev environments, and strong type systems, you often do not sit there stepping through code in a debugger. Most “debugging” becomes:

“Test failed → inspect assertion/logs → reason about code → fix → rerun tests.”

That is normal for experienced backend engineers.

Where people say “AI generated a ton of code and now I have to debug it,” they may mean several different things:

Literal debugger debugging
Stepping through generated code because they do not understand what it does. This often happens when they accepted too much AI output without reviewing architecture or invariants.

Trying to make untested code work
The AI produced plausible code, but there are no strong tests, so the person manually runs the app, clicks around, hits bugs, adds patches, repeats. That is basically “vibe QA.”

Integration debugging
Unit tests may pass, but the generated code misunderstood API contracts, DB schema, auth flow, async behavior, transaction boundaries, race conditions, nullability, etc. These are harder to catch with shallow unit tests.

Semantic debugging
The code compiles and tests pass, but it solves the wrong problem. This is very common with LLM code because it optimizes for plausible implementation, not necessarily the actual business invariant.

So yes, a lot of these complaints probably indicate weak test coverage, but not only that. It also indicates weak engineering process around AI-generated code.

The healthy AI-assisted workflow is closer to:

- define behavior/invariants
- write or ask AI to write tests
- review tests carefully
- generate implementation
- run tests
- review diff like a human wrote it
- tighten edge cases

The unhealthy workflow is:

- ask AI to build feature
- paste huge diff
- app kind of runs
- now hunt bugs manually

That second one creates the “vibeslop” problem.

Your experience probably comes from working in systems where bugs were caught by a combination of:

- types
- compiler
- tests
- logs
- metrics
- code review
- production safeguards
