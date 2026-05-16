# Z-Finance Demo — Plano de Ação Completo
> Gerado em: 2026-05-16 | Status: Em execução

---

## Objetivo
Transformar a demo atual em uma experiência de nível institucional que comunique
claramente os três públicos (Retail / Business / Institutional), o diferencial Dubai,
e a visão de plataforma global — sem confundir produto com backoffice.

---

## Mudanças Estruturais

### REMOVER da navegação principal
As seções abaixo são backoffice/ops e serão movidas para um **Admin Console separado**
acessível apenas por credencial admin:
- Reconciliação
- Alertas de threshold
- RBAC / Roles
- Precificação Avançada
- Observabilidade
- Admin (block user, plan change, etc.)

---

## Etapas de Execução

### ETAPA 1 — Switcher de Persona (Retail / Business / Institutional)
**Impacto:** Altíssimo — mostra o modelo de negócio sem precisar explicar  
**Status:** ✅ Concluída

Tarefas:
- [x] Adicionar campo `persona` ao mock store (default: `BUSINESS`)
- [x] Criar componente `PersonaSwitcher` no topbar/sidebar
- [x] Home adapta widgets conforme persona:
  - RETAIL: limites menores, sem observabilidade, sem admin, banner de upgrade, câmbio bloqueado (PRO)
  - BUSINESS: visão atual, funcionalidades completas
  - INSTITUTIONAL: treasury view, limites ilimitados, wealth panel, quick action "Treasury"
- [x] Cores/badges diferentes por persona no sidebar e balance card

---

### ETAPA 2 — Conta AED (Dirham dos Emirados)
**Impacto:** Alto — valida narrativa Dubai imediatamente  
**Status:** ✅ Concluída

Tarefas:
- [x] Adicionar `aed` em `store.accounts` no mock (saldo: AED 8.200)
- [x] Adicionar AED nas abas de conta da Home (🇧🇷 BRL / 🇺🇸 USD / 🇦🇪 AED / 📈 Invest)
- [x] Mostrar taxa AED/BRL no bloco de câmbio do Mover (FX strip com USD/AED/EUR/GBP)
- [x] Adicionar `aed` como opção nas transferências internas (com conversão automática)
- [x] Flag 🇦🇪 e formatação correta (AED X.XXX,XX)
- [x] Ação "PIX → AED" na categoria Câmbio do Mover
- [x] Fix bug: transfer_internal agora aceita fromAccount/toAccount e faz conversão multi-moeda

---

### ETAPA 3 — Página de Crédito (Z-Score + Simulador)
**Impacto:** Alto — representa Módulos 3 e 9 (120 funções) que estavam ausentes  
**Status:** ✅ Concluída

Tarefas:
- [x] Criar `src/pages/Credito.jsx`
- [x] Adicionar tab "Crédito" na navegação (ícone Landmark, substitui Investir no bottom nav)
- [x] Z-Score visual (gauge SVG circular 0–1000, gradiente colorido, 5 níveis)
- [x] Linha de crédito disponível (fiat + cripto como colateral, barra de utilização)
- [x] Simulador de empréstimo (sliders para valor/prazo/taxa, resultado em tempo real)
- [x] Histórico de crédito (loans ativos com barra de progresso + pagamento parcela)
- [x] Adicionar dados mock: `store.credit` com score, scoreHistory, linhas, ofertas
- [x] Adicionar actions simulate: `credit_simulate`, `credit_request`, `credit_pay`

---

### ETAPA 4 — Z-Pass (Identidade Financeira Digital)
**Impacto:** Alto — o produto mais diferenciado do documento, hoje invisível  
**Status:** ✅ Concluída

Tarefas:
- [x] Criar seção Z-Pass no perfil/sidebar (tela dedicada + botão Z-Pass no sidebar)
- [x] Card visual estilo "passaporte financeiro" com gradiente dinâmico por persona:
  - Nome + avatar inicial + e-mail
  - Z-Pass ID único (formato ZP-XXXX-XXXX derivado do user.id)
  - KYC FULL + VASP badge na base do card
  - Jurisdições habilitadas (🇧🇷 BRA ✅ / 🇦🇪 ARE ✅ / 🇺🇸 USA ⏳ / 🇪🇺 EUR ⏳)
  - Plano atual com badge colorido por persona
  - Data de verificação e próxima revisão
  - Flags de VASP, FATF risk, PEP, Sanções, Licença EMI, KYC Level
- [x] QR Code do Z-Pass (SVG decorativo 21×21 células, cor por persona)
- [x] Exportar Z-Pass (botão com feedback visual "Exportado!")

---

### ETAPA 5 — PIX Internacional
**Impacto:** Médio-Alto — diferencial competitivo principal  
**Status:** ✅ Concluída

Tarefas:
- [x] Adicionar categoria "PIX Internacional" no Mover (aba "PIX Intl" com ícone Plane)
- [x] Fluxo: PIX (BRL) → conversão → destino (USD/AED/EUR) com 3 ações dedicadas
- [x] Mock de países/destinos suportados (5 países, 3 live + 2 em breve)
- [x] Mostrar taxa de câmbio ao vivo no fluxo (hint com preview de conversão nos campos)
- [x] Action: `pix_international` no simulate() com taxa 1.5%, breakdown no msg/details
- [x] Painel explicativo "Como funciona" + lista de países com status e tempo de liquidação

---

### ETAPA 6 — Zion AI com Sugestões Proativas
**Impacto:** Médio-Alto — transforma de chatbot em assistente financeiro real  
**Status:** ✅ Concluída

Tarefas:
- [x] Painel inicial com 3 cards de insight proativo (computados do store real):
  - BTC valorizou/caiu R$ X hoje (+3.2%)
  - R$ X parados na conta → sugestão de rentabilizar
  - Parcela de crédito em N dias
- [x] Quick chips clicáveis: Ver saldo · Portfólio · Câmbio · PIX · Crédito · Z-Pass
- [x] Respostas contextuais expandidas: saldo multi-moeda (BRL/USD/AED), cripto, câmbio
  live, KYC por persona, crédito/Z-Score, Z-Pass, cartão JIT, PIX Internacional
- [x] "Analisando dados..." com delay aleatório 600-1000ms + nome do usuário na saudação
- [x] Histórico de conversa persistente em sessionStorage (até 20 mensagens)

---

### ETAPA 7 — Tiers de Cartão Visual
**Impacto:** Médio — reforça diferenciação por persona  
**Status:** ✅ Concluída

Tarefas:
- [x] RETAIL: cartão slate/cinza (`#1E293B → #334155 → #475569`), chip cinza, texto slate
- [x] BUSINESS: cartão azul-marinho escuro com acento índigo (atual, refinado)
- [x] INSTITUTIONAL: cartão preto-metal (`#0A0A0A`), chip dourado, reflexo áureo animado (`goldShimmer` 3.5s)
- [x] Troca automática ao mudar persona no PersonaSwitcher
- [x] Nome do tier estampado no canto superior direito do cartão (badge colorido)

---

### ETAPA 8 — Centro de Notificações
**Impacto:** Médio — polimento e sensação de produto vivo  
**Status:** ✅ Concluída

Tarefas:
- [x] Substituir sino decorativo por painel real (dropdown animado do topbar)
- [x] Feed de notificações com tipos (5 tipos com ícone emoji):
  - 💚 tx: Transação confirmada
  - 🔵 market: Alerta de mercado
  - 🟡 credit: Parcela de crédito
  - 🟠 security: Acesso detectado
  - 🔵 kyc: Revisão de compliance
- [x] Badge numérico real (contagem de não lidas, mostra "9+" se > 9)
- [x] Marcar individual como lida (clique na notificação) / Marcar todas (botão ✓✓)
- [x] Limpar lidas (botão lixeira)
- [x] Mock: 6 notificações no store (4 não lidas, 2 lidas)

---

### ETAPA 9 — Tela de Welcome / Onboarding
**Impacto:** Médio — cria narrativa em demos ao vivo  
**Status:** ✅ Concluída

Tarefas:
- [x] Após login (primeira vez na sessão), exibir welcome screen sobre o app
- [x] Animação sequencial (4 steps, 400ms de delay inicial):
  1. 🪪 "Z-Pass criado" — identidade digital ativa (1.2s)
  2. 🌍 "Contas abertas: BRL, USD, AED" (1.2s)
  3. 🛡️ "KYC: FULL verificado" — FATF LOW, não-PEP (1.2s)
  4. ✨ "Plataforma global pronta" → mensagem de boas-vindas + botão Começar
- [x] Cada step: ícone + título + subtítulo + dots bounce quando ativo, checkmark quando completo
- [x] Botão Skip (X) no canto superior direito
- [x] sessionStorage `zf_welcomed` — não repete na mesma sessão

---

### ETAPA 10 — Admin Console Separado
**Impacto:** Estrutural — limpa a nav principal  
**Status:** ✅ Concluída

Tarefas:
- [x] Criar rota/estado `adminMode` separado (sessionStorage `zf_admin`)
- [x] Mover para Admin Console: Reconciliação, Alertas, RBAC, Precificação Avançada, Observabilidade, Admin, Regulatório
- [x] Acesso via credencial diferente no Login (PIN "0000" = admin mode)
- [x] Nav do Admin Console com tema diferente (slate/cinza, banner "Acesso restrito")
- [x] Mais.jsx volta a ter apenas: KYC, Compliance, Pré-cadastro, Configurações

---

## Progresso Geral

| Etapa | Descrição | Status |
|-------|-----------|--------|
| 1 | Switcher de Persona | ✅ Concluída |
| 2 | Conta AED | ✅ Concluída |
| 3 | Página de Crédito | ✅ Concluída |
| 4 | Z-Pass | ✅ Concluída |
| 5 | PIX Internacional | ✅ Concluída |
| 6 | Zion AI Proativo | ✅ Concluída |
| 7 | Tiers de Cartão | ✅ Concluída |
| 8 | Centro de Notificações | ✅ Concluída |
| 9 | Welcome / Onboarding | ✅ Concluída |
| 10 | Admin Console Separado | ✅ Concluída |

---

## Notas Técnicas
- Stack: React 18 + Vite + Tailwind + Recharts + Lucide
- Branch: `claude/redesign-z-finance-ui-zXA7B` → push em `main`
- Cada etapa concluída = commit individual + update neste arquivo
- Build deve passar sem erros antes de cada commit
