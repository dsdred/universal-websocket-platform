# DP-018: Operational error reporting и redaction Runtime

[English version](../../en/design/DP-018-runtime-operational-error-reporting-redaction.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Planned
- **Target Milestone:** Beta — Complete the Single-Node Runtime
- **Scope:** Approved contract ARCH-004 section 19(6) для безопасной
  operational failure projection через management, activation и recovery Runtime

Этот approved документ определяет focused contract, но не реализует его и не
активирует management integration или Production Activation.

## 2. Назначение

Контракты ownership Runtime уже сохраняют точные внутренние failures и durable
lifecycle/command outcomes. Операторам также нужен стабильный безопасный способ
понимать эти failures без Secrets, payloads, raw error text, process-local
authority или cross-scope facts.

Этот design определяет границу между authoritative domain truth и
operator-safe operational report. Он сохраняет ownership failure, делает
redaction fail-closed, обеспечивает deterministic replay и оставляет logs,
metrics, traces, audit storage, alerting и public transports заменяемыми
adapters.

## 3. Источники

Этот Approved design подчиняется:

- [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md), который
  различает valid decisions и operational errors и запрещает потерю lifecycle
  failures;
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md), который
  сохраняет identity причин startup и rollback;
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  особенно section 19(6);
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md),
  который исключает Secret values из Snapshot и provenance;
- [DP-013](DP-013-runtime-management-routing.md), [DP-014](DP-014-runtime-operational-identity-persistence.md),
  [DP-015](DP-015-runtime-management-command-idempotency.md),
  [DP-016](DP-016-runtime-activation-replacement-rollback.md) и
  [DP-017](DP-017-runtime-recovery-reconciliation.md).

При конфликте с этими источниками применяется источник более высокого уровня.

## 4. Scope

Этот design определяет:

- ownership failure и boundary reporting projection;
- stable operator-safe classification и severity;
- exact scoped correlation и anti-enumeration behavior;
- allowlist redaction и forbidden content;
- создание, publication, replay и delivery-failure semantics отчёта;
- cancellation, concurrency и recovery interaction;
- acceptance proofs и implementation gates.

Он не определяет HTTP error schema, logging library, metric names, trace
format, audit database, retention period, alert policy, UI wording, localization
system или concrete authorization policy.

## 5. Термины

**Authoritative fact** — immutable domain или lifecycle fact, принадлежащий
существующей boundary: validation result, Owner outcome, durable command
outcome, attempt state, recovery classification или conditional publication
result.

**Internal cause** — исходные component-owned error identity и cause chain. Они
остаются доступны только внутри trusted execution boundary своего owner.

**Operational report** — immutable operator-safe projection одного
authoritative fact. Это evidence для observation, а не lifecycle или command
truth.

**Report key** — stable opaque correlation tuple внутри одного authorized
scope. Он идентифицирует source fact и projection version без раскрытия payload
или process-local authority.

**Category** — stable semantic class, безопасный для operator decisions. Это не
global replacement для component errors.

**Public detail** — bounded allowlisted value, type и disclosure которого
определены этим contract. Arbitrary strings не являются public detail.

**Delivery adapter** экспортирует уже безопасный report в log, metric, trace,
audit sink, alert path или будущий API mapper. Он не получает raw cause.

## 6. Граница ответственности

Component, владеющий failure, сохраняет exact error identity, wrapping,
joining и cleanup behavior. Reporting boundary не нормализует errors до того,
как owner примет authoritative decision.

Только trusted projector может временно объединить authoritative fact, его
exact scoped correlation и internal cause для выбора allowlisted category и
detail set. Projector выпускает safe value и отбрасывает любую ссылку на cause.
Delivery adapters получают только это safe value.

Reporting boundary никогда не:

- меняет desired или actual state;
- terminalize или reopen command;
- claim lifecycle, recovery или delivery work;
- превращает unknown outcome в success или failure;
- возвращает report delivery как domain outcome component;
- становится universal error или policy engine.

## 7. Модель отчёта

Каждый report содержит только:

- projection schema version;
- opaque report key;
- authorized operational domain и opaque Workspace, Configuration, Runtime
  Instance, command, attempt и generation correlations, когда применимо;
- operation и phase из closed vocabulary;
- category, severity и resolution state из closed vocabularies;
- bounded allowlisted public details;
- authoritative fact revision или immutable outcome identity, если доступно;
- никакого delivery-attempt или first-versus-replay marker в semantic content.

Absence задаётся явно. Report не использует empty string одновременно для
hidden, unknown, not applicable и absent.

Time, host name, PID, address, stack, source file и goroutine не являются core
report fields. Delivery adapter может добавить envelope time или sink metadata,
но эти metadata не являются authoritative Runtime truth и не влияют на
correlation или replay.

## 8. Стабильные категории

Initial closed category set:

- `AuthorizationUnavailable` — authorization evaluation завершилась
  operational failure без раскрытия policy reason или существования target;
- `SourceUnavailable` — operational failure чтения authorized required source;
- `PreparationFailed` — Loader, Builder, binding или pre-Host preparation
  завершились failure;
- `StartupFailed` — Host startup или readiness завершились failure после
  начала Host ownership;
- `ShutdownFailed` — exact shutdown completion не достигнут или не доказан;
- `ExecutionLost` — bound execution завершился или стал недоступен без proof
  graceful completion;
- `RecoveryBlocked` — recovery evidence отсутствует, contradictory, stale или
  indeterminate и admission остаётся closed;
- `PersistenceUnavailable` — authoritative persistence operation definitively
  failed до своего effect;
- `OutcomeIndeterminate` — authoritative mutation могла commit и требуется
  exact inspection;
- `InternalFailure` — trusted owner вернул failure без более специфичной
  безопасной классификации.

Authorization denial, malformed input, not-found, immutable-intent/revision
conflict, другие valid negative decisions и satisfied/idempotent outcomes не
являются error reports. Будущий design может представить их отдельными
operational events, но они не должны ошибочно маркироваться как failures.

## 9. Severity и resolution

Severity выводится из authoritative fact, а не raw error text:

- `Info` не используется для operational error;
- `Warning` означает, что requested operation не завершилась, но authoritative
  state coherent и safety barrier не закрыт;
- `Error` означает failure operation или execution и возможную необходимость
  operator attention;
- `Critical` означает unresolved truth, closed safety barrier или
  contradiction, мешающую safe progress.

Resolution — одно из `Terminal`, `RetryableAfterChange`,
`InspectionRequired` или `Blocked`. `RetryableAfterChange` никогда не означает
permission automatic retry. Category сама по себе не обещает retryability.

## 10. Правила классификации

Classification использует authoritative boundary и phase, stable error
identity, где он определён, и proven durable state. Она никогда не парсит raw
error messages.

Primary category выбирается по authoritative owner и phase в exact precedence:

1. operational failure authorization evaluator -> `AuthorizationUnavailable`;
2. aggregate, command, attempt, phase, binding или recovery publication,
   которая могла commit -> `OutcomeIndeterminate`;
3. recovery assessment/evidence missing, stale, contradictory или unavailable
   без indeterminate mutation -> `RecoveryBlocked`;
4. definite no-effect failure aggregate/command/recovery-store operation ->
   `PersistenceUnavailable`;
5. operational acquisition failure Configuration Source до появления detached
   load result -> `SourceUnavailable`;
6. Loader identity/schema/semantic failure после source acquisition, Builder
   failure или pre-Load execution-binding failure -> `PreparationFailed`;
7. Host Start/readiness/rollback failure -> `StartupFailed`;
8. claimed Stop/shutdown-completion failure -> `ShutdownFailed`;
9. proven loss ранее bound execution -> `ExecutionLost`;
10. любая другая trusted operational cause -> fail-closed `InternalFailure`.

Error persistence-backed Configuration Source классифицируется по Source role
как `SourceUnavailable`, а не по storage technology. Durable command-store
error — `PersistenceUnavailable`; при uncertain commit это
`OutcomeIndeterminate`. Uncertainty recovery evidence — `RecoveryBlocked`,
если conditional mutation не могла commit; иначе выигрывает
`OutcomeIndeterminate`.

Один source fact создаёт одну primary category. Cleanup или rollback causes,
остающиеся различимыми во внутренней joined chain, могут создавать bounded
secondary safe facets только при явном allowlist mapping каждой причины. Они
не заменяют primary category и не раскрывают raw text.

## 11. Correlation и scope isolation

Projection выполняется только после current authorization, разрешающей
observation exact operational domain, Workspace, Configuration, Runtime
Instance, action и target version source fact. Stored authorization никогда не
replay как authority. Caller без authorization наблюдать exact fact не получает
operational report о нём; отдельный request outcome сохраняет anti-enumeration
semantics.

Opaque correlations включаются только когда consumer report authorized
наблюдать exact object. Operational failure authorization evaluation может
report только в уже authorized parent scope consumer без narrower correlation
или disclosure существования target.

Одинаковый opaque command key в другом command scope не связан. Lookup, replay,
aggregation и delivery отчёта никогда не join scopes только по raw identifier.

## 12. Политика redaction

Redaction — construction через allowlist, а не удаление из готового raw
message. Каждое field имеет fixed safe type, maximum size и disclosure rule.
Unrecognized fields, categories, owner types и causes опускаются и
классифицируются fail closed.

Всегда запрещённое содержимое report:

- Secret values, credentials, tokens, authorization headers, cookies, keys,
  certificates и private material;
- raw ConfigurationVersion, Snapshot, request, response, message, header,
  query, environment или repository payload;
- raw internal error text, formatting, cause chain, stack trace, source path,
  SQL, storage key или vendor response;
- Host pointers, contexts, goroutines, process-local permits, memory addresses,
  unrestricted PID/process metadata или socket endpoints;
- existence или state unauthorized либо cross-Workspace object;
- user-controlled labels или identifiers без отдельного field rule для
  bounded normalized safe representation.

Opaque domain identifiers являются correlation, а не Secrets, но подчиняются
exact scope authorization.

## 13. Public details

Initial allowed detail types — closed enums и bounded facts, уже безопасные в
authoritative state: operation, lifecycle phase, expected-versus-observed
revision relation, selected schema version и необходимость exact inspection.

Public detail никогда не содержит configuration field value, Listener address,
secret reference name, storage location, provider message или arbitrary error
parameter. Для deeper diagnosis privileged internal tooling может inspect
original owner boundary по отдельному authorization и audit design; это не
ослабляет данный report.

## 14. Порядок публикации

Report projected только после authoritative source fact:

- authorization-evaluation и source-access reports после authoritative
  operational failure, но не после ordinary denial или not-found;
- lifecycle reports после exact Owner outcome;
- durable command reports после соответствующего immutable command outcome;
- activation phase reports после publication этой phase;
- recovery reports после exact conditional reconciliation publication или
  proven blocked assessment.

Report не может объявлять Running, Stopped, Failed, command completion,
recovery release или cleanup до publication/proof owning boundary. Создание
report — downstream observation, не linearization point domain state.

## 15. Replay и deduplication

Report key выводится из exact scope, source fact identity/revision, projection
schema version, operation и phase. При той же projection schema version replay
того же immutable command outcome или recovery fact создаёт byte-equivalent
semantic report и тот же report key. Новая projection schema создаёт distinct
projection и key; она не переписывает older projection и не обещает sink-level
exactly-once identity между upgrades.

At-least-once delivery разрешена. Sink может deduplicate по report key.
Duplicate delivery не является duplicate lifecycle work, новым failure или
новым command outcome. Изменённая authoritative revision или отдельная failure
phase имеет отдельный key.

Contract не требует exactly-once delivery или global report store.

## 16. Failure доставки

Delivery происходит после safe projection. Delivery adapter не получает raw
cause и не может запрашивать lifecycle replay.

Definitive или indeterminate delivery failure:

- не меняет source domain fact;
- не превращает завершённую management command в failed;
- не reopen и не close DP-015 admission;
- не claim новый attempt, phase или recovery pass;
- может быть видим через bounded local health signal только с category sink
  identity и report key, без original raw cause.

Recursive reporting в failing adapter запрещён. Adapter health, buffering,
retry, backpressure, retention и loss policy требуют concrete delivery design.

## 17. Cancellation и concurrency

Cancellation до authoritative operational failure не создаёт error report для
operation, которая не произошла. Cancellation после authoritative fact может
завершить ожидание caller, но не retract fact или report eligibility. Ordinary
caller-cancellation outcome не relabel как operational error этим design.

Concurrent projectors одного fact converges к тому же report key и semantic
content. Они не разделяют lifecycle locks и не удерживают aggregate/command
locks во время delivery. Разные Runtime Instances progress independently.

Report, observed до более позднего state change, остаётся truthful history и
не переписывается. Later authoritative fact создаёт свой report.

## 18. Взаимодействие с recovery

Recovery DP-017 может report только assessed или conditionally published
facts. Stale Running record, PID, port, probe, timeout или process absence
никогда не report как current liveness или graceful shutdown proof.

`RecoveryBlocked` отличает unresolved truth от terminal execution failure. Он
не раскрывает contradictory evidence detail сверх safe phase и inspection
requirement. Report delivery никогда не release recovery claim или command
admission.

Crash-resumed reconciliation при той же projection schema reproject те же
already-published facts с теми же report keys и никогда не повторяет lifecycle
work. После schema upgrade new-version projection имеет distinct key и также
не выполняет lifecycle work.

## 19. Матрица failures

| Source condition | Safe report | Forbidden projection |
| --- | --- | --- |
| authorization denial | no operational error report | relabel valid decision as failure |
| authorization evaluation failure | `AuthorizationUnavailable` | reason, target existence, policy internals |
| malformed, not-found или conflict outcome | no operational error report | relabel valid decision as failure |
| operational Source acquisition failure до detached result | `SourceUnavailable` | source location, vendor response, payload |
| Loader semantic/identity, Builder или pre-Load binding failure | `PreparationFailed` | Configuration/Snapshot/cause text |
| definite no-effect durable-store failure | `PersistenceUnavailable` | storage location, vendor response, retry permission |
| Host Start/rollback failure | `StartupFailed` с safe rollback facet | joined raw causes |
| unproved shutdown completion | `ShutdownFailed` | `Stopped` или graceful claim |
| proven process loss | `ExecutionLost` | liveness, adoption, automatic restart |
| recovery evidence unresolved без mutation uncertainty | `RecoveryBlocked` | guessed terminal state |
| effect любой domain mutation unknown | `OutcomeIndeterminate` | success/failure inference или retry permission |
| unknown owner cause | `InternalFailure` | raw fallback message |
| report sink failure | separate bounded adapter health | mutation original outcome или recursion |

## 20. Technology neutrality

Projection, report key, category, safe detail и delivery adapter — semantic
requirements. Они не требуют structured logging, OpenTelemetry, Prometheus,
message broker, database, outbox, queue, audit product, error registry library
или global event bus.

Implementation не должна вводить universal diagnostics service с domain
mutation authority, service locator или зависимость Runtime internals от одного
observability product.

## 21. Концептуальные операции

Design требует capability, эквивалентной:

```text
ClassifyAuthoritativeFailure
ProjectSafeOperationalReport
DeriveScopedReportKey
DeliverSafeReport
ObserveDeliveryHealth
```

Это explanatory capabilities, а не API или Go interface names. Только
projection получает internal cause; delivery получает safe report.

## 22. Acceptance proofs

Будущая implementation должна доказать минимум:

1. component error identity и cause chains не меняются у owners;
2. valid negative и idempotent outcomes не становятся operational errors;
3. каждый report downstream одного authoritative fact;
4. owner-and-phase precedence maps каждый fact к одной primary stable category
   без message parsing, включая Source/Preparation, persistence и recovery
   overlap cuts;
5. unknown causes fail closed в `InternalFailure` без raw text;
6. authorization denial и другие valid negative outcomes не создают error
   report, а authorization evaluation failure не раскрывает target existence
   или policy detail;
7. cross-scope identifiers не могут correlate или retrieve report;
8. Secrets, payloads, raw causes, stacks, paths, process authority и
   unrestricted metadata не попадают в safe report;
9. allowlisted details bounded и typed;
10. startup и rollback failures остаются internally distinguishable, а report
    раскрывает только safe primary и optional mapped facets;
11. Stopped не report без exact shutdown-completion proof;
12. indeterminate mutation не становится success, failure или retry permission;
13. recovery-blocked reporting не открывает admission и не угадывает liveness;
14. replay одного immutable fact и projection version создаёт тот же semantic
    report и key; новая projection version создаёт distinct key;
15. duplicate delivery не выполняет lifecycle или command mutation;
16. delivery failure не меняет authoritative outcome и не recurse;
17. cancellation после fact publication не retract report eligibility;
18. concurrent projection converges без удержания domain locks во время I/O;
19. разные Runtime Instances progress independently;
20. EN/RU contract, matrices, gates и Planned status остаются aligned.

Proofs включают technically available redaction corpus, failure-injection,
concurrency, replay, cross-scope и adapter-failure scenarios. Они не разрешают
Production Activation.

## 23. Gates ARCH-004 section 19

Этот Approved design закрывает focused architecture design gate ARCH-004
section 19(6). Approved DP-014–DP-017 закрывают predecessor focused design
gates sections 19(2)–(5). Dependency-ordered design set approved, но reporting,
management integration и Production Activation отсутствуют. DP-013 Ready
только для bounded isolated implementation slice.

## 24. Явно отложенное

Отложено до focused decisions или implementation tasks:

- public error DTOs, HTTP status codes, UI text, localization и client
  compatibility;
- concrete authorization и privileged diagnostic access;
- log/metric/trace/audit/alert schemas, sinks, sampling, buffering, retry,
  backpressure, loss, retention и deletion;
- storage schema, outbox, transaction mechanics, migrations и adapters;
- operator identity, audit evidence, compliance policy и data residency;
- automated remediation, restart, rollback, retry/backoff, supervision и
  Production Activation.

## 25. Граница реализации

Implementation Status — Planned. Репозиторий не содержит report model,
projector, redaction implementation, delivery adapter, management API,
durable management store, recovery executor или production wiring.

Approval закрывает design gate section 19(6), но не реализует и не подключает
contract и не меняет отдельный status DP-013.

## 26. Решение

UWP сохраняет exact failures у существующих owners и раскрывает только
immutable scoped allowlist-constructed operational projection после
соответствующего authoritative fact. Reports используют stable semantic
categories, opaque authorized correlations, bounded typed details и
projection-version-scoped deterministic replay keys. Unknown content fails
closed.

Raw errors, Secrets, payloads, stacks, process-local authority и cross-scope
facts никогда не пересекают reporting trust boundary. Report delivery —
downstream observation: его failure не может изменить lifecycle/command truth,
разрешить retry, reopen admission или повторить work. Concrete observability и
transport products остаются replaceable adapters и later decisions.
