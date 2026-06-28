"""Briefing builder — aggregate competitor data, detect content gaps via TF-IDF."""

from collections import Counter
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from lastroseo_poc.models import Briefing, CompetitorData


def build_briefing(
    competitors: list["CompetitorData"],
    keyword: str,
    paas: list[str] | None = None,
) -> "Briefing":
    """Build an SEO briefing from parsed competitor data.

    Steps:
    1. Aggregate H2/H3 frequency
    2. TF-IDF → 15 semantic + 15 secondary keywords
    3. Search intent classification (simple heuristic)
    4. Content gap detection (PAA vs extracted headings)
    """
    from lastroseo_poc.models import Briefing

    if not competitors:
        return Briefing(keyword=keyword, search_intent="informational")

    # --- 1. Aggregate structure ---
    h2_counter: Counter[str] = Counter()
    for c in competitors:
        for h2 in c.h2_h3_map:
            h2_counter[h2.lower().strip()] += 1

    proposed_structure = _rank_headings(h2_counter)

    # --- 2. TF-IDF ---
    semantic_kws, secondary_kws = _extract_keywords(competitors, keyword)

    # --- 3. Search intent ---
    search_intent = _classify_intent(competitors)

    # --- 4. Content gaps ---
    content_gaps = _detect_gaps(competitors, paas or [], h2_counter)

    # --- Target word count (median of top competitors) ---
    word_counts = [c.word_count for c in competitors if c.word_count > 0]
    target_word_count = int(
        sorted(word_counts)[len(word_counts) // 2] if word_counts else 1200
    )

    return Briefing(
        keyword=keyword,
        search_intent=search_intent,
        paas=paas or [],
        semantic_kws=semantic_kws,
        content_gaps=content_gaps,
        target_word_count=target_word_count,
        proposed_structure=proposed_structure,
        # store secondary_kws in semantic_kws for now; can split later
    )


def _rank_headings(h2_counter: Counter[str], top_n: int = 12) -> list[str]:
    """Rank H2s by frequency across competitors, filtering noise."""
    filtered = Counter()
    for h2, count in h2_counter.most_common(50):  # scan more to filter
        if not _is_noise_heading(h2):
            filtered[h2] = count
    return [h for h, _ in filtered.most_common(top_n)]


# ── Noise filter patterns ──────────────────────────────────────────────────

_NOISE_PATTERNS = [
    # English
    "footer", "newsletter", "related posts", "related articles",
    "categories", "category", "tags", "tag", "sidebar", "menu",
    "comment", "comments", "subscribe", "share this", "share",
    "download", "login", "sign up", "sign in", "register",
    "search", "popular posts", "recent posts", "archives",
    "follow us", "connect with us", "advertisement", "sponsored",
    "cookie", "privacy policy", "terms of service",
    "copyright", "rights reserved", "powered by",
    "navigation", "breadcrumb", "sitemap",
    # Portuguese
    "rodapé", "newsletter", "posts relacionados", "artigos relacionados",
    "categorias", "categoria", "tags", "menu", "barra lateral",
    "comentários", "comentario", "inscreva-se", "compartilhe",
    "baixar", "download", "entrar", "login", "cadastro",
    "buscar", "pesquisar", "posts populares", "posts recentes",
    "arquivos", "siga-nos", "anúncio", "patrocinado",
    "política de privacidade", "termos de uso",
    "direitos reservados", "navegação", "mapa do site",
    # Structural junk
    "leia mais", "leia também", "veja também", "saiba mais",
    "clique aqui", "acesse aqui", "confira", "descubra",
    "next", "previous", "anterior", "próximo",
    "back to top", "voltar ao topo",
]


def _is_noise_heading(text: str) -> bool:
    """Return True if the heading looks like nav/footer/sidebar junk."""
    t = text.strip().lower()

    # Too short or too long
    if len(t) < 4 or len(t) > 150:
        return True

    # Exact match or substring of known noise
    for pattern in _NOISE_PATTERNS:
        if pattern in t:
            return True

    # Starts with a number-only prefix (like "3." or "03 -")
    if t.split()[0].rstrip(".-)").isdigit() and len(t.split()) <= 3:
        return True

    # URL-like
    if "http" in t or "www." in t or ".com" in t:
        return True

    # Pure navigation: single-word headings that are generic
    single_word = t.strip()
    if single_word in ("home", "início", "about", "sobre", "contact", "contato",
                        "help", "ajuda", "faq", "blog", "shop", "loja"):
        return True

    return False


def _extract_keywords(
    competitors: list["CompetitorData"], keyword: str
) -> tuple[list[str], list[str]]:
    """TF-IDF extraction: 15 semantic keywords, 15 secondary keywords.

    Returns:
        (semantic_kws, secondary_kws) — each list up to 15 terms.
    """
    from sklearn.feature_extraction.text import TfidfVectorizer

    docs = [c.text_content for c in competitors if c.text_content]
    if not docs or len(docs) < 2:
        _fallback = _keyword_fallback(competitors, keyword)
        return _fallback

    try:
        vectorizer = TfidfVectorizer(
            max_features=200,
            stop_words="portuguese",
            ngram_range=(1, 2),
        )
        tfidf = vectorizer.fit_transform(docs)
        feature_names = vectorizer.get_feature_names_out()

        # Sum TF-IDF across all documents
        scores = tfidf.sum(axis=0).A1
        scored = sorted(
            zip(feature_names, scores), key=lambda x: x[1], reverse=True
        )

        # Exclude keyword itself
        keyword_lower = keyword.lower()
        terms = [(t, s) for t, s in scored if keyword_lower not in t]

        semantic = [t for t, _ in terms[:15]]
        secondary = [t for t, _ in terms[15:30]]
        return semantic, secondary
    except Exception:
        return _keyword_fallback(competitors, keyword)


def _keyword_fallback(
    competitors: list["CompetitorData"], keyword: str
) -> tuple[list[str], list[str]]:
    """Fallback: extract keywords from headings when TF-IDF fails."""
    all_headings: list[str] = []
    for c in competitors:
        all_headings.extend(c.h2_h3_map.keys())
        for h3s in c.h2_h3_map.values():
            all_headings.extend(h3s)

    from sklearn.feature_extraction.text import TfidfVectorizer

    joined = " ".join(all_headings)
    if not joined:
        return ([], [])

    try:
        vec = TfidfVectorizer(stop_words="portuguese", max_features=30)
        vec.fit([joined])
        terms = vec.get_feature_names_out().tolist()
        return (terms[:15], terms[15:30])
    except Exception:
        return ([], [])


def _classify_intent(competitors: list["CompetitorData"]) -> str:
    """Simple heuristic: if URLs contain /blog/ or /artigo/ → informational,
    if they contain /product/ or /loja/ → transactional.
    """
    info_signals = 0
    trans_signals = 0

    for c in competitors:
        lower = c.url.lower()
        if any(s in lower for s in ("/blog/", "/artigo/", "/guia/", "/post/")):
            info_signals += 1
        if any(s in lower for s in ("/product/", "/produto/", "/loja/", "/shop/", "/categoria/")):
            trans_signals += 1

    if trans_signals > info_signals:
        return "transactional"
    return "informational"


def _detect_gaps(
    competitors: list["CompetitorData"],
    paas: list[str],
    h2_counter: Counter[str],
) -> list[str]:
    """Find topics covered by PAA but missing from competitor H2s."""
    gaps: list[str] = []
    existing_h2s = set(h2_counter.keys())

    for q in paas:
        q_lower = q.lower().rstrip("?")
        # Check if any competitor covers this topic
        covered = any(
            word in " ".join(existing_h2s)
            for word in q_lower.split()
            if len(word) > 4
        )
        if not covered and q not in gaps:
            gaps.append(q)

    # Also add gaps from high-frequency terms in text not used as H2
    from sklearn.feature_extraction.text import TfidfVectorizer

    all_text = " ".join(c.text_content for c in competitors if c.text_content)
    existing_all = set()
    for c in competitors:
        existing_all.update(h.lower() for h in c.h2_h3_map.keys())
        for h3s in c.h2_h3_map.values():
            existing_all.update(h.lower() for h in h3s)

    if all_text:
        try:
            vec = TfidfVectorizer(
                max_features=50, stop_words="portuguese", ngram_range=(2, 3)
            )
            vec.fit([all_text])
            for phrase in vec.get_feature_names_out():
                if phrase.lower() not in existing_all and phrase not in gaps:
                    if len(phrase.split()) >= 2:
                        gaps.append(phrase)
        except Exception:
            pass

    return gaps[:10]  # Cap at 10 gaps
