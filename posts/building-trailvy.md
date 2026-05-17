---
title: Building Travelvy.com
slug: building-trailvy
date: 2026-05-16
excerpt: We built this website for travel planning at some point. It was Scala, Cassandra, and some CoffeeScript.
reading_time: 4 min
category: projects
---

We built Travelvy.com about 10 years ago, when, like many aspiring entrepreneurs, I decided I wanted to build something big and went with the only idea that came to mind. It was a travel-planning website with a curated set of places to visit. It was not a particularly bright idea, to be honest, though in my defense I was not very far from most of my peers at the time.

There were three of us: two founders and one engineer we hired remotely.

It was an interesting journey: finding someone in Edinburgh to photograph the places, though it turned out he did not really know how to do it, another Brit to write proper descriptions of the places, and then building the whole thing.
There were also plenty of funny stories about dealing with other people along the way, the kind I am sure any founder could tell, but that probably deserves a separate post.

The backend was Scala with Scalatra and Scalate.

The data layer was Cassandra. NoSQL was very much the hype at the time, so of course we overengineered the system.

On the frontend it was mostly jQuery, Backbone.js, and CoffeeScript. TypeScript was not really a thing yet.

Funny thing: even though I was always primarily a backend engineer, I wrote all of the UI myself and was also the primary designer, because at the time, with our limited budget, we could not find anyone even remotely competent for that work. I did it to the best of my artistic abilities, which, to be frank, I totally lack. In the end I made it look remotely alike one of the trending sites.

Content lived in YAML files that we edited manually and then uploaded to Cassandra.
Images were autoscaled using a Python library.

Later we added a few external integrations in an attempt to make the website feel more alive: Flickr for extra images, since our own stock was rather limited, and Tumblr for comments.

We managed to build an MVP and even started advertising it on Google to get traffic. We got about 1,000 visitors in total eventually, and the internet still remembers that the site existed, though I could not find any actual content in web archives.

The nail in the coffin was the realization that people did not really need this that much, and there were already other companies doing more or less the same thing, with deeper pockets.

Building it, of course, was very naive. In the postmortem we agreed that the first step should have been proper market research, to see whether anyone needed it. Even a small ad campaign for a few hundred dollars would probably have saved us from undertaking the whole journey.

Eventually we shut it down, but it was a fun and interesting experience.

Some screenshots:

- [Screenshot 1](posts/content/trailvy_1.png)
- [Screenshot 2](posts/content/trailvy_2.png)
- [Screenshot 3](posts/content/trailvy_3.png)
