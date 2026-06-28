"""Async HTML parser — extract structural features from competitor URLs."""

import asyncio
from typing import TYPE_CHECKING

import httpx
from selectolax.parser import HTMLParser

if TYPE_CHECKING:
    from lastroseo.models import CompetitorData


async def parse_urls(urls: list[str], max_concurrent: int = 5) -> list["CompetitorData"]:
    """Fetch and parse multiple URLs concurrently.

    Returns a ``CompetitorData`` per URL with headings, word count,
    media counts, and cleaned text content.
    """
    from lastroseo.models import CompetitorData

    semaphore = asyncio.Semaphore(max_concurrent)

    async def _fetch_one(url: str) -> CompetitorData:
        async with semaphore:
            return await _parse_single(url)

    tasks = [_fetch_one(u) for u in urls]
    results = await asyncio.gather(*tasks, return_exceptions=True)

    parsed: list[CompetitorData] = []
    for i, result in enumerate(results):
        if isinstance(result, Exception):
            # Graceful degradation: return a stub for failed URLs
            parsed.append(CompetitorData(url=urls[i]))
        else:
            parsed.append(result)

    return parsed


async def _parse_single(url: str) -> "CompetitorData":
    """Fetch one URL and extract all structured data."""
    from lastroseo.models import CompetitorData

    headers = {
        "User-Agent": (
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
            "AppleWebKit/537.36 (KHTML, like Gecko) "
            "Chrome/125.0.0.0 Safari/537.36"
        ),
        "Accept-Language": "pt-BR,pt;q=0.9,en;q=0.8",
    }

    async with httpx.AsyncClient(timeout=20.0, follow_redirects=True) as client:
        resp = await client.get(url, headers=headers)
        resp.raise_for_status()

    html = resp.text
    tree = HTMLParser(html)

    # --- Meta ---
    title = _meta_content(tree, "title") or _text(tree, "title")
    meta_desc = _meta_content(tree, "description")

    # --- Headings ---
    h1 = _text(tree, "h1")
    h2_h3_map: dict[str, list[str]] = {}
    for h2_node in tree.css("h2"):
        h2_text = h2_node.text(strip=True)
        if not h2_text:
            continue
        h3s: list[str] = []
        sibling = h2_node.next
        while sibling and sibling.tag != "h2":
            if sibling.tag == "h3":
                h3_text = sibling.text(strip=True)
                if h3_text:
                    h3s.append(h3_text)
            sibling = sibling.next
        h2_h3_map[h2_text] = h3s

    # --- Readability: extract main content ---
    text_content = _readable_text(html)

    # --- Counts ---
    word_count = len(text_content.split()) if text_content else 0
    img_count = len(tree.css("img"))
    video_count = len(tree.css("video"))

    return CompetitorData(
        url=url,
        title=title,
        meta_desc=meta_desc,
        h1=h1,
        h2_h3_map=h2_h3_map,
        word_count=word_count,
        media_count={"img": img_count, "video": video_count},
        text_content=text_content,
    )


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _text(tree: HTMLParser, selector: str) -> str:
    node = tree.css_first(selector)
    return node.text(strip=True) if node else ""


def _meta_content(tree: HTMLParser, name: str) -> str:
    """Extract <meta name=...> or <meta property=og:...> content."""
    selectors = [
        f'meta[name="{name}"]',
        f'meta[property="og:{name}"]',
    ]
    for sel in selectors:
        try:
            node = tree.css_first(sel)
        except Exception:
            continue
        if node and node.attributes.get("content"):
            return node.attributes["content"]
    return ""


def _readable_text(html: str) -> str:
    """Extract main article text using readability-lxml."""
    try:
        from readability import Document

        doc = Document(html)
        content_html = doc.summary()
        tree = HTMLParser(content_html)
        return tree.text(strip=True)
    except Exception:
        # Fallback: extract all body text
        tree = HTMLParser(html)
        body = tree.css_first("body")
        return body.text(strip=True) if body else ""
