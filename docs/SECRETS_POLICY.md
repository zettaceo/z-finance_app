# Politica de Segredos e Rotacao

Objetivo: reduzir risco de exposicao e atender requisitos de seguranca.

## Principios
- Segredos nunca em codigo.
- Rotacao periodica com registro em auditoria.
- Acesso minimo necessario.

## Baseline
- Rotacao a cada 90 dias (AUTH_SECRET, WEBHOOK_SECRET, DB).
- Rotacao imediata em incidente.
- Uso de secret manager em producao (Vault/SM).
- Segredos inativos revogados e auditados.
