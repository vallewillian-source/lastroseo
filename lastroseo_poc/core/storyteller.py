"""Storytelling module — generates a narrative arc from the SEO briefing.

This sits between the analyzer (Briefing) and the writer (Article).
It calls an LLM to produce a ``StoryArc`` that guides the final article
toward a compelling, human-readable narrative instead of a dry keyword dump.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from lastroseo_poc.models import BrandGuide, Briefing, StoryArc

STORYTELLER_PROMPT = """\
You are an expert editorial strategist and narrative designer. Your job is to
transform an SEO briefing into a compelling **narrative arc** for an article.

## Brand Guide
- Tone: {tone}
- Persona: {persona}
- Rules: {rules}

## SEO Briefing
{briefing_json}

{product_context}

## Instructions
Create a storytelling plan for an article about "{keyword}".

1. **angle**: A unique, provocative, or counter-intuitive hook. The article should
   NOT just define the keyword — it should make a case, reveal a truth, or
   challenge conventional wisdom. Make it specific and memorable.

2. **target_audience**: Who exactly is this for? Be specific (role, company size,
   pain point). Not "anyone interested in SEO".

3. **arguments**: 3 to 5 key arguments the article will defend, in logical
   progression. Each should build on the previous. These are NOT just H2 headings
   — they are claims the article proves with the competitor data and PAA insights.
   {product_arg_hint}

4. **emotional_arc**: The emotional journey: problem → tension → resolution →
   call-to-action. One sentence per stage.
   {product_arc_hint}

5. **section_plan**: 5 to 8 ordered sections. Each section has:
   - "h2": the section heading (compelling, not generic)
   - "goal": what this section achieves in the narrative
   - "tone": emotional register (e.g. provocativo, analítico, empático, inspirador)

{product_rule}

Output ONLY valid JSON matching this schema:
{{
  "angle": "...",
  "target_audience": "...",
  "arguments": ["...", "..."],
  "emotional_arc": "...",
  "section_plan": [
    {{"h2": "...", "goal": "...", "tone": "..."}}
  ]
}}
"""


async def build_story_arc(
    briefing: "Briefing",
    brand_guide: "BrandGuide",
    *,
    model: str = "gpt-4o",
    api_key: str | None = None,
    api_base: str | None = None,
    product_info: "ProductInfo | None" = None,
) -> "StoryArc":
    """Generate a storytelling plan from the briefing.

    Calls an LLM via LiteLLM. Falls back to a heuristic arc if no API key
    is available.

    If *product_info* is provided, the narrative arc will guide readers
    toward the product's value proposition — but NEVER mention the product
    directly in the arc itself.
    """
    from lastroseo_poc.models import StoryArc

    if not api_key:
        return _heuristic_arc(briefing, brand_guide, product_info)

    import litellm

    # Build product-specific hint strings
    product_context = ""
    product_arg_hint = ""
    product_arc_hint = ""
    product_rule = ""

    if product_info and product_info.name:
        product_context = (
            f"## Product Context (for narrative alignment ONLY — do NOT mention in the arc)\n"
            f"- Product: {product_info.name}\n"
            f"- What it does: {product_info.description}\n"
            f"- Value proposition: {product_info.value_prop}\n"
            f"The article will eventually promote this product ONLY in the final paragraph. "
            f"Your story arc should make the reader FEEL the need for {product_info.value_prop} "
            f"without ever naming {product_info.name}."
        )
        product_arg_hint = (
            "The arguments should build a case that naturally leads to needing "
            f"a solution like: {product_info.value_prop}. But DO NOT mention {product_info.name}."
        )
        product_arc_hint = (
            f"The resolution/CTA stage should create desire for {product_info.value_prop}."
        )
        product_rule = (
            f"**CRITICAL**: The article promotes {product_info.name} ONLY in the final paragraph. "
            "Your story arc must NOT include the product name. Guide the reader there naturally."
        )

    prompt = STORYTELLER_PROMPT.format(
        tone=brand_guide.tone,
        persona=brand_guide.persona,
        rules=", ".join(brand_guide.rules) if brand_guide.rules else "none",
        keyword=briefing.keyword,
        briefing_json=briefing.model_dump_json(indent=2),
        product_context=product_context,
        product_arg_hint=product_arg_hint,
        product_arc_hint=product_arc_hint,
        product_rule=product_rule,
    )

    kwargs: dict = {
        "model": _resolve_model(model),
        "messages": [
            {"role": "system", "content": "You are a narrative strategist. Output only valid JSON."},
            {"role": "user", "content": prompt},
        ],
        "temperature": 0.8,
        "max_tokens": 2000,
    }

    if api_key:
        kwargs["api_key"] = api_key
    if api_base:
        kwargs["api_base"] = api_base

    try:
        response = await litellm.acompletion(**kwargs)
        content = response.choices[0].message.content or "{}"
        # Strip markdown fences if present
        content = _strip_json_fences(content)
        arc = StoryArc.model_validate_json(content)
        # Validate: if LLM returned empty arc, use heuristic
        if not arc.arguments and not arc.section_plan:
            return _heuristic_arc(briefing, brand_guide)
        return arc
    except Exception:
        return _heuristic_arc(briefing, brand_guide)


def _heuristic_arc(
    briefing: "Briefing",
    brand_guide: "BrandGuide",
    product_info: "ProductInfo | None" = None,
) -> "StoryArc":
    """Build a StoryArc without calling an LLM — smart fallback.

    Uses the briefing's proposed_structure, PAA questions, and content gaps
    to construct a logical narrative.
    """
    from lastroseo_poc.models import StoryArc

    keyword = briefing.keyword

    # Build arguments from content gaps + PAA
    arguments: list[str] = []
    if briefing.content_gaps:
        arguments.append(
            f"O mercado ainda não aborda adequadamente: {briefing.content_gaps[0]}"
        )
    if briefing.paas:
        questions = briefing.paas[:3]
        arguments.append(f"Responder às dúvidas reais dos usuários: {', '.join(questions)}")

    if not arguments:
        if product_info and product_info.name:
            arguments = [
                f"Por que {keyword} é mais relevante agora do que nunca para empresas que buscam eficiência",
                f"Os erros mais comuns ao implementar {keyword} e como uma plataforma de coordenação resolve",
                f"Casos reais de empresas que transformaram operações com {keyword}",
                f"O que diferencia uma solução real de {keyword} de ferramentas pontuais",
            ]
        else:
            arguments = [
                f"Por que {keyword} é mais relevante agora do que nunca",
                f"Os erros mais comuns ao implementar {keyword} e como evitá-los",
                f"Casos reais de empresas que transformaram resultados com {keyword}",
                f"O futuro de {keyword}: tendências e previsões práticas",
            ]

    # Build section plan from proposed structure, filtering noise
    clean_h2s: list[str] = []
    for h2 in briefing.proposed_structure:
        h2_clean = h2.strip().lower()
        # Skip obvious noise
        if any(w in h2_clean for w in (
            "footer", "newsletter", "related", "categor", "tag",
            "sidebar", "menu", "comment", "subscribe", "share",
            "download", "login", "sign up", "search",
        )):
            continue
        if len(h2_clean) < 5 or len(h2_clean) > 120:
            continue
        clean_h2s.append(h2.strip())

    # Force meaningful sections if nothing survived
    if not clean_h2s or len(clean_h2s) < 3:
        clean_h2s = [
            f"O que realmente significa {keyword} (e o que não significa)",
            f"Por que a maioria das empresas erra ao implementar {keyword}",
            f"Os 3 pilares de uma estratégia vencedora em {keyword}",
            f"Como medir resultados de {keyword} sem precisar de um PhD em dados",
            f"O futuro de {keyword}: prepare-se agora para não ficar para trás",
        ]

    # Build section_plan with goals and tones
    tone_map = ["provocativo", "analítico", "empático", "inspirador", "prático"]
    goal_map = [
        "criar curiosidade e engajamento inicial",
        "diagnosticar o problema que o leitor enfrenta",
        "apresentar a solução com evidências",
        "contrapor objeções comuns com dados",
        "mostrar o caminho prático e o call-to-action",
    ]

    section_plan: list[dict[str, str]] = []
    for i, h2 in enumerate(clean_h2s[:7]):
        section_plan.append({
            "h2": h2,
            "goal": goal_map[i % len(goal_map)],
            "tone": tone_map[i % len(tone_map)],
        })

    return StoryArc(
        angle=(
            f"A verdade inconveniente sobre {keyword} que ninguém "
            f"está contando — e como usá-la a seu favor"
        ),
        target_audience=briefing.search_intent,
        arguments=arguments[:5],
        emotional_arc=(
            f"Reconhecimento do problema → frustração com soluções rasas → "
            f"esperança com abordagem correta → ação prática"
        ),
        section_plan=section_plan,
    )


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _resolve_model(model: str) -> str:
    if model.startswith(("openai/", "anthropic/", "ollama/")):
        return model
    return f"openai/{model}"


def _strip_json_fences(text: str) -> str:
    """Remove ```json fences from LLM output."""
    t = text.strip()
    if t.startswith("```"):
        lines = t.split("\n")
        if len(lines) > 1:
            lines = lines[1:]
        if lines and lines[-1].strip() == "```":
            lines = lines[:-1]
        t = "\n".join(lines)
    return t.strip()
