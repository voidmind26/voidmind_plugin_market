from pathlib import Path


def test_plugin_scaffold_files_exist() -> None:
    root = Path(__file__).resolve().parents[1]
    assert (root / ".claude-plugin" / "plugin.json").exists()
    assert (root / ".mcp.json").exists()
    assert (root / "README.md").exists()


def test_go_rewrite_scaffold_replaces_python_runtime() -> None:
    root = Path(__file__).resolve().parents[1]
    assert (root / "go.mod").exists()
    assert (root / "main.go").exists()
    assert (root / "server" / "router" / "http.go").exists()
    assert (root / "server" / "helpers" / "bootstrap.go").exists()
    assert (root / "frontend" / "package.json").exists()
    assert not (root / "cmd" / "gateway_platform_mcp.py").exists()
    assert not (root / "cmd" / "gateway_platform_proxy.py").exists()
    assert not (root / "cmd" / "gateway_platform_web.py").exists()


def test_readme_mentions_go_rewrite_layout() -> None:
    root = Path(__file__).resolve().parents[1]
    content = (root / "README.md").read_text()
    assert "Go 重写版本" in content
    assert "server/" in content
    assert "frontend/" in content


def test_readme_mentions_new_default_port() -> None:
    root = Path(__file__).resolve().parents[1]
    content = (root / "README.md").read_text()
    assert "18787" in content
    assert "/app" in content
    assert "/gateway/<route>" in content


def test_mcp_entry_uses_dedicated_binary() -> None:
    import json

    root = Path(__file__).resolve().parents[1]
    mcp = json.loads((root / ".mcp.json").read_text())
    args = " ".join(mcp["mcpServers"]["gateway-platform"]["args"])
    assert "./cmd/gateway-platform-mcp" in args or "bin/gateway-platform-mcp" in args
    assert 'go run .' not in args


def test_mcp_entry_builds_frontend_before_binary() -> None:
    import json

    root = Path(__file__).resolve().parents[1]
    mcp = json.loads((root / ".mcp.json").read_text())
    args = " ".join(mcp["mcpServers"]["gateway-platform"]["args"])
    assert "./build.sh" in args
    assert "go build -o" not in args


def test_mcp_entry_runs_from_plugin_root() -> None:
    import json

    root = Path(__file__).resolve().parents[1]
    mcp = json.loads((root / ".mcp.json").read_text())
    args = " ".join(mcp["mcpServers"]["gateway-platform"]["args"])
    assert 'cd "$CLAUDE_PLUGIN_ROOT";' in args
