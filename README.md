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

[stores.file]
type = "filestore"

[stores.os]
type = "osstore"
[stores.os.config]
namespace = "prod"
```

## Supported Stores

clef currently supports two built-in secret stores:

- `filestore` – A simple file-based store (local, for dev purpose only)
- `osstore` – Uses the system's native keyring (macOS, Linux via Secret Service)

Other stores may be added in the future, as long as they meet the bar for safety and maintainability.

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

Made with ❤️ to avoid exporting secrets in `.Xrc`

