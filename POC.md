## O que o Niara faz antes de escrever o artigo

O fluxo deles é dividido em duas etapas bem claras: primeiro eles constroem um **Briefing SEO** (a pauta estruturada), e só depois eles geram o artigo com IA. O briefing é o insumo principal — é o que garante que o texto não seja genérico.

Na prática, ao receber uma palavra-chave, o Niara [[8], [14], [33]]:

1. **Faz um scrape da SERP do Google em tempo real** — lê os sites que estão na primeira página para aquela palavra-chave (geralmente os top 5-10 orgânicos)
2. **Extrai dados estruturais de cada concorrente** (veja a lista abaixo)
3. **Identifica "Content Gaps"** — tópicos relevantes que os concorrentes não estão cobrindo [[14], [22]]
4. **Gera o briefing** com tudo isso estruturado
5. **Só então chama o modelo de linguagem** para escrever o artigo seguindo o briefing + o "Guia da Marca" (tom de voz, persona, regras do cliente) 

---

## Insumos ideais para coletar dos concorrentes (para o seu CLI)

Baseado no que o Niara faz [[14], [33]], aqui está o que você deveria extrair de cada concorrente top-ranking para alimentar a geração do artigo:

| Insumo | Por que importa | Como obter |
|---|---|---|
| **H1, H2, H3 de cada concorrente** | Revela a estrutura lógica que o Google está premiando. O Niara literalmente lista todas as heading tags dos concorrentes . | Parse do HTML das páginas (extração de `<h1>` a `<h3>`) |
| **Title tags e Meta Descriptions** | Entender como os concorrentes "vendem" o clique na SERP . | Meta tags no `<head>` |
| **Contagem de palavras, imagens e vídeos** | Define a extensão e multimídia esperada para ser competitivo . | Contagem simples de tokens/palavras + tags `<img>` e `<video>` |
| **"People Also Ask" (PAA)** | Perguntas que o Google associa àquela busca — excelente para enriquecer o conteúdo e capturar featured snippets [[14], [33]]. | Scrape da seção "As pessoas também perguntam" da SERP, ou API |
| **Palavras-chave secundárias e semânticas** | O Niara sugere 15 palavras secundárias + 15 semânticas para garantir cobertura tópica . | Extração de termos recorrentes no corpus dos concorrentes (TF-IDF simples) ou via API de keyword research |
| **Intenção de busca** | Classificar se é informacional, transacional, navegacional, etc. O Niara já entrega isso [[8], [33]]. | Análise do tipo de conteúdo (artigo longo vs. produto vs. lista) + IA para classificar |
| **Content Gaps** | O diferencial do Niara: identificar lacunas que ninguém está cobrindo mas são relevantes [[14], [22]]. | Comparar tópicos entre concorrentes + validação com modelo de linguagem |

---

## Proposta para a POC em CLI

Para a prova de conceito do LastroSEO, sugiro um pipeline enxuto em 3 passos:

```
lastroseo generate --keyword "automação de atendimento ao cliente"
```

**Passo 1 — SERP Scraper:** busca os top 5-7 resultados orgânicos do Google para a keyword (via SerpAPI, ScrapingBee, ou similar).

**Passo 2 — Briefing Builder:** para cada URL, extrai: headings, word count, imagens, meta tags. Depois agrega tudo em um JSON com:
- Estrutura de H2/H3 proposta (consolidada dos concorrentes + gaps)
- PAA
- Palavras semânticas sugeridas
- Meta title/description sugeridos
- Contagem de palavras alvo (para guiar o tamanho do artigo)

**Passo 3 — Article Writer:** manda o briefing estruturado + a keyword + as regras do "Guia da Marca" (configuradas num YAML ou env vars) para um LLM gerar o artigo final.
