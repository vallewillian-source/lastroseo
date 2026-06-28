"""CLI entry point — Typer + Rich interactive SEO article generator."""

from __future__ import annotations

import asyncio
from pathlib import Path

import typer
from dotenv import load_dotenv
from rich.console import Console
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.prompt import Prompt

# Load .env from project root (or current directory)
load_dotenv()

from lastroseo.config import get_api_keys, load_brand_guide, load_product_doc
from lastroseo.core.analyzer import build_briefing
from lastroseo.core.parser import parse_urls
from lastroseo.core.serp import search_serp
from lastroseo.core.storyteller import build_story_arc
from lastroseo.core.writer import generate_fallback_article, write_article

app = typer.Typer(
    name="lastroseo",
    help="🚀 LastroSEO — CLI para briefing SEO e geração de artigos via LLM",
    add_completion=False,
)

console = Console()
OUT_DIR = Path("out")

# ── helpers ──────────────────────────────────────────────────────────────────


def _emoji_progress() -> Progress:
    return Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        console=console,
    )


def _save_output(briefing, story_arc, article_md: str) -> tuple[Path, Path, Path]:
    """Persist briefing.json, story_arc.json, and artigo_final.md to OUT_DIR."""
    OUT_DIR.mkdir(parents=True, exist_ok=True)

    briefing_path = OUT_DIR / "briefing.json"
    with open(briefing_path, "w", encoding="utf-8") as fh:
        fh.write(briefing.model_dump_json(indent=2))

    story_path = OUT_DIR / "story_arc.json"
    with open(story_path, "w", encoding="utf-8") as fh:
        fh.write(story_arc.model_dump_json(indent=2))

    article_path = OUT_DIR / "artigo_final.md"
    with open(article_path, "w", encoding="utf-8") as fh:
        fh.write(article_md)

    return briefing_path, story_path, article_path


# ── main command ─────────────────────────────────────────────────────────────


@app.command()
def run(
    keyword: str = typer.Option(None, "--keyword", "-k", help="Palavra-chave alvo"),
    brand: str = typer.Option(
        None, "--brand", "-b", help="Caminho para brand_guide.yaml"
    ),
    model: str = typer.Option(
        "gpt-4o", "--model", "-m", help="Modelo LLM (ex: gpt-4o, claude-3-opus)"
    ),
    top_n: int = typer.Option(7, "--top", "-t", help="Número de concorrentes (3-10)"),
    fallback: bool = typer.Option(
        False, "--fallback", help="Usar artigo template (sem LLM)"
    ),
    product_doc: str = typer.Option(
        None, "--product-doc", "-p", help="Arquivo .md com documentação do produto/serviço"
    ),
):
    """Run the full LastroSEO pipeline: SERP → Parse → Analyze → Storytell → Generate."""
    cli_top_n = max(3, min(top_n, 10))

    # ── Banner ────────────────────────────────────────────────────────────
    console.print(
        Panel.fit(
            "[bold cyan]🚀 LastroSEO POC[/bold cyan]\n"
            "[dim]SERP Scraper · Content Gap Analyzer · Storyteller · Article Writer[/dim]",
            border_style="cyan",
        )
    )

    # ── Input ─────────────────────────────────────────────────────────────
    if not keyword:
        keyword = Prompt.ask("❓ [bold]Qual a palavra-chave alvo?[/bold]")
    console.print(f"🎯 Keyword: [bold green]{keyword}[/bold green]")

    brand_path: str | None = brand
    if not brand_path:
        brand_ask = Prompt.ask(
            "⚙️  [bold]Caminho do Brand Guide[/bold] (Enter para padrão)", default=""
        )
        brand_path = brand_ask if brand_ask.strip() else None

    # ── Progress ──────────────────────────────────────────────────────────
    progress = _emoji_progress()

    with progress:
        # ---- Step 1: SERP -------------------------------------------------
        task1 = progress.add_task("🔍 [1/5] Coletando SERP e PAA...", total=None)
        serp_resp = asyncio.run(search_serp(keyword, top_n=cli_top_n))
        urls = [r.url for r in serp_resp.organic]
        paas = serp_resp.paas

        progress.update(
            task1,
            description=f"🔍 [1/5] SERP coletada — {len(urls)} URLs, {len(paas)} PAA",
            completed=True,
        )

        if not urls:
            console.print(
                "[red]❌ Nenhum resultado encontrado na SERP. "
                "Verifique a keyword ou API key.[/red]"
            )
            raise typer.Exit(code=1)

        # ---- Step 2: Parse ------------------------------------------------
        task2 = progress.add_task("🔬 [2/5] Analisando concorrentes...", total=len(urls))
        competitors = asyncio.run(parse_urls(urls))

        parsed_count = sum(1 for c in competitors if c.text_content)
        progress.update(
            task2,
            description=f"🔬 [2/5] {parsed_count}/{len(urls)} concorrentes analisados",
            completed=True,
        )

        # ---- Step 3: Analyze ----------------------------------------------
        task3 = progress.add_task("🧠 [3/5] Construindo briefing e gaps...", total=None)
        briefing = build_briefing(competitors, keyword, paas=paas)
        progress.update(
            task3,
            description=(
                f"🧠 [3/5] Briefing pronto — "
                f"{len(briefing.proposed_structure)} H2s, "
                f"{len(briefing.content_gaps)} gaps"
            ),
            completed=True,
        )

        # ---- Step 4: Storytell --------------------------------------------
        task4 = progress.add_task("🎭 [4/5] Criando narrativa e storytelling...", total=None)

        try:
            brand_guide = load_brand_guide(brand_path)
        except FileNotFoundError:
            console.print(
                "[yellow]⚠️  Brand guide não encontrado. Usando valores padrão.[/yellow]"
            )
            from lastroseo.models import BrandGuide

            brand_guide = BrandGuide(
                tone="professional",
                persona="SEO expert",
                rules=["Be clear and direct", "Use Brazilian Portuguese"],
            )

        # Product doc (optional)
        product_info = load_product_doc(product_doc)
        if product_info:
            console.print(f"📦 Produto: [bold]{product_info.name}[/bold]")

        # API keys discovery
        api_keys = get_api_keys()
        llm_api_key = api_keys.get("llm_api_key")
        llm_model = api_keys.get("llm_model", model)
        llm_base_url = api_keys.get("llm_base_url")

        if not fallback and llm_api_key:
            story_arc = asyncio.run(
                build_story_arc(
                    briefing, brand_guide, model=llm_model, api_key=llm_api_key,
                    api_base=llm_base_url, product_info=product_info,
                )
            )
        else:
            # Fallback: heuristic arc (no LLM needed)
            story_arc = asyncio.run(
                build_story_arc(briefing, brand_guide, api_key=None, product_info=product_info)
            )

        progress.update(
            task4,
            description=(
                f"🎭 [4/5] Storytelling pronto — "
                f"{len(story_arc.arguments)} argumentos, "
                f"{len(story_arc.section_plan)} seções"
            ),
            completed=True,
        )

        # ---- Step 5: Write ------------------------------------------------
        task5 = progress.add_task("✍️  [5/5] Redigindo artigo...", total=None)

        if fallback or not llm_api_key:
            if not fallback:
                console.print(
                    "[yellow]⚠️  LASTROSEO_LLM_API_KEY não definida. "
                    "Usando template fallback.[/yellow]"
                )
                console.print(
                    "[dim]Defina a env var ou use --fallback explicitamente.[/dim]"
                )
            article_md = generate_fallback_article(briefing, story_arc, product_info)
        else:
            article_md = asyncio.run(
                write_article(
                    briefing,
                    brand_guide,
                    story_arc,
                    model=llm_model,
                    api_key=llm_api_key,
                    api_base=llm_base_url,
                    product_info=product_info,
                )
            )

        progress.update(
            task5,
            description="✍️  [5/5] Artigo gerado com sucesso!",
            completed=True,
        )

    # ── Output ────────────────────────────────────────────────────────────
    bp, sp, ap = _save_output(briefing, story_arc, article_md)

    console.print()
    console.print(Panel.fit("✅ [bold green]Sucesso![/bold green]", border_style="green"))
    console.print(f"📄 Briefing:  [bold]{bp}[/bold]")
    console.print(f"🎭 Story Arc: [bold]{sp}[/bold]")
    console.print(f"📝 Artigo:    [bold]{ap}[/bold]")
    console.print(f"📊 Palavras: [bold]{briefing.target_word_count}[/bold] "
                   f"| Intenção: [bold]{briefing.search_intent}[/bold]")
    console.print(f"🔑 Gaps:     [bold]{len(briefing.content_gaps)}[/bold] "
                   f"| Argumentos: [bold]{len(story_arc.arguments)}[/bold]")
    console.print()


# ── quick info command ───────────────────────────────────────────────────────


@app.command()
def info():
    """Display package info and service status."""
    from lastroseo import __version__

    keys = get_api_keys()

    console.print(f"🚀 [bold]LastroSEO[/bold] v{__version__}")
    console.print()

    # SearXNG status
    searxng_url = keys.get("searxng_url", "http://localhost:8080")
    console.print("Services:")
    console.print(f"  SearXNG: {searxng_url} ", end="")
    try:
        import httpx

        resp = httpx.get(f"{searxng_url}/search?q=test&format=json", timeout=5.0)
        if resp.status_code == 200:
            console.print("[green]✅ conectado[/green]")
        else:
            console.print(f"[yellow]⚠️  HTTP {resp.status_code}[/yellow]")
    except Exception as exc:
        console.print(f"[red]❌ offline — {exc}[/red]")

    console.print(f"  LLM:     {'✅' if keys['llm_api_key'] else '❌'} "
                   f"{'configurada' if keys['llm_api_key'] else 'não definida (LASTROSEO_LLM_API_KEY)'}")
    console.print(f"  Model:   {keys['llm_model']}")


if __name__ == "__main__":
    app()
