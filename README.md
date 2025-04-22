# clef 🔑

**clef** is a simple and secure command-line tool for managing secrets.

It was created to help keep sensitive data—like API keys and tokens—out of shell configuration files.
Instead of hardcoding values in `.bashrc`, `.zshrc`, or similar, you can store and retrieve them safely using `clef`.

## Features

- 🔐 Store secrets in a secret backend like OS keyring
- ⚙️ Configurable via a simple TOML file
- 🧼 Minimal and intuitive CLI
- 🔒 Designed with security and safety in mind

## Usage

```bash
clef <command> [flags]
```

### Commands

| Command                         | Aliases          | Description                          |
|----------------------------------|------------------|--------------------------------------|
| `get <key>`                      | `fetch`          | Look up a key in the store           |
| `set --key=<key> <value>`        | `put`, `store`   | Save a new key/value pair            |
| `delete <key>`                   | `rm`             | Delete a key from the store          |
| `config`                         |                  | Manage clef configuration            |
| `version`                        |                  | Print the current version            |

Run `clef <command> --help` for more details on each command.

## Example

```bash
# Store a secret
clef set --key=MY_API_KEY sk-test-abc123

# Retrieve it
clef get MY_API_KEY

# Delete it
clef delete MY_API_KEY
```

## Configuration

clef is configured via a TOML file. By default, the config is located at:

```bash
# macOS
~/Library/Application Support/clef/config.toml

# Linux
~/.config/clef/config.toml
```

Example:

```toml
# Default store if not specified
default_store = "file"
default_profile = "robot"

[stores.file]
type = "filestore"

[stores.os]
type = "osstore"
[stores.os.config]
namespace = "prod"

[stores.gcp]
type = "gcp"
[stores.gcp.config]
project-id = "project-x"

[profiles.robot]
shell = "nu"
[[profiles.robot.secrets]]
key = "fsociety-root"
store = "os"
target = "SUPER_SECRET"
```

## Supported Stores

clef currently supports two built-in secret stores:

- `filestore` – A simple file-based store (local, for dev purpose only)
- `osstore` – Uses the system's native keyring (macOS, Linux via Secret Service)
- `gcp` - Uses Google Cloud Platform [Secret Manager](https://cloud.google.com/security/products/secret-manager).

Other stores may be added in the future, as long as they meet the bar for safety and maintainability.

## Use Cases

### Exec

Clef exec enables you to run a specific command with required secrets available as environment variables.
You can either specify secrets directly using the `-s` or `--secret` flags, use a predefined profile with `-p` or `--profile`, or combine both approaches.

The command syntax follows this pattern:
```
clef exec [flags] -- <command_to_execute>

Flags:
  -h, --help                 Show context-sensitive help.
  -c, --config-file="/Users/banst/Library/Application Support/clef/config.toml"
                             Config file

  -p, --profile=STRING       Profile to load.
  -s, --secret=SECRET,...    Secrets to load into the env. Format [store.]secret[=env]. If store is empty, default store will be
                             used. If env is empty, secret name will be used as env name.
```

With the `--secret` flag, you can specify secrets in the format `[store.]secret[=ENV_VAR_NAME]`:
- If `store` is omitted, the default store will be used
- If `ENV_VAR_NAME` is omitted, the secret name will be used as the environment variable name

Examples:

```bash
# Run 'env' with secrets 'foo' as 'FOO', 'store.bar' as 'BAR', and 'baz' as 'baz'
clef exec -s foo=FOO --secret store.bar=BAR --secret baz -- env

# Run 'env' with all secrets defined in the 'stealth' profile
clef exec --profile stealth -- env

# Run 'env' with secrets from the 'stealth' profile plus an additional secret
clef exec -p stealth -s foo=ADDITIONAL_FOO -- env
```

Unlike `clef shell`, which creates an interactive shell environment, `clef exec` executes a single command and terminates afterward.
This is useful for running scripts or commands that need access to secrets without maintaining an interactive session.

### Shell

Clef shell enables you to create a shell environment with required secrets available as environment variables.
You can load a profile from your configuration, providing access to all secrets defined in that profile, or provide specific secrets.
Or both!

The command syntax follows this pattern:
```
clef shell [flags]

Flags:
  -h, --help                 Show context-sensitive help.
  -c, --config-file="/Users/banst/Library/Application Support/clef/config.toml"
                             Config file

  -p, --profile=STRING       Profile to load.
  -s, --secret=SECRET,...    Additional secrets to load into the env. Format [store.]secret[=env]. If store is empty, default store
                             will be used. If env is empty, secret name will be used as env name.
      --shell=STRING         Shell to use ($SHELL)
```

When you run `clef shell`, it launches a new shell session with the specified secrets injected as environment variables.
By default, it uses your system's default shell (`$SHELL`), but you can specify a different shell using the `--shell` flag.
With the `-s` or `--secret` flag, you can specify additional secrets in the format `[store.]secret[=ENV_VAR_NAME]`:
- If `store` is omitted, the default store will be used
- If `ENV_VAR_NAME` is omitted, the secret name will be used as the environment variable name

Examples:

```bash
# Launch a shell with secrets from the default profile
clef shell

# Launch a shell with secrets from the 'development' profile
clef shell -p development

# Launch a specific shell (bash) with secrets from the 'production' profile
clef shell -p production -s bash
```

> [!NOTE]
> While in the shell session, all specified secrets remain available as environment variables.
> Once you exit the shell (using `exit` or Ctrl+D), these secrets are removed from your environment.

> [!IMPORTANT]
> Shell mode is recommended only when you need an extended interactive session with access to secrets.
> For running specific commands, `clef exec` is the preferred approach since it automatically cleans up after execution.
> Use shell mode only when absolutely necessary, as you must remember to exit the shell to ensure secrets are removed from your environment.

## Contributing

If you’re interested in contributing—particularly by adding a new backend—you’re very welcome to open an issue or PR.

Backends are implemented in Go and must be reviewed and integrated into the codebase.
See the `osstore` implementation for an example of how a store is defined and registered.

> 🛡️ clef is intended for **safe** secret storage. Contributions should follow this principle above all else.

## Installation

Download the latest binary from the [GitHub Releases](https://github.com/b4nst/clef/releases) page and move it to a directory in your `$PATH`.

Example:

```bash
chmod +x clef
sudo mv clef /usr/local/bin/
```

## Why clef?

- Keeps secrets out of your shell history and dotfiles
- Uses secure, local backends (no external services required)
- Minimal setup, just works
- Built with a focus on **practical security**

## License

Apache – use it, modify it, share it.

---

Made with ❤️ to avoid exporting secrets in yet another `.Xrc`

