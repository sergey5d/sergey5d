---
title: Thoughts on debugging
slug: thoughts-on-debugging
date: 2026-05-17
excerpt: Step-through debugging is still useful in some cases, but if it is your normal way of understanding business logic, that may be a sign the system is already in an unhealthy state.
reading_time: 3 min
category: engineering
---

In the past few months, there has been a constant stream of messages from software engineers arguing that they are still relevant and that their skills remain necessary in the post-LLM era, where the majority of the code is written by LLMs.

I will not get into the arguments of either AI skeptics, who believe the current wave will collapse under the weight of low-quality output, or AI advocates, who often overstate the present capabilities of LLMs.

In this note, I want to discuss a common and recurring theme people raise: that it is impossible to do step-through debugging with LLM-generated code that nobody understands.

First of all, I do not deny the complexity that LLM-generated code brings to large-system development. Most large code bases follow specific patterns that allow you to navigate through the system based on these anchors and comprehend intent and logic. When developers do not exercise sufficient oversight over LLM-generated code, it may disregard these often loosely defined conventions and add another layer of complexity, increasing the comprehension burden.

What I would like to question here is the assumption that we still need to rely on step-through debugging.

It was initially quite surprising to see people seriously discussing the need for step-through debugging, and then I realized I had just forgotten the modus operandi that was valid for me about 5 years ago.

Why was I surprised? Because I rarely rely on step-through debugging now. And when I do use it, it is usually to confirm a specific mismatch between the output of a call and what a unit test expects, rather than to understand what is happening in the system code.

How did this change of paradigm - that I’m not doing step-through debugging anymore - happen?

That was due to a mixed set of reasons that shifted my mental paradigm. First, running our development environment came with real challenges. It was rough, slow to bring up, and some engineers avoided it altogether, which often led to configuration drift. That was, of course, mostly a reflection of the constraints of working in a small company with limited resources. Thus, in order to bring the system into a runnable state, I had to spend time waiting until all required containers started, and then more time verifying what had gone wrong. This was a real drag on my time, and eventually I started to operate in a “cover every possible scenario, and if something seems broken, write even more unit tests” paradigm.

Another thing that made it unfeasible was just the nature of the system: when you have calls to different microservices in the code you are working on, step-by-step debugging becomes useless. 

So these two factors - limits imposed by our dev environment and the complexity of the system - pushed me to stop using step-by-step debugging as a tool to verify the correctness of the code.

Some readers may argue that unit tests may not cover all potential issues that might arise in distributed systems, and I totally agree with this. But step-through debugging is practical mostly in local or development-like environments, so it won’t help much with those problems either. My previous points about the extensive use of unit tests do not imply that the system should only have them and nothing else.

In application development, you usually don’t need step-by-step debugging unless:

- You are working on a specific complex algorithm, and you would debug it in the context of a running unit test.

- You are working on some low-level code operating with memory allocation and pointers, i.e. in C and C++ like languages, with too many factors that can come into play and are too hard to foresee.

- You are working with a codebase that is not properly unit-tested and is overburdened by convoluted logic that you cannot comprehend.

I think that today, when most languages do not require manual memory management and engineers rarely create new complex algorithms, what most people have in mind is the last item.

And here comes the trouble: of course, if you are relying on step-through debugging, LLM-generated code will make your life extremely hard. But your system is already in an unhealthy state. It either has insufficient test coverage, is not properly modularized, has side effects here and there, has problems with logic, or, most likely, all of these.

These arguments suggest to me that you may already be approaching the boundaries of what the current system can support, and that pushing beyond them could start to break it. That may be true, but it does not have to be.

On the bright side, adding test coverage has become far less expensive. By extension, once your system is properly covered, you can move in the right direction by introducing proper abstractions and refactoring code, which would have been much harder to do prior to LLM code generation.
