---
name: kite-cli-orchestrator
description: Orchestrate Kite workflows through the local Go CLI instead of MCP transport.
---

# Kite CLI Orchestrator Skill

Use this skill to orchestrate trading and account workflows through the local Go CLI entrypoint.

## Goal

Replace MCP-first interactions with CLI-driven execution while preserving the same capabilities for auth and read operations.

## Orchestration Flow

1. Ensure credentials are available (`KITE_API_KEY`, `KITE_API_SECRET`).
2. Run CLI login once to save token state:
   - `kite-mcp-server cli login`
3. Confirm token state:
   - `kite-mcp-server cli status`
4. Execute operational commands as needed:
   - `kite-mcp-server cli profile`
   - `kite-mcp-server cli margins`
   - `kite-mcp-server cli holdings`
   - `kite-mcp-server cli positions`
   - `kite-mcp-server cli orders`
   - `kite-mcp-server cli trades`
   - `kite-mcp-server cli gtts`
   - `kite-mcp-server cli quotes --instruments NSE:INFY,NSE:SBIN`
   - `kite-mcp-server cli ltp --instruments NSE:INFY,NSE:SBIN`
   - `kite-mcp-server cli ohlc --instruments NSE:INFY,NSE:SBIN`
   - `kite-mcp-server cli order-history --order-id <ORDER_ID>`
   - `kite-mcp-server cli order-trades --order-id <ORDER_ID>`
   - `kite-mcp-server cli historical-data --instrument-token <TOKEN> --from "YYYY-MM-DD HH:MM:SS" --to "YYYY-MM-DD HH:MM:SS"`
5. Run logout when needed:
   - `kite-mcp-server cli logout`

## Token Persistence

- Default token path: `~/.kite-mcp/session.json`
- Override token path via:
  - `KITE_TOKEN_PATH` env var
  - `--token-path` CLI flag

## Notes

- CLI commands output JSON for orchestration-friendly consumption.
- Use this skill for local agent workflows where MCP transport is not preferred.
