#!/usr/bin/env python3

from __future__ import annotations

import html
import os
import re
import sys
from dataclasses import dataclass, field
from html.parser import HTMLParser


VOID_TAGS = {
    "area",
    "base",
    "br",
    "col",
    "embed",
    "hr",
    "img",
    "input",
    "link",
    "meta",
    "param",
    "source",
    "track",
    "wbr",
}


@dataclass
class TextNode:
    text: str


@dataclass
class ElementNode:
    tag: str
    attrs: list[tuple[str, str | None]] = field(default_factory=list)
    children: list[ElementNode | TextNode] = field(default_factory=list)


class TreeBuilder(HTMLParser):
    def __init__(self) -> None:
        super().__init__(convert_charrefs=True)
        self.root = ElementNode("root")
        self.stack = [self.root]

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        node = ElementNode(tag, attrs)
        self.stack[-1].children.append(node)
        if tag not in VOID_TAGS:
            self.stack.append(node)

    def handle_endtag(self, tag: str) -> None:
        for i in range(len(self.stack) - 1, 0, -1):
            if self.stack[i].tag == tag:
                del self.stack[i:]
                break

    def handle_startendtag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        self.stack[-1].children.append(ElementNode(tag, attrs))

    def handle_data(self, data: str) -> None:
        if not data:
            return
        self.stack[-1].children.append(TextNode(data))

    def handle_entityref(self, name: str) -> None:
        self.handle_data(html.unescape(f"&{name};"))

    def handle_charref(self, name: str) -> None:
        self.handle_data(html.unescape(f"&#{name};"))


def normalize_text(text: str) -> str:
    return re.sub(r"\s+", " ", text).strip()


def is_significant_text(node: TextNode) -> bool:
    return normalize_text(node.text) != ""


def format_attr_value(value: str) -> str:
    if re.fullmatch(r'[^=\s\[\]"]+', value):
        return value
    escaped = value.replace('"', '\\"')
    return f'"{escaped}"'


def split_ids(value: str) -> list[str]:
    return [part for part in value.split() if part]


def split_classes(value: str) -> list[str]:
    return [part for part in value.split() if part]


def format_node(node: ElementNode | TextNode, indent: int = 0) -> list[str]:
    if isinstance(node, TextNode):
        text = normalize_text(node.text)
        return [(" " * indent) + text] if text else []

    indent_str = " " * indent
    tag = "" if node.tag == "div" else node.tag

    ids: list[str] = []
    classes: list[str] = []
    other_attrs: list[tuple[str, str | None]] = []
    for key, value in node.attrs:
        value = "" if value is None else value
        if key == "id":
            ids = split_ids(value)
        elif key == "class":
            classes = split_classes(value)
        else:
            other_attrs.append((key, value))

    head = "["
    if tag:
        head += tag

    if ids:
        if head == "[":
            head += f"@{ids[0]}"
            ids = ids[1:]
        for item in ids:
            head += f" @{item}"

    if classes:
        if head == "[":
            head += "<: "
        else:
            head += " <: "
        head += ", ".join(classes)

    if other_attrs:
        attrs_str = " ".join(f"{key}={format_attr_value(value)}" for key, value in other_attrs)
        if head == "[":
            head += attrs_str
        else:
            head += " " + attrs_str
    has_metadata = bool(ids or classes or other_attrs)

    children = [child for child in node.children if not isinstance(child, TextNode) or is_significant_text(child)]
    if not children:
        return [indent_str + head + "]"]

    only_text = all(isinstance(child, TextNode) for child in children)
    if only_text:
        body = " ".join(normalize_text(child.text) for child in children if isinstance(child, TextNode))
        if has_metadata:
            return [f"{indent_str}{head} | {body}]"]
        if tag:
            return [f"{indent_str}{head} {body}]"]
        return [f"{indent_str}{head}{body}]"]

    lines = [indent_str + head]
    for child in children:
        lines.extend(format_node(child, indent + 2))
    lines.append(indent_str + "]")
    return lines


def convert_html_to_bark(source: str) -> str:
    parser = TreeBuilder()
    parser.feed(source)
    lines: list[str] = []
    for child in parser.root.children:
        lines.extend(format_node(child, 0))
    return "\n".join(lines) + "\n"


def main() -> int:
    if len(sys.argv) != 2:
        print(f"usage: {os.path.basename(sys.argv[0])} <source-html>", file=sys.stderr)
        return 2

    with open(sys.argv[1], "r", encoding="utf-8") as fh:
        source = fh.read()

    sys.stdout.write(convert_html_to_bark(source))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
