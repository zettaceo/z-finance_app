# Politica AML/KYT (Baseline VASP)

Este documento define a politica minima de AML/KYT para o core Zetta.
Objetivo: prevenir lavagem de dinheiro, financiamento ao terrorismo e uso indevido.

## Escopo
- Usuarios PF e PJ (KYB).
- Todas as operacoes financeiras (fiat, cripto, conversoes, card, pagamentos).

## Principios
- Risco baseado em risco (Risk-Based Approach).
- Tolerancia zero a fraude e sancoes.
- Auditoria e evidencias obrigatorias.

## Componentes obrigatorios
1. Identificacao: KYC/KYB com niveis (UNVERIFIED, BASIC, FULL).
2. Screening: checagem de sancoes (OFAC/ONU/EU) e PEPs.
3. KYT: monitoramento de transacoes por regras e scoring.
4. Escalonamento: abertura de casos de compliance.
5. Registro: logs, evidencias e trilha de auditoria.

## Regras minimas (baseline)
- Thresholds por volume e frequencia (velocity + KYC limits).
- Alertas de padrao: repeticao de destinatarios, alta frequencia, valores fracionados.
- Bloqueio preventivo quando sinais criticos forem detectados.

## Investigacao e resposta
- Casos de compliance exigem justificativa e decisao registrada.
- Possiveis respostas: liberar, manter em hold, encerrar, reportar.

## Retencao
- Logs e evidencias devem ser retidos por no minimo 7 anos.

## Responsabilidades
- COMPLIANCE: triagem, decisao e evidencias.
- AUDIT: revisao independente.
- OPS: suporte operacional sob aprovacao.
