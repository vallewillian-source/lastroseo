"""SERP integration via SearXNG — self-hosted metasearch engine (local Docker).

Uses the SearXNG JSON API at ``/search?q=keyword&format=json``.
No API key required — SearXNG runs locally at ``http://localhost:8080`` by default.

Configure via environment variables:
  ``LASTROSEO_SEARXNG_URL`` — SearXNG base URL (default: http://localhost:8080)
"""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from urllib.parse import quote_plus

import httpx

_SEARXNG_URL = os.getenv("LASTROSEO_SEARXNG_URL", "http://localhost:8080")


@dataclass
class SerpResult:
    """Raw result from SERP (SearXNG)."""

    url: str
    title: str = ""
    position: int = 0


@dataclass
class SerpResponse:
    """Parsed response from SERP."""

    organic: list[SerpResult] = field(default_factory=list)
    paas: list[str] = field(default_factory=list)


async def search_serp(keyword: str, top_n: int = 7) -> SerpResponse:
    """Query SearXNG for *keyword* and return organic results + PAA-like data.

    Returns at most *top_n* organic results (default 7).

    ``suggestions`` and ``answers`` from SearXNG are treated as
    PAA equivalents for the briefing builder.
    """
    return await _search_via_searxng(keyword, top_n)


async def _search_via_searxng(keyword: str, top_n: int) -> SerpResponse:
    """Call SearXNG JSON API and parse results."""
    params = {
        "q": keyword,
        "format": "json",
        "categories": "general",
        "language": "pt-BR",
        "safesearch": "0",
    }

    headers = {
        "Accept": "application/json",
        "User-Agent": (
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
            "AppleWebKit/537.36 (KHTML, like Gecko) "
            "Chrome/125.0.0.0 Safari/537.36"
        ),
    }

    async with httpx.AsyncClient(timeout=20.0, follow_redirects=True) as client:
        resp = await client.get(
            f"{_SEARXNG_URL}/search",
            params=params,
            headers=headers,
        )
        resp.raise_for_status()
        data = resp.json()

    # --- Organic results ---
    organic: list[SerpResult] = []
    for i, item in enumerate(data.get("results", [])[:top_n], start=1):
        url = item.get("url", "")
        title = item.get("title", "")
        if url and title:
            organic.append(SerpResult(url=url, title=title, position=i))

    # --- PAA equivalent: suggestions + answers ---
    paas: list[str] = []

    suggestions: list[str] = data.get("suggestions", [])
    if isinstance(suggestions, list):
        for s in suggestions:
            if isinstance(s, str) and s.strip():
                paas.append(s.strip())

    answers: list[str] = data.get("answers", [])
    if isinstance(answers, list):
        for a in answers:
            # answers may be strings or dicts
            if isinstance(a, str) and a.strip():
                paas.append(a.strip())
            elif isinstance(a, dict):
                text = a.get("answer") or a.get("title") or ""
                if text.strip():
                    paas.append(text.strip())

    # Deduplicate
    paas = list(dict.fromkeys(paas))

    return SerpResponse(organic=organic, paas=paas)
