# DP-020: Готовность последовательности связывания оркестрации Runtime

[English version](../../en/design/DP-020-runtime-orchestration-binding-sequence-readiness.md)

## 1. Статус

- **Статус проектирования:** Draft
- **Статус реализации:** Planned

Этот focused design разделяет оставшиеся prerequisites Approved DP-019 — точную
авторизацию оркестрации, private managed invocation и связывание
OwnerClaim-to-DP-014 — на упорядоченные, независимо проверяемые
implementation-срезы и закрывает отложенные решения проектирования, которые им
нужны.

Он не определяет production-код и не меняет ни одну принятую семантику.
TASK-026 остаётся `Blocked by Architecture`.

## 2. Назначение

Approved DP-019 фиксирует инварианты ordering, ownership, concurrency, permit и
fail-closed для prerequisites оркестрации активации, но описывает tuple
авторизации, private managed invoker, managed Flow seam и последовательность
связывания только концептуально. Этот proposal фиксирует focused,
implementable разложение именно этих seam, чтобы каждый мог быть реализован и
независимо принят до повторного рассмотрения оркестратора DP-016.

## 3. Источники полномочий

Этот proposal уточняет, не переопределяя:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  section 19;
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md) — Owner остаётся
  единственным lifecycle authority и claimant попытки запуска;
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) — синхронный путь
  `PrepareStart -> Load -> Build -> Start` и место private continuation;
- [DP-013](DP-013-runtime-management-routing.md) — exact routing,
  authorization-before-mutation и planned private seam;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) — conditional
  aggregate revision и граница membership Launch Attempt / связывания
  execution-generation;
- [DP-015](DP-015-runtime-management-command-idempotency.md) — граница claim /
  replay / permit / unresolved-barrier, включая parent/phase core TASK-028 и
  Continue gate и rendezvous pending-Stop TASK-029;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) — неизменный
  ordering activation / replacement / rollback и acceptance proofs;
- [DP-017](DP-017-runtime-recovery-reconciliation.md) — fail-closed recovery;
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md) — Approved
  prerequisite contract, который этот proposal разбивает.

Accepted ADR и Active/Frozen architecture остаются authoritative. Этот Draft
никогда не переопределяет Approved-документ и не повышает и не понижает ни один
существующий Design или Implementation Status.

## 4. Область действия

Дизайн покрывает только focused разложение концептуальных seam Approved DP-019
§7–§8 и §14–§16:

- representation и ownership policy-neutral orchestration authorizer, validated
  значение authorization request, exact failed-authorization error contract и
  адаптация DP-013 authorization-before-mutation;
- package split и направление invocation private managed invoker против
  managed Flow seam, managed construction и per-call managed Start surfaces,
  immutable per-invocation lifecycle `StartExecutionBinding` и exact
  failed private-invocation error contract;
- immutable tuple `OwnerClaimView` и последовательность связывания
  OwnerClaim-to-DP-014: composition inputs, closed четырёх-outcome результат
  continuation, порядок conditional публикации membership попытки и
  same-generation binding относительно существующего rendezvous pending-Stop,
  и размещение final gate Stop-versus-Continue;
- доказуемый инвариант адаптации активации;
- план доказательства lock/ownership, требуемый перед реализацией;
- упорядоченный список bounded implementation-срезов и порядок их зависимостей.

## 5. Не-цели

Дизайн не определяет:

- production-реализацию любого среза, public route HTTP/CLI/DTO/API или wiring;
- concrete authorization policy, модель Principal, отображение credentials или
  transport;
- storage schema, migration, external persistence, cross-process lease или
  выполнение recovery;
- реализацию оркестратора DP-016, реализацию recovery DP-017, реализацию
  reporting DP-018, automatic rollback/restart, supervision, scheduling или
  Production Activation;
- любое изменение контракта lifecycle DP-010, базового поведения Flow DP-011,
  public surfaces DP-013/DP-015 или принятых семантик
  DP-014/DP-015/DP-016/DP-017/DP-019;
- любую новую концепцию domain Runtime Instance или domain развёртывания.

## 6. Краткое решение

Оставшиеся prerequisites DP-019 фиксируются как три упорядоченных
implementation-среза плюс одна gated переоценка:

1. exact поверхность orchestration authorizer на существующей границе команд
   DP-013 / DP-019;
2. private managed invoker и managed Flow per-call seam с per-invocation
   `StartExecutionBinding`;
3. conditional публикация попытки OwnerClaim-to-DP-014 и same-generation binding
   до Load, интегрированные с существующим rendezvous pending-Stop;
4. переоценка готовности оркестратора DP-016 после того, как срезы 1–3
   реализованы и независимо приняты.

Ни один срез не обходит другой. Срез 1 может быть реализован, проверен и принят
без срезов 2–4, но срезы 2 и 3 зависят от всех предыдущих.

## 7. Отложенное решение: Представления Orchestration Authorizer и Request

### 7.1 Значение authorization request

Одно immutable, validated значение является единственным входом авторизации:

```text
OrchestrationAuthorizationRequest {
    WorkspaceID                  uint64
    ConfigurationID              uint64
    RuntimeInstanceID            runtimeconfigload.RuntimeInstanceID
    Action                       OrchestrationAction
    TargetConfigurationVersionID uint64
}
```

Validation требует, чтобы каждая identity была ненулевой и exact, а целевая
версия была действительной Published identity, принадлежащей этой Configuration.
Validation failure, denial/failure/panic авторизации, absent scope или pre-claim
cancellation выполняют нулевую mutation command, aggregate и lifecycle.

### 7.2 Исключение поля `OperationalDomain`

Tuple Approved DP-019 §7 перечисляет `OperationalDomain`. Для single-node
baseline milestone ARCH-004 и всех Approved DP операционная aggregate-модель уже
ограничивает каждый Runtime Instance своей exact парой Workspace + Configuration:

- `runtimeidentity.RuntimeInstanceView` и родственные факты связывают один
  Runtime Instance с одним Workspace и одной Configuration (aggregate facts
  DP-014);
- `runtimemanagement.Target` связывает Workspace, Configuration и Runtime
  Instance без поля domain;
- `runtimecommandidempotency.Scope.domain` — opaque string без операционного
  смысла и уже валидируется вместе с остальным scope.

Добавление скрытой константы domain по умолчанию скрывало бы решение и нарушало
бы ограничения no-hidden-default и no-magic. Поэтому authorization request
намеренно исключает `OperationalDomain`; scope domain остаётся задокументированным
single-node упрощением до будущего milestone, а не реализованной capability.
Оставшиеся пять полей несут полный вход авторизации.

### 7.3 Представление authorizer

Seam — это один policy-neutral synchronous named function type, принадлежащий
будущему composition-facing orchestration package (не хранящийся в
`runtimemanagement.Directory` и не заменяющий существующий параметр DP-015
`Authorize(...) error`). Он валидируется при каждой submission, никогда не
кэшируется и повторно вычисляется при каждой initial и replay submission.

Соответствующий authorizer возвращает либо отсутствие ошибки, либо одну exact
ошибку. Denial, unavailability, invalid scope, panic и cancellation каждый
производят truthful, specific исход ошибки и нулевую mutation. Существующая
exported поверхность авторизации DP-013 для `Directory.Start/Stop/Observe`
остаётся неизменной; orchestration actions никогда не проходят через эту
поверхность без дополнительной exact orchestration authorization.

## 8. Отложенное решение: Private Managed Invoker и Managed Flow Seam

### 8.1 Package split и направление invocation

- Private managed invoker находится на стороне composition DP-013 и валидирует
  binding против своего immutable scope и `StartRequest` перед вызовом stored
  Flow.
- Managed Flow seam находится в `internal/runtimelaunchflow`. Существующий
  unmanaged конструктор `New(owner, loader)` и его семантика
  `Start(ctx, request)` остаются неизменными.
- Invocation — один synchronous per-invocation вызов
  `Flow.StartManaged(context, StartRequest, StartExecutionBinding)`.
- `runtimelaunchflow` не импортирует `internal/runtimecommandidempotency` или
  `internal/runtimeidentity`.

### 8.2 Immutable per-invocation `StartExecutionBinding`

`StartExecutionBinding` — immutable значение, сконструированное
permit-holding stack и переданное внутрь. Оно содержит validated
`OrchestrationAuthorizationRequest`, expected aggregate revision,
composition-owned exact `ExecutionGeneration`, identity parent и phase, когда
применимо, и opaque `StartRendezvous` для этого live primitive или phase
execution. Оно не содержит ни primitive, parent, phase или Stop permit, ни
preparation token, ни Host или Snapshot, ни полномочие cancellation context, ни
mutable состояние Owner. Оно валидируется до любой mutation Owner, удерживается
только на том synchronous call stack, вызывается не более одного раза и
инвалидируется при возврате; оно никогда не сохраняется как поле Flow.

Managed construction связывает существующий stateless private
`StartClaimContinuation` ровно один раз. Binding не создаёт запись Registry,
mutable slot, goroutine, detached callback или новое состояние lifecycle.

### 8.3 Механизм opaque `StartRendezvous` через границу package

Существующий primitive rendezvous pending-Stop в
`internal/runtimecommandidempotency` остаётся единственным владельцем своих
сигналов и своих lock, согласно DP-019 §13 и §17. Seam через границу package —
opaque, exported-but-internal тип handle без exported методов, определённый в
package с наименьшей зависимостью в цепочке import Flow (соседняя цепочка
запуска `runtimelifecycle`), поэтому:

- граница команд DP-015 конструирует concrete rendezvous и выставляет его только
  как opaque handle;
- `StartExecutionBinding` хранит opaque handle;
- `runtimemanagement` и `runtimelaunchflow` передают его, не импортируя package
  границы команд.

Handle не несёт capability, permit или mutable состояние.

### 8.4 Exact failed private-invocation error contract

Validation failure binding до mutation Owner возвращает exact, distinguishable
ошибку (fail-closed) и не выполняет mutation Owner или aggregate. Ошибки Owner
или continuation возвращаются неизменными и без обёртки; на этом seam нет ни
reclassification, ни wrapping, ни recovery.

## 9. Отложенное решение: OwnerClaimView, последовательность binding и исходы

### 9.1 `OwnerClaimView` и accessor `LaunchAttemptID`

Сразу после успешного authentic `Owner.PrepareStart` и до любой работы Load,
Build или Launcher Flow собирает один immutable `OwnerClaimView`:

```text
OwnerClaimView {
    WorkspaceID                  uint64
    ConfigurationID              uint64
    RuntimeInstanceID            runtimeconfigload.RuntimeInstanceID
    LaunchAttemptID              runtimeconfigload.LaunchAttemptID
    TargetConfigurationVersionID uint64
}
```

Доказательство accessor LaunchAttemptID присутствует уже сегодня: authentic
`LaunchPreparation.LoadRequest()` возвращает `runtimeconfigload.LoadRequest`,
который несёт `LaunchAttemptID()`. Claim view источён из этого Owner-issued
`LoadRequest`; изменение формы `runtimelifecycle.StartRequest` не требуется.

### 9.2 Закрытый исход continuation

`StartClaimContinuation.AfterOwnerClaim(StartExecutionBinding, OwnerClaimView)`
возвращает ровно один закрытый исход:

- `Continue`: already-admitted rendezvous pending-Stop окончательно отсутствует;
  exact durable membership попытки и same-generation binding закоммичены при
  своих expected revision; final gate Stop-versus-Continue отпустил Flow.
- `StopConverged`: original pending Stop claimant вызвал exact Stop и свёл
  заявленную попытку; Flow не начинает Load и не выполняет публикацию или
  binding DP-014.
- `BindingFailed`: coherent exact evidence доказывает, что публикация membership
  попытки или binding generation не закоммичены и конфликта нет, и внешняя
  подготовка не начиналась; Flow подаёт failure через authentic подготовку
  Owner.
- `Blocked`: потеря permit или signal, stale или conflicting revision, другое
  generation, unavailable или unknown facts, недоказанная convergence Stop или
  indeterminate публикация; Flow не начинает Load, и связанный набор остаётся
  unresolved.

Continuation никогда сам не публикует terminal исход lifecycle, phase или
parent; terminal их могут только exact результат Owner, за которым следуют
conditional публикации DP-014 и DP-015.

### 9.3 Порядок последовательности binding и threading revision

Фиксированный порядок, после единственного claim Owner и до Load:

1. разрешить already-admitted rendezvous pending-Stop только через original Stop
   claimant; `StopConverged` выходит без обеих записей, а `Blocked` оставляет
   связанный набор unresolved;
2. только после definitive доказательства отсутствия pending Stop, conditionally
   опубликовать exact membership Launch Attempt и pin версии в DP-014 при
   expected aggregate revision;
3. прочитать committed revision, возвращённый этой записью, и conditionally
   связать exact active попытку с composition-owned execution generation при
   этом новом expected revision;
4. после любого indeterminate результата проверить exact aggregate facts через
   `ReadRuntimeInstance` и `ReadLaunchAttemptHistory` и свести к exact
   существующему terminal исходу или вернуть `Blocked`;
5. выполнить final gate Stop-versus-Continue для Stop, admitted после ранней
   проверки rendezvous, затем выпустить `Continue` только при confirmed, exact
   same-generation binding.

Уже присутствующие same-attempt/same-version membership и same-generation
binding являются zero-mutation satisfied наблюдениями convergence. Другая
попытка, другая версия, другой generation, stale revision, inactive факт или
unknown состояние никогда не auto-repaired или auto-replaced.

## 10. Инвариант адаптации активации

Initial activation использует существующий primitive путь Start неизменённым:

- `Boundary.Execute` остаётся единственным admission команды для primitive
  submission Start;
- `runtimelifecycle.StartRequest` остаётся immutable существующей формой
  `(WorkspaceID, ConfigurationID, ConfigurationVersionID)` и не несёт authority
  parent, phase или authorization;
- composition-facing authorization adapter помечает exact authorization action
  submission Start как `ActivateExactTarget`;
- parent с одной phase не создаётся ради uniformity API; формы parent
  используются только для `Replace` и `Rollback`.

Replacement и rollback отображаются на `ReplaceWithExactTarget` и
`RollbackToExactTarget` через существующий путь parent `ExecuteParent`,
представленный TASK-028 и завершённый TASK-029. Нет fallback между actions и нет
кэшированного результата авторизации.

## 11. План доказательства Lock и Ownership

Перед принятием любого implementation среза ответственные роли должны доказать
focused тестами и статическим обзором каждого сайта `Lock`/`Unlock`, что ни один
из этих lock не удерживается через названные операции:

- ledger admission команд DP-015 и lock клиента хранилища;
- lock aggregate store DP-014 и lock per-aggregate;
- состояние Directory DP-013 (которое намеренно не имеет process lock);
- managed Flow DP-011 (который намеренно не имеет process lock);
- mutex Runtime Lifecycle Owner;
- любую синхронизацию private для нового среза.

Ни один не удерживается через: (1) callback authorizer; (2) операцию storage
DP-014 или exact inspection; (3) signal или wait continuation; (4) invocation
lifecycle DP-013 или Owner; (5) Load, Build, Launcher, Host или external I/O.

Ownership остаётся: external authorization policy принадлежит composition и
заимствуется на submission; записи parent/phase DP-015 и live permits остаются в
границе команд; факты aggregate и attempt DP-014 остаются в identity store;
Lifecycle Owner остаётся единственным принимающим решения lifecycle и owner live
Host; DP-013 остаётся exact composition routing/private-invocation; DP-011
остаётся synchronous последовательностью подготовки; rendezvous pending-Stop
остаётся в bounded исполнении оркестрации. Ни одна строка не передаёт ownership
оркестратору.

## 12. Упорядоченные Implementation-срезы (выход готовности)

### Срез 1 — поверхность orchestration authorizer

- Ввести validated значение `OrchestrationAuthorizationRequest`, named
  policy-neutral тип функции `AuthorizeOrchestration`, набор
  `OrchestrationAction` и exact failed-authorization error contract на
  существующей границе команд, не меняя public поверхности DP-013 / DP-015.
- Доказать: exact initial/replay авторизация, нулевая mutation при
  denial/failure/cancellation, zero-path активации на immutable primitive
  submission Start, и отсутствие public или private bypass.
- Никакого binding DP-014, никакого изменения Flow, никакого оркестратора.

### Срез 2 — private managed invoker и managed Flow seam

- Добавить managed construction и per-call seam `StartManaged` и immutable
  значения `StartExecutionBinding` / `OwnerClaimView`, opaque handle
  `StartRendezvous` и exact failed private-invocation error contract, используя
  Slice 1.
- Доказать: validate-before-Owner-mutation, invoke-at-most-once,
  invalidate-on-return, never-stored binding; неизменные unmanaged `New` и
  `Start`; никакого goroutine, записи registry или detached callback.
- Ещё никакой логики binding DP-014.

### Срез 3 — последовательность связывания OwnerClaim-to-DP-014

- Реализовать `StartClaimContinuation.AfterOwnerClaim` с использованием
  существующих conditional операций публикации/binding `runtimeidentity.Store`,
  существующего rendezvous pending-Stop и final gate Stop-versus-Continue,
  используя Slices 1 и 2.
- Доказать: membership попытки и same-generation binding до Load; stale/different/
  unknown facts дают `Blocked` без подготовки; definitive отсутствие binding
  сходится через authentic исход Owner; Stop, admitted после ранней проверки,
  упорядочен final gate.

### Срез 4 — переоценка готовности оркестратора DP-016

- Только после того, как Slices 1–3 реализованы и независимо приняты, переоценить,
  может ли TASK-026 быть разблокирована против неизменных proofs §25 DP-016.
- Этот срез не запускается TASK-030 и может заключить, что TASK-026 остаётся
  Blocked.

Каждый срез требует собственного intake задачи, Existing Coverage Report,
Verification Matrix, Independent Review, PROCESS-002 и Coordinator Acceptance.

## 13. Acceptance Proofs

Этот дизайн сам доказывает:

1. выбранное разложение сохраняет каждое обязательство acceptance proof Approved
   DP-019 §21 и отображает каждое оставшееся proof ровно на один срез;
2. существующие поверхности DP-013, DP-014, DP-015 и DP-016 не требуют semantic
   изменения для размещения трёх implementation срезов;
3. активация остаётся на primitive immutable пути Start с нулевым continuation;
4. locks доказуемо не удерживаются через forbidden границы;
5. зеркала EN и RU семантически равны, с равными заголовками и равными количествами
   code-fence, и каждая relative ссылка разрешается.

Эти proofs проверяются проверками документации, parity, ссылок, статуса и
regression, а также свежим Independent Review этого предложения.

## 14. Граница реализации

Implementation Status остаётся Planned. ОТЛОЖЕННЫЕ выходы — упорядоченные срезы
раздела 12; ни один не реализован этой задачей. Репозиторий всё ещё не содержит
orchestration authorizer, private scoped invoker, managed continuation
Flow/OwnerClaimView, composition публикации/binding попытки, оркестратор
активации, external persistence, API, worker recovery и production wiring. Поэтому
TASK-026 остаётся Blocked; successor-задачи должны реализовать и независимо
проверить эти срезы, прежде чем TASK-026 сможет быть пересмотрена против полного
неизменного набора proofs DP-016.

## 15. Последствия

Положительные:

- каждый оставшийся prerequisite DP-019 становится независимо проверяемым и
  независимо приемлемым;
- authorization, lifecycle ownership, durability ordering и permit ownership
  остаются explicit;
- здесь не принимается никакого production или public контракта.

Стоимость:

- три implementation среза предшествуют любой переоценке готовности DP-016;
- synchronous rendezvous pending-Stop может блокировать callers;
- restart процесса по-прежнему требует Planned реализации DP-017;
- production integration по-прежнему требует external durability и аудита
  composition.

## 16. Решение

UWP фиксирует разложение готовности оставшихся prerequisites Approved DP-019 в
этом Draft/Planned предложении и будет реализовывать срезы только через
отдельные, индивидуально reviewed задачи. Она не approximates DP-016 adapter-ом,
не добавляет операции replacement/rollback Owner, не передаёт permits, не меняет
ни один Approved статус или семантику и не выдаёт planned capability за
реализованную.
