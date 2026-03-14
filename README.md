# OpenVault

Encrypted local secret manager. A replacement for `.env` files and hardcoded credentials in `~/.zshrc` — secrets are automatically injected into your shell before every command, no manual `export` needed.

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

# Make sure $GOPATH/bin is in your PATH
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc
source ~/.zshrc
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

## Command Reference

### `openvault set <KEY>`

Store a secret. Input is hidden at the terminal and never written to shell history.

```bash
openvault set OPENAI_API_KEY
openvault set AWS_SECRET_ACCESS_KEY
openvault set DATABASE_URL
```

### `openvault get <KEY>`

Print a secret's value.

```bash
openvault get OPENAI_API_KEY
```

### `openvault list`

List all stored secret names (values are never shown).

```bash
openvault list
# OPENAI_API_KEY
# AWS_SECRET_ACCESS_KEY
# DATABASE_URL
```

### `openvault delete <KEY>`

Delete a secret.

```bash
openvault delete DATABASE_URL
```

### `openvault run <command>`

Inject secrets for a single command without relying on the shell hook. Useful for CI/CD, Docker, and non-interactive environments.

```bash
openvault run npm run dev
openvault run docker push myimage
openvault run -- python -c "import os; print(os.environ['OPENAI_API_KEY'])"
```

### `openvault env`

Print all secrets as `export KEY=value` statements, suitable for `eval`.

```bash
eval "$(openvault env)"
```

---

## Security

| Property | Implementation |
|---|---|
| Encryption | AES-256-GCM with a unique random nonce per secret |
| Master key | 32-byte random key stored in macOS Keychain, never written to disk |
| Hidden input | Kernel-level echo suppression via `term.ReadPassword`, not saved to history |
| File permissions | DB file `0600`, directory `0700` |
| Memory safety | Key zeroed in memory on `Close()` |

**Verify nothing is stored in plaintext:**

```bash
strings ~/.config/openvault/vault.db | grep YOUR_SECRET
# no output
```

---

## CI/CD

OpenVault is designed for local development. In CI environments without a Keychain, use `openvault run` with secrets passed via the platform's native secret management (e.g. GitHub Actions Secrets, Doppler).

---

## Storage Locations

| Platform | Path |
|---|---|
| macOS | `~/.config/openvault/vault.db` |
| Linux | `~/.config/openvault/vault.db` (master key at `~/.config/openvault/.openvault.key`) |
| Custom | Set the `XDG_CONFIG_HOME` environment variable |

---

## Build from Source

Requires Go 1.22+.

```bash
git clone https://github.com/npc-live/openvault
cd openvault
make build
sudo mv openvault /usr/local/bin/
```

Dependencies: `github.com/spf13/cobra`, `go.etcd.io/bbolt`, `golang.org/x/term`. No CGo — fully static binary.
