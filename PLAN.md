# 📄 LastroSEO POC: Spec (PRD + SDD)

## 🎯 1. Product (PRD)

**Objetivo**: CLI que pesquisa top 5-7 concorrentes (SERP) para uma *keyword*, extrai estrutura, detecta *Content Gaps* e gera artigo SEO otimizado via LLM.

**Persona**: SEOs, Growth Hackers, Redatores.

**Inputs**: `keyword`, `brand_guide.yaml` (Tom/Voz/Regras), `api_keys` (SERP, LLM).

**Outputs**: `briefing.json`, `artigo_final.md`.

**UX**: CLI interativa, colorida, guiada por emojis. *Fallback* para `Prompt` caso args faltarem.

## 🛠️ 2. System (SDD)

**Stack Core**: `Python 3.12+`, `AsyncIO`.

**Padrão**: Pipeline Assíncrono (Scrape -> Parse -> Analyze -> Generate).

### 📦 3. Stack & Libs (Best-in-Class)

| Domínio | Lib | Motivo |
| :--- | :--- | :--- |
| **CLI / UX** | `Typer` + `Rich` | Roteamento rápido, tabelas, cores, prompts nativos. |
| **HTTP / Async** | `httpx` | Cliente async robusto, suporta HTTP/2, timeouts. |
| **SERP** | `serpapi` / `scrapingbee` | Escapes de bloqueio Google, PAA estruturado. |
| **HTML Parse** | `selectolax` | Parser C-based (Lexbor). 10x+ rápido que BS4. |
| **Limpeza** | `readability-lxml` | Extrai corpo principal, ignora nav/footers. |
| **NLP / Semântica** | `scikit-learn` | `TfidfVectorizer` rápido para gaps e palavras secundárias. |
| **LLM Router** | `LiteLLM` | Interface unificada (OpenAI, Anthropic, Ollama). |
| **Data / Config** | `Pydantic` (v2) + `PyYAML` | Validação estrita, dump JSON, leitura de guia da marca. |

## ⚙️ 4. Módulos e Pipeline

### 🟢 Módulo 1: `cli_io` (Rich + Typer)

- **Ação**: Captura inputs.
- **Lógica**: Lê args (`--keyword`). Se `None`, invoca `Rich.Prompt`. Carrega `brand_guide.yaml`. Exibe progresso (`Rich.Progress`).
- **Emojis/UX**: 🚀 Início, 🔍 Scrape, 🧠 Análise, ✍️ Geração, ✅ Conclusão.

### 🟡 Módulo 2: `serp_scraper`

- **Ação**: Top 5-7 URLs + PAA (People Also Ask).
- **Lógica**: Query na API SERP. Retorna `List[CompetitorData]` + `List[PAAQuestion]`.

### 🟠 Módulo 3: `html_parser`

- **Ação**: Extrai *features* estruturais.
- **Lógica**:
  1. `httpx` GET async em batch (`asyncio.gather`).
  2. `selectolax` para mapear `H1, H2, H3`, `Title`, `MetaDesc`.
  3. `readability` para isolar texto core.
  4. Contagens: `len(text.split())`, tags `<img>`, tags `<video>`.
- **Output**: `CompetitorData` enriquecido.

### 🔴 Módulo 4: `briefing_builder` (Analyzer)

- **Ação**: Consolida dados e acha Gaps.
- **Lógica**:
  1. **Agregação**: Frequência de H2/H3 entre concorrentes.
  2. **TF-IDF**: Extrai 15 palavras-chave secundárias + 15 semânticas (corpus vs keyword).
  3. **Intent**: Classifica via LLM rápido (Informacional/Transacional) baseado no tipo de URL (Blog vs E-comm).
  4. **Gaps**: Compara PAA + Tópicos esperados vs H2s extraídos. O que não está nos concorrentes?
- **Output**: `Briefing` (JSON estruturado).

### 🟣 Módulo 5: `article_writer`

- **Ação**: Gera rascunho final.
- **Lógica**: Monta prompt sistêmico (Brand Guide + Briefing JSON + Rules). Chama LLM via `LiteLLM`.
- **Output**: Markdown formatado (H1, H2, meta tags no topo).

## 🧱 5. Data Models (Pydantic v2)

```python
from pydantic import BaseModel, Field


class CompetitorData(BaseModel):
    url: str
    title: str
    meta_desc: str
    h1: str
    h2_h3_map: dict[str, list[str]]  # {h2: [h3s]}
    word_count: int
    media_count: dict  # {"img": int, "video": int}
    text_content: str  # Para TF-IDF


class Briefing(BaseModel):
    keyword: str
    search_intent: str
    paas: list[str]
    semantic_kws: list[str]
    content_gaps: list[str]
    target_word_count: int
    proposed_structure: list[str]  # Lista de H2s ordenados


class BrandGuide(BaseModel):
    tone: str
    persona: str
    rules: list[str]
```

## 🚀 6. CLI Spec (Exemplo de Execução)

```bash
# Execução Direta
$ lastroseo run --keyword "automação de atendimento" --brand ./brand.yaml

# Execução Interativa (Fallback)
$ lastroseo run
🚀 LastroSEO POC
❓ Qual a palavra-chave alvo? ▸ automação de atendimento
⚙️  Caminho do Brand Guide (Enter para padrão) ▸ 
🔍 [1/3] Coletando SERP e PAA...
🧠 [2/3] Analisando Top 5 Concorrentes e Gaps...
✍️  [3/3] Redigindo artigo (GPT-4o / Claude)...
✅ Sucesso! Arquivos salvos em ./out/briefing.json e ./out/artigo.md
```

## 📝 7. Estrutura de Arquivos

```text
lastroseo/
├── cli.py           # Typer app & Rich prompts
├── models.py        # Pydantic schemas
├── config.py        # YAML & ENV loading
├── core/
│   ├── serp.py      # SERP API integration
│   ├── parser.py    # selectolax & readability
│   ├── analyzer.py  # TF-IDF & Gap logic
│   └── writer.py    # LiteLLM & Prompt builder
└── out/             # Output directory
```
