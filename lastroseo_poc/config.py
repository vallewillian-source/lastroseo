"""Configuration loader — YAML brand guide + environment variables."""

import os
from pathlib import Path

import yaml

from lastroseo_poc.models import BrandGuide

_ENV_PREFIX = "LASTROSEO_"

_DEFAULT_BRAND = Path(__file__).parent / "brand_guide.yaml"


def load_brand_guide(path: str | Path | None = None) -> BrandGuide:
    """Load brand guide from a YAML file.

    Falls back to LASTROSEO_BRAND_GUIDE env var, then to the bundled
    ``brand_guide.yaml`` inside the package.
    """
    resolved = _resolve_path(path)
    if not resolved or not resolved.exists():
        msg = (
            f"Brand guide not found at {resolved}. "
            f"Provide --brand or set {_ENV_PREFIX}BRAND_GUIDE."
        )
        raise FileNotFoundError(msg)

    with open(resolved, encoding="utf-8") as fh:
        raw = yaml.safe_load(fh) or {}

    return BrandGuide(
        tone=raw.get("tone", "professional"),
        persona=raw.get("persona", "SEO expert"),
        rules=raw.get("rules", []),
    )


def get_api_keys() -> dict[str, str | None]:
    """Return API keys and service URLs from environment variables."""
    return {
        "searxng_url": os.getenv(f"{_ENV_PREFIX}SEARXNG_URL", "http://localhost:8080"),
        "llm_api_key": os.getenv(f"{_ENV_PREFIX}LLM_API_KEY"),
        "llm_model": os.getenv(f"{_ENV_PREFIX}LLM_MODEL", "gpt-4o"),
        "llm_base_url": os.getenv(f"{_ENV_PREFIX}LLM_BASE_URL"),
    }


def _resolve_path(path: str | Path | None) -> Path | None:
    if path:
        return Path(path).expanduser().resolve()
    env_path = os.getenv(f"{_ENV_PREFIX}BRAND_GUIDE")
    if env_path:
        return Path(env_path).expanduser().resolve()
    if _DEFAULT_BRAND.exists():
        return _DEFAULT_BRAND
    return None


def load_product_doc(path: str | Path | None = None) -> "ProductInfo | None":
    """Parse a product/service .md file into a ProductInfo model.

    Extracts:
      - ``name`` from the first ``# Heading``
      - ``description`` from the first paragraph after the heading
      - ``value_prop`` from sections mentioning "proposta de valor" or "posicionamento"
      - ``cta`` auto-generated from name + description
      - ``raw_markdown`` — full file content for LLM context
    """
    from lastroseo_poc.models import ProductInfo

    if not path:
        return None

    resolved = Path(path).expanduser().resolve()
    if not resolved.exists():
        return None

    raw = resolved.read_text(encoding="utf-8")

    name = ""
    description = ""
    value_prop = ""

    lines = raw.split("\n")
    in_first_para = False
    para_lines: list[str] = []

    for line in lines:
        stripped = line.strip()

        # Extract name from first # heading
        if not name and stripped.startswith("# ") and not stripped.startswith("## "):
            name = stripped[2:].strip()
            # Remove common suffixes
            for suffix in (" — Documentação", " — Doc", " — Overview", " — Resumo"):
                if suffix in name:
                    name = name[: name.index(suffix)]
            in_first_para = True
            continue

        # Collect first paragraph after heading — skip metadata and separators
        if in_first_para:
            if stripped.startswith("#"):
                in_first_para = False
            elif stripped in ("", "---", "***", "___"):
                if para_lines:
                    in_first_para = False
            elif stripped.startswith(">"):
                # Blockquote — skip for description (usually meta)
                pass
            elif stripped and not stripped.startswith("|") and not stripped.startswith("```"):
                para_lines.append(stripped)
            elif not stripped and para_lines:
                in_first_para = False

    description_candidate = " ".join(para_lines).strip()

    # If first paragraph is meta/boilerplate, try finding synthesis section
    if not description_candidate or len(description_candidate) < 60 or \
       any(w in description_candidate.lower() for w in ("documento interno", "confidenciais")):
        for section_start in ("## 1. Síntese Executiva", "## 1.", "## Overview", "## Resumo", "## Descrição"):
            idx = raw.find(section_start)
            if idx < 0:
                continue
            chunk = raw[idx: idx + 600]
            for cl in chunk.split("\n")[2:]:
                cl = cl.strip()
                if cl and not cl.startswith("#") and not cl.startswith(">") and len(cl) > 40:
                    description_candidate = cl
                    break
            if description_candidate:
                break

    description = description_candidate[:300] if description_candidate else ""

    # Extract value prop — look for positioning sentences
    raw_lower = raw.lower()
    value_prop = ""
    for marker in ("frase de posicionamento", "posicionamento", "diferenciais centrais",
                    "o que o produto entrega", "mensagem central", "proposta de valor"):
        idx = raw_lower.find(marker)
        if idx >= 0:
            # Grab a few lines after the marker heading
            chunk_lines = raw[idx: idx + 600].split("\n")
            for cl in chunk_lines[1:]:
                cl = cl.strip()
                if cl and not cl.startswith("#") and not cl.startswith("-") and len(cl) > 30:
                    value_prop = cl
                    break
            if value_prop:
                break

    # Fallback: use description
    if not value_prop and description:
        value_prop = description

    # Generate CTA
    short_desc = (description or value_prop or f"coordenação de atendimento com IA")[:100]
    short_desc = short_desc.rstrip(".").lower()
    cta = (
        f"Conheça o {name} e descubra como {short_desc}. "
        f"Transforme conversas em fluxos rastreáveis e ganhe controle total da sua operação."
    ) if name else ""

    return ProductInfo(
        name=name,
        description=description,
        value_prop=value_prop,
        cta=cta,
        raw_markdown=raw[:4000],  # Truncate for prompt size
    )
