"""Article writer — LiteLLM-powered article generation from StoryArc + briefing.

Produces a professional, human-readable article — not a dry keyword list.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from lastroseo_poc.models import BrandGuide, Briefing, StoryArc


SYSTEM_PROMPT = """\
You are an expert SEO journalist. Write articles in **Brazilian Portuguese**
that rank on Google AND people actually enjoy reading.

## Brand Guide
- **Tone**: {tone}
- **Persona**: {persona}
- **Rules**: {rules}

## Storytelling Arc (MANDATORY — follow this EXACTLY)
- **Angle**: {angle}
- **Target Audience**: {audience}
- **Arguments to defend** (in order):
{arguments}
- **Emotional journey**: {emotional_arc}

## Section Plan
{section_plan}

## Reference Data (from SERP analysis)
- **Keyword**: {keyword}
- **Search intent**: {intent}
- **People Also Ask** (use as FAQ material): {paas}
- **Semantic keywords** (sprinkle naturally): {sem_kws}
- **Content gaps** (address these — competitors missed them): {gaps}
- **Target word count**: {word_count} words (±15%)

## Writing Rules
1. **Lead paragraph** must hook the reader — use the angle. Start with a question,
   a surprising stat, or a provocative statement. No "Neste artigo vamos...".
2. Each H2 section must be 2-4 paragraphs with substance — NO single-sentence sections.
3. Use concrete examples, comparisons, and mini-stories. Not just definitions.
4. **Readability**: short paragraphs (3-5 sentences max). Mix sentence lengths.
5. Include a counterpoint/objection section before the conclusion.
6. End with a strong, actionable conclusion + FAQ section using PAA questions.
7. Write in Brazilian Portuguese natural to a native speaker.
8. Output ONLY the final article in valid Markdown.
9. Start with HTML comments: meta title and meta description.
10. H1 must be compelling, NOT just the keyword — incorporate the angle.

## Article Structure Template
<!-- meta tags -->
# [Compelling H1 that includes the angle, not just the keyword]

[Lead paragraph — hook, problem, promise]

## [Context/Problem section — why this matters NOW]

## [Argument 1 — from the story arc]

## [Argument 2 — from the story arc]

...

## [Counterpoint — what skeptics say, and why they're wrong]

## [Conclusion — actionable summary + next steps]

## Perguntas Frequentes (FAQ)
...use PAA questions...

{product_rule}
"""


async def write_article(
    briefing: "Briefing",
    brand_guide: "BrandGuide",
    story_arc: "StoryArc",
    *,
    model: str = "gpt-4o",
    api_key: str | None = None,
    api_base: str | None = None,
    product_info: "ProductInfo | None" = None,
) -> str:
    """Generate a Markdown article from briefing, brand guide, and story arc.

    Uses LiteLLM to route to the configured model. Set
    ``LASTROSEO_LLM_API_KEY`` env var or pass *api_key* directly.
    """
    import litellm

    rules_block = "\n".join(f"  - {r}" for r in brand_guide.rules) or "  - (none)"

    # Format arguments as numbered list
    args_block = "\n".join(
        f"  {i}. {arg}" for i, arg in enumerate(story_arc.arguments, 1)
    ) if story_arc.arguments else "  (none)"

    # Format section plan
    sp_lines: list[str] = []
    for i, sec in enumerate(story_arc.section_plan, 1):
        sp_lines.append(
            f"  {i}. **{sec.get('h2', '???')}** "
            f"(goal: {sec.get('goal', 'inform')}, tone: {sec.get('tone', 'neutro')})"
        )
    section_plan_block = "\n".join(sp_lines) if sp_lines else "  (none)"

    prompt = SYSTEM_PROMPT.format(
        tone=brand_guide.tone,
        persona=brand_guide.persona,
        rules=rules_block,
        angle=story_arc.angle or f"Guia definitivo sobre {briefing.keyword}",
        audience=story_arc.target_audience or "profissionais de marketing e negócios",
        arguments=args_block,
        emotional_arc=story_arc.emotional_arc or "problema → solução → ação",
        section_plan=section_plan_block,
        keyword=briefing.keyword,
        intent=briefing.search_intent,
        paas=", ".join(briefing.paas[:8]) if briefing.paas else "none",
        sem_kws=", ".join(briefing.semantic_kws[:5]) if briefing.semantic_kws else "none",
        gaps=", ".join(briefing.content_gaps[:5]) if briefing.content_gaps else "none",
        word_count=briefing.target_word_count,
        product_name=product_info.name if product_info else "",
        product_desc=product_info.description if product_info else "",
        product_cta=product_info.cta if product_info else "",
        product_rule=_build_product_rule(product_info),
    )

    kwargs: dict = {
        "model": _resolve_model(model),
        "messages": [
            {"role": "system", "content": prompt},
            {
                "role": "user",
                "content": (
                    f"Escreva o artigo completo sobre '{briefing.keyword}' "
                    f"seguindo o story arc e a estrutura definidos. "
                    f"Seja provocativo, original e útil."
                ),
            },
        ],
        "temperature": 0.75,
        "max_tokens": 4096,
    }

    if api_key:
        kwargs["api_key"] = api_key
    if api_base:
        kwargs["api_base"] = api_base

    try:
        response = await litellm.acompletion(**kwargs)
        content: str = response.choices[0].message.content or ""
        # Post-processing: strip product mentions before last paragraph
        if product_info and product_info.name and product_info.name in content:
            content = _enforce_cta_only_in_last_paragraph(content, product_info.name)
        return content.strip()
    except Exception:
        # If LLM call fails, fall back to template
        return generate_fallback_article(briefing, story_arc, product_info)


def write_article_sync(
    briefing: "Briefing",
    brand_guide: "BrandGuide",
    story_arc: "StoryArc",
    *,
    model: str = "gpt-4o",
    api_key: str | None = None,
    api_base: str | None = None,
    product_info: "ProductInfo | None" = None,
) -> str:
    """Synchronous wrapper around ``write_article``."""
    import asyncio

    return asyncio.run(
        write_article(briefing, brand_guide, story_arc, model=model, api_key=api_key,
                       api_base=api_base, product_info=product_info)
    )


def _build_product_rule(product_info: "ProductInfo | None") -> str:
    """Build the product placement rule for the system prompt."""
    if not product_info or not product_info.name:
        return ""
    return (
        f"## 🚨 PRODUCT PLACEMENT RULE — PUNISHABLE BY REWRITE 🚨\n"
        f"The product **{product_info.name}** EXISTS ONLY IN THE LAST PARAGRAPH.\n\n"
        f"**WHERE IT MUST APPEAR:**\n"
        f"- ONLY in the FINAL paragraph of the entire article (after FAQ, before the end).\n\n"
        f"**WHERE IT MUST NEVER APPEAR:**\n"
        f"- NOT in any H2 section body\n"
        f"- NOT in the FAQ answers (even when discussing tools/pricing)\n"
        f"- NOT in the introduction or lead\n"
        f"- NOT in meta title or meta description\n"
        f"- NOT in the conclusion before FAQ\n"
        f"- NOT in any comparison or example\n\n"
        f"**FINAL PARAGRAPH FORMAT:**\n"
        f"Write ONE short paragraph at the VERY END that says:\n"
        f"\"{product_info.description} Conheça o {product_info.name}: {product_info.cta}\"\n\n"
        f"**CHECK BEFORE OUTPUT:** Search for \"{product_info.name}\". "
        f"It must appear EXACTLY ONCE — in the final paragraph only. "
        f"If it appears anywhere else, DELETE those mentions."
    )


# ---------------------------------------------------------------------------
# Fallback: rich template — no LLM, but still professional
# ---------------------------------------------------------------------------

FALLBACK_TEMPLATE = """\
<!--
  meta_title: {meta_title}
  meta_description: {meta_desc}
-->

# {h1_title}

{lead}

{body}

## Conclusão: O Que Fazer Agora

{conclusion}

{product_cta}

## Perguntas Frequentes (FAQ)

{faq}
"""


def generate_fallback_article(
    briefing: "Briefing",
    story_arc: "StoryArc | None" = None,
    product_info: "ProductInfo | None" = None,
) -> str:
    """Generate a professional article without calling an LLM.

    Uses the StoryArc for structure, competitor insights for substance,
    and produces a readable, narrative-driven article — NOT generic filler.
    """
    from datetime import datetime

    arc = story_arc or _default_arc(briefing)
    year = datetime.now().year
    kw = briefing.keyword

    # ── Meta ───────────────────────────────────────────────────────────
    meta_title = f"{kw}: {arc.angle[:60] if arc.angle else 'Guia Completo'} [{year}]"
    meta_desc = (
        f"Descubra {kw} com uma abordagem prática e baseada em dados. "
        f"{' '.join(arc.arguments[:2]) if arc.arguments else ''}"
    )[:155]

    # ── H1 ─────────────────────────────────────────────────────────────
    h1_title = f"{kw}: {arc.angle.split('.')[0] if arc.angle else 'O Guia Definitivo'}"

    # ── Lead ───────────────────────────────────────────────────────────
    lead = _build_lead(kw, arc, briefing)

    # ── Body — one section per section_plan entry ──────────────────────
    body_parts: list[str] = []
    for i, sec in enumerate(arc.section_plan):
        h2 = sec.get("h2", f"Seção {i + 1}")
        goal = sec.get("goal", "")
        tone = sec.get("tone", "")

        body_parts.append(f"## {h2}\n")

        # Build a substantive paragraph based on the section goal
        para = _build_section_paragraph(h2, goal, tone, kw, briefing, i)
        body_parts.append(f"{para}\n")

        # Add a second paragraph with a concrete angle
        detail = _build_detail_paragraph(h2, kw, briefing, i)
        body_parts.append(f"{detail}\n")

    # ── Conclusion ─────────────────────────────────────────────────────
    conclusion = (
        f"Implementar **{kw}** não é apenas uma questão de tecnologia — "
        f"é uma decisão estratégica que impacta diretamente a experiência "
        f"dos seus clientes e os resultados do seu negócio.\n\n"
        f"Comece pequeno: escolha um processo, automatize, meça os resultados "
        f"e escale a partir dos dados. As empresas que estão ganhando hoje "
        f"não são as que mais investem — são as que melhor executam."
    )

    # ── Product CTA (only if product_info provided) ─────────────────────
    product_cta_block = ""
    if product_info and product_info.name:
        product_cta_block = (
            f"\n\n---\n\n"
            f"Se você quer dar o próximo passo e transformar a forma como sua "
            f"empresa lida com **{kw}**, conheça o **{product_info.name}**. "
            f"{product_info.description}. "
            f"{product_info.cta}"
        )

    # ── FAQ ────────────────────────────────────────────────────────────
    faq_parts: list[str] = []
    faq_sources = briefing.paas[:6] if briefing.paas else [
        f"O que é {kw}?",
        f"Como implementar {kw}?",
        f"Quais os benefícios de {kw}?",
        f"{kw} vale a pena para pequenas empresas?",
    ]
    for q in faq_sources:
        faq_parts.append(f"### {q}\n")
        faq_parts.append(
            f"Esta é uma das perguntas mais comuns sobre **{q.lower().rstrip('?')}**. "
            f"Com base na análise dos principais conteúdos sobre **{kw}**, "
            f"a resposta envolve três fatores principais: contexto do seu negócio, "
            f"maturidade da equipe e ferramentas disponíveis. "
            f"O mais importante é começar com um piloto controlado e expandir "
            f"com base em métricas reais de satisfação do cliente.\n"
        )

    return FALLBACK_TEMPLATE.format(
        meta_title=meta_title,
        meta_desc=meta_desc,
        h1_title=h1_title,
        lead=lead,
        body="\n".join(body_parts),
        conclusion=conclusion,
        product_cta=product_cta_block,
        faq="\n".join(faq_parts),
    )


# ---------------------------------------------------------------------------
# Paragraph builders — generate real-sounding content from briefing data
# ---------------------------------------------------------------------------


def _enforce_cta_only_in_last_paragraph(content: str, product_name: str) -> str:
    """Post-process: remove all product mentions except the last one.

    Ensures *product_name* appears only in the final paragraph.
    """
    import re

    paragraphs = content.split("\n\n")
    if len(paragraphs) < 2:
        return content

    # Find the last occurrence
    last_idx = -1
    for i in range(len(paragraphs) - 1, -1, -1):
        if product_name.lower() in paragraphs[i].lower():
            last_idx = i
            break

    if last_idx < 0:
        return content  # No mentions at all — fine

    # Remove all mentions before the last occurrence
    cleaned: list[str] = []
    for i, para in enumerate(paragraphs):
        if i == last_idx:
            cleaned.append(para)
        elif product_name.lower() in para.lower():
            # Remove the product mention but keep the paragraph
            pattern = re.compile(re.escape(product_name), re.IGNORECASE)
            cleaned_para = pattern.sub("[ferramenta de automação]", para)
            cleaned.append(cleaned_para)
        else:
            cleaned.append(para)

    return "\n\n".join(cleaned)


def _build_lead(keyword: str, arc: "StoryArc", briefing: "Briefing") -> str:
    """Craft a compelling lead paragraph."""
    angle_hook = arc.angle[:100] if arc.angle else f"dominar {keyword}"

    pain_point = (
        briefing.paas[0] if briefing.paas
        else f"como realmente obter resultados com {keyword}"
    )

    return (
        f"Você já parou para pensar por que algumas empresas conseguem "
        f"resultados extraordinários com **{keyword}** enquanto outras "
        f"patinam com as mesmas ferramentas?\n\n"
        f"A resposta não está no orçamento nem no tamanho da equipe. "
        f"Está na **abordagem**. Este artigo vai além do óbvio sobre *{angle_hook}* "
        f"e mostra o que os líderes do mercado fazem diferente — "
        f"com exemplos práticos e um caminho claro para você aplicar hoje.\n\n"
        f"Se você já se perguntou *{pain_point}*, "
        f"continue lendo. Este é o artigo que faltava."
    )


def _build_section_paragraph(
    h2: str,
    goal: str,
    tone: str,
    keyword: str,
    briefing: "Briefing",
    index: int,
) -> str:
    """Build a substantive opening paragraph for an H2 section."""
    starters = [
        f"O conceito de **{h2.lower()}** vai muito além do que a maioria "
        f"dos conteúdos superficiais entrega. Na prática, estamos falando de...",

        f"Quando analisamos os dados reais de empresas que implementaram "
        f"**{h2.lower()}**, um padrão fica claro: o sucesso não vem da "
        f"ferramenta, mas da estratégia por trás dela.",

        f"Existe um mito persistente sobre **{h2.lower()}** que precisa "
        f"ser desfeito. Não se trata de substituir pessoas por máquinas, "
        f"mas de potencializar o que já funciona bem.",

        f"A diferença entre empresas medianas e extraordinárias em "
        f"**{keyword}** está exatamente aqui: **{h2.lower()}**. "
        f"E a boa notícia é que não é complexo.",

        f"Se você perguntar para 10 especialistas em **{keyword}** qual é "
        f"o fator mais negligenciado, 8 vão mencionar **{h2.lower()}**. "
        f"Veja por quê.",
    ]
    return starters[index % len(starters)]


def _build_detail_paragraph(
    h2: str,
    keyword: str,
    briefing: "Briefing",
    index: int,
) -> str:
    """Build a follow-up detail paragraph with concrete framing."""
    angles = [
        (
            f"Pense em uma empresa de médio porte que decidiu implementar "
            f"**{keyword}** começando exatamente por **{h2.lower()}**. "
            f"Em 90 dias, os resultados já eram visíveis: redução de custos, "
            f"aumento de satisfação e uma equipe mais engajada. O segredo? "
            f"Não tentaram resolver tudo de uma vez."
        ),
        (
            f"Os números não mentem: empresas que priorizam **{h2.lower()}** "
            f"dentro da estratégia de **{keyword}** reportam resultados "
            f"consistentemente superiores. Não é coincidência — é método."
        ),
        (
            f"Um erro clássico ao abordar **{keyword}** é pular direto para "
            f"a implementação sem antes entender **{h2.lower()}**. "
            f"As consequências? Retrabalho, frustração e desperdício de budget."
        ),
    ]
    return angles[index % len(angles)]


def _default_arc(briefing: "Briefing") -> "StoryArc":
    """Minimal arc when no StoryArc is provided."""
    from lastroseo_poc.models import StoryArc

    return StoryArc(
        angle=f"O que realmente importa sobre {briefing.keyword}",
        target_audience=briefing.search_intent,
        arguments=[f"Entender {briefing.keyword} profundamente"],
        emotional_arc="curiosidade → compreensão → ação",
        section_plan=[
            {"h2": f"O que é {briefing.keyword}?", "goal": "contextualizar", "tone": "analítico"},
            {"h2": "Por que isso importa agora?", "goal": "criar urgência", "tone": "provocativo"},
            {"h2": "Como aplicar na prática", "goal": "orientar ação", "tone": "prático"},
        ],
    )


def _resolve_model(model: str) -> str:
    if model.startswith(("openai/", "anthropic/", "ollama/")):
        return model
    return f"openai/{model}"
