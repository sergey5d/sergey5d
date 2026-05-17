---
title: Thoughts on debugging
slug: thoughts-on-debugging
date: 2026-05-17
excerpt: Step-through debugging is still useful in some cases, but if it is your normal way of understanding business logic, that may be a sign the system is already in an unhealthy state.
reading_time: 3 min
category: engineering
---

In the past few months, there has been a flood of messages from SWEs trying to defend their stance that they are still relevant and their skills are needed in the post-LLM era, where the majority of the code is written by LLMs.

I won’t delve into the arguments of either AI denialists, who state that AI will collapse under the weight of the slop it produces, nor will I defend the side of AI pushers who overstate current LLM abilities.

In this note, I want to discuss a common and recurring theme people raise: that it is impossible to do step-through debugging with LLM-generated code that nobody understands.

First of all, I do not deny the complexity that LLM-generated code brings to large-system development. Most large code bases follow specific patterns that allow you to navigate through the system based on these anchors and comprehend intent and logic. When LLM-generated code breaks these often loosely defined conventions, it adds another layer of complexity and increases the comprehension burden.

What I would like to question here is the assumption that we still need to rely on step-through debugging.

It was initially quite surprising to see people seriously discussing the need for step-through debugging, and then I realized I had just forgotten the modus operandi that was valid for me about 5 years ago.

Why was I surprised? Because I rarely do step-through debugging now. And if I ever happen to do it, it is to verify what’s wrong by comparing the output of a specific call against what is expected in unit tests, not to figure out what is happening in the system code.

How did this change of paradigm - that I’m not doing step-through debugging anymore - happen?

That was due to a mixed set of reasons that shifted my mental paradigm. First, our dev environment was extremely hard to bring up, and every other time something was broken. That, of course, was just a reflection of the constraints of working in a small company with limited resources to make everything right. Thus, in order to bring the system into a runnable state, I had to spend time waiting until all required containers started, and then more time verifying what had gone wrong. This was a real drag on my time, and eventually I started to operate in a “cover every possible scenario, and if something seems broken, write even more unit tests” paradigm.

Another drag that made it unfeasible was just the nature of the system: when you have calls to different microservices in the code you are working on, step-by-step debugging becomes extremely time-consuming. So these two factors - limits imposed by our dev environment and the complexity of the system - pushed me to stop using step-by-step debugging as a tool to verify the correctness of my code.

Some readers may argue that unit tests may not cover all potential issues that might arise in distributed systems, and I wholeheartedly agree with this. But step-through debugging is practical mostly in local or development-like environments, so it won’t help much with those problems either. My previous points about the extensive use of unit tests do not imply that the system should only have them and nothing else.

In modern application development, you usually don’t need step-by-step debugging unless:

You are working on a specific complex algorithm, and you would debug it in the context of a running unit test.

You are working on some low-level code operating with memory allocation and pointers, i.e. in C and C++ like languages, with too many factors that can come into play and are too hard to foresee.

You are working with a codebase that is not properly unit-tested and is overburdened by convoluted logic that you cannot comprehend.

I think that today, when most languages do not require manual memory management and engineers rarely create new complex algorithms, what most people have in mind is the last item.

To conclude: if step-through debugging is your normal way of understanding business logic, that may be a sign that the system is already in an unhealthy state - insufficiently covered by tests, poorly modularized, or reliant on convoluted logic.
