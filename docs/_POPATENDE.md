# POP Atende — Documentação de Negócio V2

> Documento interno. Contém informações confidenciais de estratégia, produto, arquitetura e operação. Não deve ser compartilhado externamente nem utilizado como peça publicitária sem adaptação.

---

## 1. Síntese Executiva

POP Atende é uma plataforma de automação de fluxos de atendimento coordenados por agentes de IA.

A proposta é clara:

> POP Atende coordena fluxos completos entre seus clientes, funcionários, fornecedores e agentes de IA.

A plataforma permite criar agentes inteligentes, conectar esses agentes a múltiplos canais, dar a eles bases de conhecimento próprias, definir regras de encaminhamento para humanos, associar agendas individuais e, principalmente, configurar fluxos operacionais previsíveis que os agentes conduzem de ponta a ponta.

Cada execução de um fluxo é chamada de **Ficha**. A Ficha é a unidade operacional do POP Atende. Ela mostra o que está acontecendo, em qual etapa o processo está, quais dados já foram coletados, quem precisa agir, quais alertas foram disparados e onde um administrador pode intervir.

Essa mudança reposiciona o POP Atende de uma solução de atendimento automatizado para uma camada de coordenação operacional.

---

## 2. Mudança de Posicionamento

### 2.1 Posicionamento anterior

A V1 descrevia o POP Atende principalmente como um SaaS de atendimento ao cliente via WhatsApp, com agendamento automático, base de conhecimento, agentes de IA e gestão de fichas.

Esse posicionamento fazia sentido na primeira fase. O WhatsApp era o ponto mais visível do produto. O cliente enviava mensagem, o agente respondia, consultava uma base de conhecimento, agendava horários e encaminhava para humanos quando necessário.

O problema desse enquadramento é que ele reduz a percepção de valor do produto.

Quando o POP Atende é vendido como “atendimento pelo WhatsApp com IA”, ele entra em uma categoria lotada: bots, chatbots, CRMs de WhatsApp, plataformas de atendimento, automações simples, ferramentas de disparo e soluções de suporte. A comparação vira preço, canal e resposta automática.

A plataforma já faz mais do que isso.

### 2.2 Novo posicionamento

O novo posicionamento é:

> Plataforma de automação de fluxos de atendimento coordenados por agentes de IA.

Esse enquadramento muda a categoria percebida. O POP Atende deixa de competir apenas com bots de WhatsApp e passa a ocupar um espaço mais estratégico: operações de atendimento que precisam seguir processos, coletar dados, envolver pessoas diferentes, acionar fornecedores, cumprir regras e manter rastreabilidade.

O WhatsApp continua sendo uma porta de entrada forte. Mas o produto não termina nele. Um fluxo pode começar no WhatsApp, continuar no Slack, envolver um atendente humano dentro da plataforma, disparar um alerta por e-mail para um gerente e aguardar a resposta de um fornecedor em outro canal.

O valor está na coordenação.

### 2.3 Frase de posicionamento

Versão curta:

> Automatize fluxos completos de atendimento com agentes de IA.

Versão mais explícita:

> O POP Atende coordena processos entre clientes, funcionários e fornecedores usando agentes de IA, múltiplos canais e fluxos rastreáveis.

Versão comercial:

> Seu atendimento deixa de ser uma sequência de conversas soltas e passa a operar como um fluxo organizado, rastreável e automatizado.

---

## 3. Tese do Produto

Empresas pequenas e médias não sofrem apenas por responder mensagens manualmente. O problema maior é que boa parte da operação acontece dentro de conversas dispersas, sem estado claro, sem dono evidente, sem etapa definida e sem visão consolidada.

Um cliente chama no WhatsApp. Um atendente coleta dados. Outro funcionário precisa aprovar. Um fornecedor precisa responder. Um técnico precisa verificar agenda. Um gerente precisa ser avisado se algo sair do padrão. Parte disso fica no WhatsApp, parte fica na cabeça da equipe, parte fica em planilhas, parte fica em áudios perdidos.

O POP Atende transforma esse caos em fluxo.

A tese central é que o atendimento moderno não deve ser apenas conversacional. Ele precisa ser processual. Conversas são a interface. Fluxos são a estrutura.

A IA do POP Atende não existe apenas para responder perguntas. Ela existe para conduzir processos: entender contexto, coletar informações, decidir a próxima etapa, acionar pessoas, respeitar regras, consultar conhecimento, atualizar fichas e manter a operação avançando.

---

## 4. Proposta de Valor

### 4.1 Mensagem central

> POP Atende coordena seus fluxos de atendimento do início ao fim.

### 4.2 O que o produto entrega

O POP Atende entrega uma camada operacional para empresas que precisam lidar com atendimentos repetitivos, processos previsíveis e múltiplos participantes.

A plataforma permite:

- criar múltiplos agentes de IA diretamente pela UI;
- configurar bases de conhecimento específicas para cada agente;
- conectar agentes a canais como WhatsApp, Slack, Telegram, e-mail e outros canais suportados;
- definir quando o agente deve continuar sozinho e quando deve encaminhar para humanos;
- configurar agendas próprias para agentes e usuários;
- criar fluxos personalizados em canvas visual;
- executar esses fluxos por meio de Fichas rastreáveis;
- envolver clientes, funcionários, fornecedores e agentes no mesmo processo;
- acompanhar cada execução em uma visão administrativa unificada;
- intervir manualmente em qualquer etapa quando necessário.

### 4.3 Valor prático

O cliente compra menos “um robô que responde mensagens” e mais uma operação organizada.

Antes: atendimento espalhado em chats, planilhas e memória operacional.

Depois: processos com etapas claras, responsáveis, dados coletados, histórico, alertas e intervenção administrativa.

Essa diferença deve orientar produto, discurso comercial, landing page, demonstrações e roadmap.

---

## 5. Conceitos Centrais do Produto

### 5.1 Agentes

Agentes são unidades inteligentes configuráveis dentro do POP Atende.

Cada agente pode ter objetivo, instruções, base de conhecimento, canais, agenda, regras de encaminhamento e permissões próprias. O cliente pode criar múltiplos agentes pela interface, sem depender de edição manual de código ou configuração técnica feita pela SofaCoding.

Exemplos:

- agente de recepção;
- agente de suporte técnico;
- agente de triagem comercial;
- agente de pós-venda;
- agente financeiro;
- agente de fornecedores;
- agente interno para apoio ao time operacional.

O design do produto deve evitar a ideia de “um bot genérico da empresa”. Cada agente existe para uma função operacional específica.

### 5.2 Base de conhecimento

Cada agente possui sua própria base de conhecimento.

Essa base pode conter textos, regras, documentos, procedimentos, perguntas frequentes, políticas comerciais, informações técnicas, instruções de atendimento e qualquer material necessário para que o agente execute seu papel com segurança.

A base de conhecimento tem duas funções:

1. permitir que o agente responda dúvidas com base em informação aprovada;
2. orientar decisões operacionais dentro dos fluxos.

Ela não é apenas um FAQ. É parte da memória operacional da empresa.

### 5.3 Canais

Os agentes podem ser conectados a múltiplos canais.

O WhatsApp continua sendo o canal prioritário para o mercado brasileiro, mas o produto deve ser pensado como multicanal desde a estratégia. A inspiração é próxima ao modelo do OpenClaw: agentes conectáveis a diferentes superfícies de conversa e ação.

Canais previstos ou desejados:

- WhatsApp;
- Slack;
- Telegram;
- e-mail;
- webchat;
- canais internos da própria plataforma;
- outros conectores futuros.

A decisão de negócio é importante: o canal não define o produto. O fluxo define o produto.

### 5.4 Encaminhamento humano

O POP Atende permite configurar em quais situações um agente deve encaminhar o atendimento para uma pessoa.

Esse encaminhamento pode acontecer por regra explícita, por incerteza, por solicitação do cliente, por etapa do fluxo ou por condição operacional.

Exemplos:

- cliente demonstra irritação;
- dúvida fora da base de conhecimento;
- solicitação comercial sensível;
- necessidade de aprovação interna;
- problema técnico sem solução automática;
- conflito de agenda;
- valor acima de determinado limite;
- etapa que exige validação humana.

O ponto de produto: o agente não substitui todos os humanos. Ele organiza quando humanos devem entrar.

### 5.5 Agendas

Cada agente e cada usuário pode ter uma agenda própria.

Isso permite que agentes façam agendamentos, consultem disponibilidade, respeitem calendários individuais e coordenem compromissos associados aos fluxos.

A agenda deixa de ser apenas uma funcionalidade de marcação de horários. Ela passa a ser um recurso operacional que pode ser usado dentro dos fluxos.

Exemplos:

- marcar visita técnica;
- agendar consulta;
- reservar horário com especialista;
- bloquear agenda de um agente;
- organizar disponibilidade de um fornecedor;
- acionar um profissional apenas quando houver janela disponível.

---

## 6. Fluxos — A Funcionalidade Central

### 6.1 Definição

Fluxos são processos personalizados e previsíveis criados dentro da UI do POP Atende.

Um fluxo define as etapas de um atendimento ou operação, as saídas possíveis de cada etapa, as regras de negócio, as condicionais, os alertas, os participantes envolvidos e os caminhos que uma execução pode seguir.

Fluxos substituem e expandem o conceito anterior de “templates de fichas”.

Na V1, a ideia ainda estava próxima de modelos pré-configurados mantidos pela SofaCoding. Na V2, o fluxo vira um ativo gerenciável pelo próprio cliente, editado visualmente em um canvas e usado pelos agentes para coordenar a operação.

### 6.2 Decisão estratégica

Fluxos são a estrela do POP Atende.

Eles devem ocupar o centro da narrativa de produto, da demonstração comercial e da evolução técnica. Agentes, canais, agendas e base de conhecimento são componentes poderosos, mas ficam mais fortes quando conectados a fluxos.

O diferencial não é “ter IA no atendimento”. O diferencial é ter IA conduzindo processos reais, com controle administrativo e rastreabilidade.

### 6.3 O que um fluxo pode coordenar

Um fluxo pode envolver:

- clientes externos;
- funcionários internos;
- gerentes;
- atendentes humanos;
- agentes de IA;
- fornecedores externos;
- técnicos de campo;
- profissionais com agenda própria;
- departamentos;
- canais diferentes;
- alertas automáticos;
- decisões condicionais;
- ações manuais de administradores.

O fluxo pode atravessar múltiplas conversas e múltiplos canais.

Exemplo simples:

1. cliente entra em contato pelo WhatsApp;
2. agente coleta dados iniciais;
3. agente consulta base de conhecimento;
4. agente identifica necessidade de avaliação técnica;
5. agente aciona suporte interno pelo Slack;
6. funcionário responde com orientação;
7. agente agenda visita na agenda de um técnico;
8. se o cliente for prioritário, gerente recebe alerta por e-mail;
9. Ficha registra tudo;
10. administrador acompanha e pode intervir.

### 6.4 Diferença entre conversa e fluxo

Uma conversa é uma sequência de mensagens.

Um fluxo é uma estrutura de trabalho.

O POP Atende usa conversas como interface, mas registra o trabalho como fluxo. Essa distinção deve aparecer no produto e no discurso comercial. A empresa não precisa apenas saber o que foi dito. Ela precisa saber o que já foi feito, o que falta fazer, quem é responsável e o que acontece se determinada condição for satisfeita.

---

## 7. Editor Visual de Fluxos

### 7.1 Visão geral

O editor de fluxos é uma tela baseada em canvas, semelhante em espírito a ferramentas como n8n. O administrador cria e organiza componentes visuais, conecta saídas, define regras e estrutura a lógica operacional.

O canvas deve transmitir uma ideia clara: o cliente está desenhando o processo da empresa, não programando um chatbot.

### 7.2 Componentes principais

O editor trabalha com três tipos principais de componentes:

1. etapas;
2. condicionais;
3. alertas.

Esses componentes podem ser conectados livremente conforme as regras do produto.

### 7.3 Etapas

Etapas são os blocos principais do fluxo.

Cada etapa representa um momento operacional: coletar dados, aguardar uma resposta, pedir aprovação, executar uma ação, agendar um horário, validar informações, solicitar documentos ou encaminhar uma demanda.

Em cada etapa, o administrador pode definir:

- nome da etapa;
- descrição operacional;
- participante responsável;
- instruções para o agente;
- dados que devem ser coletados;
- critérios de conclusão;
- saídas possíveis;
- mensagens associadas;
- ações disponíveis;
- próximos destinos.

Uma etapa não precisa ser executada apenas por IA. Ela pode depender de um humano, de um fornecedor, de um cliente ou de combinação entre participantes.

### 7.4 Saídas de etapa

Cada etapa pode ter uma ou mais saídas.

Uma saída representa o resultado possível daquela etapa. O administrador pode criar saídas como:

- aprovado;
- recusado;
- aguardando cliente;
- precisa de humano;
- visita agendada;
- dados incompletos;
- orçamento aceito;
- problema resolvido;
- encaminhar para suporte;
- escalar para gerente.

Cada saída pode ser ligada a outra etapa, uma condicional ou um alerta.

Esse modelo permite que o fluxo seja previsível sem ser rígido. O processo tem caminhos definidos, mas o agente pode escolher o caminho correto conforme o contexto e as regras configuradas.

### 7.5 Condicionais

Condicionais são componentes em formato de losango.

Elas representam pontos de decisão no fluxo. Cada condicional possui uma regra específica que o agente deve considerar para bifurcar o atendimento em dois caminhos possíveis.

Exemplos de regras:

- se o cliente já for cadastrado, seguir para agendamento; caso contrário, coletar cadastro;
- se o valor estimado for maior que R$ 1.000, solicitar aprovação do gerente;
- se a solicitação envolver garantia, encaminhar para suporte especializado;
- se o fornecedor não responder em determinado prazo, disparar alerta;
- se o cliente tiver urgência, priorizar agenda mais próxima;
- se houver documentação incompleta, retornar para coleta de dados.

As saídas de uma condicional podem levar a:

- outra etapa;
- outra condicional;
- um alerta.

A condicional deve ser tratada como lógica de negócio configurável. Não é um if técnico exposto ao usuário. É uma regra operacional descrita em linguagem compreensível.

### 7.6 Alertas

Alertas são componentes dead-end.

Eles podem ser conectados a qualquer saída de etapa ou condicional. Quando acionados, executam uma comunicação automática e encerram aquele ramo específico do fluxo.

Existem dois tipos iniciais de alerta:

- alerta por WhatsApp;
- alerta por e-mail.

Ao configurar um alerta, o administrador define:

- canal do alerta;
- destinatário;
- template da mensagem;
- variáveis usadas no texto;
- condição de disparo, quando aplicável;
- identificação da Ficha associada.

Os templates podem usar dados variáveis da execução.

Exemplo de template:

```text
Atenção: a ficha {{ficha.codigo}} do cliente {{cliente.nome}} foi escalada para aprovação.
Motivo: {{etapa.motivo_escalacao}}
Valor estimado: {{orcamento.valor}}
```

Alertas não substituem etapas de trabalho. Eles servem para notificação, escalada e visibilidade.

### 7.7 Regras de conexão

A lógica visual deve permitir liberdade suficiente para representar processos reais, mas com restrições para evitar fluxos inválidos.

Regras recomendadas:

- etapas podem se conectar a etapas, condicionais ou alertas;
- condicionais podem se conectar a etapas, outras condicionais ou alertas;
- alertas são dead-end;
- uma etapa pode ter múltiplas saídas;
- cada saída deve ter destino explícito;
- o fluxo deve ter um ponto inicial;
- o fluxo deve ter pelo menos um caminho de conclusão;
- componentes sem conexão devem ser destacados como incompletos;
- loops devem ser permitidos com cuidado, quando fizerem sentido operacional.

---

## 8. Fichas — Execuções dos Fluxos

### 8.1 Definição

Uma Ficha é uma execução concreta de um fluxo.

Se o fluxo é o desenho do processo, a Ficha é o processo acontecendo.

Exemplo:

- Fluxo: “Solicitação de visita técnica”.
- Ficha: “Visita técnica do cliente João Silva, aberta em 17/05/2026, atualmente aguardando confirmação do técnico”.

A Ficha é o objeto operacional mais importante para administradores.

### 8.2 O que uma Ficha mostra

A Ficha deve mostrar, com clareza:

- fluxo de origem;
- cliente ou entidade principal;
- etapa atual;
- status geral;
- participantes envolvidos;
- histórico de eventos;
- conversas relacionadas;
- dados coletados por etapa;
- decisões tomadas pelo agente;
- alertas disparados;
- ações pendentes;
- próximos passos possíveis;
- responsável atual;
- agenda relacionada, quando houver;
- anexos ou documentos associados;
- intervenções manuais realizadas.

A Ficha deve funcionar como uma visão de controle. Ela precisa responder rapidamente: o que está acontecendo, por que está parado, quem precisa agir e o que pode ser feito agora.

### 8.3 Dados coletados por etapa

Cada etapa pode produzir informações importantes.

Esses dados não devem ficar perdidos em mensagens. Devem ser extraídos, estruturados e exibidos na Ficha.

Exemplos:

- nome do cliente;
- telefone;
- endereço;
- tipo de serviço;
- urgência;
- descrição do problema;
- orçamento estimado;
- profissional responsável;
- horário agendado;
- documentos enviados;
- parecer do suporte;
- motivo de recusa;
- aprovação do gerente;
- observações internas.

A decisão de produto aqui é forte: conversa gera dado operacional. O POP Atende deve capturar esse dado e torná-lo visível.

### 8.4 Intervenção manual

Administradores podem executar manualmente ações no lugar de qualquer participante do fluxo.

Isso mantém controle total sobre a operação. A IA coordena, mas a empresa não fica refém dela.

Exemplos de intervenção:

- avançar uma etapa;
- corrigir um dado coletado;
- escolher uma saída diferente;
- reenviar alerta;
- atribuir responsável;
- responder pelo agente;
- executar ação que seria de um fornecedor;
- encerrar uma Ficha;
- reabrir uma etapa;
- ajustar horário de agenda.

Toda intervenção deve ser registrada no histórico da Ficha.

### 8.5 Visão administrativa

Administradores devem ter uma visão geral de todas as Fichas.

Essa tela precisa funcionar como painel operacional da empresa. Não é apenas uma lista. Deve permitir filtros, agrupamentos e leitura rápida.

Filtros desejáveis:

- por fluxo;
- por status;
- por etapa atual;
- por responsável;
- por agente;
- por cliente;
- por canal;
- por atraso;
- por prioridade;
- por data de criação;
- por data de atualização;
- por alerta disparado.

Agrupamentos úteis:

- Fichas abertas;
- Fichas atrasadas;
- Fichas aguardando humano;
- Fichas aguardando cliente;
- Fichas aguardando fornecedor;
- Fichas concluídas;
- Fichas com intervenção recente.

A visão de Fichas deve ser vendida como “controle da operação”.

---

## 9. Coordenação Multicanal

### 9.1 Princípio

O POP Atende deve operar sobre múltiplos canais sem perder o estado do processo.

A empresa pode falar com o cliente pelo WhatsApp, com a equipe pelo Slack, com o gerente por e-mail e com um fornecedor por Telegram, mantendo tudo conectado à mesma Ficha.

Esse é um ponto de diferenciação. Muitas ferramentas têm canais. Poucas mantêm um processo coordenado entre canais.

### 9.2 Exemplo de fluxo multicanal

Cenário: assistência técnica residencial.

1. Cliente chama no WhatsApp informando problema no ar-condicionado.
2. Agente coleta endereço, modelo do aparelho e sintomas.
3. Agente consulta base de conhecimento e identifica possível tipo de atendimento.
4. Fluxo cria uma Ficha de visita técnica.
5. Agente consulta agenda dos técnicos.
6. Agente aciona time interno no Slack para confirmar disponibilidade de peça.
7. Se peça estiver indisponível, agente dispara alerta para fornecedor externo.
8. Se o cliente for premium, gerente recebe e-mail.
9. Técnico confirma horário.
10. Agente confirma visita com cliente pelo WhatsApp.
11. Ficha fica visível para administrador com todos os eventos.

O cliente vê uma conversa simples. A empresa vê um fluxo coordenado.

### 9.3 Conversas vinculadas

Uma Ficha pode ter várias conversas associadas.

Exemplos:

- conversa com cliente;
- conversa com atendente;
- conversa com suporte interno;
- conversa com fornecedor;
- conversa de aprovação gerencial;
- thread de Slack;
- alerta por e-mail;
- conversa de acompanhamento pós-atendimento.

A UI deve permitir que o administrador transite entre essas conversas sem perder o contexto da Ficha.

---

## 10. Experiência de Chat

### 10.1 Chat como superfície operacional

O POP Atende possui uma experiência de chat completa.

Isso continua importante. A diferença é que o chat não é mais a categoria principal do produto. Ele é uma superfície de execução dos fluxos.

Administradores e atendentes podem visualizar, responder e intervir em conversas conectadas aos agentes e às Fichas.

### 10.2 Intervenção em conversas

Admins podem visualizar e intervir em todas as conversas.

Isso inclui:

- assumir uma conversa;
- responder como humano;
- orientar o agente;
- corrigir rota de atendimento;
- mover conversa para outro responsável;
- vincular conversa a uma Ficha;
- abrir uma nova Ficha a partir da conversa;
- encerrar atendimento;
- consultar histórico do cliente.

A experiência deve deixar claro quando o atendimento está com IA, com humano ou em modo híbrido.

### 10.3 Histórico do cliente

O histórico do cliente deve consolidar conversas, Fichas, agendamentos e dados relevantes.

A empresa precisa saber não apenas “o que ele falou”, mas quais processos já aconteceram com aquele cliente.

---

## 11. Agendas

### 11.1 Agendas por agente e usuário

Cada agente e cada usuário pode ter sua própria agenda.

Administradores podem visualizar agendas dos agentes e de qualquer usuário do sistema.

Isso permite que a operação seja coordenada sem depender de calendários externos dispersos ou planilhas paralelas.

### 11.2 Agendas dentro dos fluxos

Fluxos podem incluir etapas de agenda.

Exemplos:

- selecionar profissional disponível;
- reservar horário;
- reagendar atendimento;
- confirmar presença;
- bloquear agenda;
- aguardar data futura para avançar etapa;
- disparar lembrete antes do compromisso;
- avançar automaticamente após horário agendado.

A agenda deve ser entendida como parte do motor de fluxo.

### 11.3 Visão administrativa de agendas

Admins devem conseguir visualizar:

- agenda de cada agente;
- agenda de cada usuário;
- compromissos associados a Fichas;
- disponibilidade;
- conflitos;
- histórico de alterações;
- responsáveis por cada compromisso.

Essa visão reforça o POP Atende como sistema operacional de atendimento, não apenas como inbox.

---

## 12. Gestão de Agentes pela UI

### 12.1 Criação de múltiplos agentes

A plataforma permite criar múltiplos agentes de IA pela interface.

Essa é uma mudança relevante em relação à lógica inicial, em que a criação de agentes era mais controlada pela SofaCoding. Na V2, a direção de produto é dar ao administrador mais autonomia.

Cada agente deve possuir:

- nome;
- descrição;
- função operacional;
- instruções principais;
- canais conectados;
- base de conhecimento;
- agenda própria;
- regras de encaminhamento;
- fluxos que pode iniciar ou coordenar;
- permissões;
- status de ativação.

### 12.2 Padrão moderno de construção de agentes

Os agentes seguem padrões modernos de construção:

- raciocínio orientado a ferramentas;
- instruções explícitas de função;
- uso de base de conhecimento;
- memória operacional limitada e controlada;
- execução de ações dentro do sistema;
- integração com canais externos;
- coordenação com humanos;
- participação em fluxos estruturados.

O agente não é apenas um prompt. É uma entidade operacional com ferramentas, contexto e responsabilidades.

### 12.3 Relação entre agente e fluxo

Um agente pode:

- iniciar uma Ficha;
- avançar uma etapa;
- escolher uma saída;
- consultar uma condicional;
- acionar um alerta;
- conversar com participantes;
- atualizar dados coletados;
- consultar agenda;
- encaminhar para humano;
- pausar execução quando faltar informação;
- pedir aprovação.

A IA coordena o fluxo, mas o fluxo limita e orienta a IA. Essa combinação é parte central do produto: autonomia com trilhos.

---

## 13. Base de Conhecimento

### 13.1 Propósito

A base de conhecimento dá substância ao agente.

Ela reduz improviso, padroniza respostas e permite que cada agente opere com informações específicas da empresa.

Pode conter:

- textos livres;
- documentos;
- regras internas;
- políticas comerciais;
- procedimentos técnicos;
- materiais de treinamento;
- catálogos de serviço;
- FAQs;
- scripts de atendimento;
- instruções de triagem;
- normas de encaminhamento.

### 13.2 Base por agente

Cada agente tem sua própria base.

Isso evita mistura de contexto e permite especialização. Um agente financeiro não precisa consultar a mesma base do suporte técnico. Um agente de fornecedores não precisa responder como recepção.

### 13.3 Relação com fluxos

A base de conhecimento também pode orientar decisões dentro dos fluxos.

Exemplo: em uma etapa de triagem, o agente pode consultar uma regra técnica para decidir se o caso exige visita presencial, suporte remoto ou encaminhamento humano.

---

## 14. Encaminhamento para Humanos

### 14.1 Princípio

O POP Atende não deve vender uma fantasia de automação total.

O valor está em automatizar o que é previsível e estruturar a entrada humana quando ela for necessária.

### 14.2 Critérios de encaminhamento

Encaminhamentos podem ocorrer por:

- regra configurada;
- baixa confiança;
- solicitação explícita do cliente;
- tipo de etapa;
- perfil do cliente;
- criticidade;
- valor financeiro;
- atraso;
- exceção operacional;
- falha de canal;
- necessidade de aprovação.

### 14.3 Encaminhamento dentro de fluxos

O encaminhamento humano não deve ser visto apenas como “transferir conversa”.

Ele pode ser uma etapa formal do fluxo.

Exemplo:

- etapa: “Aprovação do gerente”;
- responsável: gerente humano;
- saída 1: aprovado;
- saída 2: recusado;
- saída 3: solicitar mais informações;
- alerta: enviar e-mail se pendente por mais de 24 horas.

Esse modelo aumenta previsibilidade e reduz abandono operacional.

---

## 15. Público-Alvo e Casos de Uso

### 15.1 Perfil de empresa

O POP Atende é indicado para empresas que:

- recebem alto volume de solicitações repetitivas;
- operam com processos previsíveis;
- dependem de WhatsApp ou chat como porta de entrada;
- envolvem múltiplos funcionários no atendimento;
- precisam acionar fornecedores ou terceiros;
- trabalham com agendas;
- perdem controle operacional em conversas soltas;
- usam planilhas ou grupos para acompanhar demandas;
- precisam de rastreabilidade sem implantar um sistema pesado.

### 15.2 Verticais prioritárias

#### Serviços técnicos e residenciais

Um dos melhores fits para fluxos.

Exemplos:

- assistência técnica;
- manutenção residencial;
- instalação;
- limpeza;
- reformas;
- climatização;
- suporte em campo.

Fluxos comuns:

- triagem de problema;
- coleta de endereço;
- validação de disponibilidade;
- agendamento de visita;
- aprovação de orçamento;
- acionamento de fornecedor;
- confirmação de execução;
- pós-atendimento.

#### Saúde e bem-estar

Forte fit com agenda, triagem e pós-atendimento.

Exemplos:

- clínicas médicas;
- odontologia;
- fisioterapia;
- psicologia;
- veterinária;
- estética avançada.

Fluxos comuns:

- triagem inicial;
- agendamento;
- confirmação de consulta;
- coleta de documentos;
- orientação pré-atendimento;
- acompanhamento pós-consulta;
- encaminhamento administrativo.

#### Beleza e estética

Fit em operações com agenda intensa e relacionamento recorrente.

Exemplos:

- salões;
- barbearias;
- clínicas estéticas;
- spas;
- studios especializados.

Fluxos comuns:

- agendamento;
- reagendamento;
- confirmação;
- indicação de procedimento;
- pacote recorrente;
- pós-atendimento.

#### B2B operacional

Empresas que coordenam atendimento com fornecedores e equipes internas.

Exemplos:

- distribuidoras;
- pequenos ERPs verticais;
- suporte técnico B2B;
- logística local;
- facilities;
- empresas com rede de parceiros.

Fluxos comuns:

- abertura de solicitação;
- validação interna;
- acionamento de fornecedor;
- aprovação de orçamento;
- acompanhamento de SLA;
- escalada gerencial.

### 15.3 Perfil do decisor

Prioridade comercial:

| Prioridade | Decisor |
|---|---|
| 1 | Sócio, founder ou dono da operação |
| 2 | Diretor ou gerente operacional |
| 3 | Coordenador de atendimento |
| 4 | Responsável administrativo |
| 5 | Head comercial ou pós-venda |

A venda deve falar com quem sente a desorganização do processo, não apenas com quem responde mensagens.

---

## 16. Narrativa Comercial

### 16.1 O problema

Empresas operacionais não perdem dinheiro apenas por demorar a responder.

Elas perdem dinheiro porque atendimentos viram fios soltos:

- ninguém sabe em que etapa está;
- dados precisam ser perguntados de novo;
- gerente só descobre problema tarde;
- fornecedor responde fora do contexto;
- cliente cobra atualização;
- funcionário esquece follow-up;
- informação importante fica presa em conversa.

### 16.2 A solução

O POP Atende cria fluxos claros para esses atendimentos.

A IA conversa, coleta dados, consulta conhecimento, aciona pessoas, agenda horários, dispara alertas e mantém uma Ficha atualizada. Humanos entram quando necessário, com contexto completo.

### 16.3 Demonstração ideal

A demo deve mostrar menos “olha o bot respondendo” e mais “olha o processo andando”.

Roteiro recomendado:

1. apresentar um fluxo visual no canvas;
2. mostrar etapas, condicionais e alertas;
3. simular entrada de cliente pelo WhatsApp;
4. mostrar agente coletando dados;
5. abrir a Ficha criada;
6. mostrar etapa atual e dados coletados;
7. disparar uma condicional;
8. acionar um alerta por e-mail ou WhatsApp;
9. mostrar intervenção manual do admin;
10. mostrar agenda associada;
11. concluir o fluxo.

A mensagem final da demo: o cliente não está comprando respostas automáticas. Está comprando controle operacional.

---

## 17. Arquitetura de Produto

### 17.1 Componentes principais

O POP Atende combina:

- backend de agentes;
- engine de fluxos;
- sistema de Fichas;
- base de conhecimento;
- conectores de canais;
- chat operacional;
- calendários;
- painel administrativo;
- editor visual de fluxos;
- camada de permissões;
- armazenamento de eventos e histórico.

### 17.2 Engine de fluxos

A engine de fluxos é responsável por:

- interpretar definição do fluxo;
- criar Fichas;
- controlar etapa atual;
- validar saídas;
- processar condicionais;
- disparar alertas;
- registrar eventos;
- permitir intervenção manual;
- manter histórico;
- conectar conversas e ações à Ficha;
- expor estado para a UI;
- fornecer contexto para os agentes.

A engine deve ser tratada como núcleo do produto.

### 17.3 Agentes e ferramentas

Os agentes devem acessar ferramentas internas para operar o sistema.

Ferramentas possíveis:

- consultar base de conhecimento;
- criar Ficha;
- atualizar Ficha;
- avançar etapa;
- escolher saída;
- consultar fluxo;
- enviar mensagem;
- criar alerta;
- consultar agenda;
- criar agendamento;
- encaminhar para humano;
- registrar dado coletado;
- pedir intervenção.

A arquitetura ideal separa bem decisão, ação e registro. O agente pode decidir, mas o sistema valida e registra.

### 17.4 Chat e Matrix

A base de chat pode continuar apoiada em Matrix/Synapse e na experiência derivada do Element, com customizações para a operação do POP Atende.

O chat deve servir a três funções:

1. comunicação com clientes e participantes;
2. superfície de intervenção humana;
3. histórico conversacional vinculado às Fichas.

### 17.5 Conectores

A camada de conectores deve permitir expansão progressiva.

WhatsApp via EvolutionAPI continua sendo prioridade inicial. Slack, Telegram, e-mail e outros canais entram como expansão natural do posicionamento multicanal.

O produto deve evitar acoplamento conceitual com WhatsApp. A arquitetura pode começar por ele, mas a documentação e o modelo mental devem ser multicanais.

---

## 18. Modelo de Dados Conceitual

### 18.1 Entidades principais

Entidades de negócio:

- Organização;
- Usuário;
- Agente;
- Canal;
- Integração;
- Base de conhecimento;
- Documento;
- Fluxo;
- Etapa;
- Saída;
- Condicional;
- Alerta;
- Ficha;
- Evento da Ficha;
- Conversa;
- Participante;
- Agenda;
- Agendamento;
- Permissão.

### 18.2 Fluxo

Um Fluxo contém:

- nome;
- descrição;
- versão;
- status;
- etapa inicial;
- componentes;
- conexões;
- regras;
- variáveis disponíveis;
- permissões de uso;
- agentes autorizados;
- metadados de criação e edição.

Fluxos devem ser versionados. Uma Ficha em andamento não deve quebrar porque o administrador editou o fluxo original.

### 18.3 Ficha

Uma Ficha contém:

- fluxo e versão de origem;
- estado atual;
- cliente ou entidade principal;
- participantes;
- dados coletados;
- eventos;
- conversas vinculadas;
- agenda vinculada;
- alertas disparados;
- status;
- responsável atual;
- timestamps;
- histórico de intervenção.

### 18.4 Evento

Eventos são registros imutáveis do que aconteceu na Ficha.

Exemplos:

- Ficha criada;
- etapa iniciada;
- mensagem recebida;
- dado coletado;
- saída escolhida;
- condicional avaliada;
- alerta disparado;
- humano assumiu;
- agente respondeu;
- agendamento criado;
- etapa concluída;
- Ficha encerrada.

A linha do tempo de eventos é fundamental para auditoria e confiança.

---

## 19. Permissões e Papéis

### 19.1 Papéis principais

Papéis sugeridos:

| Papel | Função |
|---|---|
| Admin | Controla organização, fluxos, agentes, usuários, agendas e Fichas |
| Manager | Acompanha operação, intervém em Fichas e gerencia parte do time |
| Operador | Atua em conversas e etapas atribuídas |
| Agente de IA | Executa ações conforme permissões e ferramentas disponíveis |
| Participante externo | Cliente, fornecedor ou terceiro envolvido em uma Ficha |

### 19.2 Permissões relevantes

Permissões devem cobrir:

- criar agente;
- editar agente;
- conectar canal;
- editar base de conhecimento;
- criar fluxo;
- editar fluxo;
- publicar fluxo;
- pausar fluxo;
- visualizar Fichas;
- intervir em Fichas;
- executar ações manuais;
- visualizar conversas;
- assumir conversa;
- visualizar agendas;
- editar agendas;
- gerenciar usuários;
- configurar alertas.

A regra geral: quanto mais o POP Atende vira sistema operacional, mais permissões importam.

---

## 20. Onboarding

### 20.1 Onboarding ideal

O onboarding deve levar o cliente a um primeiro fluxo funcional, não apenas a um agente respondendo mensagens.

Etapas recomendadas:

1. mapear processo principal do cliente;
2. identificar canais usados;
3. definir participantes;
4. criar primeiro agente;
5. carregar base de conhecimento;
6. configurar canal inicial, geralmente WhatsApp;
7. desenhar primeiro fluxo no canvas;
8. configurar alertas;
9. configurar agendas;
10. executar teste interno;
11. publicar fluxo;
12. acompanhar primeiras Fichas reais.

### 20.2 Primeiro fluxo recomendado

O primeiro fluxo deve ser pequeno e de alto valor.

Exemplos:

- agendamento de visita técnica;
- triagem de novo cliente;
- abertura de solicitação;
- confirmação de consulta;
- aprovação de orçamento;
- suporte inicial com escalada humana.

Evitar começar com o fluxo mais complexo da empresa. O objetivo inicial é provar coordenação, visibilidade e controle.

---

## 21. Modelo de Negócio

### 21.1 Fontes de receita

Fontes principais:

- assinatura mensal ou anual;
- cobrança por número de agentes;
- cobrança por volume de execuções de Fichas;
- cobrança por canais conectados;
- cobrança por usuários internos;
- implantação assistida;
- criação de fluxos personalizados;
- integrações sob demanda;
- consultoria operacional.

A V2 abre espaço para precificação mais defensável que a V1. Bots de WhatsApp sofrem pressão de preço. Automação de fluxos operacionais permite ticket maior.

### 21.2 Direção de pricing

Possível estrutura futura:

| Plano | Perfil | Limites prováveis |
|---|---|---|
| Starter | Pequenas operações | 1–2 agentes, poucos fluxos, WhatsApp, volume limitado de Fichas |
| Growth | Operações em expansão | múltiplos agentes, mais fluxos, agendas, alertas, canais adicionais |
| Operations | Empresas com operação mais complexa | alto volume de Fichas, permissões avançadas, múltiplos canais, suporte prioritário |
| Custom | Casos sob medida | integrações, fluxos avançados, implantação consultiva |

A cobrança por Fichas pode ser explorada com cuidado. Ela aproxima preço de valor operacional entregue, mas precisa ser simples para não gerar ansiedade no cliente.

### 21.3 Serviços profissionais

A criação de fluxos pode virar frente importante de receita.

Muitas PMEs não sabem desenhar bem seus processos. O POP Atende pode vender implantação assistida como vantagem, não como fricção.

Serviços possíveis:

- desenho de fluxo;
- revisão de processo;
- implantação de base de conhecimento;
- criação de agentes especializados;
- integração com sistemas externos;
- treinamento de equipe;
- otimização de operação após uso real.

---

## 22. Estratégia de Produto

### 22.1 Princípios

Princípios de produto:

- fluxo acima de canal;
- Ficha acima de conversa;
- humano no controle;
- IA com trilhos claros;
- dados operacionais extraídos da conversa;
- multicanal por arquitetura, não por discurso vazio;
- editor visual como ativo estratégico;
- intervenção manual sempre disponível;
- rastreabilidade como fonte de confiança.

### 22.2 O que priorizar

Prioridades da V2:

1. editor visual de fluxos;
2. Fichas ricas e operacionais;
3. criação e gestão de agentes pela UI;
4. base de conhecimento por agente;
5. integração com WhatsApp estável;
6. agendas por agente e usuário;
7. alertas por WhatsApp e e-mail;
8. intervenção administrativa;
9. suporte progressivo a Slack e Telegram;
10. visão consolidada de conversas vinculadas a Fichas.

### 22.3 O que evitar

Evitar dispersão em funcionalidades genéricas de CRM antes de consolidar o núcleo.

O POP Atende não precisa virar um CRM completo no curto prazo. Também não precisa competir em disparo de mensagens, funil comercial tradicional ou help desk genérico.

A força está em fluxos coordenados por IA.

---

## 23. Roadmap V2

### 23.1 Núcleo obrigatório

- CRUD de agentes pela UI;
- base de conhecimento por agente;
- conexão de canais por agente;
- regras de encaminhamento humano;
- agenda por agente e usuário;
- editor visual de fluxos;
- componentes de etapa, condicional e alerta;
- execução de fluxos como Fichas;
- visão geral de Fichas;
- detalhe rico da Ficha;
- intervenção manual em Fichas;
- visualização de conversas vinculadas;
- alertas por WhatsApp e e-mail.

### 23.2 Expansões naturais

- Slack como canal operacional interno;
- Telegram como canal externo ou alternativo;
- e-mail como canal bidirecional;
- templates prontos por vertical;
- métricas de gargalo por etapa;
- SLA por fluxo;
- automações temporais;
- biblioteca de variáveis para alertas;
- permissões avançadas;
- versionamento visual de fluxos;
- marketplace interno de conectores;
- analytics de Fichas e conversas;
- API pública para integrações.

### 23.3 Métricas de produto

Métricas relevantes:

- número de fluxos criados por organização;
- número de Fichas abertas;
- taxa de conclusão de Fichas;
- tempo médio por etapa;
- etapas com maior gargalo;
- percentual de Fichas com intervenção humana;
- percentual de Fichas concluídas sem intervenção;
- alertas disparados;
- canais usados por fluxo;
- agentes ativos;
- uso da base de conhecimento;
- agendamentos criados;
- conversas vinculadas por Ficha.

Essas métricas contam a história certa: operação, não apenas atendimento.

---

## 24. Concorrência e Diferenciação

### 24.1 Categorias concorrentes

O POP Atende pode ser comparado com diferentes categorias:

- bots de WhatsApp;
- plataformas de atendimento;
- CRMs conversacionais;
- help desks;
- ferramentas de automação como n8n;
- plataformas de agentes de IA;
- sistemas verticais de agendamento;
- softwares operacionais específicos por nicho.

A estratégia não deve aceitar comparação direta com bots simples. Quando isso acontecer, a venda deve voltar para fluxos, Fichas e coordenação multiator.

### 24.2 Diferenciais

Diferenciais centrais:

- agentes de IA configuráveis pela UI;
- base de conhecimento por agente;
- fluxos visuais personalizados;
- Fichas como execuções rastreáveis;
- coordenação entre clientes, funcionários e fornecedores;
- múltiplos canais dentro do mesmo processo;
- agendas por agente e usuário;
- intervenção humana completa;
- alertas configuráveis com variáveis;
- visão operacional unificada.

### 24.3 Defesa estratégica

A defesa do produto vem da combinação, não de uma feature isolada.

Um concorrente pode ter WhatsApp. Outro pode ter IA. Outro pode ter automação visual. Outro pode ter agenda. O POP Atende combina esses elementos em torno de uma unidade operacional própria: a Ficha.

A Ficha é o ponto de amarração do produto.

---

## 25. Exemplos de Fluxos

### 25.1 Visita técnica residencial

Etapas:

1. atendimento inicial;
2. coleta de dados do cliente;
3. diagnóstico preliminar;
4. condicional: exige visita presencial?;
5. consulta de agenda;
6. confirmação com técnico;
7. confirmação com cliente;
8. alerta para gerente se cliente premium;
9. execução da visita;
10. fechamento e feedback.

Participantes:

- cliente pelo WhatsApp;
- agente de IA;
- técnico;
- suporte interno;
- gerente.

### 25.2 Clínica odontológica

Etapas:

1. recepção do cliente;
2. identificação do tipo de consulta;
3. coleta de dados básicos;
4. condicional: primeira consulta ou retorno?;
5. escolha de profissional;
6. agendamento;
7. confirmação automática;
8. alerta se houver cancelamento;
9. pós-consulta;
10. fechamento administrativo.

Participantes:

- paciente;
- agente de recepção;
- atendente;
- dentista;
- administrativo.

### 25.3 Aprovação de orçamento

Etapas:

1. cliente solicita orçamento;
2. agente coleta escopo;
3. suporte valida informações;
4. condicional: valor acima do limite?;
5. se sim, gerente aprova;
6. se não, agente envia proposta;
7. cliente aceita ou recusa;
8. agenda execução;
9. Ficha encerrada.

Alertas:

- e-mail para gerente em orçamentos altos;
- WhatsApp para vendedor se cliente aceitar;
- e-mail interno se cliente não responder.

### 25.4 Acionamento de fornecedor

Etapas:

1. cliente abre solicitação;
2. agente identifica necessidade de terceiro;
3. agente envia solicitação ao fornecedor;
4. condicional: fornecedor respondeu?;
5. se sim, confirma prazo;
6. se não, alerta gerente;
7. agente atualiza cliente;
8. Ficha acompanha execução.

Esse caso demonstra bem a força multicanal do POP Atende.

---

## 26. Riscos e Decisões em Aberto

### 26.1 Complexidade do editor de fluxos

O editor precisa ser poderoso sem virar ferramenta técnica demais.

Risco: aproximar-se demais de um n8n e assustar o público PME.

Decisão recomendada: começar com componentes poucos, claros e orientados a atendimento. Etapas, condicionais e alertas bastam para a primeira versão forte.

### 26.2 Autonomia excessiva dos agentes

Agentes conduzindo fluxos podem errar escolhas.

Mitigação:

- regras explícitas;
- saídas limitadas;
- validação pela engine;
- logs de decisão;
- intervenção manual;
- encaminhamento humano;
- modo de simulação antes de publicar fluxo.

### 26.3 Multicanal amplo demais

Adicionar muitos canais cedo pode diluir foco.

Decisão recomendada: WhatsApp como canal principal de entrada, Slack como canal interno forte, e-mail como alerta, Telegram como expansão. O discurso é multicanal; a execução deve ser incremental.

### 26.4 Precificação

Cobrar apenas por usuário pode subvalorizar automação. Cobrar apenas por mensagem aproxima o produto de chatbot. Cobrar por Ficha pode capturar valor, mas precisa ser entendido pelo cliente.

Decisão recomendada: combinar planos com limites simples de agentes, fluxos, canais e Fichas.

---

## 27. Direção de Marca e Comunicação

### 27.1 Tom

A comunicação deve ser direta, operacional e concreta.

Evitar linguagem excessivamente futurista sobre IA. O cliente não compra “agentes autônomos”. Ele compra menos bagunça, menos retrabalho e mais controle.

### 27.2 Mensagens possíveis

- “Transforme conversas em processos rastreáveis.”
- “Automatize fluxos de atendimento com agentes de IA.”
- “Coordene clientes, equipe e fornecedores em uma única operação.”
- “Cada atendimento vira uma Ficha. Cada Ficha tem etapa, responsável e histórico.”
- “A IA conversa. O fluxo organiza. O administrador controla.”

### 27.3 O que não enfatizar demais

- “Chatbot”;
- “bot de WhatsApp”;
- “respostas automáticas”;
- “IA que substitui atendentes”;
- “CRM completo”.

Esses termos reduzem o produto ou criam expectativa errada.

---

## 28. Resumo Estratégico

O POP Atende V2 deve ser entendido como uma plataforma de coordenação operacional baseada em fluxos, agentes de IA e Fichas.

O WhatsApp continua importante, mas deixa de ser a tese. A tese é que empresas precisam transformar conversas em processos. Agentes de IA são os coordenadores desses processos. Fluxos são o desenho. Fichas são a execução. Canais são as superfícies. Humanos continuam no controle.

A estrela do produto são os Fluxos.

A unidade operacional são as Fichas.

O diferencial é coordenar clientes, funcionários, fornecedores e agentes de IA em processos rastreáveis, multicanais e configuráveis pela própria empresa.

Essa é a categoria que o POP Atende deve ocupar.
