# ledger-service

A small Go service used to exercise an end-to-end CI/CD pipeline locally.

This repo is the **service** in a three-repo simulation:

- `ledger-service` (this repo) — the application code, unit/integration/contract tests, and the per-service `ci.yml` that calls the reusable workflows.
- [`platform-golden-paths`](https://github.com/chrissbo/platform-golden-paths) — reusable GitHub Actions workflows, Rego policies, Kyverno policies, the centrally pinned linter config.
- [`platform-deploy`](https://github.com/chrissbo/platform-deploy) — GitOps deploy repo. Argo CD watches this; promotions are PRs here.

## What this simulates

Pipelines 1–4 of the architecture in
[`research/cicd-toolchain/pipeline-architecture.md`](https://github.com/chrissbo/upvest-platform/blob/main/research/cicd-toolchain/pipeline-architecture.md)
end-to-end on free, open tooling: Cosign keyless signing, SLSA L3 provenance,
Pact contracts, OPA verdict, Kyverno admission, Argo CD + Argo Rollouts in a
local kind cluster. No cloud, no paid SaaS.

Plan and feasibility analysis:
[`research/cicd-toolchain/local-simulation-feasibility.md`](https://github.com/chrissbo/upvest-platform/blob/main/research/cicd-toolchain/local-simulation-feasibility.md).

## Status

Phase 0 — bootstrap. Service code, workflows, and local stack still to come.
