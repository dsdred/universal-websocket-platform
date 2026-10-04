$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot '../publisher-capability-preflight.ps1')

# No network, credentials or Git mutation: replace only internal readers in this
# test process. Production has no fixture switch or fake-identity argument.
$nativeIdentity = Get-PublisherIdentity
$script:calls = New-Object 'System.Collections.Generic.List[string]'
$script:scenario = ''
$script:oid = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
$script:route = @{
    ExpectedPrincipal = 'TEST\publisher'; ExpectedSid = 'S-1-5-21-123-1002'
    ExpectedOrigin = 'https://github.com/example/project.git'
    ExpectedRepository = 'example/project'; ExpectedGitHubLogin = 'publisher'
    ExpectedBranch = 'docs/task-077'; RepositoryPath = $PSScriptRoot
    Mode = 'Discovery'; MergeMethod = 'merge'
}

function Invoke-PublisherRead {
    param([string]$Tool, [string[]]$Arguments, [string]$Path)
    $command = $Tool + ' ' + ($Arguments -join ' ')
    $script:calls.Add($command)
    $value = $null
    $status = 0
    if ($command -match 'git.exe remote get-url') { $value = $script:route.ExpectedOrigin; if ($script:scenario -eq 'origin') { $value = 'https://github.com/other/project.git' } }
    elseif ($command -match 'git.exe symbolic-ref') { $value = 'docs/task-077' }
    elseif ($command -match 'git.exe rev-parse') { $value = $script:oid }
    elseif ($command -match 'git.exe status') { $value = ''; if ($script:scenario -eq 'dirty') { $value = '?? pending.txt' } }
    elseif ($command -match 'git.exe ls-remote') {
        if ($script:route.ExpectedOrigin -match '^git@') {
            Assert-Test ($env:GIT_SSH_COMMAND -ceq 'ssh -o BatchMode=yes' -and $env:GIT_SSH_VARIANT -ceq 'ssh') 'SSH origin requires BatchMode.'
        }
        $value = $script:oid + "`trefs/heads/main"; if ($script:scenario -eq 'gitread') { $status = 1 }
    }
    elseif ($command -match 'GET user$') {
        $value = '{"login":"publisher"}'
        if ($script:scenario -eq 'user') { $value = '{"login":"other"}' }
        if ($script:scenario -eq 'auth') { $status = 401 }
    }
    elseif ($command -match 'GET repos/example/project$') {
        $repo = @{ full_name = 'example/project'; html_url = 'https://github.com/example/project'; default_branch = 'main'; archived = $false; disabled = $false; permissions = @{ pull = $true; push = $true }; allow_merge_commit = $true; allow_squash_merge = $true; allow_rebase_merge = $true }
        switch ($script:scenario) {
            'repo' { $repo.full_name = 'other/project' }
            'default' { $repo.default_branch = 'other' }
            'archive' { $repo.archived = $true }
            'disabled' { $repo.disabled = $true }
            'permission' { $repo.permissions.push = $false }
            'permissionmissing' { $repo.Remove('permissions') }
            'merge' { $repo.allow_merge_commit = $false }
        }
        $value = $repo | ConvertTo-Json -Depth 5 -Compress
    }
    elseif ($command -match 'GET repos/example/project/branches/main$') { $value = '{"name":"main","protected":true,"commit":{"sha":"' + $script:oid + '"}}' }
    elseif ($command -match 'GET repos/example/project/rules/branches/main$') {
        $value = '[{"type":"pull_request"}]'
        if ($script:scenario -eq 'rules') { $status = 403 }
        if ($script:scenario -eq 'rulesinvalid') { $value = '{}' }
    }
    elseif ($command -match 'GET repos/example/project/branches/main/protection$') {
        $value = '{"required_status_checks":null}'
        if ($script:scenario -in @('graphnull', 'grapherror', 'graphmissing', 'graphwrong', 'graphvisible', 'graphfail')) { $status = 404 }
        if ($script:scenario -eq 'protection') { $status = 403 }
    }
    elseif ($command -match 'POST graphql') {
        if ($command -match 'mutation' -or $command -notmatch 'query=') { throw 'Forbidden GraphQL operation' }
        $value = '{"data":{"repository":{"nameWithOwner":"example/project","ref":{"prefix":"refs/heads/","name":"main","target":{"oid":"' + $script:oid + '"},"branchProtectionRule":null}}},"errors":null}'
        if ($script:scenario -eq 'grapherror') { $value = '{"data":null,"errors":[{"message":"secret-not-to-print"}]}' }
        if ($script:scenario -eq 'graphmissing') { $value = $value.Replace(',"branchProtectionRule":null', '') }
        if ($script:scenario -eq 'graphwrong') { $value = $value.Replace('example/project', 'other/project') }
        if ($script:scenario -eq 'graphvisible') { $value = $value.Replace('"branchProtectionRule":null', '"branchProtectionRule":{"id":"123"}') }
        if ($script:scenario -eq 'graphfail') { $status = 403 }
    }
    elseif ($command -match 'GET repos/example/project/rules/branches/docs%2Ftask-077$') {
        $value = '[]'
        if ($script:scenario -eq 'targetrules') { $status = 403 }
    }
    else { throw ('Unexpected command: ' + $command) }
    return @{ ExitCode = $(if ($status -eq 0) { 0 } else { 1 }); Output = $value; HttpStatus = $status; Failure = 'CommandFailed' }
}

function Assert-Test([bool]$Condition, [string]$Message) {
    if (-not $Condition) { throw $Message }
}

# Real current identity is rejected before even our external-reader stub executes.
$report = Get-PublisherCapability $script:route
Assert-Test ($report.Failure -eq 'ExecutionIdentityMismatch' -and $script:calls.Count -eq 0) 'Native wrong identity must reject before commands.'
Write-Output ('PASS native identity rejection: ' + $nativeIdentity.Principal)

function Get-PublisherIdentity { return @{ Principal = 'TEST\publisher'; Sid = 'S-1-5-21-123-1002'; ProcessId = $PID } }
$count = 0
foreach ($case in @(
    @('', 'PASS'), @('origin', 'OriginMismatch'), @('auth', 'GitHubReadFailed:User:HTTP401'),
    @('user', 'GitHubIdentityMismatch'), @('repo', 'RepositoryIdentityMismatch'),
    @('default', 'RepositoryIdentityMismatch'), @('archive', 'RepositoryUnavailable'),
    @('disabled', 'RepositoryUnavailable'), @('permission', 'RepositoryPermissionUnavailable'),
    @('permissionmissing', 'RepositoryPermissionUnavailable'), @('merge', 'MergeMethodUnavailable'),
    @('gitread', 'OriginReadFailed'), @('rules', 'PolicyVisibilityUnavailable:Rules:HTTP403'),
    @('rulesinvalid', 'RulesResponseInvalid'), @('protection', 'PolicyVisibilityUnavailable:Protection:HTTP403'),
    @('graphnull', 'PASS'), @('grapherror', 'ProtectionVisibilityUnavailable'),
    @('graphmissing', 'ProtectionVisibilityUnavailable'), @('graphwrong', 'ProtectionVisibilityUnavailable'),
    @('graphvisible', 'ProtectionVisibilityUnavailable'), @('graphfail', 'ProtectionVisibilityUnavailable'),
    @('targetrules', 'TargetPolicyVisibilityUnavailable'), @('dirty', 'PASS')
)) {
    $script:scenario = $case[0]; $script:calls.Clear()
    $before = [Environment]::GetEnvironmentVariable('GIT_TERMINAL_PROMPT', 'Process')
    $report = Get-PublisherCapability $script:route
    $actual = $(if ($report.Capability -eq 'PASS') { 'PASS' } else { $report.Failure })
    Assert-Test ($actual -eq $case[1]) ('Scenario ' + $case[0] + ': expected ' + $case[1] + ', got ' + $actual)
    Assert-Test (-not $report.FullP0 -and -not $report.OwnershipAssigned -and -not $report.PublicationAuthority -and -not $report.FutureMutationGuaranteed) 'Collector may never assign authority.'
    Assert-Test ([Environment]::GetEnvironmentVariable('GIT_TERMINAL_PROMPT', 'Process') -ceq $before) 'Prompt setting must restore.'
    Assert-Test (($report | ConvertTo-Json -Depth 8) -notmatch 'secret-not-to-print') 'Tool error payload must not escape.'
    Assert-Test (($script:calls -join "`n") -notmatch '\b(push|fetch|pull|merge|checkout|update-ref|auth)\b') 'Only read-only probes allowed.'
    $count++
}
$script:scenario = 'dirty'; $script:route.Mode = 'P0Capability'; $script:calls.Clear()
$report = Get-PublisherCapability $script:route
Assert-Test ($report.Failure -eq 'RepositoryDirty' -and -not ($script:calls -match 'gh.exe')) 'Dirty P0Capability rejects before remote probes.'
$script:route.Mode = 'Discovery'; $script:route.ExpectedSid = 'S-1-5-21-123-9999'; $script:calls.Clear()
$report = Get-PublisherCapability $script:route
Assert-Test ($report.Failure -eq 'ExecutionIdentityMismatch' -and $script:calls.Count -eq 0) 'SID mismatch rejects before commands.'
$script:route.ExpectedSid = 'S-1-5-21-123-1002'; $script:scenario = ''
foreach ($method in @('merge', 'squash', 'rebase')) {
    $script:route.MergeMethod = $method
    $report = Get-PublisherCapability $script:route
    Assert-Test ($report.Capability -eq 'PASS') ('Real API merge property mapping failed: ' + $method)
}
$script:route.MergeMethod = 'invalid'; $script:calls.Clear()
$report = Get-PublisherCapability $script:route
Assert-Test ($report.Failure -eq 'RouteInvalid' -and $script:calls.Count -eq 0) 'Invalid internal merge enum must reject before reads.'
$script:route.MergeMethod = 'merge'; $script:route.Mode = 'invalid'; $script:calls.Clear()
$report = Get-PublisherCapability $script:route
Assert-Test ($report.Failure -eq 'RouteInvalid' -and $script:calls.Count -eq 0) 'Invalid internal mode enum must reject before reads.'
Write-Output ("PASS $count capability scenarios; dirty P0 separation; SID guard; no authority; prompt restore; safe output; read-only command inventory.")
Write-Output 'PASS real API merge/squash/rebase property mappings and internal enum guards.'

# Exercise actual process environment restoration, not a second mocked setter.
$script:route.Mode = 'Discovery'; $script:scenario = ''
$promptNames = @('GIT_TERMINAL_PROMPT', 'GCM_INTERACTIVE', 'GH_PROMPT_DISABLED', 'GIT_OPTIONAL_LOCKS', 'GIT_SSH_COMMAND', 'GIT_SSH_VARIANT')
$originalPrompt = @{}
foreach ($name in $promptNames) { $originalPrompt[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
try {
    foreach ($state in @('absent', 'present', 'empty')) {
        foreach ($name in $promptNames) {
            if ($state -eq 'absent') { Remove-Item -LiteralPath ('Env:' + $name) -ErrorAction SilentlyContinue }
            else { [Environment]::SetEnvironmentVariable($name, $(if ($state -eq 'empty') { '' } else { 'task077-sentinel' }), 'Process') }
        }
        $beforeValues = @{}
        foreach ($name in $promptNames) { $beforeValues[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
        foreach ($origin in @('https://github.com/example/project.git', 'git@github.com:example/project.git')) {
            $script:route.ExpectedOrigin = $origin
            $report = Get-PublisherCapability $script:route
            Assert-Test ($report.Capability -eq 'PASS') 'Environment scenario capability failed.'
            foreach ($name in $promptNames) {
                Assert-Test ([Environment]::GetEnvironmentVariable($name, 'Process') -ceq $beforeValues[$name]) ('Environment restore failed: ' + $state + '/' + $name)
            }
        }
        # Windows PowerShell/.NET Framework may represent empty as absence; compare
        # the actually representable original value rather than invent a state.
        Write-Output ('PASS process environment restoration: ' + $state)
    }
}
finally {
    foreach ($name in $promptNames) {
        if ($null -eq $originalPrompt[$name]) { Remove-Item -LiteralPath ('Env:' + $name) -ErrorAction SilentlyContinue }
        else { [Environment]::SetEnvironmentVariable($name, $originalPrompt[$name], 'Process') }
    }
}
