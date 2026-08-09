# Third-party notices

`obs-cli` is derived from [Yakitrak/notesmd-cli](https://github.com/Yakitrak/notesmd-cli), originally authored by Kartikay Jainwal and distributed under the MIT License. The original copyright and MIT permission notice remain in [LICENSE](./LICENSE).

## Upstream fork point

Provenance is recorded here so it stays verifiable without relying on GitHub fork metadata.

| Field | Value |
|---|---|
| Upstream repository | `https://github.com/Yakitrak/notesmd-cli` |
| Last common commit | `cae9aa84eed47ce23d139526ea3184ce09100450` (`cae9aa8`, 2026-05-13) |
| Upstream version at that commit | `v0.3.6` (`git describe`: `v0.3.6-1-gcae9aa8`) |
| First divergent commit | `7379326` (2026-06-04) `feat: rename to obs-cli, add Chinese comments, update docs` |

Every commit up to and including `cae9aa8` is shared with upstream and remains under Kartikay Jainwal's copyright. Everything from `7379326` onward is original to this project. No upstream changes have been merged since the fork point; the two code bases have diverged and are no longer merge-compatible.

The project also vendors the Go modules below. This file is a summary; the complete, controlling license and notice texts are distributed from the corresponding paths under `vendor/`.

| Module | Version | License | License files |
|---|---|---|---|
| `github.com/BurntSushi/toml` | v0.3.1 | MIT | `vendor/github.com/BurntSushi/toml/COPYING` |
| `github.com/adrg/frontmatter` | v0.2.0 | MIT | `vendor/github.com/adrg/frontmatter/LICENSE` |
| `github.com/davecgh/go-spew` | v1.1.1 | ISC | `vendor/github.com/davecgh/go-spew/LICENSE` |
| `github.com/google/jsonschema-go` | v0.4.2 | MIT | `vendor/github.com/google/jsonschema-go/LICENSE` |
| `github.com/inconshreveable/mousetrap` | v1.1.0 | Apache-2.0 | `vendor/github.com/inconshreveable/mousetrap/LICENSE` |
| `github.com/pmezard/go-difflib` | v1.0.0 | BSD-3-Clause | `vendor/github.com/pmezard/go-difflib/LICENSE` |
| `github.com/spf13/cobra` | v1.10.2 | Apache-2.0 | `vendor/github.com/spf13/cobra/LICENSE.txt` |
| `github.com/spf13/pflag` | v1.0.9 | BSD-3-Clause | `vendor/github.com/spf13/pflag/LICENSE` |
| `github.com/stretchr/testify` | v1.11.1 | MIT | `vendor/github.com/stretchr/testify/LICENSE` |
| `golang.org/x/sys` | v0.44.0 | BSD-3-Clause | `vendor/golang.org/x/sys/LICENSE` |
| `gopkg.in/yaml.v2` | v2.3.0 | Apache-2.0 and MIT | `vendor/gopkg.in/yaml.v2/LICENSE`, `LICENSE.libyaml`, `NOTICE` |
| `gopkg.in/yaml.v3` | v3.0.1 | MIT and Apache-2.0 | `vendor/gopkg.in/yaml.v3/LICENSE`, `NOTICE` |

No license in this file replaces or modifies the terms in the original license files.
