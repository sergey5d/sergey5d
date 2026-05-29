#!/usr/bin/env python3

from __future__ import annotations

import re
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Iterable


ROOT = Path(__file__).resolve().parent
POSTS_DIR = ROOT / "posts"
BLOG_INDEX_BARK = ROOT / "blog.bark"

SITE_TITLE = "Sergey Denisov"
COLLOPHON = "© 2026 · built across OpenAI and Anthropic datacenters"
FOOTER_LINKS = [
    ("https://github.com/sergey5d", "github"),
    ("https://www.linkedin.com/in/sadenisov/", "linkedin"),
    ("contact.html", "more →"),
]


@dataclass
class Post:
    slug: str
    title: str
    date: datetime
    excerpt: str
    reading_time: str
    category: str
    body_lines: list[str]

    @property
    def date_display(self) -> str:
        return self.date.strftime("%d %b %Y")

    @property
    def file_name(self) -> str:
        return f"{self.slug}.html"

    @property
    def bark_file_name(self) -> str:
        return f"{self.slug}.bark"


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


def indent_lines(lines: Iterable[str], spaces: int) -> list[str]:
    prefix = " " * spaces
    return [prefix + line for line in lines]


def format_attr_value(value: str) -> str:
    if re.fullmatch(r'[^=\s\[\]"]+', value):
        return value
    return '"' + value.replace('"', '\\"') + '"'


def bark_node(
    tag: str = "",
    *,
    classes: list[str] | None = None,
    attrs: dict[str, str] | None = None,
    body: str | None = None,
    children: list[str] | None = None,
    use_pipe: bool | None = None,
) -> list[str]:
    classes = classes or []
    attrs = attrs or {}
    children = children or []

    head = "["
    if tag and tag != "div":
        head += tag

    if classes:
        head += (" <: " if head != "[" else "<: ") + ", ".join(classes)

    if attrs:
        attr_bits = " ".join(f"{key}={format_attr_value(value)}" for key, value in attrs.items())
        head += (" " if head != "[" else "") + attr_bits

    has_metadata = bool(classes or attrs)

    if children:
        lines = [head]
        lines.extend(indent_lines(children, 2))
        lines.append("]")
        return lines

    if body is None or body == "":
        return [head + "]"]

    if use_pipe is None:
        use_pipe = has_metadata

    if use_pipe:
        return [f"{head} | {body}]"]
    if head == "[":
        return [f"{head}{body}]"]
    return [f"{head} {body}]"]


INLINE_PATTERN = re.compile(
    r"\[([^\]]+)\]\(([^)]+)\)|`([^`]+)`|\*\*([^*]+)\*\*|\*([^*]+)\*"
)


def render_inline(text: str) -> str:
    parts: list[str] = []
    pos = 0
    for match in INLINE_PATTERN.finditer(text):
        start, end = match.span()
        if start > pos:
            parts.append(text[pos:start])

        link_label, link_href, code_text, strong_text, em_text = match.groups()
        if link_label is not None:
            parts.append(f"[a href={format_attr_value(link_href)} | {render_inline(link_label)}]")
        elif code_text is not None:
            parts.append(f"[code {code_text}]")
        elif strong_text is not None:
            parts.append(f"[strong {render_inline(strong_text)}]")
        elif em_text is not None:
            parts.append(f"[em {render_inline(em_text)}]")

        pos = end

    if pos < len(text):
        parts.append(text[pos:])

    return "".join(parts)


def render_markdown(markdown: str) -> list[str]:
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
            blocks.extend(bark_node("p", body=render_inline(" ".join(paragraph).strip()), use_pipe=False))
            paragraph = []

    def flush_list() -> None:
        nonlocal list_items
        if list_items:
            children: list[str] = []
            for item in list_items:
                children.extend(bark_node("li", body=render_inline(item), use_pipe=False))
            blocks.extend(bark_node("ul", children=children))
            list_items = []

    def flush_quote() -> None:
        nonlocal quote_lines
        if quote_lines:
            blocks.extend(
                bark_node("blockquote", body=render_inline(" ".join(quote_lines).strip()), use_pipe=False)
            )
            quote_lines = []

    def flush_code() -> None:
        nonlocal code_lines
        if code_lines:
            blocks.extend(bark_node("pre", children=bark_node("code", body="\n".join(code_lines), use_pipe=False)))
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
            blocks.extend(bark_node("h2", body=render_inline(stripped[3:]), use_pipe=False))
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

    return blocks


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
                body_lines=render_markdown(body),
            )
        )
    return sorted(posts, key=lambda post: post.date, reverse=True)


def footer_lines(back_href: str) -> list[str]:
    link_lines: list[str] = []
    for href, label in FOOTER_LINKS:
        actual_href = back_href if label == "more →" else href
        attrs = {"href": actual_href}
        if actual_href.startswith("http"):
            attrs["target"] = "_blank"
            attrs["rel"] = "noreferrer"
        link_lines.extend(bark_node("a", attrs=attrs, body=label))

    return bark_node(
        "footer",
        classes=["site-footer"],
        children=bark_node(
            "",
            classes=["shell"],
            children=
            bark_node("span", classes=["colophon"], body=COLLOPHON)
            + bark_node("", classes=["links"], children=link_lines),
        ),
    )


def page_shell(title: str, main_attrs: dict[str, str], main_children: list[str], active_nav: str, *, include_lucide: bool) -> list[str]:
    nav_children: list[str] = []
    for label, href in [
        ("Home", "index.html"),
        ("Experience", "projects.html"),
        ("Blog", "blog.html"),
        ("About", "about.html"),
        ("Contact", "contact.html"),
    ]:
        classes = ["is-active"] if label == active_nav else []
        nav_children.extend(bark_node("a", classes=classes, attrs={"href": href}, body=label))

    head_children = [
        *bark_node("meta", attrs={"charset": "utf-8"}),
        *bark_node("meta", attrs={"name": "viewport", "content": "width=device-width, initial-scale=1"}),
        *bark_node("title", body=title, use_pipe=False),
        *bark_node("link", attrs={"rel": "icon", "type": "image/png", "href": "icon.png"}),
        *bark_node("link", attrs={"rel": "stylesheet", "href": "site.css"}),
    ]
    if include_lucide:
        head_children.extend(
            bark_node("script", attrs={"src": "https://unpkg.com/lucide@latest/dist/umd/lucide.min.js"})
        )

    page_lines = bark_node(
        "html",
        attrs={"lang": "en"},
        children=
        bark_node("head", children=head_children)
        + bark_node(
            "body",
            children=
            bark_node(
                "header",
                classes=["site-header"],
                attrs={"data-screen-label": "Header"},
                children=bark_node(
                    "",
                    classes=["shell"],
                    children=
                    bark_node(
                        "a",
                        classes=["wordmark"],
                        attrs={"href": "index.html", "aria-label": "Home"},
                        children=bark_node("span", classes=["dot"]) + bark_node("span", body=SITE_TITLE, use_pipe=False),
                    )
                    + bark_node("nav", classes=["nav"], attrs={"aria-label": "Primary"}, children=nav_children),
                ),
            )
            + bark_node(
                "main",
                classes=[main_attrs.pop("class")] if "class" in main_attrs else [],
                attrs=main_attrs,
                children=bark_node("", classes=["shell"], children=main_children),
            )
            + footer_lines("contact.html"),
        ),
    )

    return page_lines


def render_post_page(post: Post) -> list[str]:
    prose_children = post.body_lines
    main_children = bark_node(
        "",
        classes=["page-stack"],
        children=
        bark_node(
            "a",
            classes=["back-link"],
            attrs={"href": "blog.html"},
            children=bark_node("i", attrs={"data-lucide": "arrow-left"}) + ["  all posts"],
        )
        + bark_node(
            "",
            classes=["page-head"],
            children=
            bark_node("h1", classes=["title"], body=post.title)
            + bark_node(
                "",
                classes=["post-meta"],
                children=
                bark_node("span", classes=["tnum"], body=post.date_display)
                + bark_node("span", body="·", use_pipe=False)
                + bark_node("span", body=post.reading_time, use_pipe=False)
                + bark_node("span", body="·", use_pipe=False)
                + bark_node("span", body=post.category, use_pipe=False),
            ),
        )
        + bark_node("", classes=["prose"], children=prose_children)
        + bark_node("script", body="lucide.createIcons();", use_pipe=False),
    )

    return page_shell(
        title=f"{post.title} — {SITE_TITLE}",
        main_attrs={"class": "post", "data-screen-label": "Blog post"},
        main_children=main_children,
        active_nav="Blog",
        include_lucide=True,
    )


def group_by_year(posts: Iterable[Post]) -> list[tuple[int, list[Post]]]:
    grouped: dict[int, list[Post]] = {}
    for post in posts:
        grouped.setdefault(post.date.year, []).append(post)
    return sorted(grouped.items(), key=lambda item: item[0], reverse=True)


def render_blog_index(posts: list[Post]) -> list[str]:
    sections: list[str] = []
    for idx, (year, items) in enumerate(group_by_year(posts)):
        title = "Recent" if idx == 0 else "Earlier"
        section_classes = ["section"]
        if idx == 0:
            section_classes.append("blog-list-section")

        list_children: list[str] = []
        for post in items:
            list_children.extend(
                bark_node(
                    "li",
                    children=
                    bark_node("span", classes=["when", "tnum"], body=post.date_display)
                    + bark_node(
                        "",
                        classes=["what"],
                        children=
                        bark_node("h3", children=bark_node("a", attrs={"href": post.file_name}, body=post.title))
                        + bark_node("span", classes=["excerpt"], body=post.excerpt),
                    )
                    + bark_node("span", classes=["kind"], body=post.reading_time),
                )
            )

        sections.extend(
            bark_node(
                "section",
                classes=section_classes,
                children=
                bark_node(
                    "",
                    classes=["section-head"],
                    children=
                    bark_node("span", classes=["label"], body=f"// {year}")
                    + bark_node("h2", body=title, use_pipe=False)
                    + bark_node("span", classes=["count", "tnum"], body=f"{len(items)} posts"),
                )
                + bark_node("ul", classes=["log"], children=list_children),
            )
        )

    main_children = bark_node(
        "",
        classes=["page-stack"],
        children=
        bark_node(
            "",
            classes=["page-head"],
            children=
            bark_node("", classes=["eyebrow"], body="// writing")
            + bark_node("h1", classes=["title"], body="Notes on anything of substance or lack thereof")
            + bark_node(
                "p",
                classes=["lede"],
                body="Random thoughts about software, systems, AI, and everything else.",
            ),
        )
        + sections,
    )

    return page_shell(
        title=f"Blog — {SITE_TITLE}",
        main_attrs={"data-screen-label": "Blog"},
        main_children=main_children,
        active_nav="Blog",
        include_lucide=False,
    )


def write_bark(path: Path, lines: list[str]) -> None:
    path.write_text("\n".join(lines) + "\n")


def main() -> None:
    posts = load_posts()
    write_bark(BLOG_INDEX_BARK, render_blog_index(posts))
    for post in posts:
        write_bark(ROOT / post.bark_file_name, render_post_page(post))
    print(f"Generated {len(posts)} post bark files and blog index bark.")


if __name__ == "__main__":
    main()
