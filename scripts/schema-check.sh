#!/usr/bin/env bash

set -euo pipefail

jq empty \
  docs/spec/schema/response-v2.schema.json \
  docs/spec/schema/capabilities-v2.schema.json \
  docs/spec/schema/search-audit-report-v2.schema.json \
  docs/spec/schema/compare-synthesis-report-v2.schema.json \
  docs/spec/schema/project-status-report-v2.schema.json \
  docs/spec/schema/agent-handoff-v1.schema.json \
  docs/spec/schema/agent-result-v1.schema.json \
  docs/spec/schema/compatibility-v1.schema.json \
  docs/compatibility.json \
  skills/evals/scenarios.schema.json

go test ./pkg/protocol ./cmd -run 'Schema|CapabilitiesGolden|JointCompatibility'
