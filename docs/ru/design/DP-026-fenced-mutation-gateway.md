# DP-026: Fenced Mutation Gateway

[English version](../../en/design/DP-026-fenced-mutation-gateway.md)

## 1. Статус и полномочия

- **Design Status:** Draft
- **Implementation Status:** Planned (not implemented)
- **Task:** [TASK-098](../../tasks/TASK-098-FENCED-MUTATION-GATEWAY-DESIGN.md)
- **Исходное исследование:** [DP-025](DP-025-controlled-publisher-recovery.md), без изменений, Draft / Planned.

Документ фиксирует candidate B по handoff Architect для нового требования владельца проекта,
а не confirmed governance gap. Он не изменяет PROCESS-001/002 или agent contracts.
C, существующий fail-closed STOP, остаётся единственным действующим baseline при потере owner.
A Direct Publisher остаётся UNKNOWN относительно полного fencing; B CONDITIONALLY FEASIBLE
для prospective research при выполнении приведённых proof obligations, но не готов к deployment.
Все состояния, поля, permits и transitions здесь — предлагаемые design labels,
не native event formats и не разрешение создавать события. Acceptance/publication дизайна
не активирует DP-025/026, не восстанавливает прежний run и не разрешает recovery.

## 2. Проблема, threat model и non-goals

После отказа Publisher могут сохраняться credentials, дочерние процессы, queued commands
и принятые Git/GitHub requests. Новый дизайн разделяет user authority, native ownership,
claim reservation, gateway admission, фактическую mutation eligibility и terminal outcome.
Native CAS или чтение epoch не являются fencing внешних mutations. В частности,
проверка epoch перед API call не мешает приостановленному executor вызвать API позднее.

Угрозы: revival агента/gateway, competing successors, stale reads, retained credentials,
ABA, lost responses, partitions, отказ процесса/host и backup rollback.
Candidate предполагает доверенных host administrator, credential authority и authenticated
admission issuer; защита от их компрометации находится вне указанной модели.
Будущий дизайн должен отклонить activation, если эти assumptions не выполняются.

Safety означает exclusive authorized mutation, сохранение provenance и отсутствие blind replay.
Liveness условна: неустранимая неопределённость может блокировать навсегда. Timeout, lease expiry,
completed-run status и отсутствие записи не доказывают отсутствие effects.
Non-goals: implementation, provider configuration, mutation PoC, live TASK-095/096 exercise,
normative amendment, automatic HA, retrospective source Release и history rewrite.
Локальные Runtime модели DP-023/024 не квалифицируют Publisher provider capabilities.

## 3. Evidence и provider capability ledger

Приоритет repository не меняется: [PROCESS-001](../../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md)
и [Publisher](../../engineering/agents/publisher.md) определяют native ownership и P0–P10;
[PROCESS-002](../../engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md) — durable navigation.
PROCESS-001 owning-context tuple (с1208), source-authored handoff (с1484),
canonical store/readback (с1559) и terminal owner prerequisites (с1714)
являются действующими contracts, не gateway capabilities. Line anchors относятся к baseline
этой задачи; смысл ограничен указанными разделами.

Ledger сохраняет bounded prior observations Provider Capability Assessment.
Он не заявляет fresh deployment qualification. PROVEN — узкая документированная
гарантия либо bounded observation; NOT SUPPORTED относится только к названному интерфейсу;
UNKNOWN означает недостаток evidence, а не отсутствие функции.

| ID | Authoritative source / observation | Квалифицированный domain и ограничение | Deployment qualification |
| --- | --- | --- | --- |
| G01 | [Codex App Server](https://learn.chatgpt.com/docs/app-server): thread/read, thread/resume, turn/interrupt | Чтение/reopening/cancellation не документируют восстановление original owning execution binding или fencing внешних requests | UNKNOWN; нужны exact supported endpoint/version и identity bridge |
| G02 | [git-update-ref](https://git-scm.com/docs/git-update-ref): old OID и transactions | Conditional local refs; concurrent multi-ref readers могут увидеть subset, а не общий snapshot | UNKNOWN для gateway authority store и filesystem fencing |
| G03 | [git-push](https://git-scm.com/docs/git-push): explicit expected-OID lease, atomic при поддержке server | Только ref transport predicates; это не permission на force-push и не API/credential fence | UNKNOWN atomic capability exact server |
| G04 | [GitHub GraphQL Git](https://docs.github.com/en/graphql/reference/git): updateRefs / beforeOid / afterOid | Документированный atomic multi-ref conditional update/delete, не PR creation/merge или native events | UNKNOWN end-to-end deployment; read-only schema observation 192e51, 2026-10-17T22:45:29Z — bounded evidence |
| G05 | [REST best practices](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api) | Generic unsafe-method conditional requests NOT SUPPORTED без explicit endpoint exception; ETag GET не write CAS | Общая conditional mutation guarantee отсутствует |
| G06 | [PR API](https://docs.github.com/en/rest/pulls/pulls), merge head predicate | Expected head не Publisher epoch, owner или fixed-base CAS; async merge request UUID/status документирован, но здесь не квалифицирован | UNKNOWN deployment и lost-response correlation |
| G07 | [GraphQL mutations](https://docs.github.com/en/graphql/reference/mutations), clientMutationId | Correlation не доказывает deduplication или exactly-once PR creation | UNKNOWN exact operation retry guarantee |
| G08 | [REST refs](https://docs.github.com/en/rest/git/refs), ref deletion | REST DELETE не имеет expected-OID input; read-then-delete имеет race; G04 — conditional candidate | UNKNOWN выбранный delete path; mutation test отсутствует |
| G09 | [GitHub App installation token](https://docs.github.com/en/rest/apps/installations) | Credential revocation отличается от cancellation принятых requests и fencing всех cached/alternate credentials | UNKNOWN полная capability revocation/drain |
| G10 | [SQLite isolation](https://www.sqlite.org/isolation.html), [atomic commit](https://www.sqlite.org/atomiccommit.html), [WAL](https://www.sqlite.org/wal.html) | Только candidate transactional store; filesystem/configuration assumptions, WAL не network-filesystem protocol | UNKNOWN selected store/configuration; external cross-system transaction отсутствует |
| G11 | Действующие contracts native persisted readback и trusted route | Actual actor/history/run/route/principal и independently read-back items, не только SID/sessionId | UNKNOWN новый gateway issuer/native bridge и conditional append |

Latest documentation и schema не сертифицируют installed-client compatibility.
В assessment наблюдались Desktop 0.162.0-alpha.17.2 и npm CLI 0.128.0; это historical
version observations, а не утверждение о текущей паре. App Server здесь не запускается/resume.
Native identity, durable readback, store CAS, linearizability, old-executor fence,
trusted admission и crash recovery требуют отдельных evidence. Conditional Git/GitHub
refs не сертифицируют остальные domains. Universal exactly-once, immutable storage,
atomic external acceptance и automatic cancellation не предполагаются.

## 4. Trust boundary и полный mutation-path inventory

Candidate — один single-active gateway с mutation credentials и отдельными
Git workspace, index и refs. Caller передаёт authenticated intents, а не shell commands,
произвольные URLs или credentials. Только gateway валидирует, записывает, исполняет и reconciles.
Read-only observer не имеет write capability. Обязательны и credential, и filesystem
isolation: API mediation сама по себе не защищает checkout или local refs.

Coverage inventory должен закрыть Git CLI, REST, GraphQL, browser/MCP, hooks, subprocesses,
automation, alternate workspaces, cached/shared credentials и direct native emitters.
Privileged bypass либо устраняется, либо явно включается в trusted enforcement boundary.
UNKNOWN bypass coverage запрещает activation. Hooks/child processes требуют bounded approved
behavior и lifetime tracking; отключение одного observed path не доказывает отсутствие других.

| Path | Executor / credential boundary | Predicate и observable effect | In-flight, retry и fencing obligation |
| --- | --- | --- | --- |
| Push | Gateway Git transport credential / remote ref policy | Exact Target range и task ref; remote OID равен authorized head | Delayed old push должен быть fenced/drained; uncertain reply требует exact ref/request readback |
| PR creation | Gateway API credential / PR-write scope | Сначала exact repo/head/base discovery; сохранить unique PR ID и exact head/base | Lost reply не разрешает create retry; clientMutationId только correlation |
| Merge | Gateway API credential / contents-write и effective rules | Fresh head/base/checks, MERGEABLE / CLEAN, разрешённый merge commit; наблюдать MERGED SHA/parents | Expected head не owner/base CAS; accepted request может завершиться после revoke; сначала reconcile |
| Remote ref deletion | Gateway Git/API credential / exact ref | Текущий authorized head; conditional G04 candidate либо отдельно квалифицированный path; наблюдать absence | Recreated/moved ref сохраняется; REST read/delete race нельзя назвать fencing |
| Fetch/prune | Gateway remote-read credential и exclusive local writer | Exact origin; objects, tracking refs и FETCH_HEAD наблюдаются раздельно | Local writes являются effects; partial state и subprocess lifetime требуют reconciliation |
| Checkout main | Gateway exclusive workspace/index/HEAD writer | Exact clean prerequisites; наблюдать main, index и worktree | Git locks не fence всех writers; caller/stale gateway не имеет filesystem access |
| Fast-forward | Gateway transport и exclusive local writer | Только ff-only, разрешённая ancestry; наблюдать main/origin/main и worktree | Fetch и local update — разные crash domains; reset/rebase не используются для repair |
| Local branch deletion | Gateway exclusive ref writer | Exact task branch, merged proof, только safe -d; наблюдать absence | Recreated/moved ref блокирует delete; ref lock не global ownership |
| Terminal event | Future approved attributed native transport / authenticated issuer | Proven current native owner, P9 и exact terminal payload; independent persisted readback | Journal не фабрикует Publisher event; lost append требует exact readback до retry |

Inventory включает всех P0–P10 writers, а не только remote operations. Gateway credential
не передаётся agents. Final credential choice, scopes, issuance, rotation, revocation,
secret storage и process isolation остаются UNKNOWN до отдельной qualification.
Host path или SID не являются authenticated caller identity либо exclusive mutation authority.

## 5. Identity, user authority и immutable Target

Admission должен связать actual actor, original history/branch identity, native owning run,
authenticated route, actual principal и gateway caller identity. Trusted transport/issuer
evidence должен устанавливать эти поля; client-supplied tuple — лишь assertion.
Read-only transcript access, same conversation/session, same SID или completed run не
устанавливают continuity. Gateway не может impersonate прежнего Publisher.

Раздельные axes: scoped user recovery authorization и version; revocation version;
native ownership; unique claim/predecessor version; fence evidence; operation state;
terminal disposition. Old publication permission не переносится автоматически.
Требуются future approved carry-over policy и explicit scoped recovery authorization;
Draft не разрешает эти вопросы молчаливым принятием historical permission.

Immutable Target включает publication class/task/repository, exact branch, fixed base,
ordered commits, head/tree, accepted manifest и acceptance identity. Каждый intent
связан с exact payload, phase и Target digest. Target change invalidates affected authority
и queued work; implicit retargeting или source commit rewrite отсутствуют.
Original records неизменны; successor создаёт только self-attributed records с
original/replacement provenance и explicit observation cuts.

## 6. Candidate state machine, claim и admission

Все transitions требуют independent authoritative readback; Unknown или conflict блокирует
execution в любом состоянии. Proposed admission point не меняет действующий
InitialDispatch / Release / Route / Accept ownership protocol.

| State / actor | Exact predecessor и proof | Crash / rejection rule |
| --- | --- | --- |
| BlockedObserved / observer | Complete transaction chain, exact Target, last proven phase и operation inventory | Missing lineage остаётся UNKNOWN; удобный owner не выбирается |
| Authorized / authority recorder | Actual scoped user recovery authorization, candidate, authority/revoke versions | Permission carry-over не предполагается; revoke/Target change отклоняет |
| ClaimReserved / conditional store | Unique claim и expected predecessor/version, Target, authority и revoke versions в одном qualified serialization domain | Reservation не permission; один winner, lost reply проверяется readback |
| PriorExecutorFenced / enforcer | Proven rejection всех prior capabilities, children, queues и stale gateway instances; inventory accepted requests | Partial fence, retained credential или missing request блокирует admission |
| EffectsReconciled / reconciler | Authoritative outcomes/drain каждого potentially in-flight request и partial local effect | Uncertain остаётся blocked; completed effect не повторяется |
| SuccessorAdmitted / admission issuer | Conditional commit проверяет unchanged claim, fence, outcome inventory, Target и authority/revoke versions | Proposed linearization point: durable admission commit и independent readback; native bridge требует future normative approval |
| Executing / gateway | Exact current admitted epoch/context, fresh phase prerequisites и durable operation intent | Stale caller/queued intent отклоняется; после interruption inspect-first |
| TerminalVerified / attributed emitter и reader | Current native owner, all phases proven, clean synchronized main, independently persisted terminal | Нет ownerless P10/fabricated source event; uncertain append reconciles |

Unique claim identity и monotonic version должны исключать ABA. OID-only predecessor check
не доказывает отсутствия промежуточных изменений. Claim/admission/revoke order атомарен
только внутри qualified store serialization domain, никогда одновременно GitHub и native provider.
Cancellation claim после fencing не восстанавливает prior owner и не размораживает credentials.

Отказ agent при proven active gateway отличается от отказа gateway.
В первом случае допустимо проектировать rejection stale agent intents и reconciliation
существующего journal до отдельно authorized successor; во втором нужны evidence fencing/drain
прежнего gateway, children, credentials и accepted requests до нового dispatch.
Automatic failover или TTL-based takeover не предлагаются.

## 7. Durable journal и revoke/dispatch ordering

Proposed journal authoritative только после qualification store, durability, rollback и access
controls. Logical operation ID отличается от physical record ID.
Required fields: transaction/Target digest, scoped authorization/version, native actor/context,
claim/predecessor, epoch/incarnation, phase, exact request payload/digest, preconditions,
intent ID, dispatch/result state, provider request locator, observed cut, result digest,
uncertainty и independent readback locator. Secrets не попадают в journal.

| Journal state | Required durable meaning | Recovery action |
| --- | --- | --- |
| IntentRecorded | Exact operation предложена, dispatch decision не доказано | Revalidate current authority; intent не доказывает effect |
| DispatchReserved | Serialized authority/revoke/epoch/precondition decision | Independent worker не dispatch после authority change |
| MayHaveDispatched | Persist до external call, включая crash-before-call window | Считать potentially accepted; не auto-resend при restart |
| OutcomeVerified | Exact authoritative outcome связан с intent/request | Сохранить completed effect либо documented rejection |
| Finalized | Outcome и evidence durably recorded | Перейти только к proven next phase |
| RejectedBeforeDispatch / Uncertain | Proven rejection before send отличается от insufficient evidence | Первое можно re-plan под fresh authority; второе блокирует conflicting work |

**Действующее authority rule:** actual observed native user revoke invalidates current
authority и требует STOP по existing contracts. Ни DispatchReserved, ни MayHaveDispatched
не дают entitlement на отправку после такого revoke. Ни одно состояние не доказывает
external acceptance; reservation не является уже принятым GitHub request.
Draft не может переопределить или отложить действие current actual user revoke.

**Неутверждённая candidate policy — activation blocker:** future gateway может предложить
durable RevokeCommitted record в том же execution serialization domain как revoke
linearization point. Trusted receipt/message arrival, effective commit и acknowledgement
тогда различаются. Acknowledgement следует только после commit и independent readback.
Receipt-versus-commit authority policy здесь НЕ решена и НЕ утверждена; она требует
отдельного explicit user/normative approval до implementation или activation.

Только при этом conditional proposal RevokeCommitted до dispatch отклоняет queued/new
work; dispatch до RevokeCommitted может иметь accepted external request для reconciliation.
Gateway, paused до отправки при удержании execution lock, может оставить trusted revoke
receipt pending навсегда: такой receipt не committed revoke и не proof accepted request.
Его нельзя сообщать как completed revocation, external acceptance или terminal closure.
Reservation сама по себе не даёт post-effective-revoke send permission.
Нет atomic transaction между gateway revocation и external API acceptance.

Если policy требует immediate effect при trusted receipt, candidate B не доказывает
это свойство без более сильного enforced external fence; admission blocked, C действует.
Execution lock должен охватывать dispatch gap: release после token check до call
недостаточен. Удержание до завершения call может навсегда отложить proposed commit/acknowledgement;
это явная liveness cost, не исключение из действующего authority rule.
Любая uncertain reserved operation требует drain/reconciliation; нельзя заявлять
absence или автоматически отправлять повторно.

Crash после MayHaveDispatched, но до отправки observationally ambiguous с
accepted-call/response-loss. Journal ordering сознательно исключает automatic retry.
Transport cancellation, timeout и credential revoke не доказывают, что request
не был accepted. Proven terminal/ref/PR state связывается с intent; отсутствие записей
слабее authoritative no-effects evidence.

## 8. Phase-specific obligations P0–P10

Gateway — proposed executor boundary, не разрешение пропускать действующие Publisher gates.
Каждая фаза сохраняет current authority/context checks и exact operation journal linkage.
Исследуется только merge commit; Draft не разрешает squash/rebase/force.

| Phase | Fresh prerequisites / executor | Proven completion и recovery boundary |
| --- | --- | --- |
| P0 | Native owner и proposed admission bridge, trusted identity/route, API+Git probes, exact Target/branch/base/range, permissions/effective rules | Full current P0 proof, не только collector capability; context change останавливает |
| P1 | Gateway exact authorized head/ref, нет conflicting remote effect | Remote OID равен Target; discover existing completion до push |
| P2 | Gateway exact head/base/repo PR discovery | Один matching PR ID/URL; ambiguous create сначала reconciles |
| P3 | Gateway/observer fresh checks и conflict inspection | Required Success или No CI отдельно от CI PASS; Pending/Failed блокирует; No CI требует MERGEABLE / CLEAN |
| P4 | Gateway fresh head/base/checks, MERGEABLE / CLEAN и effective merge-commit policy | MERGED и exact merge SHA/parents; accepted/unknown reply inspect, без blind remerge |
| P5 | Gateway exact remote task ref ещё authorized head; отдельные delete/fetch-prune intents | Deletion/absence и refreshed refs доказаны раздельно; moved/recreated ref не удаляется |
| P6 | Gateway exclusive local workspace, clean phase facts | Current main и exact index/worktree; partial checkout reconciles |
| P7 | Gateway ff-only и ancestry proof | Synchronized main/origin/main; partial transport/local update inspect раздельно |
| P8 | Gateway merged exact local task branch, только safe -d | Exact local ref absence; без glob/-D/cleanup других tasks |
| P9 | Gateway/independent observer fresh main/ref/worktree facts | main == origin/main == authoritative remote main, task refs absent, clean index/worktree |
| P10 | Proven current native owner и все terminal prerequisites; future approved self-attributed native emission | Persisted terminal и independent exact readback; gateway journal сам по себе не P10 |

P5–P9 failure не откатывает P4. Local mutation и GitHub acceptance имеют разные durability
domains. Fresh read может квалифицировать outcome, но не восстановить потерянного native owner.
Terminal execution заблокировано до explicit approval, implementation и qualification
будущего gateway/native ownership bridge.

## 9. Crash consistency, rollback и HA limitations

| Crash point | Безопасное candidate continuation |
| --- | --- |
| До durable authorization/claim | Нет нового owner; inspect exact authoritative history |
| После reservation до fence | Non-executing reservation; effect permission отсутствует |
| Во время partial fence | STOP; доказать fencing, не предполагать completion |
| После fence до reconciliation | Сохранить fence; reconcile known claim, не вернуть old owner |
| После intent/reservation до call | Различить proven not dispatched и MayHaveDispatched; uncertain блокирует |
| После external acceptance до reply | Discover exact request/effect; не automatic resend |
| После reply до journal completion | Reconstruct exact outcome и связать с intent |
| После admission до acknowledgement | Читать exact winning admission; не второй owner |
| После terminal append до receipt | Independent exact persisted readback; не synthetic duplicate |
| Restart из старого backup | STOP до proof non-rollback authority/fence и current external outcomes |

Epoch counter внутри того же store откатывается вместе с backup. Hash chain выявляет
часть modifications, но не доказывает freshness и не предотвращает replay полного old snapshot.
Candidate restart поэтому требует independently trusted non-rollback high-water mark,
unique incarnation и external fence evidence либо отдельно доказанный equivalent.
Такой provider mechanism здесь не квалифицирован: UNKNOWN блокирует reactivation.

Single-active — enforcement property, не имя процесса. Stale gateway replica с
credentials или writable workspace является unfenced executor. Supervisor, child lifetime,
credential revocation и accepted-request drain требуют proof до restart.
HA, automatic leader replacement и partition tolerance требуют отдельного protocol;
DB lease/TTL недостаточно. Initial single-active design явно содержит SPOF.
Provider outage сохраняет uncertain outcomes и может блокировать recovery навсегда.
Local Git locks, store transactions и remote conditional refs имеют разные domains;
ни один не делает весь journal/Git/GitHub/native pipeline атомарным.

## 10. Покрытие formal invariants

Это design proof obligations, не runtime test results.

| ID | Invariant DP-025 | Candidate mechanism / proof и rejection |
| --- | --- | --- |
| I01 | Не более одного фактически mutation-eligible owner во всех epochs | Exclusive boundary + prior capability fence/drain + conditional admission; любой bypass/stale gateway отклоняет |
| I02 | Immutable Target и accepted scope | Digest в authority, claim, journal и per-phase predicate; mismatch отклоняет queued/executing intent |
| I03 | Self-attribution и сохранение history | Authenticated issuer/native bridge, immutable original locators; fabricated Release/continuity отклоняется |
| I04 | Reservation не permission | Из ClaimReserved нет dispatch до fence/reconciliation/admission/readback |
| I05 | Revoke блокирует new/queued execution; accepted requests reconciled | Current native revoke требует STOP; future receipt/commit policy не утверждена; serialized execution и exact in-flight proof, reservation не acceptance |
| I06 | Каждый effect связан с current authority, intent и outcome | Durable journal, authenticated execution linkage и independent authoritative result |
| I07 | Unknown не retry; completed effects не повторяются | MayHaveDispatched запрещает resend; operation-specific reconciliation |
| I08 | P10 требует current owner, clean synchronized main и terminal readback | P9, native bridge и independent persisted terminal receipt; missing proof отклоняет |

Formal/model safety properties должны охватывать stale agents, gateway incarnations,
queues, subprocesses и accepted requests, а не только ledger rows. Liveness claims требуют
explicit fairness, provider availability и outcome observability assumptions.
Manual trace coverage демонстрирует согласованность документа, не universal provider behavior.

## 11. Adversarial traces T01–T30

Каждая строка определяет future proof/rejection obligation; live test PASS не заявлен.

| ID | Scenario и state/predecessor | Invariants | Expected result / evidence |
| --- | --- | --- | --- |
| T01 | Competing ClaimReserved на same predecessor | I01,I04 | Один conditional winner; loser не dispatch; exact winning readback |
| T02 | Claim CAS unavailable / Authorized | I01,I04 | B admission недоступно; retain C, capability UNKNOWN |
| T03 | Claim succeeds, old credentials retained / fence | I01,I04 | Reject admission; complete revocation/bypass proof |
| T04 | Old run revives after claim / Executing | I01,I06 | Stale caller/capability отклонён на actual mutation boundary |
| T05 | Child/cached credential bypass / fence | I01 | Mediation proof failed; STOP до complete inventory/fence |
| T06 | Gateway replicas partition / admission | I01 | Stale replica fenced либо STOP; без lease-only HA |
| T07 | Crash before claim commit / Authorized | I04,I07 | Owner не выводится; exact store readback до reissue |
| T08 | Lost successful claim reply / ClaimReserved | I01,I07 | Прочитать unique winner/version, не duplicate claim |
| T09 | Crash after reservation before fence | I04 | Reservation остаётся non-executing |
| T10 | Crash during partial fence | I01,I04 | Unknown fence блокирует successor; persist covered/uncovered paths |
| T11 | Crash after fence before reconciliation | I01,I07 | Сохранить fence, resume exact inventory; не old-owner restoration |
| T12 | Lost admission reply / SuccessorAdmitted | I01,I06 | Independent winning admission readback определяет candidate |
| T13 | ABA/reused claim or predecessor | I01,I02 | Unique identity и monotonic non-rollback version отклоняют reuse |
| T14 | Merge accepted, reply lost / P4 uncertain | I06,I07 | Exact PR/request/merge SHA reconciliation, не blind merge |
| T15 | Queued intent survives revoke / dispatch | I05 | Current actual native revoke требует STOP; reject unaccepted work; reservation/MayHaveDispatched не acceptance или post-revoke send permission; future receipt/commit policy требует отдельного approval |
| T16 | Revoke races claim/admission | I01,I05 | Serialization order/authority version отклоняет post-revoke admission |
| T17 | Revoke races accepted API request | I05,I07 | Сохранить uncertain effect; не fabricated no-effects claim |
| T18 | Target changes after enqueue | I02,I05 | Reject stale intent/admission; не silent retarget |
| T19 | Delayed old push / fence/drain | I01,I06,I07 | Proven rejection либо exact outcome до remaining work |
| T20 | TASK-095 post-merge offline fixture | I02,I07,I08 | Сохранить P0–P4, не replay merge; P5–P10 blocked без proof |
| T21 | TASK-096 pre-merge offline fixture | I02,I07 | Сохранить P0–P3, нужны fresh P4 facts/admission; без live merge |
| T22 | Approved design without deployment | I01,I04 | Нет operational owner/activation; retain C |
| T23 | Legacy capabilities cannot be fenced | I01,I03 | Historical migration rejected; retain C |
| T24 | P10 without proven current owner | I08 | Reject terminalization даже после external merge |
| T25 | Terminal append reply lost | I03,I07,I08 | Независимо прочитать exact persisted item, не duplicate terminal |
| T26 | Только TTL/completed run как loss proof | I01,I04 | Fence/drain evidence недостаточно; STOP |
| T27 | Task ref recreated at different OID / P5 | I02,I07 | Не удалять new ref; exact conditional predicate/reconcile |
| T28 | Observer fabricates source Release | I03 | Reject provenance; без ownership transition |
| T29 | Journal committed, API outcome unknown | I06,I07 | Сохранить MayHaveDispatched/Uncertain; без automatic resend |
| T30 | Claim cancelled after old executor fenced | I01,I04 | Не возвращать old capability; нужен new authorized disposition |

Дополнительные future traces должны включать old-backup rollback, process pause внутри
dispatch gap и stale gateway restart; они дополняют, а не заменяют T01–T30.
Каждый future proof фиксирует exact predecessor, request/epoch, observation cut,
authoritative result и limitation.

## 12. Activation, fixtures, PoCs и следующие gates

Default activation prospective-only после отдельных architecture decision, normative
alignment, implementation, independent verification, deployment qualification и explicit
activation scope. Required impact inventory: PROCESS-001/002, AGENT, Coordinator/Publisher
contracts, general/Publisher recovery scenarios, task template и mirrored process guides.
Design task их не меняет. Existing native admission/terminal contracts действуют,
пока approved bridge не существует; gateway admission не может их заменить.

Offline fixtures используют [immutable TASK-096 evidence](https://github.com/dsdred/universal-websocket-platform/blob/82cf49627e54881599bb60a55f28c8e9c6b0c4db/docs/tasks/TASK-096-FAILED-OWNER-RECOVERY-ASSESSMENT.md)
и DP-025 E07–E10. TASK-095 PR #99 merge c8047fd08caa79480c2bc73b5c382f4aa5079db7
имеет historical P0–P4 proof и P5–P10 obligations. TASK-096 имеет historical P0–P3,
P4 blocker и remaining P4–P10; current PR/checks outcome здесь UNKNOWN.
Неисполнение конкретной команды не доказывает отсутствие всех последующих действий.
Live fixture read/recovery, reuse old permission и terminal inference отсутствуют.

Historical migration требует separate exact authorization, full original lineage/attempt/request
inventory, complete legacy credential/workspace/child fencing, authoritative in-flight outcomes,
approved native bridge и successor self-attribution. Gateway deployment не доказывает
legacy mediation задним числом. Missing source history/unfenceable capability сохраняет
transaction blocked. Нет synthetic Release, InitialDispatch, continuity или historical terminal rewrite.

| Proposed isolated PoC | Required evidence / STOP boundary |
| --- | --- |
| Conditional claim/ref contention | Concurrent winner, ABA, lost reply и exact version/readback; sample не universal guarantee |
| Credential/workspace isolation | Каждый перечисленный bypass запрещён, включая children/alternate workspace; unknown path блокирует |
| Crash-point injection | Каждая boundary раздела9, uncertain dispatch и journal restart; без automatic resend |
| Revival/partition | Paused old agent/gateway и children не могут mutate; без automatic HA assumption |
| Backup rollback | Independent freshness/incarnation/fence evidence сохраняется после old snapshot; UNKNOWN блокирует restart |
| Merge request UUID/status | Exact supported deployment, intent correlation, async outcome и lost response; cancellation не предполагается |
| Conditional delete | Exact recreated-ref race отклонён, documented predicate и lost-response reconciliation |
| Formal/model verification | I01–I08, T01–T30 и rollback/dispatch-gap traces при explicit assumptions |

PoCs требуют отдельного разрешения и disposable resources. Mutation probe,
implementation или provider setup здесь не выполняются. Read-only research может сузить
capability gap, но не заменить proof exclusive enforcement.

Design acceptance требует Architect Confirmation, independent Tester, non-author Review,
EN/RU parity, valid links, PROCESS-002, eight-path Scope Audit, canonical accepted manifest
и post-decision integrity. Manual design verification не runtime/provider certification.
B остаётся conditional; A UNKNOWN; C действует.
Next candidate — отдельно согласованный evidence/architecture gate, Not Activated, без
automatic Task ID. STOP при unknown enforcement, stale predecessor, competing claim,
Target/provenance mismatch, uncertain outcome, bypass, rollback ambiguity или требовании
unapproved normative/runtime work.
