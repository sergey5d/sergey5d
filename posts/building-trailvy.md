---
title: Building Trailvy
slug: building-trailvy
date: 2026-05-16
excerpt: I built this website for trail planning at some point. It was Scala, Cassandra, and some JavaScript.
reading_time: 4 min
category: projects
---

I built Trailvy at some point as a website for travel planning. The idea was that when people travel, provide them website with curated set of places to visit.

It was 3 of us (2 founder and one engineer we hired remotely).

It was interesting journey - to find person in Edinborough to make photos ofthe places, another Brit to create proper desriptions and to develop it.

The backend was Scala 2 with an SBT build, Scalatra for HTTP endpoints, Jetty as the embedded web server, Scalate for server-side templates, and Guice with scala-guice for dependency injection.
Lift JSON handled JSON, Dispatch was there for HTTP client calls, Scalaz showed up in core and service code, JSoup handled HTML parsing and cleanup, Ehcache was used for caching, and Akka was present for some async and background work.

The data layer was Cassandra (NO SQL was a hype at that time and of course we overengineered the system).

On the frontend it was jQuery, Backbone.js, Underscore.js, CoffeeScript compiled to JavaScript, Mustache templates in some places, and Masonry for layout. JavaScript tests ran through Karma and PhantomJS.

Funny thing, that though I always was primary backend engineer, I wrote all of the UI myself).

Content came from YAML-based POI files, with a custom `poi-builder` importer loading them into Cassandra.
There was also a custom Python plus ImageMagick image pipeline to scale images.

It also had a few external integrations added later in attempt to make website alive: Tumblr and Flickr for image loading,
Amazon S3 for original image hosting and CDN duties, and email sending through AWS Simple Email Server.

We managed to build MVP and even started to advertizing it on google.com to gain traffic. The funny thing was that we had about 1k visitors though most of them did not stay on the site for long.

The coffin to the project was realization that people don't really need this much and there were some other companies doing exactly the same as we did.

That was of course very naive and post mortem we discussed that first step should have been a thorough market research instead of building it and seing if somebody would need it.

Eventually we shut it down, but that was fun and interesting experience.

Some screenshots from that version:

- [Screenshot 1](posts/content/trailvy_1.png)
- [Screenshot 2](posts/content/trailvy_2.png)
- [Screenshot 3](posts/content/trailvy_3.png)
