# grok-cli-go

A command-line interface (CLI) for interacting with xAI's Grok API. This tool enables multi-turn conversations with Grok models (e.g., Grok 4), supports agentic workflows via tool calling (ReAct-style looping for tasks like calculations and file operations), and includes live search integration.

Built in Go for performance and cross-platform compatibility, it's designed for developers wanting a lightweight, customizable CLI to leverage Grok's capabilities without relying on web or mobile apps.

## Features

- **Multi-Turn Conversations**: Maintain context across interactions with Grok models.
- **Agentic Tool Calling**: Automatically handles tool executions in a loop until resolution. Includes built-in tools for math expressions (`calculate`), file reading (`read_file`), and file writing (`write_file`).
- **Live Search**: Enabled via Grok API's `search_parameters` for real-time web/X searches, with citations (handling requires parsing response metadata).
- **Model Support**: Defaults to `grok-4` for advanced reasoning; configurable for `grok-3` or `grok-3-mini` via environment variable.
- **Extensibility**: Easily add more tools (e.g., git operations) to support codebase workflows like bug fixing or PR handling.
- **CI/CD Integration**: GitHub Actions workflows for testing PRs and releasing multi-arch binaries via Goreleaser.
- **No External Dependencies Beyond Go**: Simple setup with standard libraries.

Note: File operations are limited to the current working directory for security. Live search uses a custom request to include Grok-specific parameters.

## Installation

1. **Prerequisites**:
   - Go 1.24 or later.
   - A Grok API key from [xAI API](https://console.x.ai/team/default/api-keys).

2. **Clone the Repo**:
   ```
   git clone https://github.com/yourusername/grok-cli-go.git
   cd grok-cli-go
   ```

3. **Install Dependencies**:
   ```
   go mod tidy
   ```

4. **Build and Run**:
   ```
   go build -o grok-cli
   export GROK_API_KEY=your_api_key_here
   # Optional: export GROK_MODEL=grok-3  # Defaults to grok-4 if not set
   ./grok-cli
   ```

For multi-arch binaries, use the Goreleaser workflow by tagging a release (e.g., `git tag v1.0.0 && git push --tags`).

## Usage

On macOS, if you encounter a Gatekeeper security warning when running grok-cli, remove the quarantine attribute:

   ```
   sudo xattr -d com.apple.quarantine ~/bin/grok-cli
   ```

Run the CLI and interact via prompts:

```
Grok CLI: Enter your query (type 'exit' to quit). Tools and live search are enabled.
You: What is the latest news on xAI?
Grok: [Response with live search results]
You: Read the file example.txt
Grok: [File contents, via tool]
You: exit
```

### Tool Examples (Agentic Workflow)

- **Calculation**: "Calculate sqrt(16) + 5" – Invokes `calculate` tool.
- **File Read**: "Read content from file.txt" – Invokes `read_file` tool.
- **File Write**: "Write 'Hello' to newfile.txt" – Invokes `write_file` tool.
- **Live Search**: Queries requiring external knowledge will trigger search automatically.

Extend tools in `main.go` for more features, e.g., adding a git tool for codebase interactions.

## Configuration

- **Environment Variables**:
  - Set `GROK_API_KEY` for API access.
  - Set `GROK_MODEL` to select the model (e.g., `grok-3` or `grok-3-mini`; defaults to `grok-4`).
- **Adding Tools**: Modify the `tools` slice and `toolFunctions` map to include custom functions.

## Contributing

Contributions are welcome! Fork the repo, create a PR, and ensure it passes the CI workflow.

1. Fork and clone.
2. Create a feature branch: `git checkout -b feature/new-tool`.
3. Commit changes: `git commit -m "Add new tool"`.
4. Push: `git push origin feature/new-tool`.
5. Open a PR.

## License

MIT License. See [LICENSE](LICENSE) for details.
