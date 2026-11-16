# ADR 0005: Production Runtime Observation

[English version](../../en/adr/0005-production-runtime-observation.md)

## Статус

**Design Status: Draft. Implementation Status: Not Implemented.**

Предложение Architect A132-001, [TASK-132](../../tasks/TASK-132-PRODUCTION-RUNTIME-OBSERVATION-DESIGN.md).
Admission production implementation: **BLOCKED** до решений ниже.
Review design не утверждает ADR, не меняет действующие контракты и не активирует
TASK-131, Stage B/B1/B2, Production C2/C3 или C2-Exposure. Gate 11 остаётся OPEN.

## Контекст и repository evidence

На baseline `88fb333453ab9563285d65cd44a8b0a4019f447a`
[`Owner.Observe`](../../../internal/runtimelifecycle/owner.go) возвращает coherent
mutex-protected snapshot lifecycle одного Instance; авторизованный
[`Directory.Observe`](../../../internal/runtimemanagement/directory.go) реализован
изолированно. Actual states: stopped/starting/running/stopping/failed.
[`Control Service`](../../../cmd/control-service/main.go) создаёт configuration
services и HTTP/Studio routes без Owner/Directory/managed Runtime graph.
[`Publish`](../../../internal/configurationversion/service.go) обновляет versions;
[`health`](../../../internal/http/server.go) сообщает доступность Control.
Они не доказывают Runtime execution. Dashboard уже показывает Runtime Unavailable.

[`Host.Running/Ready`](../../../internal/runtime/host.go) читают host.state.
[`Listener.serve`](../../../internal/listener/listener.go) сообщает unexpected
Serve errors, но не меняет Host/Owner lifecycle при возврате Serve.
[DP-010 section 27](../design/DP-010-runtime-lifecycle-owner-contract.md) запрещает
выводить Running -> Failed через polling Running(). Process containment evidence
доказывает только exact declared generation facts, а не readiness Host.
Forensic table task record содержит symbols, baseline lines и изученные tests.

[ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) замораживает
readiness, lifecycle ownership, startup/rollback и shutdown order, оставляя
focused supervision design открытым. [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md)
разделяет Instance, Attempt, version и process identity; initial topology —
in-process. [DP-013](../design/DP-013-runtime-management-routing.md) откладывает
production composition, HTTP и concrete policy. Этот Draft не переопределяет
их или [ADR-0002](0002-configuration-dsl.md)/[ADR-0003](0003-runtime-architecture.md).

## Сравнение вариантов

| Критерий | A: existing supervisor observation | B: source-owned evidence | C: read-only validity adapter |
|---|---|---|---|
| Изменения | Отдельно допустить real containment/evidence composition; готового production supervisor нет | Retained completion capability на Host/Listener | Exact-target capture, identity, validity и transport boundary |
| Source of Truth | Existing exact-generation authority | Owner lifecycle; Host-owned component facts | Заимствует Owner/B/A, не владеет Runtime state |
| Failure detection | Только generation termination в declared guarantee | Возврат underlying Serve; не смерть всего процесса | Самостоятельно отсутствует |
| Concurrency | Exact-generation authority revalidation | One-shot retained fact, без lifecycle callbacks | Revision/attempt recheck; fail-closed при overlap |
| Stale state | Durable binding не равен свежему Host health | Completion доступен без потери очереди | Expiry снимает current presentation |
| Performance | Platform/authority I/O требует отдельных bounds | O(1) bounded local publication/read | Bounded local capture; без polling worker |
| API | Recovery evidence не становится health API | Private read-only capability, без lifecycle commands | Versioned allowlisted DTO, no-store, concealed denial |
| Testability | Generation/unknown/platform qualification | Deterministic Serve/start/stop/reporter races | Identity, age, timeout, out-of-order, policy tests |
| Compatibility | DP-017/022 guarantees неизменны | Ready, Owner state, cleanup и Stop неизменны | Existing Start/Stop/Observe signatures сохраняются |
| Protected dependencies | Собственный production admission; TASK-102 не обязателен автоматически | Без diagnostics C2/C3 или Stage B | Отдельному observation path Stage B необязателен |
| Ограничения | Current production supervisor отсутствует; живой process не равен readiness Host | Serve return не равен полному shutdown Host или deadlock detection | Не создаёт отсутствующее source evidence |

A применим как future supplemental generation source, а не готовый supervisor.
C отдельно способен показать lifecycle facts с execution Unknown, но не решает
unexpected execution loss. Надёжность требует source evidence и явной validity,
а не выбора простого endpoint wrapper.

## Предлагаемое решение: B плюс C

Lifecycle authority остаётся у существующего Owner. Добавляются bounded retained
source fact и read-only adapter. Не добавляются второе state repository,
event-sourced lifecycle registry, diagnostics consumer, background supervisor
или automatic terminalization. Всё ниже — предложение, а не current API.

### Семантика состояний и evidence

| Значение | Значение и источник |
|---|---|
| Running | Owner actual running: опубликован успешный startup. Не continuous liveness, traffic health или отсутствие deadlock. |
| Starting / Stopping | Existing Owner launch/shutdown work; progress и bounded completion не выводятся. |
| Stopped | Authoritative Owner stopped outcome с его resource-ownership semantics; не missing data/PID. |
| Failed | Existing Owner failure outcome; observation не создаёт этот transition из component event. |
| Unknown | Observation boundary доступна, но evidence недостаточен, stale, mismatched или contradictory. Не новый Owner state. |
| Unavailable | Observation/source недоступен, не настроен или не авторизован; утверждения об execution нет. Не Stopped. |

Результат разделяет `lifecycleState`, `desiredState`, `executionEvidence` и
`validity`. `executionEvidence` — unknown либо allowlisted fact, например
`listenerServeReturned`, с явным proof scope. `liveConfirmed` запрещён:
отсутствие completion marker не доказывает полезное execution. Lifecycle
running может сосуществовать с Serve-return evidence; UI обязан раскрывать
различие evidence level и не показывать live-success badge.

Process liveness, Host lifecycle readiness, configuration publication и
Control availability разделены. Initial design не измеряет application progress
или continuous health. Даже future process-alive proof не докажет usability
Listener. Published != Running; Health != Runtime liveness;
Snapshot != current liveness; Unknown != Stopped.

### Identity, source ownership и restart

Target — exact `(workspaceID, configurationID, runtimeInstanceID)`.
Observation явно различает active/no-active attempt; last attempt исторический
и не подменяет active. Version принадлежит pinned Attempt. Одну version могут
исполнять несколько Instances. Запрещены latest-version inference, implicit
list/discovery, arbitrary PID/address probe и фиктивный stopped Owner.

`launchAttemptID` и execution generation включаются только при exact binding.
Generation заимствуется из sanctioned execution composition и не выделяется
observation. Existing DP-017/022 generation authority отличается от Host/Attempt
и не доказывает health. Без binding generation claim отсутствует. Restart
Control invalidates process-local observation; request не запускает recovery
или reconciliation. Old response не связывается с новым Attempt даже при
повторном использовании PID/version/address.

Каждому observation, включая no-active-attempt, дополнительно нужен
`observationSourceIncarnation`: trusted composition-issued identity одного
source process lifetime, без reuse при замене source. Это observation provenance
epoch, а не Runtime identity, execution/containment generation, recovery
authority или proof alive/ready. Observation использует qualified identity,
не выделяя independent execution generation. Uniqueness/nonreuse issuer и
authenticated current-source bootstrap должны быть доказаны при production
admission; PID, clock или assertion unqualified random identifier недостаточны.
Missing issuer или unverifiable provenance дают Unknown без current-success
presentation.

Request содержит `expectedObservationSourceIncarnation` из authenticated current-source
bootstrap; сам bootstrap не даёт successful Runtime-status claim.
Capture и каждый response связываются с той же identity. Mismatch
даёт Unknown, а не silent rebinding. Client binding — `(source incarnation,
exact target, attempt/no-attempt, request generation)`; new incarnation или
bootstrap failure снимает old presentation до rebinding. Old-incarnation
response отбрасывается даже при одинаковых Target/Attempt strings. Unknown
current-source identity запрещает current claim. Это provenance на bootstrap
и capture, а не гарантия выживания source до display.
Source binding immutable внутри adapter. Смена incarnation инвалидирует pending
requests; новый bootstrap требует нового pinned request. Delayed bootstrap
response не заменяет current binding вне своей request generation.

B сохраняет один typed completion fact на lifetime Listener/Host. Запись один
раз сразу после **любого** underlying Serve return, включая nil/closed, до
diagnostics и arbitrary legacy reporter. Publication bounded, non-consuming,
без callbacks, I/O, cancellation, Stop или Owner state publication. Host
связывает capability со своим actual component graph. Missing coverage даёт
Unknown; blocked/panicking reporter не скрывает публикацию. Serve completion
не доказывает handlers drained, Host cleanup или resources released. Retention
связан с тем же owned Host reference; нет replay/history store или cross-Attempt
reuse capability.

### Capture и concurrency

1. Authentication и exact observation authorization предшествуют target-aware
   external disclosure. Это proposed transport gate: current Directory ищет
   internal scope до Authorize, хотя доступ к Owner выполняет после него.
2. Qualified source incarnation pin-ится; Owner snapshot и proposed private monotonic observation revision захватываются
   под его mutex. Revision растёт при existing relevant lifecycle/attempt
   publications, а не при чтении. Overflow даёт fail-closed.
3. Exact bound Host evidence читается вне Owner mutex. Запрещены nested
   cross-owner locks, external callbacks, network I/O и lifecycle waits.
4. Повторная проверка source incarnation, Owner revision/Attempt и source revision. Любое изменение,
   missing binding или неоднозначный порядок дают Unknown(`ConcurrentChange`/
   `CoverageMissing` и т.п.) без retry/spin. Stable sample покрывает bounded
   capture interval, а не global atomic process/Owner cut или future execution.
5. Cancellation, policy failure и timeout отбрасывают sample; прежний success
   не используется. Commands, state repair и shutdown отсутствуют.

### Предлагаемая read-only boundary и ошибки

До endpoint implementation отдельно утверждаются exact route/schema,
target provisioning/discovery, real authentication/authorization и ingress
bounds. Existing trusted-local unauthenticated Studio не даёт автоматически
полномочий на operational target disclosure. Minimal conceptual response:
schema version, mandatory source incarnation, exact target/attempt binding, optional bound generation,
lifecycle/desired state, evidence/proof scope, validity/closed reason,
`observedAt`, capture duration и maximum presentation age. Без raw Host,
cancellation capability, Configuration payload, Secret, PID, path и raw error.

Proposed request budget — 2 секунды; максимум 8 admitted requests globally и
1 per authenticated principal, immediate refusal при превышении. Logical bounds
не ограничивают pre-handler ingress или hostile backend: network limits,
shutdown и prompt local callbacks требуют отдельных production proofs. Нельзя
оборачивать blocking source безлимитными goroutines. Без retry, queue и worker.

Closed reasons: NotConfigured, Unavailable, Cancelled, Busy, ConcurrentChange,
ScopeMismatch, GenerationMismatch, SourceIncarnationMismatch, CoverageMissing, Stale. Unauthorized и
nonexistent targets дают одинаковый generic external failure; concrete HTTP
codes/DTO/resource limits остаются production-admission blockers, а не
реализованным endpoint. Ошибка не возвращает old successful sample.

### Время и presentation validity

`observedAt` — server UTC в конце capture; capture duration использует server
monotonic clock. UTC — информация, а не generation/lease или authority.
Proposed maximum presentation age: 5000 ms, UI policy, а не proof liveness.
Нельзя устанавливать freshness сравнением browser/server wall clocks.
Client консервативно считает age от monotonic request start, включая весь
round trip; over-age response не становится current. Source validation и
response serialization должны укладываться в request budget.

Каждый refresh/target selection имеет client request generation; obsolete
responses отбрасываются. New request, error, timeout, потеря coverage, смена
identity, tab resume и expiry снимают current presentation. После expiry —
Unknown и optional явно historical last sample; manual Refresh достаточен.
Client timers не гарантируют работу при suspended page: на resume claim
снимается до redisplay. Даже unexpired response означает лишь «Lifecycle
running — observed at …» с execution proof level, а не «жив сейчас». Exit
между sample и display принципиально возможен.

## Анализ отказов

| Сценарий | Обязательное правдивое поведение |
|---|---|
| Unexpected Runtime exit | Смерть всего in-process процесса лишает observer/API. Serve return публикует source completion, а не automatic Failed/Stopped или cleanup success. |
| Supervisor недоступен | Optional authority missing: execution Unknown/Unavailable; доступный lifecycle fact маркируется отдельно. |
| Restart Control | New qualified source incarnation; одинаковые Target/Attempt strings не восстанавливают old binding. Без snapshot hydration или implicit DP-017 recovery; old samples historical. |
| Restart Runtime | Новый Attempt/Host, old completion/response отбрасываются; in-place Host restart не появляется. |
| Stale snapshot | Expiry снимает current presentation независимо от UTC label или прошлого HTTP success. |
| Observation потеряно | Retained cell не теряет событие как lossy queue; missing capability/coverage даёт Unknown, не default alive. |
| Exit между read и display | Абсолютной гарантии нет; показываются sampled facts и age, а не promise current liveness. |
| Concurrent start/stop/exit | Existing lifecycle arbitration неизменен; mismatch даёт Unknown, а не второй Stop или fabricated ordering. |
| PID reuse | PID не identity/routing; нужны exact Instance/Attempt и доступный generation binding. |
| Dashboard сохраняет Running | Generation guard, no-store и expiry; error/refresh/resume снимает current claim; history явно отделена. |

## Admission, решения Owner и условный roadmap

**BLOCKED для полной production implementation.** Additive B+C способен
сохранить ARCH-002: Ready == Host Running, startup commit, Gate/context,
rollback, shutdown, resource ownership и Owner transitions неизменны. Focused
read-only capability возможна без blanket frozen-foundation amendment.
Но этот Draft не approved implementation contract; отсутствуют actual
production graph, target provisioning и concrete security/transport admission.

Нужны решения Owner: выбрать sampled lifecycle с явно unknown execution либо
отдельно потребовать stronger supervision; отдельно утвердить focused design;
выбрать real managed deployment/target graph и authentication/authorization
boundary. «Failed сразу после exit» и «Running значит healthy сейчас» требуют
отдельного lifecycle/progress design. Универсальный prerequisite TASK-102,
Stage B или diagnostics C2/C3 не установлен. Observation не выполняет и не
обходит Gate 11 diagnostics.

| Условный slice (без Task ID/activation) | Acceptance criteria до downstream use |
|---|---|
| R1: approval design и production boundary decisions | Exact proof vocabulary, frozen compatibility, реальные graph/security owners; без implicit implementation permission. |
| R2: isolated retained source capability | Deterministic all-Serve-return, pre/post-start, Stop overlap, blocked/panicking reporter и retention proofs; без lifecycle/cancellation effects; full Go/race/stress/frozen regressions. |
| R3: isolated capture adapter | Exact tuple/revision/generation/source-incarnation, overflow, cancellation, capacity, redaction и missing evidence proofs; same Target/Attempt после source restart, delayed bootstrap/old response, missing issuer, wrong pin, duplicate/reused identity и source replacement во время capture; без retry/worker/new state store. |
| R4: отдельный production composition/transport admission | Real Owner graph/grants, qualified nonreused source-incarnation issuer и authenticated bootstrap, approved route/DTO/ingress/shutdown limits, security/API/source-to-operator proofs; STOP без real graph/provenance. R2/R3 не заменяют R4. |
| R5: Dashboard после R4 | Browser actual-contract/expiry/resume/error/old-response tests, publication/health separation и Studio regression; без controls. |

Каждому slice нужны fresh intake, coverage report и Size Guard; split при >500
production lines, >15 paths или нескольких independent contracts. Roadmap —
условный dependency plan, а не Ready implementation tasks или successor queue.

## Последствия

Authoritative lifecycle остаётся у одного owner, source completion не теряется
из-за diagnostics consumption, stale data не превращается незаметно в новый
live claim. Цена — дополнительная focused capability, revision/recheck discipline
и явный UI evidence vocabulary. Progress/health, whole-process supervision,
durable recovery и automatic terminalization остаются открытыми.
Existing Approved/Active/Frozen sources и accepted ADR не меняются.
Independent review и exact documentation manifest фиксируются в TASK-132;
acceptance design evidence не повышает Draft и не активирует работу.
