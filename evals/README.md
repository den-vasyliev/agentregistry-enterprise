# Master Agent Evaluation Suite

Python-based eval suite for testing the master agent via HTTP API and A2A protocol.

## Prerequisites

1. Controller running with a `MasterAgentConfig` resource applied
2. A `ModelCatalog` configured so the agent has a working LLM backend
3. Python 3.11+

## Setup

```bash
cd evals
pip install -e .
```

## Running Evals

### HTTP event pipeline tests (default)

```bash
pytest -v
```

### A2A evals (experimental, requires A2A server on :8084)

```bash
pytest -m a2a -v
```

### All tests including A2A

```bash
pytest -m '' -v
```

## Configuration

| Environment Variable | Default | Description |
|---|---|---|
| `MASTER_AGENT_API_URL` | `http://localhost:8080` | HTTP API base URL |
| `MASTER_AGENT_A2A_URL` | `http://localhost:8084` | A2A server base URL |

## Eval Datasets

- `eval_datasets/triage_basic.test.json` - Basic triage: infrastructure event triggers incident creation
- `eval_datasets/multi_event.test.json` - Multi-event lifecycle: create then resolve incidents
- `eval_datasets/test_config.json` - Criteria thresholds for eval scoring

## Adding New Evals

1. Create a new `.test.json` file in `eval_datasets/` using the ADK EvalSet schema
2. Each eval case needs a `conversation` array with `user_content`, `final_response`, and `intermediate_data.tool_uses`
3. Add a test function in `test_a2a_eval.py` pointing to the new dataset
