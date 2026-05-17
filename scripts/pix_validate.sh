#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://95.111.247.134:8080}"
DB_CONTAINER="${DB_CONTAINER:-zetta-postgres}"
DB_USER="${DB_USER:-zetta}"
DB_NAME="${DB_NAME:-zetta_core}"
DB_PASSWORD="${DB_PASSWORD:-zetta_change_me}"
WEBHOOK_SECRET="${WEBHOOK_SECRET:-zetta_webhook_secret}"

USER_ID="00000000-0000-0000-0000-000000000101"
ACCOUNT_ID="00000000-0000-0000-0000-000000000201"

echo "==> Seed usuario/conta e deposito (R$ 2.000,00)"
docker exec -e PGPASSWORD="${DB_PASSWORD}" -i "${DB_CONTAINER}" psql -U "${DB_USER}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 <<'SQL'
BEGIN;
INSERT INTO users (id, external_id, email, full_name, status, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000101', 'conta_teste_01', 'teste@zetta.local', 'Conta Teste', 'ACTIVE', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO accounts (id, user_id, currency, scale, balance, status, created_at)
VALUES ('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000101', 'BRL', 2, 0, 'ACTIVE', NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO transactions (id, account_id, user_id, type, status, amount, fee, net_amount, idempotency_key, occurred_at, created_at)
VALUES ('00000000-0000-0000-0000-000000000302', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000101', 'DEPOSIT', 'CONFIRMED', 200000, 0, 200000, 'seed-deposit-2000', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
COMMIT;
SQL

echo "==> Cadastrar chave PIX"
curl -s -X POST "${API_BASE_URL}/pix/keys" \
  -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"${USER_ID}\",\"account_id\":\"${ACCOUNT_ID}\",\"type\":\"EMAIL\",\"key\":\"teste5@pix.local\"}"
echo

echo "==> Enviar PIX (PENDING_PARTNER)"
PIX_SEND_RESP=$(curl -s -X POST "${API_BASE_URL}/pix/send" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: pix-send-0001' \
  -d "{\"account_id\":\"${ACCOUNT_ID}\",\"user_id\":\"${USER_ID}\",\"amount\":10000,\"fee\":0,\"net_amount\":10000,\"external_ref\":\"PIX-TESTE-1\"}")
echo "${PIX_SEND_RESP}"

TRANSFER_ID=$(printf '%s' "${PIX_SEND_RESP}" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p' | head -n 1)
if [ -z "${TRANSFER_ID}" ]; then
  echo "Falha ao extrair transfer_id do envio PIX."
  exit 1
fi

echo "==> Webhook CONFIRMED"
WEBHOOK_BODY="{\"transfer_id\":\"${TRANSFER_ID}\",\"status\":\"CONFIRMED\"}"
SIG=$(printf '%s' "${WEBHOOK_BODY}" | openssl dgst -sha256 -hmac "${WEBHOOK_SECRET}" -hex | awk '{print $2}')
curl -s -X POST "${API_BASE_URL}/pix/webhook" \
  -H 'Content-Type: application/json' \
  -H "X-Signature: ${SIG}" \
  -d "${WEBHOOK_BODY}"
echo

echo "==> Enviar PIX 2 (para rejeicao)"
PIX_SEND_RESP_2=$(curl -s -X POST "${API_BASE_URL}/pix/send" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: pix-send-0002' \
  -d "{\"account_id\":\"${ACCOUNT_ID}\",\"user_id\":\"${USER_ID}\",\"amount\":5000,\"fee\":0,\"net_amount\":5000,\"external_ref\":\"PIX-TESTE-2\"}")
echo "${PIX_SEND_RESP_2}"

TRANSFER_ID_2=$(printf '%s' "${PIX_SEND_RESP_2}" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p' | head -n 1)
if [ -z "${TRANSFER_ID_2}" ]; then
  echo "Falha ao extrair transfer_id do envio PIX 2."
  exit 1
fi

echo "==> Webhook REJECTED"
WEBHOOK_BODY_2="{\"transfer_id\":\"${TRANSFER_ID_2}\",\"status\":\"REJECTED\"}"
SIG_2=$(printf '%s' "${WEBHOOK_BODY_2}" | openssl dgst -sha256 -hmac "${WEBHOOK_SECRET}" -hex | awk '{print $2}')
curl -s -X POST "${API_BASE_URL}/pix/webhook" \
  -H 'Content-Type: application/json' \
  -H "X-Signature: ${SIG_2}" \
  -d "${WEBHOOK_BODY_2}"
echo

echo "==> Saldo final no ledger (confirmados)"
docker exec -e PGPASSWORD="${DB_PASSWORD}" -i "${DB_CONTAINER}" psql -U "${DB_USER}" -d "${DB_NAME}" -c \
  "SELECT COALESCE(SUM(CASE type WHEN 'DEPOSIT' THEN net_amount WHEN 'WITHDRAWAL' THEN -net_amount WHEN 'PAYMENT' THEN -net_amount WHEN 'TRADE_BUY' THEN -net_amount WHEN 'TRADE_SELL' THEN net_amount WHEN 'CARD_AUTH' THEN -net_amount WHEN 'REVERSAL' THEN net_amount ELSE 0 END), 0) AS balance FROM transactions WHERE account_id = '${ACCOUNT_ID}' AND status = 'CONFIRMED';"
