# 🐍 eel-cli

CLI utility for creating and managing Eel projects with Vite.

## Features

- **Create projects**: Generate new Eel projects with web frontend
- **Package management**: Manage both Python and web dependencies
- **Development server**: Start development with hot reload
- **Build**: Create standalone executables with PyInstaller
- **Multiple package managers**: Support for npm, yarn, pnpm, bun

## Installation

```powershell
irm https://raw.githubusercontent.com/JuanBrotenelle/eel-cli/main/scripts/install.ps1 | iex
```

## Usage

### Create a new project

[Available templates](https://vite.dev/guide/#trying-vite-online)

```bash
# Create with interactive prompts
eel create my-project

# Create with specific package manager
eel create my-project --manager npm

# Create with init command
eel create my-project --manager bun --init vanilla
```

### Install dependencies

```bash
# Install all dependencies (Python + web)
eel install
```

### Manage web packages

```bash
# Add a package
eel web add react
eel web add --dev typescript

# Remove a package
eel web remove react
```

### Manage Python packages

```bash
# Add a package
eel py add requests
eel py add --dev pytest

# Remove a package
eel py remove requests
```

### Development

```bash
# Start development server (URL mode - recommended)
eel dev --mode url

# Start development server (watch mode)
eel dev --mode watch
```

### Build

```bash
# Build with default settings
eel build

# Build with custom name and icon
eel build --name "My App" --icon "icon.ico"

# Build as directory (not single file)
eel build --name "My App" --no-console
```

## Project Structure

```
my-project/
├── main.py              # Eel application entry point
├── pyproject.toml       # Python dependencies
├── eel.cli.json         # CLI configuration
├── web/                 # Web frontend
│   ├── package.json     # Web dependencies
│   ├── eel.d.ts         # TypeScript definitions
│   └── ...              # Your web files
└── dist/                # Build output
```

## Configuration

The `eel.cli.json` file stores project configuration:

```json
{
  "manager": "npm",
  "dev": {
    "mode": "url"
  },
  "build": {
    "appName": "my-project",
    "icon": "",
    "noConsole": true,
    "oneFile": true
  }
}
```

## Requirements

- Go 1.24.5+
- Python 3.10+
- uv (Python package manager)
- Node.js package manager (npm, yarn, pnpm, or bun)

## License

MIT
