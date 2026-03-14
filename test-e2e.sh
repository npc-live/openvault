#!/usr/bin/env bash
# E2E test for openvault cloud sync.
# Run this in your terminal (requires a real TTY for password prompts).
#
# Usage:
#   ./test-e2e.sh your@email.com yourpassword
#
# What it tests:
#   1. register   → sends verification email
#   2. (you paste the code)
#   3. sync       → pushes local secrets
#   4. logout
#   5. Simulate new device: wipe vault, init, login --secret-key, sync, verify secrets match

set -euo pipefail

EMAIL="${1:-}"
PASSWORD="${2:-}"

if [[ -z "$EMAIL" || -z "$PASSWORD" ]]; then
  echo "Usage: $0 <email> <password>"
  exit 1
fi

BIN="./openvault"
DB="$HOME/.config/openvault/vault.db"
SK="$HOME/openvault-secret-key.txt"

# ── Colours ───────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; RED='\033[0;31m'; BOLD='\033[1m'; NC='\033[0m'
ok()   { echo -e "${GREEN}✓${NC} $*"; }
fail() { echo -e "${RED}✗${NC} $*"; exit 1; }
step() { echo -e "\n${BOLD}▶ $*${NC}"; }

# ── Preflight ─────────────────────────────────────────────────────────────────
step "Preflight"
[[ -f "$BIN" ]] || fail "binary not found — run: CGO_ENABLED=0 go build -o openvault ."
[[ -f "$DB"  ]] || fail "vault not found — run: ./openvault init"

SECRETS_BEFORE=$(./openvault list 2>/dev/null | sort)
COUNT=$(echo "$SECRETS_BEFORE" | grep -c . || true)
ok "Vault has $COUNT secret(s): $(echo $SECRETS_BEFORE | tr '\n' ' ')"

# ── Step 1: register ──────────────────────────────────────────────────────────
step "Register (email: $EMAIL)"
echo "  → Server will send a 6-digit code to $EMAIL"

# Drive register with expect
expect -f - <<EXPECT
set timeout 60
spawn $BIN register
expect "Email: "       { send "$EMAIL\r" }
expect "Password: "    { send "$PASSWORD\r" }
expect "Confirm"       { send "$PASSWORD\r" }
expect "code"          {
    send_user "\n\n>>> CHECK YOUR EMAIL for the 6-digit code, then paste it below:\n"
    expect_user -re "(\[0-9\]{6})" {
        set code \$expect_out(1,string)
        send "\$code\r"
    }
}
expect {
    "Registered successfully" { send_user "\nRegistered OK\n" }
    "error"                   { send_user "\nERROR\n"; exit 1 }
}
expect eof
EXPECT

ok "register complete"

# ── Step 2: verify secret key file ───────────────────────────────────────────
step "Verify secret key file"
[[ -f "$SK" ]] || fail "secret key file not found at $SK"
PERM=$(stat -f "%OLp" "$SK")
[[ "$PERM" == "600" ]] || fail "secret key file has wrong permissions: $PERM (want 600)"
SK_HEX=$(cat "$SK" | tr -d '\n')
[[ ${#SK_HEX} -eq 64 ]] || fail "secret key is not 64 hex chars (got ${#SK_HEX})"
ok "Secret key: ${SK_HEX:0:8}...${SK_HEX:56:8} (permissions: $PERM)"

# ── Step 3: sync ──────────────────────────────────────────────────────────────
step "Sync (push local secrets)"
OUTPUT=$("$BIN" sync 2>&1)
echo "  $OUTPUT"
echo "$OUTPUT" | grep -q "pushed" || fail "sync output missing 'pushed'"
ok "sync complete"

# ── Step 4: logout ────────────────────────────────────────────────────────────
step "Logout"
"$BIN" logout
ok "logged out"

# ── Step 5: simulate new device ───────────────────────────────────────────────
step "Simulate new device (wipe vault + keychain)"
echo "  Backing up vault to /tmp/vault.db.bak"
cp "$DB" /tmp/vault.db.bak

security delete-generic-password -s openvault        2>/dev/null || true
security delete-generic-password -s openvault-token  2>/dev/null || true
rm -f "$DB"
ok "Wiped vault and keychain"

step "Init fresh vault"
"$BIN" init
ok "init done"

step "Login with secret key"
expect -f - <<EXPECT
set timeout 30
spawn $BIN login --secret-key $SK
expect "Email: "    { send "$EMAIL\r" }
expect "Password: " { send "$PASSWORD\r" }
expect {
    "Logged in"  { send_user "\nLogin OK\n" }
    "error"      { send_user "\nERROR\n"; exit 1 }
}
expect eof
EXPECT
ok "login complete"

step "Sync (pull secrets from cloud)"
OUTPUT=$("$BIN" sync 2>&1)
echo "  $OUTPUT"
echo "$OUTPUT" | grep -q "pulled" || fail "sync output missing 'pulled'"
ok "sync complete"

# ── Step 6: verify secrets match ─────────────────────────────────────────────
step "Verify secrets match original"
SECRETS_AFTER=$("$BIN" list 2>/dev/null | sort)
if [[ "$SECRETS_BEFORE" == "$SECRETS_AFTER" ]]; then
  ok "Secret names match: $(echo $SECRETS_AFTER | tr '\n' ' ')"
else
  echo "  Before: $SECRETS_BEFORE"
  echo "  After:  $SECRETS_AFTER"
  fail "Secret list mismatch after sync"
fi

# Check each secret value is decryptable
while IFS= read -r KEY; do
  [[ -z "$KEY" ]] && continue
  VAL=$("$BIN" get "$KEY" 2>/dev/null)
  [[ -n "$VAL" ]] || fail "Could not decrypt secret: $KEY"
  ok "  $KEY = ${VAL:0:4}..."
done <<< "$SECRETS_AFTER"

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}All tests passed!${NC}"
echo ""
echo "Restoring original vault from backup (optional):"
echo "  cp /tmp/vault.db.bak ~/.config/openvault/vault.db"
echo "  ./openvault login --secret-key ~/openvault-secret-key.txt"
