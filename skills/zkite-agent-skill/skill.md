---
name: zkite-agent-skill
description: Orchestrate Kite workflows through the local zkite CLI instead of MCP transport.
---

# zkite Agent Skill

Use this skill to orchestrate trading and account workflows through the local `zkite` CLI entrypoint.

## Goal

Replace MCP-first interactions with CLI-driven execution while preserving the same capabilities for auth and read operations.

## Orchestration Flow

1. Ensure credentials are available (`KITE_API_KEY`, `KITE_API_SECRET`).
2. Run CLI login once to save token state:
   - `zkite cli login`
3. Confirm token state:
   - `zkite cli status`
4. Execute operational commands as needed:
   - `zkite cli profile`
   - `zkite cli margins`
   - `zkite cli holdings`
   - `zkite cli positions`
   - `zkite cli orders`
   - `zkite cli trades`
   - `zkite cli gtts`
   - `zkite cli quotes --instruments NSE:INFY,NSE:SBIN`
   - `zkite cli ltp --instruments NSE:INFY,NSE:SBIN`
   - `zkite cli ohlc --instruments NSE:INFY,NSE:SBIN`
   - `zkite cli order-history --order-id <ORDER_ID>`
   - `zkite cli order-trades --order-id <ORDER_ID>`
   - `zkite cli historical-data --instrument-token <TOKEN> --from "2025-01-01 09:15:00" --to "2025-01-31 15:30:00"`
5. Run logout when needed:
   - `zkite cli logout`

## Token Persistence

- Default token path: `~/.kite-mcp/session.json`
- Override token path via:
  - `KITE_TOKEN_PATH` env var
  - `--token-path` CLI flag

## Notes

- CLI commands output JSON for orchestration-friendly consumption.
- Use this skill for local agent workflows where MCP transport is not preferred.
