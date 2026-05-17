OPS Checklist — Z-FINANCE

Objetivo
- Garantir deploy previsivel e validacao rapida em VPS.

Pre-deploy (local)
1) Rodar testes:
   - go test ./...
2) Revisar env vars obrigatorias:
   - DB_URL, AUTH_SECRET, WEBHOOK_SECRET
3) Ajustar env vars opcionais conforme ambiente:
   - CORS_ORIGINS, HTTP_PORT, APP_NAME
   - WEBHOOK_ALLOWED_IPS, WEBHOOK_RATE_LIMIT_PER_MINUTE
   - OTEL_SERVICE_NAME, OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_INSECURE
   - ALERT_*_THRESHOLD
4) Confirmar roles admin:
   - Usuario operador deve ter role ADMIN no ambiente.
5) Confirmar monitoramento de dependencias externas:
   - Exchange/Custody com health-check validado antes de operacoes criticas.

Deploy (VPS)
1) Subir codigo para /root/zetta-bank-core (git pull ou rsync).
2) Build:
   - go build -o zetta ./cmd
3) Rodar em background (exemplo simples):
   - nohup ./zetta > /var/log/zetta.log 2>&1 &
4) Validar logs iniciais:
   - tail -n 200 /var/log/zetta.log

Smoke checks (curl)
- Health:
  - GET http://<HOST>:8080/health
  - GET http://<HOST>:8080/health/db
- Observabilidade:
  - GET http://<HOST>:8080/debug/vars
  - GET http://<HOST>:8080/admin/observability/summary
- Alertas:
  - GET http://<HOST>:8080/admin/alerts/check?older_than_minutes=30
- Auditoria:
  - GET http://<HOST>:8080/admin/audit/logs?limit=5

Rollback (manual)
1) Encerrar processo atual.
2) Reexecutar binario anterior.
3) Revalidar health + logs.
4) Verificar reconciliacao pendente:
   - GET http://<HOST>:8080/admin/reconcile/pending

Compliance checks (pre-prod)
- Confirmar politica AML/KYC/KYB ativa
- Confirmar matriz RBAC e separacao de funcoes
- Confirmar politica de retencao de logs (7 anos)
- Confirmar backup/restore e DR documentados
- Confirmar rotacao de segredos e responsaveis
