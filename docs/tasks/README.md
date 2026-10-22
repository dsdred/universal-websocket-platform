# Task Records

Каталог содержит внутренние task records, handoff и постоянные отчёты. Эти
operational документы ведутся на русском языке и не требуют EN-зеркал.

Текущая engineering task — [TASK-107](TASK-107-RUNTIME-TLS-ENFORCEMENT.md) —
TLS Runtime enforcement v1, verification-stable `In Progress`, branch
`codex/task-107-runtime-tls-enforcement`, baseline
`9151db9e32e0b351d403087e4f1361b31b134fac`. Actual checkpoints and closure
resolve only from newest matching task Recovery Evidence Envelope. No staging,
commit or publication authorized. Product implementation requires Architecture PASS.
TASK-106 source `909abc7f67a36f6709ba343d0cf5b63bc3509c85` published through
PR #106, merge `9151db9e32e0b351d403087e4f1361b31b134fac`; merge and source
have identical tree `dd11a05d473683ea53faf67d197eb65362f8d8c4`.
Required PR go-ci run39728769463 on source and push run39729079012 on merge
succeeded, including Linux race/tidy/module guards. This does not qualify
Windows/native TASK-102 or serve as TASK-107 verification/publication authority.
TASK-105 source `8de5b7da4ce5ba802ca9fc2d803c48bc1a5f6924` published through
PR #105, merge `48a6c359ab5604add3c309248ba39dfaff7a0656`. Baseline hosted
Go CI succeeded, including race/tidy/module guards; main ruleset requires go-ci.
These are historical baseline facts, not TASK-107 verification or authority.
TASK-104 ReadSeconds source `63d99906924195115cb9dc260762ca050cb675b8`
published through PR #104, merge `fb916df77e697fb7a45dc7bb6e8d0150411f2d82`.
TASK-103 WriteSeconds source `a2742ab927f7a23fd10a92db986ce36ff6a987e8`
published through PR #103, merge `aed7cc8f1693cf8bb48d71532fe543f09d2d1191`.
Historical race NOT RUN and R-103-N1 nonblocking remain limitations; these
facts provide no fresh publication authority or TASK-107 verification.
TASK-102 remains source-only Accepted, committed/unpublished on its separate
branch at `8bfa2e22a4b6310a71f6d36c8724b1adedd727ad`; native build/link NOT RUN,
toolchain NOT QUALIFIED (Owner checkpoint). C active fail-closed; B/PoC inactive.

Historical design task — [TASK-098](TASK-098-FENCED-MUTATION-GATEWAY-DESIGN.md) —
Fenced Mutation Gateway, Documentation-only / Design,
verification-stable `In Progress`, branch
`docs/task-098-fenced-mutation-gateway-design`, baseline
`3e41932b4cd7d30acc1f61ec3d80d9700e3d8b84`. Actual gates/closure resolve only
from newest valid matching Recovery Evidence Envelope in its task record.
Scope — новое design requirement владельца; confirmed governance gap не установлен.
DP-026 Draft / Planned проектирует prospective single-active gateway:
exclusive credentials/workspace, durable journal, fencing и uncertain outcomes.
Все неквалифицированные provider capabilities остаются UNKNOWN.
Native CAS не доказывает fencing внешних mutations. DP-025 остаётся Draft / Planned;
C остаётся действующим fail-closed baseline. Normative contracts, implementation,
deployment, PoC, activation и live recovery не входят в scope.
Следующие normative/evidence/implementation/activation gates — Not Activated,
без автоматически назначенных Task IDs.

TASK-097 — historical Coordinator Accepted, replacement R-097-007, manifest
`5a583be8e6e935c1f556f4bc4bdb58b8c6c975b6`; isolated source
`a3aa6a4a71422fdc38ee6a77feeb5af2719f1240`, PR #101 terminally published,
merge `3e41932b4cd7d30acc1f61ec3d80d9700e3d8b84`.
Это durable historical outcome, не fresh publication gate/owner или authority TASK-098.
Published DP-025 Draft / Planned не активировал controlled recovery.
Original commit/branch сохранены; TASK-096 source не включён в publication range.

TASK-096 — historical Completed — Coordinator Accepted по E-096-007/E-096-008;
accepted source `82cf49627e54881599bb60a55f28c8e9c6b0c4db`, accepted manifest
`1e065935ded73cd69c89b7f60a8359488509697d`. Authoring Completion не доказывает
terminal publication. TASK-095/096 — bounded forensic fixtures design; их permission,
ownership, immutable sources и operational transactions не меняются этой задачей.

TASK-095 — historical Completed — Coordinator Accepted по E-095-006/E-095-007;
accepted source `6a35b76f7678a06f65c24eb06de9bc8e0f6149bd`, accepted manifest
`82fac22f0766fbad81cdd4568230cfd5240cb537`. Immutable accepted assessment source
сохраняется. Authoring Completion не доказывает terminal publication; current
assessment не переносит authority и не изменяет TASK-095 record/transaction.

TASK-094 — historical Completed — Coordinator Accepted по E-094-006/E-094-007;
accepted source `0354eaaa9375328f8bf8d89d115a9f6ebfc16abc`, PR #98 local merge
`8cc432c3b6c751daa3e2daf4d791e54d99a74c75`. Assessment установил correct original
dispatch и исключил TargetMismatch; conditional ordinary recovery не активирован.
Immutable local metadata не fresh remote/P0–P10 evidence.

[TASK-093](TASK-093-INITIAL-DISPATCH-TARGET-MISMATCH-RECONCILIATION.md) —
published governance, historical Completed — Coordinator Accepted
по E-093-009/E-093-010, accepted manifest
`76d2f016f30b0ebba5d149f02d4f5ccfff36c599`, source
`8df57a598460ad677b0f128d5609d1913bc3e5ed`, PR #97 merge
`bfb29d3bc6da58e05e921700e8e311b399f5fc97` по immutable Git/read-only GitHub
evidence. Generic append-only mismatch design не доказывает incident eligibility,
current ownership/gate или authority применить его к TASK-092.

TASK-092 — immutable historical Completed — Coordinator Accepted по
E-092-007/E-092-008, source commit
`b25119ac94f21c115ca5dce7759447fbbd07fe3e`, accepted manifest
`51e60028f19d7bd2e87f86908aab8c839b713abf`. Source читается через
`git show b25119ac94f21c115ca5dce7759447fbbd07fe3e:docs/tasks/TASK-092-JOINT-SESSION-READINESS.md`;
его файл/DP-024 §22 не включены в этот main-based governance checkout. Это
unpublished accepted source; task Completion не доказывает publication и не
переносит authority. Его publication/recovery не активируются этой задачей.

TASK-091 — последняя опубликованная A1 implementation, historical Completed —
Coordinator Accepted, accepted manifest `27ebcd9465b1e4f078e6b64cd80fcc492a81efd2`,
source `108a461409421af36011ea5ca7a6b2e6a2ddca14`, PR #96 merge
`7f993c810ecd105f9f6d1b8c93fb2f173f832879` по local immutable Git metadata.
DP-024 §21 сохраняет isolated source/issuer, one-time generation binding,
detached observation и actual keeper inspection. A2 implementation/readiness
не активируется. B persistent-owner/claim/provenance остаётся unresolved;
successful durable qualification/full item7 Not Ready, integrated P1–P14 Planned.
DP-024 Approved / Partial, DP-017 Approved / Planned overall. No startup opening,
claim, permit, lifecycle/reconciliation/release, production wiring или promotion.
Historical Acceptance/cached facts не являются fresh gates/P0–P10 evidence.

Последняя опубликованная readiness —
[TASK-090](TASK-090-CLOSED-LOCAL-SESSION-READINESS.md). TASK-090 — historical Completed — Coordinator Accepted, accepted source manifest
`dacc667c07fa126151fe49ea5b356a4edc7d9d06`, source commit
`9a2aa9f33e9e9c7b4244024ab12437a27891a74b`, PR #95 merge
`587b55bdda0aa4ee92f2ce1827481e8dab51da86` by local immutable Git metadata.
Its additive DP-024 §20 establishes exact A1 implementation readiness/decomposition
and dependent A2 limitations; publication added no runtime capability. Historical
acceptance is not a fresh gate; cached facts are not fresh remote or Publisher
P0–P10 evidence. Original readiness work is not resumed.

Последний опубликованный design —
[TASK-089](TASK-089-CURRENT-AUTHORITY-JOINT-SESSION-DESIGN.md). TASK-089 — historical Completed — Coordinator Accepted, accepted source manifest
`4bdbdeb29093339c9adcb6d42f19391c5c3868e2`, source commit
`b010b00fc517e58177a1cbe46355168c558f471f`, PR #94 merge
`0a9b5d4e9d33cbf30216999c5e6e8abc6dff4ccb` by local immutable Git metadata.
Its additive DP-024 §19 supplies the private current-authority handoff/joint
observation design prerequisite; publication added no runtime capability.
Historical acceptance is not a fresh gate; cached facts are not fresh remote or
Publisher P0–P10 evidence. Original design work is not resumed.

Предыдущая опубликованная readiness —
[TASK-088](TASK-088-RECOVERY-QUALIFICATION-READINESS.md). TASK-088 — historical Completed — Coordinator Accepted, accepted source manifest
`5d66cadf5a65144df49356532f40806f787558b3`, source commit
`2870188eb59c843886369cefa8ab589535bc77e4`, PR #93 merge
`5aa367713c5e93abb752dcf25ff1d7b41ae7b5f0` by local immutable Git metadata.
Its additive DP-024 §18 readiness/decomposition supplies this design prerequisite;
publication added no runtime capability. Assessment Ready; full item7/P1–P14
integration Split Required / Not Ready as one bounded task. Historical acceptance
is not a fresh gate; cached facts are not fresh remote or Publisher P0–P10 evidence.

Последняя опубликованная implementation —
[TASK-087](TASK-087-FATAL-SOURCE-CLOSED-ASSEMBLY.md). TASK-087 — historical Completed — Coordinator Accepted, accepted source manifest
`11119b8e440809756783360455c8cc75472d6182`, source commit
`e1e03cb9fa898d4a784de06886070ba6c6a49cae`, PR #92 merge
`ca01f7cc5b94c59d6402aeac0eba33e2070f003a` by local immutable Git metadata.
Keeper-associated source roots, complete registration and typed fatal publication/
completion joins cover all seven actual sites; opaque closed assembly supplies the
same roots/participant to real DP-014/015 owners. Capture remains sequential,
not joint/current-authority or qualified no-claim proof. Coordinated writers stay
closed; authorized replay has zero lifecycle/provider callbacks. Authentic durable
origins/stores, claim/resume/issuance and production wiring remain absent.
Historical acceptance does not transfer to current gates; cached local facts are
not fresh remote or Publisher P0–P10 evidence.

Предыдущая опубликованная readiness —
[TASK-086](TASK-086-FATAL-SOURCE-COMPOSITION-READINESS.md). TASK-086 — historical Completed — Coordinator Accepted, accepted source manifest
`539567e4dd250b32094a708f919cbcf0649b50c8`, source commit
`f159fb9bbd6ba07a49465ac20c56b54aafd47954`, PR #91 merge
`fac149e1fdcfcd402e5872ae60336b083f1e3b22` by local immutable Git metadata.
Its Approved DP-024 §16 plan is the implementation prerequisite for TASK-087;
publication added no runtime capability. Historical acceptance does not transfer
to current gates. Cached local facts are not fresh remote or Publisher P0–P10 evidence.

Последний опубликованный implementation prerequisite —
TASK-085 — Completed — Coordinator Accepted, accepted source manifest
`2b399f887e312eef51e38845f53799115ee40e44`, source commit
`0d9c433086a96641eaff63dfa6d506f311bd69bb`, PR #90 merge
`7daefc0b96c8127a7f884b521686998a2181d43a` by local immutable Git metadata.
Its closed complete DP-015 slice is the prerequisite for TASK-086; historical
acceptance does not transfer to current gates. Cached local facts are not fresh
remote or Publisher P0–P10 evidence.

Предыдущий опубликованный design —
TASK-083 — Completed — Coordinator Accepted; accepted source manifest
`7a982b3b997cbe4d5c499829a355ce8824d8ce34`, source commit
`bafe45972bf0da17b10aaa8096cac9d8d59cdbe3`, PR #88 merge
`fb906b335523b933d7cdf3c634ca9946bcb410f0` по local immutable Git metadata.
Design-only publication добавила concrete DP-024 §15 plan, без runtime capability.
Это historical published source; исходная работа не возобновляется. Cached local
metadata не является fresh remote или Publisher P0–P10 evidence.

Предыдущая опубликованная readiness —
[TASK-082](TASK-082-REAL-OWNER-PARTICIPATION-READINESS.md), Completed — Coordinator
Accepted, accepted source manifest `aa97244d3f46cb7f535af69d6a86babe12780f26`,
commit `6cfb17bff4246d17e41107755ace0fa3b6d96850`, PR #87 merge
`cf17b1728349bca9ac3916c88a00445f021cc4a9` по local immutable Git metadata.
Assessment READY; full integration SPLIT REQUIRED / NOT READY AS ONE BOUNDED TASK;
its implementation intake was NOT READY pending the TASK-083 seam/decomposition.
Cached local metadata is not fresh network/P0–P10 proof. Durable provenance/stores
remain full item7 prerequisites; publication created no runtime capability.

Последняя опубликованная local foundation — TASK-081, accepted source manifest
`acefebe25d8a1053d6af0af00bbc1625014876c4`, commit
`f5818c6e0500ede49fc1f3c4b314db63e16affc1`, PR #86 merge
`4a2a575577f180527fa7a9b5499da037b294bb06` по local immutable Git metadata.
`internal/runtimerecoverycoordination` содержит только closed test-owner models;
На этом publication checkpoint real-owner integration, durable recovery claim и
production recovery отсутствовали; текущий closed DP-014 slice отдельно описан выше.
Cached refs/local merge metadata не являются fresh network evidence.

Предыдущая опубликованная design boundary — TASK-080 accepted subject `70302b93b3908b0216ce79298ade7ae0e10d123b`, source
`eac6f6eac0a19fbea4d4c47e978124e7fc6cd6eb` through PR #85 merge
`ac82aa29b9aaf7d25a0822b7ef3b3785b9149b18` by local immutable Git metadata.
Approved DP-024 design delivers no runtime or storage capability.

Предыдущая опубликованная design boundary —
[TASK-079](TASK-079-RECOVERY-CLAIM-ADMISSION-BOUNDARY.md), accepted subject
`15047580e43ee0a177bc35619f9f0dfaee2b2ae5`, source
`4ead8a82ac04a227121cd53b35b190a9a4a0182c`, PR #84 merged как
`0d018c42ff2244e80b05f7ea9e0e58b0966f6413`; Draft refined by TASK-080.

Предыдущая опубликованная readiness — TASK-078, Completed — Coordinator Accepted,
subject `4c73bc9fd72ca61d28c2e67f454e936178d46b2a`, source `8687758`, PR #83,
merge `4dad1af14102f516784cb1ae0ef39a59fec6b0ec` по local Git metadata.

Последний опубликованный process repair —
[TASK-077](TASK-077-TRUSTED-PUBLISHER-ROUTE.md), Coordinator Accepted subject
`bb2bd18df470021a8dc5bd6751a292c0d00c2e1c`, commit
`c70c398b2a7b5c85765f53d663088d117ca1e21e`, PR #82 merged как
`5e40742c74860c7446941cffbd8747309b3b1053`; product capability unchanged.

Последняя опубликованная product task —
[TASK-076](TASK-076-LOCAL-CONFIGURATION-STUDIO.md), Coordinator Accepted source
manifest `c1dd8f72c67588ca151b5dc32374ccd7ec0bfadf`, task commit
`b8c2b8672343aefccfe03000144ebdd4df93f8ea`, PR #81 merged как
`e9a87b5a35c880ac71444be9960876d3fbd02fbc`. Local/dev Configuration Studio
использует existing APIs, сохраняет truthful Published != Running/in-memory/
preliminary URL limitations и не меняет Runtime/DP-017.

Предыдущая опубликованная product task —
[TASK-075](TASK-075-RUNTIME-READ-ONLY-RECOVERY-ASSESSMENT.md), `Completed —
Coordinator Accepted`. Task commit
`f7b80ac976a144aa4918bd4ed74140e48fd31310` опубликован через PR #80 и merged
как `5eb54efea3f3df5c7967b2c1e2d0894c6289e50a`. Private mutation-free
assessment DP-017 и required DP-015 detached snapshot seam реализованы
изолированно. Durable recovery claim/permit/barrier/release, reconciliation,
Control Service Runtime wiring/reporting и Production Activation остаются `Not
Activated`.

Предыдущая опубликованная product task —
[TASK-074](TASK-074-RUNTIME-CONTAINMENT-COMPOSITION-GATE.md), `Completed —
Coordinator Accepted`. Task commit
`2d78d417784012e4e945e60f8d721246ea5bdcbf` опубликован через PR #79 и merged
как `bfab084c1a9664181027650b092bf240e04af435`. Isolated private composition
gate существует в `internal/runtimecontainmentcomposition`, но не wired в
Control Service и не реализует DP-017 workflow.

Предыдущая опубликованная product task —
[TASK-073](TASK-073-RUNTIME-SHUTDOWN-EVIDENCE-COMPOSER.md), `Completed —
Coordinator Accepted`. Task commit
`37ca6d6b5065aeb1a4831bc30aa54503ae06ff64` опубликован через PR #78 и merged
как `a84098284b0202f2fea9a61e089a74f564a05405`. Isolated private full-tuple
composer и invocation-scoped use-once handle теперь существуют в
`internal/runtimeexecutionevidence`; TASK-073 не добавила admission/provider
composition, Control Service wiring, DP-017 recovery или Production Activation.

Предыдущая опубликованная product task —
[TASK-071](TASK-071-RUNTIME-OWNER-SHUTDOWN-PROVENANCE.md), `Completed —
Coordinator Accepted`. Integration commit
`e1234a8447d75d1ffddf07dd953d444956b15ff1` опубликован через PR #77 и merged
как `0cec13d8e2b310545d5e2af40286158fda9820e9`; local/remote task refs
отсутствуют, `main == origin/main`. TASK-071 добавляет только isolated Owner
terminal provenance и exact-attempt snapshot/revalidation prerequisite; full
composer впоследствии реализован и опубликован TASK-073.

Опубликованный process repair —
[TASK-072](TASK-072-PUBLISHER-HANDOFF-STORE-RECOVERY.md), bounded docs-only
Publisher handoff-store repair на `docs/task-072-handoff-store-recovery` от
`c058da69f2296e52a8e32cc25e195190389dbca7`. Историческая Acceptance относится
к immutable source subject `00108732096fc68ad06391dab709edd962c1ebf1` task
commit `60ee9bec4dde3df08620f1a9d54b28da3f5b4c62`, опубликованному через PR #76
в `main@fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`. Это historical source, не
current live resolver TASK-072 в изменившемся integration tree. Product/readiness
scope и publication authority другого Target этим repair не создаются.

Предыдущая опубликованная product task —
[TASK-070](TASK-070-RUNTIME-CONTAINMENT-EVIDENCE-READER.md), `Completed —
Coordinator Accepted`; task commit `82b7cce` merged в
`main@c058da69f2296e52a8e32cc25e195190389dbca7`. Isolated generation reader не
добавляет full shutdown composer, DP-017 recovery, production wiring или
Activation.

Предыдущая завершённая product work —
[TASK-069](TASK-069-RUNTIME-PROCESS-CONTAINMENT-BOOTSTRAP.md), `Completed —
Coordinator Accepted (2026-09-25)` по newest valid matching Recovery Evidence
Envelope. Windows-only DP-023 bootstrap реализован изолированно; task commit
`046ddcfaa0a5f73e8821de71b578a0db4d2a3ecd` опубликован и merged в
`main@8eebcbc065f0aeb1c88a3be1c76ba586dfd34d19`. DP-023 — `Approved /
Implemented in isolation`, DP-022 — `Approved / Partial` для опубликованных
reader TASK-070, prerequisite provenance/read TASK-071, full-tuple composer
TASK-073 и private composition gate TASK-074. TASK-075 активирует только
read-only assessment item 6. Control Service wiring, recovery claim/barrier,
reconciliation, reporting и Production Activation отсутствуют.

Последняя завершённая documentation-only work —
[TASK-068](TASK-068-RUNTIME-CONTAINMENT-IMPLEMENTATION-READINESS.md),
`Completed — Coordinator Accepted (2026-09-21)`, Design-only / Readiness:
Independent Architect утвердил зеркальный DP-023 (`Approved / Planned`) и
первый bounded code slice без изменения product capability. Последующая
публикация подтверждается локальным Git: task commit
`924f71c6a7ee3a89d4f15bd74dda6ee4ba4c4df7` опубликован через PR #73 и merged
как `d3666ffea13a1866d86d7e8df9d4125aaade20c1`; прежние local/remote task refs
отсутствуют. TASK-069 активирована отдельным intake.

Предыдущая завершённая documentation-only work —
[TASK-067](TASK-067-RUNTIME-EXECUTION-CONTAINMENT-DESIGN-STATUS-DECISION.md),
`Completed — Coordinator Accepted (2026-09-21)`, `Design-update`: задача
приняла отдельное формальное Design Status решение по зеркальному DP-022
(execution-containment/evidence boundary, требуемый DP-017 section 11) —
Design Status повышен Draft → Approved при Implementation Status `Planned` —
на stage Architecture Confirmation по прецеденту TASK-021, с синхронизацией
downstream readiness. Task не создала новый архитектурный контракт и не
повысила Implementation Status: на момент её closure containment capability,
containment ledger и evidence adapter не были реализованы, termination proof
предыдущей generation не мог быть получен ни одним компонентом
(`Unknown(GuaranteeNotDeclared)` при default level `None`); DP-017 и DP-018
implementation, production integration и Production Activation не
активированы. Decision subject принят по manifest
`ab95d3d1c460e88df3222c3dfaf4dfb0048cf87c` после независимых Verification,
PROCESS-002, Scope Audit `13/0/0` и final Review в двух rounds (blocking
finding C1 первого review adjudicated `Disproven`). На момент closure commit и
publication не были авторизованы и не выполнялись — это historical closure
state. Последующая публикация подтверждена локальным Git: task commit
`a7218683c34c1097f20065c1e1e03e24e07e122d` опубликован через PR #72 и merged
как `82a03be49635690cec06d90679eb4d8b4801bade`.

Предыдущая завершённая documentation-only work —
[TASK-066](TASK-066-RUNTIME-EXECUTION-CONTAINMENT-EVIDENCE-DESIGN.md),
`Completed — Coordinator Accepted (2026-09-20)`, `Design-only`: задача
зафиксировала bounded technology-neutral
execution-containment/evidence boundary, который требуется DP-017 section 11,
зеркальным DP-022. Containment capability, containment ledger и evidence
adapter не реализованы, поэтому наблюдение завершения execution generation
ничем не обеспечивается; DP-017 и DP-018 implementation, production integration
и Production Activation не активированы. На момент closure commit и publication
не выполнялись; это historical closure state, а не live gate: задача позднее
опубликована — task commit `a727562ec948686ba494305f00bf2a49d72a420c` через
PR #71 (state `MERGED`, mergedAt `2026-09-19T22:51:16Z`) и merged как
`2c899a2069167c0d83b7c2d3cb8ffb13e862bfb9`, что подтверждено read-only
reconstruction из Git/GitHub; local и remote task-refs отсутствуют,
`main == origin/main` равны merge commit. Design Status DP-022 повышен
Draft → Approved отдельным решением TASK-067, а не Acceptance этой задачи.

Предыдущая завершённая documentation-only work —
[TASK-065](TASK-065-TASK-026-PUBLICATION-STATE-RECONCILIATION.md), `Completed —
Coordinator Accepted (2026-09-14)`: синхронизирован стабильный publication
outcome TASK-026 после PR #69; product capability не изменилась, следующая
task не активирована.

Предыдущая завершённая product work —
[TASK-026](TASK-026-RUNTIME-ACTIVATION-REPLACEMENT-ROLLBACK.md), `Completed —
Coordinator Accepted (2026-09-13)`: отдельная
[TASK-064](TASK-064-RUNTIME-COMMAND-SATISFIED-OUTCOME.md) terminally published
через PR #68 distinct durable primitive `Satisfied` outcome. Isolated
orchestrator TASK-026 и 19/19 proof coverage реализованы; independent Tester,
PROCESS-002, Scope Audit 31/0/0 и final Reviewer проходят. Task commit
`77c9ee1abe5e34ec20e49fd3c1bc760f9d9d605a` опубликован через PR #69 и merged
как `84c7dac0d6c323de2e46b931d918f076932e74df`; следующая task не активирована.

- [TASK-000 — Repository Synchronization](TASK-000-REPOSITORY-SYNCHRONIZATION.md)
- [TASK-000 — Repository Synchronization Report](TASK-000-REPOSITORY-SYNCHRONIZATION-REPORT.md)
- [TASK-001 — DP-008 Snapshot Builder поверх DetachedLoadResult](TASK-001-DP-008-SNAPSHOT-BUILDER.md)
- [TASK-002 — Упрощённое автономное продолжение проекта](TASK-002-AUTONOMOUS-PROJECT-CONTINUATION.md)
- [TASK-003 — Уточнение implementation prerequisites Draft DP-009](TASK-003-DP-009-IMPLEMENTATION-PREREQUISITES.md)
- [TASK-004 — Изолированная реализация Runtime Bootstrap DP-009](TASK-004-DP-009-RUNTIME-BOOTSTRAP.md)
- [TASK-005 — Implementation prerequisites Runtime Launcher](TASK-005-RUNTIME-LAUNCHER-PREREQUISITES.md)
- [TASK-006 — Изолированная реализация Runtime Launcher](TASK-006-RUNTIME-LAUNCHER.md)
- [TASK-007 — Implementation prerequisites Runtime Lifecycle Owner](TASK-007-RUNTIME-LIFECYCLE-OWNER-PREREQUISITES.md)
- [TASK-008 — Publisher pipeline governance](TASK-008-PUBLISHER-PIPELINE-GOVERNANCE.md)
- [TASK-009 — Runtime Lifecycle Owner implementation](TASK-009-RUNTIME-LIFECYCLE-OWNER.md)
- [TASK-010 — Production Launch Pipeline Design](TASK-010-PRODUCTION-LAUNCH-PIPELINE-DESIGN.md)
- [TASK-011 — Runtime Launch Flow implementation](TASK-011-RUNTIME-LAUNCH-FLOW.md)
- [TASK-012 — Engineering Process Hardening](TASK-012-ENGINEERING-PROCESS-HARDENING.md)
- [TASK-013 — Runtime Source Composition Design](TASK-013-RUNTIME-SOURCE-COMPOSITION-DESIGN.md)
- [TASK-014 — Runtime Source Implementation](TASK-014-RUNTIME-SOURCE-IMPLEMENTATION.md)
- [TASK-015 — Runtime Management Routing Design](TASK-015-RUNTIME-MANAGEMENT-ROUTING-DESIGN.md)
- [TASK-016 — Runtime Operational Identity Persistence Design](TASK-016-RUNTIME-OPERATIONAL-IDENTITY-PERSISTENCE-DESIGN.md)
- [TASK-017 — Runtime Management Command Idempotency Design](TASK-017-RUNTIME-MANAGEMENT-COMMAND-IDEMPOTENCY-DESIGN.md)
- [TASK-018 — Runtime Activation, Replacement, and Rollback Design](TASK-018-RUNTIME-ACTIVATION-REPLACEMENT-ROLLBACK-DESIGN.md)
- [TASK-019 — Runtime Recovery and Reconciliation Design](TASK-019-RUNTIME-RECOVERY-RECONCILIATION-DESIGN.md)
- [TASK-020 — Runtime Operational Error Reporting and Redaction Design](TASK-020-RUNTIME-OPERATIONAL-ERROR-REPORTING-REDACTION-DESIGN.md)
- [TASK-021 — Runtime Management Readiness Assessment](TASK-021-RUNTIME-MANAGEMENT-READINESS-ASSESSMENT.md) — Completed, Coordinator Accepted
- [TASK-022 — Root README Runtime Status Synchronization](TASK-022-ROOT-README-RUNTIME-STATUS-SYNCHRONIZATION.md) — Completed, Coordinator Accepted
- [TASK-023 — Runtime Management Routing Implementation](TASK-023-RUNTIME-MANAGEMENT-ROUTING.md) — Completed, Coordinator Accepted
- [TASK-024 — Runtime Operational Identity Persistence Implementation](TASK-024-RUNTIME-OPERATIONAL-IDENTITY-PERSISTENCE.md) — Completed, Coordinator Accepted
- [TASK-025 — Runtime Management Command Idempotency Implementation](TASK-025-RUNTIME-COMMAND-IDEMPOTENCY.md) — Completed, Coordinator Accepted
- [TASK-026 — Runtime Activation, Replacement, and Rollback Implementation](TASK-026-RUNTIME-ACTIVATION-REPLACEMENT-ROLLBACK.md) — Completed, Coordinator Accepted (2026-09-13); isolated orchestrator implemented; independent Tester PASS WITH ENVIRONMENT LIMITATION 0/0, PROCESS-002 Synchronized, Scope Audit 31/0/0, final Reviewer APPROVED 0/0; task commit `77c9ee1abe5e34ec20e49fd3c1bc760f9d9d605a` published through PR #69 and merged as `84c7dac0d6c323de2e46b931d918f076932e74df`
- [TASK-027 — Runtime Activation Orchestration Prerequisites Design](TASK-027-RUNTIME-ACTIVATION-ORCHESTRATION-PREREQUISITES-DESIGN.md) — Completed, Coordinator Accepted
- [TASK-028 — Runtime Command Parent/Phase Prerequisites Implementation](TASK-028-RUNTIME-COMMAND-PARENT-PHASE-PREREQUISITES.md) — Completed, Coordinator Accepted; TASK-026 remains Blocked
- [TASK-029 — Runtime Command Continue and Pending-Stop Prerequisite](TASK-029-RUNTIME-COMMAND-CONTINUE-PENDING-STOP.md) — Completed, Coordinator Accepted; TASK-026 remains Blocked
- [TASK-030 — Runtime Orchestration Binding Sequence Design](TASK-030-RUNTIME-ORCHESTRATION-BINDING-SEQUENCE-DESIGN.md) — Completed, Coordinator Accepted; DP-020 Draft/Planned, no implementation slice started; TASK-026 remains Blocked
- [TASK-031 — Runtime Orchestration Authorization Surface](TASK-031-RUNTIME-ORCHESTRATION-AUTHORIZATION-SURFACE.md) — Completed, Coordinator Accepted; historically accepted partial isolated DP-020 slice 1 prerequisite; TASK-026 remains Blocked
- [TASK-032 — Runtime Private Managed Invoker and Managed Flow Seam](TASK-032-RUNTIME-PRIVATE-MANAGED-INVOKER.md) — Completed, Coordinator Accepted after rework; historically accepted partial isolated DP-020 slice 2 prerequisite; TASK-026 remains Blocked
- [TASK-033 — TASK-032 Closure Synchronization](TASK-033-TASK-032-CLOSURE-SYNCHRONIZATION.md) — Completed, Coordinator Accepted; documentation drift corrected; DP-020 slice 3 is not activated
- [TASK-034 — Managed Binding Contract Reconciliation](TASK-034-MANAGED-BINDING-CONTRACT-RECONCILIATION.md) — Completed, Coordinator Accepted; repeat review Approved 0/0; at TASK-034 closure Slice 2R remained unactivated and Slice 3 blocked
- [TASK-035 — Managed Binding Repair](TASK-035-MANAGED-BINDING-REPAIR.md) — Completed, Coordinator Accepted; repeat review Approved 0/0; Slice 2R implemented in isolation; Slice 3 was unactivated at closure
- [TASK-036 — Slice 3 Readiness Reconciliation](TASK-036-SLICE-3-READINESS-RECONCILIATION.md) — Completed, Coordinator Accepted; Independent Review Approved 0/0; Slice 3 implementation was the accepted next candidate at closure
- [TASK-037 — Runtime Owner Claim Binding](TASK-037-RUNTIME-OWNER-CLAIM-BINDING.md) — Completed, Coordinator Accepted; Slice 3 implemented and accepted in isolation; Independent Reviewer APPROVED 0/0
- [TASK-038 — Slice 4 Orchestrator Readiness Reassessment](TASK-038-SLICE-4-ORCHESTRATOR-READINESS.md) — Completed, Coordinator Accepted; repeat Reviewer APPROVED 0/0; verdict `TASK-026 REMAINS BLOCKED`; at TASK-038 closure design-only atomic expected-attempt Owner Stop was the first unactivated candidate
- [TASK-039 — Atomic Expected-Attempt Runtime Owner Stop Design](TASK-039-EXPECTED-ATTEMPT-OWNER-STOP-DESIGN.md) — Completed, Coordinator Accepted; repeat Reviewer APPROVED 0/0; accepted Draft DP-010 design recorded; implementation was absent at closure and subsequently completed as TASK-040; TASK-026 remained Blocked at closure
- [TASK-040 — Expected-Attempt Runtime Owner Stop Implementation](TASK-040-EXPECTED-ATTEMPT-OWNER-STOP.md) — Completed, Coordinator Accepted; isolated implementation and proof tests verified; repeat final Reviewer APPROVED 0/0; its later invoker prerequisite is implemented by TASK-043; TASK-026 remained Blocked pending later readiness reassessment
- [TASK-041 — Managed Continuation Documentation Baseline Reconciliation](TASK-041-PRIVATE-EXACT-SCOPE-INVOKER-DESIGN.md) — Completed, Coordinator Accepted (2026-08-20); final Reviewer APPROVED 0/0; documentation-only live status repair synchronized; its then-unactivated invoker design became TASK-042 and isolated implementation TASK-043; TASK-026 remained Blocked at closure
- [TASK-042 — Private Exact-Scope Composition Invoker Design](TASK-042-PRIVATE-EXACT-SCOPE-INVOKER-DESIGN.md) — Completed, Coordinator Accepted (2026-08-20); Tester PASS 0/0; repeat Reviewer APPROVED 0/0; Scope Audit 17/0/0; Draft DP-021 recorded and later implemented partially/in isolation by TASK-043; TASK-026 remained Blocked at closure
- [TASK-043 — Private Exact-Scope Managed Start Invoker Implementation](TASK-043-PRIVATE-EXACT-SCOPE-INVOKER.md) — Completed, Coordinator Accepted (2026-08-21); final Reviewer APPROVED 0/0; Scope Audit 21/0/0; concrete DP-021 invoker implemented in isolation; later TASK-044 Ready-to-reactivate outcome is historical and superseded by TASK-026 recheck
- [TASK-044 — Runtime Terminal / Orchestrator Readiness Reassessment](TASK-044-RUNTIME-TERMINAL-ORCHESTRATOR-READINESS.md) — Completed, Coordinator Accepted (2026-08-24); historical Architect `UNBLOCK TASK-026`; repeat Reviewer APPROVED 0/0; Scope Audit 16/0/0; PROCESS-002 Synchronized; later superseded by TASK-026 `CONFIRMED BLOCKER`
- [TASK-045 — TASK-026 Reactivation Status Reconciliation](TASK-045-TASK-026-STATUS-RECONCILIATION.md) — Completed, Coordinator Accepted (2026-08-24); repeat Tester PASS 0/0/0; Independent Reviewer APPROVED 0/0; Scope Audit 7/0/0; PROCESS-002 Synchronized; closure-time Ready-to-reactivate outcome later superseded by TASK-026 recheck
- [TASK-046 — Tracked-Start Managed-Parent Admission Design](TASK-046-TRACKED-START-MANAGED-PARENT-ADMISSION-DESIGN.md) — Completed, Coordinator Accepted (2026-08-25); repeat Reviewer Approved 0/0; Scope Audit 15/0/0; PROCESS-002 Synchronized; additive DP-015 planned contract documented without implementation; TASK-026 remains Blocked; implementation prerequisite is the next candidate, Not Activated, without a Task ID
- [TASK-047 — Tracked-Start Managed-Parent Admission Implementation](TASK-047-TRACKED-START-MANAGED-PARENT-ADMISSION.md) — Completed, Coordinator Accepted (2026-08-25); independent Tester PASS WITH ENVIRONMENT LIMITATION 0/0; final Reviewer APPROVED 0/0; Scope Audit 18/0/0; isolated DP-015 prerequisite implemented; TASK-026 remains Blocked pending separate reassessment
- [TASK-048 — Execution Interruption Recovery Governance](TASK-048-EXECUTION-INTERRUPTION-RECOVERY-GOVERNANCE.md) — Completed, Coordinator Accepted (2026-08-26); Tester PASS 0/0/1; Repeat Independent Review Approved 0/0; Scope Audit 19/0/0; canonical manifest uses exact unsigned UTF-8 path-byte order; TASK-026 remains Blocked and is not activated
- TASK-049 — Replay-First Orchestration Admission Design — Completed,
  Coordinator Accepted (2026-08-28); its record is contained in immutable
  target `4a040b4e86ec2f4361ec765657e46cd0f36bf349` on branch
  `docs/task-049-replay-first-late-generation-design`, not in current `main`;
  publication ended `InvalidatedByTargetChange`, so the old authorization is
  terminal, non-transferable and non-reusable; TASK-026 remains Blocked and its
  implementation candidate is Not Activated
- [TASK-050 — Publisher Execution Environment Capability and Trusted-Context Handoff](TASK-050-PUBLISHER-EXECUTION-CONTEXT-HANDOFF.md) — Completed — Coordinator Accepted (2026-08-30); published as task commit `794ce5f350649115900ab8c88f34a91cf181e1c8` through PR #52 and merged as `ae76c8385ac3241946267272e4468d74fcee9cb4`; exact-context capability, trusted-context handoff and stable-live-state governance are current
- [TASK-051 — TASK-049 Publication-Invalidation Reconciliation](TASK-051-TASK-049-PUBLICATION-RECONCILIATION.md) — Completed, Coordinator Accepted (2026-08-31); task commit `171d002322e965e98f7af75722338858b4d255d1` published through PR #53 and merged as `3f7c23dc3acc6a4007a8f6a081f91386aefaa2ff`; historical TASK-049 old target remains immutable and its ended publication authority did not transfer
- [TASK-052 — Blocking Documentation Live-State Reconciliation](TASK-052-BLOCKING-DOCUMENTATION-LIVE-STATE-RECONCILIATION.md) — Completed, Coordinator Accepted (2026-08-31); task commit `c71b1deef0c5cf0ff42c685db7274bb77a846666` published through PR #54 and merged as `44a0283e6667b79fdc63882afe0dc2b1f136a9bc`; TASK-026 remains Blocked and the implementation candidate remains Not Activated
- [TASK-053 — Root README Runtime Status Reconciliation](TASK-053-ROOT-README-RUNTIME-STATUS.md) — Completed, Coordinator Accepted (2026-08-31); task commit `bfc1d579f1dbf353d1a74dfb4e98b61132ff1bf2` published through PR #55 and merged as `0e0d02ad976db0fc05346d1085932b658e5bbb0b`; TASK-026 remains Blocked and its implementation prerequisite remains Not Activated
- [TASK-054 — Documentation Home User Guidance Reconciliation](TASK-054-DOCS-HOME-USER-GUIDANCE.md) — Completed, Coordinator Accepted (2026-09-01); task commit `07f9cc7e8a5aac4e24b52795f21fdeca5b9d5b14` published through PR #56 and merged as `64601fc26dd57c0ebd09f1db39f9851b6b16643e`; TASK-026 remains Blocked and its implementation prerequisite remains Not Activated
- [TASK-055 — Mirrored MASTER_PLAN Governance Freshness Reconciliation](TASK-055-MASTER-PLAN-GOVERNANCE-FRESHNESS.md) — Completed, Coordinator Accepted (2026-09-02); task commit `da44e0ab22aa94a223628afdc9b20e61a1337e02` published through PR #57 and merged as `9884d8458bc99ce61439c286c810d7e2cd2f91ae`; TASK-026 remains Blocked and runtime/DP/proposal candidates remain Not Activated
- [TASK-056 — Wiki Knowledge-Map Freshness Reconciliation](TASK-056-WIKI-KNOWLEDGE-MAP-FRESHNESS.md) — Completed, Coordinator Accepted (2026-09-02); task commit `bd87bbb8526efe1413899e8125e847d80aade09a` published through PR #58 and merged as `934a7137d4c75598df4cbf9c28fc09c0fa665e5e`; TASK-026 remained Blocked and deferred candidates were not activated by that documentation task
- [TASK-057 — Replay-First Late-Generation Admission Implementation](TASK-057-REPLAY-FIRST-LATE-GENERATION-ADMISSION.md) — immutable task commit `50c24ea209f69aae7a531aa72080bcf53542cdb5` published through PR #59; Historical Equivalence remains Not Proven, while exact four bounded claims are prospectively accepted by TASK-060 event `9199e91e-82cf-4b94-8e9d-c81ba91015b6`
- [TASK-058 — Runtime Activation Readiness Reassessment](TASK-058-RUNTIME-ACTIVATION-READINESS-REASSESSMENT.md) — Sealed Negative Disposition through PR #60 on `main@8ce7b9095b2d56e065034bb031b6d5806eab87c8`; chronology and historical accepted-to-published equivalence remain Not Proven; no Acceptance/BCC/Completion or positive prerequisite proof
- [TASK-059 — Prospective Published-Subject Acceptance Governance](TASK-059-PROSPECTIVE-PUBLISHED-SUBJECT-ACCEPTANCE-GOVERNANCE.md) — Completed, Coordinator Accepted (2026-09-05); task commit `c4c858941ceb9544e6a30454e6309a9ef159b875` published through PR #61 by local merge metadata and merged as `600dc10b737ce2dfda550379a6ec68e3b680f959`; general IPSPA protocol only
- [TASK-060 — TASK-057 Published Subject Prospective Acceptance](TASK-060-TASK-057-PUBLISHED-SUBJECT-PROSPECTIVE-ACCEPTANCE.md) — Completed, Coordinator Accepted (2026-09-06); task commit `bd7356152ee62d5b8de7b8e9a7fa49c3890a4ae5` published through PR #62 and merged as `cc5e7598029a659ee0f4c382cd01627a753a3200`; exact event accepts only four bounded claims, Historical Equivalence remains Not Proven
- [TASK-061 — Runtime Activation Readiness Reassessment](TASK-061-RUNTIME-ACTIVATION-READINESS-REASSESSMENT.md) — Completed — Coordinator Accepted (2026-09-07); task commit `eb4717dd295988c6633668c04141082991c7636f` published through PR #63 and merged as `67e81f8b5b86cfbd9da1e616af9d27fda9c093f8`; immutable historical verdict `READY — UNBLOCK`, matrix 7 Direct / 10 Compositional / 2 Missing core / 0 Missing prerequisite / 0 Missing external / 0 Deferred; fresh TASK-026 evidence does not rewrite this accepted record
- [TASK-062 — Runtime Command Parent Terminalization Repair](TASK-062-RUNTIME-COMMAND-PARENT-TERMINALIZATION.md) — Completed — Coordinator Accepted (2026-09-08); task commit `07f70f0a4bb076a3b47324c41de28a80e14ed73f` published through PR #65 and merged as `2f1de022f7821cf9a4b65fe408c42349060f523e`; bounded DP-015 repair implemented and accepted in isolation
- [TASK-063 — Runtime Activation Readiness Reassessment after Parent Terminalization Repair](TASK-063-RUNTIME-ACTIVATION-READINESS-REASSESSMENT.md) — Completed — Coordinator Accepted (2026-09-09); task commit `499054be6745c967848b5aac0f3ff6a64dfbe7a9` published through PR #66 and merged as `2404c3439f44b9b0b22f87d87029695f921fee62`; independent Architect verdict `READY — UNBLOCK TASK-026`, matrix 7/10/2/0/0/0
- [TASK-064 — Runtime Command Durable Satisfied Outcome](TASK-064-RUNTIME-COMMAND-SATISFIED-OUTCOME.md) — Completed — Coordinator Accepted in isolation; task commit `60b3dd03a23a995227726fa2f4094295c8f15f0f` published through PR #68 and merged as `9483232bde262cd69ba7146a44ee9f251a65cb40`; bounded DP-015 prerequisite only; TASK-026 subsequently reactivated by separate intake
- [TASK-065 — TASK-026 Publication State Reconciliation](TASK-065-TASK-026-PUBLICATION-STATE-RECONCILIATION.md) — Completed — Coordinator Accepted (2026-09-14); documentation-only synchronization of task commit `77c9ee1abe5e34ec20e49fd3c1bc760f9d9d605a`, PR #69, and merge `84c7dac0d6c323de2e46b931d918f076932e74df`; product capability unchanged; no next task activated
- [TASK-066 — Runtime Execution Containment and Evidence Design](TASK-066-RUNTIME-EXECUTION-CONTAINMENT-EVIDENCE-DESIGN.md) — Completed — Coordinator Accepted (2026-09-20); Design-only; зеркальный DP-022 зафиксирован как Draft/Planned на момент closure, независимые Verification, PROCESS-002, Scope Audit, final Review и Coordinator Acceptance пройдены; task commit `a727562ec948686ba494305f00bf2a49d72a420c` опубликован через PR #71 (`MERGED`, mergedAt `2026-09-19T22:51:16Z`) и merged как `2c899a2069167c0d83b7c2d3cb8ffb13e862bfb9` — на момент closure commit и publication не выполнялись, это historical closure state; Design Status DP-022 повышен Draft → Approved отдельным решением TASK-067, а не Acceptance этой задачи; на момент closure containment capability, ledger и evidence adapter не были реализованы; current TASK-069 candidate не активирует DP-017/DP-018, production integration или Production Activation
- [TASK-067 — Runtime Execution Containment Design Status Decision](TASK-067-RUNTIME-EXECUTION-CONTAINMENT-DESIGN-STATUS-DECISION.md) — Completed — Coordinator Accepted (2026-09-21); `Design-update`; Design Status DP-022 повышен Draft → Approved при Implementation Status `Planned`; независимые Verification, PROCESS-002, Scope Audit 13/0/0 и final Review (две rounds, C1 `Disproven`) пройдены, decision subject `ab95d3d1c460e88df3222c3dfaf4dfb0048cf87c`; containment capability, ledger и evidence adapter не были реализованы на момент closure; current TASK-069 candidate не активирует DP-017/DP-018, production integration или Production Activation; commit/publication не были авторизованы и не выполнялись на момент closure — это historical closure state; task commit `a7218683c34c1097f20065c1e1e03e24e07e122d` позднее опубликован через PR #72 и merged как `82a03be49635690cec06d90679eb4d8b4801bade`
- [TASK-068 — Initial Runtime Containment Implementation Boundary and Slice Decomposition](TASK-068-RUNTIME-CONTAINMENT-IMPLEMENTATION-READINESS.md) — Completed — Coordinator Accepted (2026-09-21); Design-only / Readiness; DP-023 Approved/Planned and first bootstrap slice accepted without product capability change; task commit `924f71c6a7ee3a89d4f15bd74dda6ee4ba4c4df7` published through PR #73 and merged as `d3666ffea13a1866d86d7e8df9d4125aaade20c1`
- [TASK-069 — Runtime Process-Containment Bootstrap Implementation](TASK-069-RUNTIME-PROCESS-CONTAINMENT-BOOTSTRAP.md) — Completed — Coordinator Accepted (2026-09-25) by newest valid matching envelope; task commit `046ddcfaa0a5f73e8821de71b578a0db4d2a3ecd` merged into `main@8eebcbc065f0aeb1c88a3be1c76ba586dfd34d19`; DP-023 Approved/Implemented in isolation; not wired into production
- [TASK-070 — Runtime Containment Exact Generation Evidence Reader](TASK-070-RUNTIME-CONTAINMENT-EVIDENCE-READER.md) — Completed, Coordinator Accepted; task commit `82b7cce29ea9bca350b7945ac51a2531135d1a21` published through PR #75 and merged as `c058da69f2296e52a8e32cc25e195190389dbca7`; isolated reader only
- [TASK-071 — Runtime Owner Shutdown Provenance and Exact Attempt Snapshot](TASK-071-RUNTIME-OWNER-SHUTDOWN-PROVENANCE.md) — Completed, Coordinator Accepted; integration commit `e1234a8447d75d1ffddf07dd953d444956b15ff1` published through PR #77 and merged as `0cec13d8e2b310545d5e2af40286158fda9820e9`; isolated DP-014 prerequisite only
- [TASK-072 — Publisher Handoff Store Bootstrap and Recovery](TASK-072-PUBLISHER-HANDOFF-STORE-RECOVERY.md) — Completed, Coordinator Accepted; docs-only process repair commit `60ee9bec4dde3df08620f1a9d54b28da3f5b4c62` published through PR #76 and included in TASK-071 integration
- [TASK-073 — Runtime Shutdown-Completion Evidence Composer](TASK-073-RUNTIME-SHUTDOWN-EVIDENCE-COMPOSER.md) — Completed, Coordinator Accepted; task commit `37ca6d6b5065aeb1a4831bc30aa54503ae06ff64` published through PR #78 and merged as `a84098284b0202f2fea9a61e089a74f564a05405`; isolated full-tuple composer only
- [TASK-074 — Runtime Containment Composition, Admission, and Generation Provider Gate](TASK-074-RUNTIME-CONTAINMENT-COMPOSITION-GATE.md) — Completed, Coordinator Accepted; task commit `2d78d417784012e4e945e60f8d721246ea5bdcbf` published through PR #79 and merged as `bfab084c1a9664181027650b092bf240e04af435`; isolated private composition gate only
- [TASK-075 — Runtime Read-Only Recovery Assessment](TASK-075-RUNTIME-READ-ONLY-RECOVERY-ASSESSMENT.md) — Completed, Coordinator Accepted; task commit `f7b80ac976a144aa4918bd4ed74140e48fd31310` published through PR #80 and merged as `5eb54efea3f3df5c7967b2c1e2d0894c6289e50a`; isolated read-only assessment only; durable recovery authority and production wiring remain Not Activated
- [TASK-077 — Trusted Publisher Initial Route and Capability Preflight](TASK-077-TRUSTED-PUBLISHER-ROUTE.md) — Coordinator Accepted; commit c70c398 published through PR #82 as merge 5e40742; product capability unchanged
- [TASK-078 — Runtime Recovery Claim and Admission Barrier Readiness](TASK-078-RECOVERY-CLAIM-READINESS.md) — Completed, Coordinator Accepted; source 8687758 merged through PR #83 as 4dad1af; item7 code remains Not Ready
- [TASK-080 — Runtime Recovery Claim Design Status and Conformance](TASK-080-RECOVERY-CLAIM-DESIGN-STATUS.md) — verification-stable In Progress; matching envelope resolves gates/closure; published source eac6f6e through PR #85 merge ac82aa2; DP024 design Approved, no code activation by TASK080
- [TASK-079 — Runtime Recovery Claim and Admission Implementation Boundary](TASK-079-RECOVERY-CLAIM-ADMISSION-BOUNDARY.md) — accepted subject15047580 published source4ead8a82 through PR84 merge0d018c42; Draft refined by TASK080
- [TASK-076 — Local Configuration Studio](TASK-076-LOCAL-CONFIGURATION-STUDIO.md) — Coordinator Accepted source; committed as b8c2b86 and merged through PR #81 as e9a87b5; existing-API local/dev UI, no Runtime/DP-017 change

- [TASK-081 — Runtime Recovery Ordering and Fence Foundation](TASK-081-RECOVERY-ORDERING-FOUNDATION.md) — historical Coordinator Accepted source acefebe25d8a1053d6af0af00bbc1625014876c4; commit f5818c6e0500ede49fc1f3c4b314db63e16affc1 / PR #86 / merge 4a2a575577f180527fa7a9b5499da037b294bb06 by local immutable Git metadata; local test-owner foundation without real-owner/durable recovery claim at publication.

- [TASK-082 — Runtime Real-Owner and Generation Participation Readiness](TASK-082-REAL-OWNER-PARTICIPATION-READINESS.md) — Completed — Coordinator Accepted, exact accepted source aa97244d3f46cb7f535af69d6a86babe12780f26; commit 6cfb17bff4246d17e41107755ace0fa3b6d96850 / PR #87 / merge cf17b1728349bca9ac3916c88a00445f021cc4a9 by local immutable Git metadata; readiness only, full integration SPLIT REQUIRED.
- [TASK-083 — Runtime Real-Owner Participation Design](TASK-083-REAL-OWNER-PARTICIPATION-DESIGN.md) — historical Completed — Coordinator Accepted source 7a982b3b997cbe4d5c499829a355ce8824d8ce34; commit bafe45972bf0da17b10aaa8096cac9d8d59cdbe3 / PR #88 / merge fb906b335523b933d7cdf3c634ca9946bcb410f0 by local immutable Git metadata; design-only, no work resume.
- [TASK-084 — Runtime DP-014 Coordination Participation](TASK-084-DP014-COORDINATION-PARTICIPATION.md) — Completed — Coordinator Accepted; source `4ddaf8272b537b27b825d854d8cd0b325f49bd2e` / PR #89 merge `20f5ba532ba9a1dbaa16c00796662fc6191d0ff5` by local immutable Git facts; closed leaf/DP014 prerequisite, no integrated recovery.
- [TASK-085 — Runtime DP-015 Coordination Participation](TASK-085-DP015-COORDINATION-PARTICIPATION.md) — historical Completed — Coordinator Accepted; source `0d9c433086a96641eaff63dfa6d506f311bd69bb`, PR #90 merge `7daefc0b96c8127a7f884b521686998a2181d43a` by local immutable metadata; closed DP015 membership/client replacement, inert roots.
- [TASK-086 — Runtime Fatal-Source Coupling and Composition Supply Readiness](TASK-086-FATAL-SOURCE-COMPOSITION-READINESS.md) — historical Completed — Coordinator Accepted; accepted source `539567e4dd250b32094a708f919cbcf0649b50c8`, commit `f159fb9bbd6ba07a49465ac20c56b54aafd47954` / PR #91 merge `fac149e1fdcfcd402e5872ae60336b083f1e3b22` by local immutable metadata; design/readiness only.
- [TASK-087 — Runtime Fatal-Source Coupling and Closed Assembly](TASK-087-FATAL-SOURCE-CLOSED-ASSEMBLY.md) — historical Completed — Coordinator Accepted; source `e1e03cb9fa898d4a784de06886070ba6c6a49cae`, accepted manifest `11119b8e440809756783360455c8cc75472d6182`, PR #92 merge `ca01f7cc5b94c59d6402aeac0eba33e2070f003a` by local immutable metadata; source-bound fatal coupling and opaque closed assembly only, no integrated recovery readiness.
- [TASK-088 — Runtime Recovery Qualification and Durable Prerequisite Readiness](TASK-088-RECOVERY-QUALIFICATION-READINESS.md) — historical Completed — Coordinator Accepted; accepted manifest `5d66cadf5a65144df49356532f40806f787558b3`, source `2870188eb59c843886369cefa8ab589535bc77e4`, PR #93 merge `5aa367713c5e93abb752dcf25ff1d7b41ae7b5f0` by local immutable metadata; additive DP-024 §18 readiness only, no capability promotion.
- [TASK-089 — Current-Authority Handoff and Joint Observation Session Design](TASK-089-CURRENT-AUTHORITY-JOINT-SESSION-DESIGN.md) — historical Completed — Coordinator Accepted; accepted manifest `4bdbdeb29093339c9adcb6d42f19391c5c3868e2`, source `b010b00fc517e58177a1cbe46355168c558f471f`, PR #94 merge `0a9b5d4e9d33cbf30216999c5e6e8abc6dff4ccb` by local immutable metadata; additive DP-024 §19 design only, no runtime capability.
- [TASK-090 — Closed-Local Joint Session Implementation Readiness](TASK-090-CLOSED-LOCAL-SESSION-READINESS.md) — historical Completed — Coordinator Accepted; accepted manifest `dacc667c07fa126151fe49ea5b356a4edc7d9d06`, source `9a2aa9f33e9e9c7b4244024ab12437a27891a74b`, PR #95 merge `587b55bdda0aa4ee92f2ce1827481e8dab51da86` by local immutable metadata; additive DP-024 §20 readiness only, no runtime capability.
- [TASK-091 — Source-Minted Current-Authority Handoff](TASK-091-CURRENT-AUTHORITY-HANDOFF.md) — current bounded A1 Implementation, verification-stable In Progress; newest valid matching envelope resolves gates/closure; isolated source/issuer and actual keeper inspection, no durable qualification; A2 fresh dependent readiness/Size Guard Not Activated.

Новый агент начинает с корневого [`AGENTS.md`](../../AGENTS.md), а не с
отдельного task record.
