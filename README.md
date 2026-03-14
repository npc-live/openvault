# OpenVault

Encrypted secret manager with optional E2EE cloud sync. A replacement for `.env` files and hardcoded credentials in `~/.zshrc` — secrets are automatically injected into your shell before every command, no manual `export` needed.

```
openvault set OPENAI_API_KEY    # hidden input, encrypted at rest
node -e "console.log(process.env.OPENAI_API_KEY)"  # just works
```

---

## Install & Setup

Complete setup takes about one minute.

### Step 1 — Install the binary

**Download a pre-built binary** (no Go required):

Go to [Releases](https://github.com/npc-live/openvault/releases), download the binary for your platform, and move it to your PATH:

```bash
# macOS Apple Silicon
curl -L https://github.com/npc-live/openvault/releases/latest/download/openvault-darwin-arm64 -o openvault
chmod +x openvault
sudo mv openvault /usr/local/bin/

# macOS Intel
curl -L https://github.com/npc-live/openvault/releases/latest/download/openvault-darwin-amd64 -o openvault
chmod +x openvault
sudo mv openvault /usr/local/bin/

# Linux x86_64
curl -L https://github.com/npc-live/openvault/releases/latest/download/openvault-linux-amd64 -o openvault
chmod +x openvault
sudo mv openvault /usr/local/bin/
```

**Or install with Go:**

```bash
go install github.com/npc-live/openvault@latest
```

### Step 2 — Initialize the vault

```bash
openvault init
```

Creates an encrypted database at `~/.config/openvault/vault.db`. The master key is stored in your macOS Keychain — no password to remember.

### Step 3 — Enable automatic injection

```bash
# zsh
echo 'eval "$(openvault shell-init --shell zsh)"' >> ~/.zshrc
source ~/.zshrc

# bash
echo 'eval "$(openvault shell-init --shell bash)"' >> ~/.bashrc
source ~/.bashrc
```

This installs a shell hook that injects all your secrets into the environment before every command you run.

### Step 4 — Store your first secret

```bash
openvault set OPENAI_API_KEY
# Enter value for OPENAI_API_KEY: (input hidden)
# Secret "OPENAI_API_KEY" stored.
```

### Step 5 — Use it, no prefix needed

```bash
node -e "console.log(process.env.OPENAI_API_KEY)"  # ✓
npm run dev                                          # ✓
docker push myimage                                  # ✓
python train.py                                      # ✓
```

---

## Cloud Sync (optional)

Sync secrets across devices with end-to-end encryption. The server stores only ciphertext — it cannot decrypt anything.

### How it works

```
vault_key = PBKDF2-SHA256(password, secret_key, 100k iter)
```

Your `secret_key` is a random 32-byte file saved to `~/.config/openvault/secret-key.txt`. It never leaves your machine. Without it (and your password), the ciphertext on the server is useless.

### Register

```bash
openvault register
# Email: you@example.com
# Password: (hidden)
# Confirm password: (hidden)
# → verification code sent to your email
# Verification code: 123456
# Registered successfully!
# Secret key saved to: ~/.config/openvault/secret-key.txt
```

**Back up `~/.config/openvault/secret-key.txt` somewhere safe.** You need it to log in on a new device.

### Sync

```bash
openvault sync
# ↓ 0 pulled, ↑ 3 pushed
```

Pull + merge + push in one step. Last-write-wins by timestamp.

### Login on a new device

```bash
openvault init
openvault login --secret-key ~/.config/openvault/secret-key.txt
# Email: you@example.com
# Password: (hidden)

openvault sync
# ↓ 3 pulled, ↑ 3 pushed
```

### Other commands

```bash
openvault logout          # revokes token on server + clears local keychain
openvault forgot-password # reset password via email code
```

---

## Command Reference

### Local

| Command | Description |
|---|---|
| `openvault init` | Create a new vault |
| `openvault set <KEY>` | Store a secret (hidden input) |
| `openvault get <KEY>` | Print a secret's value |
| `openvault list` | List all secret names |
| `openvault delete <KEY>` | Delete a secret (also removes from server if logged in) |
| `openvault run <cmd>` | Run a command with secrets injected |
| `openvault env` | Print all secrets as `export KEY=value` |
| `openvault shell-init` | Print shell hook code |

### Cloud sync

| Command | Description |
|---|---|
| `openvault register` | Create account, re-encrypt vault, push secrets |
| `openvault login` | Authenticate on this device (`--secret-key <file>` for new devices) |
| `openvault logout` | Revoke JWT on server and clear local credentials |
| `openvault sync` | Pull + merge + push (last-write-wins) |
| `openvault forgot-password` | Reset password via email OTP |

---

## Security

### Local

| Property | Implementation |
|---|---|
| Encryption | AES-256-GCM, unique random nonce per secret |
| Master key | 32-byte random key in macOS Keychain, never written to disk |
| Hidden input | Kernel-level echo suppression via `term.ReadPassword` |
| File permissions | DB `0600`, directory `0700` |
| Memory | Key zeroed on `Close()`, password bytes zeroed after use |

### Cloud sync

| Property | Implementation |
|---|---|
| Key derivation | PBKDF2-SHA256, 100,000 iterations |
| Server storage | Ciphertext only — server cannot decrypt secrets |
| Secret key | Never sent to server; required to derive vault key on a new device |
| Password hashing | PBKDF2-SHA256 server-side (separate salt) |
| Auth tokens | HS256 JWT with `jti`; revoked server-side on logout |
| OTP security | SHA-256 hashed at rest; max 5 attempts before invalidation |
| Transport | TLS (Cloudflare Workers) |

**Verify nothing is stored in plaintext:**

```bash
strings ~/.config/openvault/vault.db | grep YOUR_SECRET
# no output
```

---

## Storage Locations

| File | Path |
|---|---|
| Vault database | `~/.config/openvault/vault.db` |
| Secret key (cloud) | `~/.config/openvault/secret-key.txt` |
| Master key (macOS) | macOS Keychain, service `openvault` |
| Master key (Linux) | `~/.config/openvault/.openvault.key` |

Set `XDG_CONFIG_HOME` to override the config directory.

---

## CI/CD

OpenVault is designed for local development. In CI environments without a Keychain, use `openvault run` with secrets passed via the platform's native secret management (e.g. GitHub Actions Secrets).

---

## Build from Source

Requires Go 1.22+.

```bash
git clone https://github.com/npc-live/openvault
cd openvault
CGO_ENABLED=0 go build -o openvault .
sudo mv openvault /usr/local/bin/
```

Dependencies: `cobra`, `bbolt`, `golang.org/x/term`, `golang.org/x/crypto`. No CGo — fully static binary.
