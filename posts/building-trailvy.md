---
title: Building Travelvy.com
slug: building-trailvy
date: 2026-05-16
excerpt: A retrospective on building an early travel-planning product with Scala, Cassandra, Backbone.js, and CoffeeScript.
reading_time: 4 min
category: projects
---

About a decade ago, I co-founded Travelvy.com, a travel-planning website built around a curated collection of places to visit. It was my first serious attempt to build a product from the ground up.

The team consisted of two founders and one remote engineer.

Building the product involved more than software. We coordinated local photography and editorial content in Edinburgh while designing and implementing the platform itself.

The backend was Scala with Scalatra and Scalate.

The data layer was Cassandra. We chose it during the height of enthusiasm for NoSQL and built more complexity than the product needed at that stage.

On the frontend it was mostly jQuery, Backbone.js, and CoffeeScript. TypeScript was not really a thing yet.

Although I was primarily a backend engineer, our limited budget meant that I also designed and built the user interface. It pushed me well outside my usual role and gave me a much better appreciation for product and frontend work.

Content lived in YAML files that we edited manually and then uploaded to Cassandra.
Images were autoscaled using a Python library.

Later, we added Flickr for additional images and Tumblr for comments.

We shipped an MVP and used Google ads to test acquisition, eventually attracting about 1,000 visitors.

The central lesson was that we had started with implementation instead of market validation. Demand was limited, and established competitors were already serving the same space with substantially greater resources. A small acquisition test before development would have challenged our assumptions much earlier.

We eventually shut the product down, but the experience shaped how I think about scope, validation, and the difference between building software and building a business.

Some screenshots:

- [Screenshot 1](posts/content/trailvy_1.png)
- [Screenshot 2](posts/content/trailvy_2.png)
- [Screenshot 3](posts/content/trailvy_3.png)
