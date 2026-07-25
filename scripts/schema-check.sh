#!/usr/bin/env bash

set -euo pipefail

jq empty \
  docs/spec/schema/response-v2.schema.json \
  docs/spec/schema/capabilities-v2.schema.json \
  docs/spec/schema/search-audit-report-v2.schema.json \
  docs/spec/schema/compare-synthesis-report-v2.schema.json

go test ./pkg/protocol ./cmd -run 'Schema|CapabilitiesGolden'
