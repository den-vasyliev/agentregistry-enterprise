"""A2A-based evaluations using ADK AgentEvaluator.

These tests use the experimental RemoteA2aAgent and require:
- A2A server running on MASTER_AGENT_A2A_URL (default: http://localhost:8084)
- google-adk[eval] installed

Run separately: pytest test_a2a_eval.py -v -m a2a
"""

import os

import pytest
from google.adk.evaluation.agent_evaluator import AgentEvaluator

EVAL_DIR = os.path.join(os.path.dirname(__file__), "eval_datasets")

pytestmark = pytest.mark.a2a


@pytest.mark.asyncio
async def test_triage_basic():
    """Evaluate basic infrastructure event triage via A2A."""
    await AgentEvaluator.evaluate(
        agent_module="master_agent",
        eval_dataset_file_path_or_dir=os.path.join(EVAL_DIR, "triage_basic.test.json"),
        num_runs=1,
    )


@pytest.mark.asyncio
async def test_multi_event():
    """Evaluate multi-event processing and incident lifecycle via A2A."""
    await AgentEvaluator.evaluate(
        agent_module="master_agent",
        eval_dataset_file_path_or_dir=os.path.join(EVAL_DIR, "multi_event.test.json"),
        num_runs=1,
    )
