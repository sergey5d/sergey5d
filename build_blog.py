#!/usr/bin/env python3

from __future__ import annotations

import html
import re
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Iterable


ROOT = Path(__file__).resolve().parent
POSTS_DIR = ROOT / "posts"
BLOG_INDEX = ROOT / "blog.html"
HOME_PAGE = ROOT / "index.html"

SITE_TITLE = "Sergey Denisov"
SITE_HANDLE = ""
COLLOPHON = "© 2026 · built across OpenAI and Anthropic datacenters"
FOOTER_LINKS = [
    ('https://github.com/sergey5d', 'github'),
    ('https://www.linkedin.com/in/sadenisov/', 'linkedin'),
    ('contact.html', 'more →'),
]


@dataclass
class Post:
    slug: str
    title: str
    date: datetime
    excerpt: str
    reading_time: str
    category: str
    body: str

    @property
    def date_display(self) -> str:
        return self.date.strftime("%d %b %Y")

    @property
    def file_name(self) -> str:
        return f"{self.slug}.html"


def parse_frontmatter(text: str) -> tuple[dict[str, str], str]:
    if not text.startswith("---\n"):
        raise ValueError("Markdown file is missing frontmatter")

    _, rest = text.split("---\n", 1)
    frontmatter, body = rest.split("\n---\n", 1)

    metadata: dict[str, str] = {}
    for raw_line in frontmatter.splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        key, value = line.split(":", 1)
        metadata[key.strip()] = value.strip().strip('"')

    return metadata, body.lstrip()


def render_inline(text: str) -> str:
    escaped = html.escape(text)
    escaped = re.sub(r"`([^`]+)`", lambda m: f"<code>{m.group(1)}</code>", escaped)
    escaped = re.sub(r"\[([^\]]+)\]\(([^)]+)\)", lambda m: f'<a href="{html.escape(m.group(2), quote=True)}">{m.group(1)}</a>', escaped)
    escaped = re.sub(r"\*\*([^*]+)\*\*", r"<strong>\1</strong>", escaped)
    escaped = re.sub(r"\*([^*]+)\*", r"<em>\1</em>", escaped)
    return escaped


def render_markdown(markdown: str) -> str:
    lines = markdown.splitlines()
    blocks: list[str] = []
    paragraph: list[str] = []
    list_items: list[str] = []
    quote_lines: list[str] = []
    code_lines: list[str] = []
    in_code = False

    def flush_paragraph() -> None:
        nonlocal paragraph
        if paragraph:
            blocks.append(f"<p>{render_inline(' '.join(paragraph).strip())}</p>")
            paragraph = []

    def flush_list() -> None:
        nonlocal list_items
        if list_items:
            items = "".join(f"<li>{render_inline(item)}</li>" for item in list_items)
            blocks.append(f"<ul>{items}</ul>")
            list_items = []

    def flush_quote() -> None:
        nonlocal quote_lines
        if quote_lines:
            blocks.append(f"<blockquote>{render_inline(' '.join(quote_lines).strip())}</blockquote>")
            quote_lines = []

    def flush_code() -> None:
        nonlocal code_lines
        if code_lines:
            code = html.escape("\n".join(code_lines))
            blocks.append(f"<pre><code>{code}</code></pre>")
            code_lines = []

    for raw_line in lines:
        line = raw_line.rstrip("\n")
        stripped = line.strip()

        if stripped.startswith("```"):
            flush_paragraph()
            flush_list()
            flush_quote()
            if in_code:
                flush_code()
                in_code = False
            else:
                in_code = True
            continue

        if in_code:
            code_lines.append(line)
            continue

        if not stripped:
            flush_paragraph()
            flush_list()
            flush_quote()
            continue

        if stripped.startswith("## "):
            flush_paragraph()
            flush_list()
            flush_quote()
            blocks.append(f"<h2>{render_inline(stripped[3:])}</h2>")
            continue

        if stripped.startswith("> "):
            flush_paragraph()
            flush_list()
            quote_lines.append(stripped[2:])
            continue

        if stripped.startswith("- "):
            flush_paragraph()
            flush_quote()
            list_items.append(stripped[2:])
            continue

        paragraph.append(stripped)

    flush_paragraph()
    flush_list()
    flush_quote()
    if in_code:
        flush_code()

    return "\n\n      ".join(blocks)


def load_posts() -> list[Post]:
    posts: list[Post] = []
    for path in sorted(POSTS_DIR.glob("*.md")):
        metadata, body = parse_frontmatter(path.read_text())
        posts.append(
            Post(
                slug=metadata["slug"],
                title=metadata["title"],
                date=datetime.strptime(metadata["date"], "%Y-%m-%d"),
                excerpt=metadata["excerpt"],
                reading_time=metadata["reading_time"],
                category=metadata["category"],
                body=render_markdown(body),
            )
        )
    return sorted(posts, key=lambda post: post.date, reverse=True)


def footer_html(back_href: str) -> str:
    links = []
    for href, label in FOOTER_LINKS:
        actual_href = back_href if label == "more →" else href
        target = ' target="_blank" rel="noreferrer"' if actual_href.startswith("http") else ""
        links.append(f'      <a href="{actual_href}"{target}>{label}</a>')
    return "\n".join(links)


def page_shell(title: str, main_label: str, main_content: str, active_nav: str) -> str:
    nav = {
        "Home": "",
        "About": "",
        "Experience": "",
        "Blog": "",
        "Contact": "",
    }
    nav[active_nav] = ' class="is-active"'
    return f"""<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>{title}</title>
  <link rel="icon" type="image/png" href="icon.png" />
  <link rel="stylesheet" href="site.css" />
  <script src="https://unpkg.com/lucide@latest/dist/umd/lucide.min.js"></script>
</head>
<body>

<header class="site-header" data-screen-label="Header">
  <div class="shell">
    <a href="index.html" class="wordmark" aria-label="Home">
      <span class="dot"></span>
      <span>{SITE_TITLE}</span>
    </a>
    <nav class="nav" aria-label="Primary">
      <a href="index.html"{nav["Home"]}>Home</a>
      <a href="projects.html"{nav["Experience"]}>Experience</a>
      <a href="blog.html"{nav["Blog"]}>Blog</a>
      <a href="about.html"{nav["About"]}>About</a>
      <a href="contact.html"{nav["Contact"]}>Contact</a>
    </nav>
  </div>
</header>

<main{main_label}>
  <div class="shell">
{main_content}
  </div>
</main>

<footer class="site-footer">
  <div class="shell">
    <span class="colophon">{COLLOPHON}</span>
    <div class="links">
{footer_html("contact.html")}
    </div>
  </div>
</footer>

<script>lucide.createIcons();</script>
</body>
</html>
"""


def render_post_page(post: Post) -> str:
    content = f"""
    <a href="blog.html" class="back-link"><i data-lucide="arrow-left"></i> all posts</a>

    <h1 class="title">{html.escape(post.title)}</h1>

    <div class="post-meta">
      <span class="tnum">{post.date_display}</span>
      <span>·</span>
      <span>{html.escape(post.reading_time)}</span>
      <span>·</span>
      <span>{html.escape(post.category)}</span>
    </div>

    <div class="prose">
      {post.body}
    </div>
"""
    return page_shell(
        title=f"{post.title} — {SITE_TITLE}",
        main_label=' class="post" data-screen-label="Blog post"',
        main_content=content,
        active_nav="Blog",
    )


def group_by_year(posts: Iterable[Post]) -> list[tuple[int, list[Post]]]:
    grouped: dict[int, list[Post]] = {}
    for post in posts:
        grouped.setdefault(post.date.year, []).append(post)
    return sorted(grouped.items(), key=lambda item: item[0], reverse=True)


def render_blog_index(posts: list[Post]) -> str:
    sections: list[str] = []
    year_groups = group_by_year(posts)
    for idx, (year, items) in enumerate(year_groups):
        title = "Recent" if idx == 0 else "Earlier"
        extra_class = " blog-list-section" if idx == 0 else ""
        list_items = "\n".join(
            f"""        <li>
          <span class="when tnum">{post.date_display}</span>
          <div class="what">
            <h3><a href="{post.file_name}">{html.escape(post.title)}</a></h3>
            <span class="excerpt">{html.escape(post.excerpt)}</span>
          </div>
          <span class="kind">{html.escape(post.reading_time)}</span>
        </li>"""
            for post in items
        )
        sections.append(
            f"""
    <section class="section{extra_class}">
      <div class="section-head">
        <span class="label">// {year}</span>
        <h2>{title}</h2>
        <span class="count tnum">{len(items)} posts</span>
      </div>

      <ul class="log">
{list_items}
      </ul>
    </section>"""
        )

    content = f"""
    <div class="page-stack">
      <div class="page-head">
        <div class="eyebrow">// writing</div>
        <h1 class="title">Notes on anything of substance or lack thereof</h1>
        <p class="lede">Random thoughts about software, systems, AI, and everything else.</p>
      </div>
{''.join(sections)}
    </div>
"""

    return page_shell(
        title=f"Blog — {SITE_TITLE}",
        main_label=' data-screen-label="Blog"',
        main_content=content,
        active_nav="Blog",
    )


def update_home_links(posts: list[Post]) -> None:
    if not HOME_PAGE.exists():
        return

    content = HOME_PAGE.read_text()
    replacements = {
        "href=\"blog.html\" class=\"title\">Notes on idempotency keys that won't haunt you</a>":
            f'href="{posts[0].file_name}" class="title">Notes on idempotency keys that won&#39;t haunt you</a>',
        "href=\"blog.html\" class=\"title\">The case for boring databases</a>":
            'href="the-case-for-boring-databases.html" class="title">The case for boring databases</a>',
        "href=\"blog.html\" class=\"title\">A small theory of on-call</a>":
            'href="a-small-theory-of-on-call.html" class="title">A small theory of on-call</a>',
    }
    for old, new in replacements.items():
        content = content.replace(old, new)
    HOME_PAGE.write_text(content)


def main() -> None:
    posts = load_posts()
    BLOG_INDEX.write_text(render_blog_index(posts))
    for post in posts:
        (ROOT / post.file_name).write_text(render_post_page(post))
    update_home_links(posts)
    print(f"Generated {len(posts)} posts and blog index.")


if __name__ == "__main__":
    main()
