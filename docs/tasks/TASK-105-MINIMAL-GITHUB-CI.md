# TASK-105 — Minimal GitHub CI

## Status

In Progress. Verification-stable projection; actual checkpoints and decision
resolve only from newest matching Recovery Evidence Envelope. No commit or
publication authorized.

## Task Contract

- Task Mode: Implementation (engineering CI configuration).
- Why Now: explicit Owner authorization 2026-10-21; repository has no workflow.
  One independently reviewable CI slice; Idle/TLS/native work is not selected.
- Repository: universal-websocket-platform; branch
  `codex/task-105-minimal-github-ci`; trusted baseline/current HEAD
  `fb916df77e697fb7a45dc7bb6e8d0150411f2d82`, initially clean main.
  Branch prefix follows desktop instruction; no main/history modification.
- Definition of Done: Architecture Confirmation before workflow mutation;
  exactly one `.github/workflows/go-ci.yml`, PR/main and push/main, ubuntu-24.04
  x64, Go 1.26.5, job go-ci. Check tracked Go formatting, readonly build/test/vet,
  CGO race, tidy diff and unchanged module files. Least privileges, no persisted
  checkout credentials, official actions at full SHA, bounded timeout, no path
  filters or suppressed failures. Offline and negative verification, independent
  Tester/non-author Reviewer, PROCESS-002, exact manifest and integrity check.
- Out of Scope: module changes, new dependencies, downloads/installations, host
  installation, GitHub settings/required checks, hosted execution, commit/stage,
  publication, Approved/Frozen changes, product behavior changes, Architecture B,
  PoC, unpublished TASK-102. Only the two named orchestration_admission Go files
  may receive CRLF-to-LF formatting corrections, with exact semantics retained.
- Verification Plan: existing Go 1.26.5 windows/amd64; GOTOOLCHAIN=local,
  GOPROXY=off, GOSUMDB=off. Offline YAML structure and script validation using
  already installed tools, negative fault checks in disposable copies, full
  readonly build/test/vet and tidy diff. Race only with existing usable CGO
  tooling; unavailable checks explicitly limited. No acquisition workaround.
- Stop conditions: Architecture FAIL, contract/product changes, missing dependency
  or required download/install, blocking findings, scope expansion, commit-ready.
  Hosted GitHub execution requires a separate Publication Gate.
- Next candidate: separate Publication Gate after independently authorized commit;
  Not Activated. Product/native work remains Not Activated.

## Sources and documentation baseline

AGENTS/AGENT, PROCESS-001/002, role contracts, PROJECT_CONTEXT, spec README,
current-state/decisions, MASTER_PLAN EN/RU, Runtime architecture sources and Go
module/source/tests. Repository precedence preserved. Baseline TASK-104 record
contains accepted manifest and integrity evidence; local Git shows source
`63d9990` merged by PR #104 at trusted baseline. Historical task records stay
immutable. TASK-102 stays committed/unpublished on its separate branch.

## Roles and ordered stages

Coordinator /root; Architect /root/architect105 (read-only); Developer then
Documentation Agent /root/implementation105 as nonoverlapping assigned roles;
independent Tester /root/tester105; non-author Reviewer /root/reviewer105.
Publisher N/A.
Intake -> Documentation Baseline -> Architecture Confirmation -> implementation
-> offline/negative verification -> independent Tester -> PROCESS-002 -> Scope
Audit/Size Guard -> non-author final Review -> Coordinator decision -> independent
post-decision integrity -> STOP AT COMMIT GATE.

## Existing Coverage Report

- Existing Coverage: existing Runtime Go suites and full module checks; no CI
  workflow or workflow verification suite exists. Go module baseline is 1.25.0;
  selected CI execution toolchain is 1.26.5, without module directive changes.
- Coverage Gap: workflow trigger/security/command contract and negative rejection
  checks; Linux hosted CGO execution not demonstrated locally.
- Added Proof/Regression Tests: planned offline structural/negative checks in
  disposable local artifacts; no new product tests needed for CI-only scope.
- Remaining Limitations: Windows host CGO_ENABLED=0, gcc/clang not found;
  actionlint not installed. Module files have checkout CRLF but baseline Git LF;
  tidy must distinguish checkout-only newline drift without changing files.

## Documentation applicability

Task record/index, current-state and PROJECT_CONTEXT: required and synchronized
for current-task and CI-source status; TASK-104 durable PR #104 merge reflected,
historical records and TASK-102 preserved. MASTER_PLAN EN/RU: Not applicable,
no milestone boundary, engineering dependency or roadmap status changes, as
confirmed by Architect. Design Proposals/ADR/ARCH/spec decisions: Not applicable,
no design/implementation status or normative contract change. CHANGELOG:
Not applicable, internal tooling without user-facing/release changes. General
README/spec/architecture indexes and EN/RU guides: Not applicable, no new
public document, architecture or navigation target; internal task record has
no mandatory EN mirror. No critical drift identified.

## Scope Audit preparation

Scope Audit candidate classification: Required 7 / Questionable 0 / Removable 0.
- `.github/workflows/go-ci.yml`: Owner CI contract.
- `internal/runtimecommandidempotency/orchestration_admission.go`: authorized
  CRLF-to-LF correction only.
- `internal/runtimecommandidempotency/orchestration_admission_test.go`: authorized
  CRLF-to-LF correction only.
- `docs/tasks/TASK-105-MINIMAL-GITHUB-CI.md`: PROCESS-001 contract/evidence.
- `docs/tasks/README.md`: PROCESS-002 current-task navigation.
- `spec/current-state.md`: PROCESS-002 factual CI-source/task state.
- `.ai/PROJECT_CONTEXT.md`: PROCESS-002 current-task/fundamental state.
Size Guard candidate: seven paths, no new package or architecture contract,
one CI behavior, zero semantic production delta; apparent large Go line diff
is solely authorized CRLF-to-LF correction, proven by baseline-byte equality.
Independent offline proof artifacts remain disposable/outside candidate.
Final audit confirmation, verification/review/decision and exact manifest
remain Coordinator gates; no Acceptance or completed check is claimed here.

## Recovery Evidence Envelope

### E-105-001 — Intake

Clean main at exact Owner baseline proven before branch creation. Initial switch
failed with `.git/index.lock` permission denied, with no lock or branch created.
This record was created as the first content change; immediate read-only
reconstruction confirmed only this untracked path, unchanged main/HEAD and no
index.lock. Supported escalated retry then created the authorized task branch
without stage/history mutation. No other task changes preceded the branch.
Architecture Confirmation not yet completed; workflow absent. No final verdict
or Acceptance claimed. First incomplete stage: Architecture Confirmation.

### E-105-002 — Architecture Confirmation, before workflow mutation

Architect /root/architect105: PASS / APPROVED — NO NEW PREREQUISITE; blocking
findings 0. Read-only baseline inspection at `fb916df` confirmed automation-only
scope; ADR-0003, Frozen ARCH-002 and Active ARCH-004 unchanged. Go 1.26.5 is the
CI toolchain, not a Go module migration; retain go1.25.0 directive and ADR-0001.
All Owner workflow/security requirements accepted; formatting must equal baseline
bytes with only CRLF replaced by LF. Linux does not prove Windows-only containment
adapter correctness. No critical architecture/documentation drift. MASTER_PLAN,
ADR/ARCH/DP and CHANGELOG amendments N/A for this tooling slice. TASK-104 current
pointer may be synchronized to its durable merge fact without historical edits.
Pre-implementation contract update N/A: no contract changes. Implementation is
now admitted within the Task Contract; no hosted execution authorized.

### E-105-003 — Developer and Documentation handoff

Developer /root/implementation105 implemented one workflow after durable
Architecture PASS E-105-002. Job go-ci uses ubuntu-24.04, timeout 15 minutes,
Go '1.26.5', GOTOOLCHAIN=local, explicit bash, readonly build/test/vet/race,
CGO_ENABLED=1 for race, tidy -diff and final module diff guard. Formatting uses
a NUL-delimited tracked-file list in a temporary file; git list and each gofmt
failure propagate, with no process-substitution failure hiding. PR/main and
push/main have no path filters; contents:read, persist-credentials:false,
cache:false, full official action SHAs, no suppressed failures.
Pin provenance (official release/commit sources read-only verified by Coordinator):
- https://github.com/actions/checkout/releases/tag/v4.2.2
- https://github.com/actions/checkout/commit/11bd71901bbe5b1630ceea73d27597364c9af683
- https://github.com/actions/setup-go/releases/tag/v5.0.2
- https://github.com/actions/setup-go/commit/0a12ed9d6a96ab950c8f026ed9f722fe0da7ef32
- Runner x64 label: https://docs.github.com/en/actions/reference/runners/github-hosted-runners
No local download/installation occurred; future hosted action/tool acquisition
belongs only to the separately authorized Publication Gate.

Exact raw implementation OIDs (git hash-object --no-filters):
- .github/workflows/go-ci.yml: 2bde3fdea3ffed66f2f0abea17c4de870c045fef
- internal/runtimecommandidempotency/orchestration_admission.go:
  11fab2353ee24adb5aaa3cae2194ccaca2cbd9bf
- internal/runtimecommandidempotency/orchestration_admission_test.go:
  29adb32f6745b70b55a78b2bcec603c2a10bf69a
For each Go file, raw git-show baseline bytes with only CRLF replaced by LF
exactly equal current bytes (471/786 CRLF pairs respectively; now zero).
No semantic or module change. Reproduction: Python subprocess.check_output
(['git','show',baseline+':'+path]).replace(b'\r\n',b'\n') == Path(path).read_bytes().

Developer role ended before Documentation role began. Documentation Agent
synchronized only current pointers and CI-source facts in task index,
current-state and PROJECT_CONTEXT; bytes from historical TASK-098 onward
retained exactly. TASK-104 source/merge facts verified from local immutable Git;
TASK-102 untouched. Mandatory applicability disposition is in projected body.
Documentation synchronization candidate ready for independent verification;
no Tester/Reviewer/Coordinator verdict claimed. First incomplete gate:
offline/negative verification and independent Tester.

### E-105-004 — Frozen candidate and Coordinator Scope Audit

Anchor HEAD/base `fb916df77e697fb7a45dc7bb6e8d0150411f2d82`, branch
`codex/task-105-minimal-github-ci`; Git object format sha1. Exact candidate
manifest `c8e533bfb8d82cdbdb209b20ff0c2f9dd2dcecd6`; task-record-v1 projection
`5493716e1d49be75be275f22e5958e8906e81167`. Ordered rows below use
path | projection | state | intended mode | projected Git blob OID.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | fa88e31b85fda50b0f3ffb7103704cbf17520a1d
.github/workflows/go-ci.yml | full | present | 100644 | 2bde3fdea3ffed66f2f0abea17c4de870c045fef
docs/tasks/README.md | full | present | 100644 | 37e1c344876046aa283a620a0cb627ba18fb9c0a
docs/tasks/TASK-105-MINIMAL-GITHUB-CI.md | task-record-v1 | present | 100644 | 5493716e1d49be75be275f22e5958e8906e81167
internal/runtimecommandidempotency/orchestration_admission.go | full | present | 100644 | 11fab2353ee24adb5aaa3cae2194ccaca2cbd9bf
internal/runtimecommandidempotency/orchestration_admission_test.go | full | present | 100644 | 29adb32f6745b70b55a78b2bcec603c2a10bf69a
spec/current-state.md | full | present | 100644 | 895d05943918b412d3db91f2d0c5eddeec2291ef
```

Reproduction follows PROCESS-001 Persistent Recovery Anchor byte-exact algorithm:
read raw files, apply task-record-v1 only to this task (unique ordered Status,
Task Contract, terminal Recovery Evidence Envelope headings); sort paths by
unsigned UTF-8 bytes; serialize each row as path NUL projection NUL state NUL
mode NUL OID NUL. Hash projected files and manifest with git hash-object --stdin,
no filters or normalization. Envelope is excluded and cannot attest its own bytes.
Intended mode 100644 matches existing paths/new regular files; no deletions.
Future authorized staging must preserve exact raw bytes rather than normalize
mixed/CRLF files; no stage or intent-to-add is authorized here.

Coordinator Scope Audit: Required 7, Questionable 0, Removable 0. Workflow is the
single requested CI behavior; two exact formatting repairs make tracked gofmt
pass; task record provides mandatory recovery/evidence; three navigation sources
provide mandatory PROCESS-002 current-task/CI facts. No path can be removed while
preserving the corresponding DoD/process obligation. Size Guard ACCEPT: seven
paths, one CI behavior, zero new packages/contracts/dependencies, zero semantic
production delta; large line counts are solely the two authorized newline repairs.
Documentation applicability/synchronization is explicit in projected body; no
EN/RU public documentation changes or semantic parity delta. Historical records,
Approved/Frozen documents and TASK-102 were not modified.

Root checks: git -c core.whitespace=cr-at-eol diff --check exit 0; module guard
exit 0; index empty. CR-at-EOL accommodation only recognizes preserved baseline
mixed line endings and does not change Git config or suppress content failures.
Raw module observations: go.mod aa82a60c0769256f13c86b31382a8994cf6e73e4;
go.sum f1912accf367972c7c6278b7a8d7de478028139b. main/cached origin/main
remain baseline; TASK-102 branch remains 8bfa2e22a4b6310a71f6d36c8724b1adedd727ad.
Next unfinished checkpoints: independent Tester then final non-author Review.


### E-105-005 — Independent Tester handoff

Tester /root/tester105, independent of implementation: PASS WITH LIMITATION;
blocking findings 0, nonblocking findings 0. Subject anchor HEAD/base
fb916df77e697fb7a45dc7bb6e8d0150411f2d82; branch
codex/task-105-minimal-github-ci; canonical seven-row manifest
c8e533bfb8d82cdbdb209b20ff0c2f9dd2dcecd6 (rows E-105-004).
Own independent hashlib Git-blob implementation below reproduces exact manifest,
raw module invariance and empty index; not merely reuse of Coordinator assertion.
Existing Coverage Report above preceded disposable proof additions; no product
or permanent test files added. Applicable coverage: 266 tracked Go files,
full existing module tests, seven Bash run scripts, exact workflow/security
structure, ten injected rejection scenarios and module/format semantics proofs.

Commands used existing Go1.26.5 windows/amd64, PyYAML6.0.3, Git2.50.1 and
Git Bash5.2.37 at C:/tools/Git/bin/bash.exe. No acquisition or config mutation.
Every Go invocation used GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off.
Native repository commands/results:
- go build -mod=readonly ./...: exit0.
- go test -mod=readonly -count=1 ./...: initial sandbox exit1 (localhost
  connectex access denied across fixtures); supported escalated exact rerun
  completed exit0, all packages PASS. This was environment recovery, no rework.
- go vet -mod=readonly ./...: exit0.
- go mod tidy -diff: native exit1, solely go.sum checkout CRLF-to-LF diff;
  no module bytes modified. Same command in disposable LF Git-object source
  with exact candidate Go overlays: exit0, no output.
- git diff --exit-code -- go.mod go.sum: exit0; raw module OIDs E-105-004
  independently unchanged. git -c core.whitespace=cr-at-eol diff --check: exit0.
- CGO_ENABLED=1 go test -mod=readonly -race -count=1 ./...: exit1,
  runtime/cgo reports C compiler gcc not found in PATH; no race tests executed.
  No compiler download/install attempted.
- GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=readonly ./... in
  LF disposable source: exit0. This is cross-build proof, not Linux runtime/race.

Offline proof below executed as python -u <saved-script.py> using supported
escalated exec only because Git Bash otherwise failed NtCreateDirectoryObject;
its final complete run exit0. PyYAML BaseLoader retains literal on and scalar
strings, avoiding YAML1.1 boolean-key coercion. Exact key/value checks verify
main-only triggers/no path filters, one job/runner/version/timeout, contents:read,
persist-credentials:false/cache:false, full official action SHAs and all commands;
step allowlist excludes suppressed failure/conditional overrides. It is a targeted
contract validator, not actionlint or the GitHub Actions service validator.
All seven extracted scripts pass actual bash -n.
Ten negative scenarios PASS: malformed tracked filename containing spaces exits1;
injected gofmt failure47 and git enumeration failure48 propagate; each of five
Go commands injected failure49 stops before marker; go.mod and go.sum drift
individually reject exit1. Positive fixture and clean module guard exit0.
The two authorized files independently equal baseline raw bytes with only CRLF
replaced by LF (no semantics delta).

Clarification of E-105-004 formatting statement: the exact formatting script
passes all 266 files in the LF Git-object source with the two candidate overlays.
The raw Windows checkout rejects cmd/control-service/main.go because other
baseline files are checked out CRLF; no authorization to normalize those files
was inferred. Initial disposable archive inherited global core.autocrlf and was
correctly rejected; final successful reproduction explicitly uses per-command
-c core.autocrlf=false for archive and temporary git add. No Git config changed.

Remaining limitations: no hosted execution authorized/performed, no Linux runtime
or CGO/race proof; native CGO compiler and actionlint absent. No offline check
claims action download/setup execution, hosted tool availability, Windows-native
LPAC compilation or production activation. Publication Gate must inspect actual
hosted results separately. Parser/script proofs and native regression/cross-build
are available substitutes; unavailable checks are not silently PASS.

Reproducible offline proof source follows. Save the exact code as UTF-8 with LF,
no added terminal newline (the stored script has none). Git blob OID: 7848a3cf28989faf976b52e59683638e6c731ae6.

```python
import os,pathlib,subprocess,tempfile,yaml,re,zipfile,io
R=pathlib.Path('E:/wikiPRJ/universal-websocket-platform'); B='C:/tools/Git/bin/bash.exe'
def run(a,c=R): return subprocess.run(a,cwd=c,capture_output=True,text=True)
def bash(s,c=R): return run([B,'--noprofile','--norc','-e','-o','pipefail','-c',s],c)
w=yaml.load((R/'.github/workflows/go-ci.yml').read_bytes(),Loader=yaml.BaseLoader)
assert set(w)=={'name','on','permissions','jobs'} and w['on']=={'pull_request':{'branches':['main']},'push':{'branches':['main']}}
assert w['permissions']=={'contents':'read'} and set(w['jobs'])=={'go-ci'}
j=w['jobs']['go-ci']; assert set(j)=={'runs-on','timeout-minutes','env','defaults','steps'}
assert j['runs-on']=='ubuntu-24.04' and j['timeout-minutes']=='15' and j['env']=={'GOTOOLCHAIN':'local'} and j['defaults']=={'run':{'shell':'bash'}}
s=j['steps']; assert len(s)==9
assert s[0]['uses']=='actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683' and s[0]['with']=={'persist-credentials':'false'}
assert s[1]['uses']=='actions/setup-go@0a12ed9d6a96ab950c8f026ed9f722fe0da7ef32' and s[1]['with']=={'go-version':'1.26.5','cache':'false'}
cmds=['go build -mod=readonly ./...','go test -mod=readonly -count=1 ./...','go vet -mod=readonly ./...','go test -mod=readonly -race -count=1 ./...','go mod tidy -diff','git diff --exit-code -- go.mod go.sum']
assert [a['run'] for a in s[3:]]==cmds and s[6]['env']=={'CGO_ENABLED':'1'}
assert all(set(a).issubset({'name','uses','with','run','env'}) for a in s)
assert len(list((R/'.github/workflows').iterdir()))==1
print('YAML structural/security PASS')
f=s[2]['run']; r=bash(f); assert r.returncode!=0 and 'cmd/control-service/main.go' in r.stderr; print('native formatter CRLF limitation',r.returncode)
print('native checkout limitation recorded')
for name in ['orchestration_admission.go','orchestration_admission_test.go']:
 p='internal/runtimecommandidempotency/'+name; old=subprocess.check_output(['git','show','HEAD:'+p],cwd=R); assert (R/p).read_bytes()==old.replace(b'\r\n',b'\n')
print('two files exact CRLF-to-LF-only PASS')
with tempfile.TemporaryDirectory(prefix='task105-') as td:
 d=pathlib.Path(td); assert run(['git','init','-q'],d).returncode==0
 (d/'ok.go').write_text('package proof\n',newline='\n'); assert run(['git','add','ok.go'],d).returncode==0; assert bash(f,d).returncode==0
 (d/'bad file.go').write_text('package proof\nfunc bad( ){ }\n',newline='\n'); assert run(['git','add','bad file.go'],d).returncode==0
 r=bash(f,d); assert r.returncode!=0 and 'bad file.go' in r.stderr; print('negative bad tracked spaced name PASS',r.returncode)
 (d/'bad file.go').write_text('package proof\n',newline='\n')
 for prefix,label in [('gofmt() { return 47; }; export -f gofmt;\n','gofmt failure'),('git() { return 48; }; export -f git;\n','git failure')]:
  r=bash(prefix+f,d); assert r.returncode!=0; print('negative',label,'PASS',r.returncode)
 for c in cmds[:-1]:
  r=bash('go() { return 49; }; export -f go;\n'+c+'\necho SHOULD_NOT_RUN',d); assert r.returncode!=0 and 'SHOULD_NOT_RUN' not in r.stdout
 print('negative five exact Go commands fail closed PASS')
 (d/'go.mod').write_text('module proof\n\ngo 1.25.0\n',newline='\n'); (d/'go.sum').write_text('',newline='\n'); assert run(['git','add','go.mod','go.sum'],d).returncode==0; assert bash(cmds[-1],d).returncode==0
 for name in ['go.mod','go.sum']:
  p=d/name; before=p.read_bytes(); p.write_bytes(before+b'drift\n'); assert bash(cmds[-1],d).returncode!=0; p.write_bytes(before)
 print('negative module drift both files PASS')
 for a in s:
  if 'run' in a: assert bash("bash -n <<'PROOF'\n"+a['run']+'\nPROOF\n',d).returncode==0
 print('all seven Bash syntax checks PASS')
with tempfile.TemporaryDirectory(prefix='task105-lf-') as td:
 d=pathlib.Path(td); zipfile.ZipFile(io.BytesIO(subprocess.check_output(['git','-c','core.autocrlf=false','archive','--format=zip','HEAD'],cwd=R))).extractall(d)
 for p in ['internal/runtimecommandidempotency/orchestration_admission.go','internal/runtimecommandidempotency/orchestration_admission_test.go']:
  (d/p).write_bytes((R/p).read_bytes())
 assert run(['git','init','-q'],d).returncode==0
 assert run(['git','-c','core.autocrlf=false','add','.'],d).returncode==0
 r=bash(f,d); assert r.returncode==0,(r.stdout,r.stderr); print('LF Git-object source exact formatter PASS')
 env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off')
 r=subprocess.run(['go','mod','tidy','-diff'],cwd=d,env=env,capture_output=True,text=True); print('LF tidy exit',r.returncode,r.stdout,r.stderr); assert r.returncode==0
 env.update(GOOS='linux',GOARCH='amd64',CGO_ENABLED='0')
 r=subprocess.run(['go','build','-mod=readonly','./...'],cwd=d,env=env,capture_output=True,text=True); print('Linux cross-build exit',r.returncode,r.stdout,r.stderr); assert r.returncode==0
```

Independent canonical-manifest/raw-module/index proof, same save convention.
Git blob OID: 40c255f85638e6b6e553ad2fa45c75f6bf9f8f64. Execute python <saved-proof.py> from existing tools;
observed exit0 with canonical c8e533bfb8d82cdbdb209b20ff0c2f9dd2dcecd6.

```python
import pathlib,hashlib,re,subprocess
r=pathlib.Path('E:/wikiPRJ/universal-websocket-platform')
paths=['.ai/PROJECT_CONTEXT.md','.github/workflows/go-ci.yml','docs/tasks/README.md','docs/tasks/TASK-105-MINIMAL-GITHUB-CI.md','internal/runtimecommandidempotency/orchestration_admission.go','internal/runtimecommandidempotency/orchestration_admission_test.go','spec/current-state.md']
def blob(b):return hashlib.sha1(b'blob '+str(len(b)).encode()+b'\0'+b).hexdigest()
rows=[]
for p in sorted(paths,key=lambda s:s.encode()):
 b=(r/p).read_bytes(); proj='full'
 if 'TASK-105-' in p:
  headings=[list(re.finditer(b'^'+h+b'\r?\n',b,re.M)) for h in [b'## Status',b'## Task Contract',b'## Recovery Evidence Envelope']]; assert all(len(x)==1 for x in headings)
  a,c,e=[x[0] for x in headings]; assert a.end()<c.start()<e.start() and not re.search(b'^## ',b[e.end():],re.M)
  b=b[:a.end()]+b'STATUS-EVIDENCE-EXCLUDED\0'+b[c.start():e.start()];proj='task-record-v1'
 rows.append([p,proj,'present','100644',blob(b)])
m=b''.join(('\0'.join(row)+'\0').encode() for row in rows)
assert blob(m)=='c8e533bfb8d82cdbdb209b20ff0c2f9dd2dcecd6'
for p,h in [('go.mod','aa82a60c0769256f13c86b31382a8994cf6e73e4'),('go.sum','f1912accf367972c7c6278b7a8d7de478028139b')]:assert blob((r/p).read_bytes())==h
assert subprocess.check_output(['git','diff','--cached','--name-only'],cwd=r)==b''
print('independent hashlib canonical seven rows/module invariance/index PASS',blob(m))
```

No subject mutation, stage/commit or publication performed by Tester. First
unfinished checkpoint: final non-author Review then Coordinator decision/integrity.


### E-105-006 — Independent non-author final Reviewer

Reviewer /root/reviewer105, not implementation/documentation author: Approved;
blocking findings 0, nonblocking findings 0. Reviewed exact seven paths/ordered
rows E-105-004 at branch codex/task-105-minimal-github-ci, HEAD/base
fb916df77e697fb7a45dc7bb6e8d0150411f2d82, object format sha1. Independent
raw-byte Python implementation of PROCESS-001 reproduced task-record-v1
5493716e1d49be75be275f22e5958e8906e81167 and canonical manifest
c8e533bfb8d82cdbdb209b20ff0c2f9dd2dcecd6 before this append. No subject change
or staging is authorized by this verdict; future Commit Gate must preserve raw
bytes, without LF/CRLF equivalence or Git-filter normalization.

Architecture PASS E-105-002 preceded workflow implementation. Reviewed one
main-only PR/push workflow, ubuntu-24.04 x64, Go 1.26.5, go-ci/15-minute job,
all seven required checks, NUL-safe tracked-file formatting, explicit Bash
fail-closed commands, contents:read, credential persistence disabled and full
SHA official actions. Release links E-105-003 independently resolved to the
exact checkout and setup-go commit URLs. No suppressed failures/path filters,
module/dependency/API/lifecycle/contract changes, settings mutation, TASK-102
change, installation, commit or publication found. Raw Go files independently
equal baseline with only 471/786 CRLF pairs replaced by LF.

Durable independent Tester E-105-005 reviewed in full: exact manifest,
commands/results, scope/coverage counts, executable proof source and explicit
limitations. Extracting its two Python code blocks with documented LF/no-final-
newline save convention independently reproduces proof-source Git blob OIDs
7848a3cf28989faf976b52e59683638e6c731ae6 and
40c255f85638e6b6e553ad2fa45c75f6bf9f8f64. Native build/test/vet pass; local
socket-restriction retry is recorded. Ten negative rejection scenarios and seven
Bash syntax checks support failure propagation. No actual Linux runtime or
CGO/race execution PASS is claimed: race probe fails for absent gcc, actionlint
is absent, hosted Actions have not run. Native formatter/tidy rejection from
checkout CRLF is explicitly separated from LF Git-object fixture formatter
266-file/tidy PASS and Linux amd64 CGO-disabled cross-build PASS. E-105-004
formatting statement is interpreted only with E-105-005 clarification.

Reviewer Scope Audit independently applied the mandatory removal question,
'Можно ли удалить изменение и сохранить DoD?', to every path:
- workflow: no, the requested CI deliverable would disappear;
- each of the two Go files: no, committed CRLF would fail LF-source gofmt;
- task record: no, contract/recovery/exact-evidence obligations would disappear;
- task index: no, current task navigation would remain stale;
- current-state: no, factual CI-source/current task status would remain stale;
- PROJECT_CONTEXT: no, current task/fundamental state would remain stale.
Required 7 / Questionable 0 / Removable 0; Size Guard ACCEPT, one behavior,
zero semantic production delta/new packages/contracts. Documentation
synchronization/applicability is adequate; historical facts remain distinct
from fresh authority, Approved/Frozen sources unchanged, public EN/RU parity
has no delta. No needless documentation scope identified. Independent
cr-at-eol diff check exit0; status/contract remains verification-stable
In Progress with envelope-resolution rule, no premature closure claim.

First incomplete checkpoint: Coordinator decision followed by separate
post-decision integrity, then STOP AT COMMIT GATE. This verdict supplies no
commit/publication permission and does not attest its own final envelope bytes.

### E-105-007 — Coordinator Acceptance R-105-001

Coordinator /root: ACCEPT TASK-105 configuration candidate, decision R-105-001,
2026-10-21. Exact branch codex/task-105-minimal-github-ci; baseline/current HEAD
fb916df77e697fb7a45dc7bb6e8d0150411f2d82; canonical seven-path manifest
c8e533bfb8d82cdbdb209b20ff0c2f9dd2dcecd6; task-record-v1 projection
5493716e1d49be75be275f22e5958e8906e81167; exact ordered rows E-105-004.

Architecture Confirmation PASS preceded implementation. Required workflow,
commands/security and two exact newline repairs are implemented within Owner
scope. Independent Tester E-105-005 PASS WITH LIMITATION, findings 0/0;
non-author Reviewer E-105-006 Approved, findings 0/0. PROCESS-002 Synchronized;
Scope Audit Required 7 / Questionable 0 / Removable 0; Size Guard ACCEPT.
Final root readback reproduced the exact manifest/rows and seven-path set,
unchanged module raw OIDs, empty index, no raw trailing whitespace/conflicts;
three new current-task links resolve. No subsequent subject mutation identified.

Accepted verification scope: native offline Go1.26.5 build/test/vet, targeted YAML
contract validation, seven actual Bash syntax proofs, ten fail-closed negative
scenarios, LF Git-object source formatting of 266 tracked Go files and tidy,
Linux amd64 CGO0 cross-build. Explicit accepted limitations: hosted execution
not run/authorized; no Linux runtime/CGO/race PASS; local race probe exits1 for
missing gcc, no race tests run; actionlint absent. Native formatter/tidy CRLF
rejections are not PASS and do not justify additional repository normalization;
LF source fixture PASS is separate evidence. CI source acceptance does not
claim Windows LPAC/native toolchain qualification or product activation.

Latest disposition: Coordinator Accepted with recorded environment limitations;
exact candidate ready for separately authorized Commit Gate after independent
post-decision integrity. Verification-stable projected In Progress resolves by
this matching envelope, without changing historical records. Next recommended
step is Owner Commit Gate for this exact candidate, then a separate Publication
Gate including actual hosted results; neither activated here. TASK-102 remains
unchanged and unpublished, Architecture B/PoC inactive, Approved/Frozen unchanged.
No stage, commit, push, settings/required-check changes or publication performed.

First incomplete checkpoint: independent post-decision integrity, then STOP AT
COMMIT GATE. This decision does not attest its own final envelope bytes and does
not grant commit or hosted execution/publication permission.

### E-105-008 - Independent post-decision integrity

Reviewer /root/reviewer105: PASS, findings 0/0. Independently checked after
Coordinator decision R-105-001 / E-105-007: exact seven-path set, all ordered
E-105-004 path/projection/present/100644/blob rows, raw Git-blob identity and
PROCESS-001 task-record-v1 projection. Current manifest remains
c8e533bfb8d82cdbdb209b20ff0c2f9dd2dcecd6; task projection remains
5493716e1d49be75be275f22e5958e8906e81167. Decision matches prior E-105-006
Approved exact subject and E-105-005 PASS WITH LIMITATION. No projected subject
mutation or post-decision scope expansion found. Excluded status remains
verification-stable In Progress with matching-envelope resolution; Acceptance
and environment limitations are explicit, without contract/status conflict.

Raw module OIDs independently unchanged: go.mod
aa82a60c0769256f13c86b31382a8994cf6e73e4, go.sum
f1912accf367972c7c6278b7a8d7de478028139b. HEAD/main/cached origin/main remain
fb916df77e697fb7a45dc7bb6e8d0150411f2d82; current branch remains
codex/task-105-minimal-github-ci. Actual TASK-102 ref
feature/task-102-lpac-process-admission remains
8bfa2e22a4b6310a71f6d36c8724b1adedd727ad. Index remains empty. No task commit
or publication was performed; no settings/required-check mutation authorized
or executed in this engineering cycle.

Metadata provenance checked separately: pre-Reviewer prefix length 24719 bytes
still has SHA256 ddd4fb30bd0303092147f92a8130ac0865c7e06f8313305c9adef97924464303.
The E-105-006 body was compared byte-for-byte with its original authored
UTF-8/CRLF payload and matched. Coordinator E-105-007 was written append-only;
its LF separator differs from Reviewer CRLF and is preserved without rewriting.
A preliminary separator assertion rejected that newline assumption only;
read-only reconciliation confirmed unchanged E-105-006 content and no unknown
mutation. No nonexistent pre-E-105-007 raw snapshot/hash is claimed. This E-105-008
append preserves its complete predecessor prefix byte-for-byte; metadata does
not self-attest final raw bytes. No interruption or subject rework occurred.

All E-105-005/E-105-007 limitations remain: hosted execution not performed,
Linux runtime and CGO/race not proven, local gcc/actionlint absent; native
checkout CRLF formatter/tidy rejection distinguished from LF-source substitute
proof. No additional product tests/installations needed for this integrity gate.
First incomplete checkpoint: separately authorized Owner Commit Gate.
STOP AT COMMIT GATE. No staging, commit or publication permission is granted.
