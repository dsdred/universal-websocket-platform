# Read-only evidence collector. Native authorization/ownership/Target checks stay
# with Publisher; this script cannot grant publication authority or full P0 PASS.
[CmdletBinding()]
param(
    [string]$ExpectedPrincipal,
    [string]$ExpectedSid,
    [string]$ExpectedOrigin,
    [string]$ExpectedRepository,
    [string]$ExpectedGitHubLogin,
    [string]$ExpectedBranch,
    [ValidateSet('Discovery', 'P0Capability')][string]$Mode = 'Discovery',
    [ValidateSet('merge', 'squash', 'rebase')][string]$MergeMethod = 'merge',
    [string]$RepositoryPath = (Get-Location).Path
)

function Get-PublisherIdentity {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    try { return @{ Principal = $identity.Name; Sid = $identity.User.Value; ProcessId = $PID } }
    finally { $identity.Dispose() }
}

function Invoke-PublisherRead {
    param([string]$Tool, [string[]]$Arguments, [string]$Path)
    # Arguments are fixed read-only commands plus validated route facts, never a shell.
    $start = New-Object Diagnostics.ProcessStartInfo
    $start.FileName = $Tool
    $start.WorkingDirectory = $Path
    $start.UseShellExecute = $false
    $start.CreateNoWindow = $true
    $start.RedirectStandardOutput = $true
    $start.RedirectStandardError = $true
    $start.Arguments = ($Arguments | ForEach-Object {
        '"' + ($_ -replace '(\\*)"', '$1$1\"' -replace '(\\+)$', '$1$1') + '"'
    }) -join ' '
    $process = New-Object Diagnostics.Process
    $process.StartInfo = $start
    try {
        [void]$process.Start()
        $stdout = $process.StandardOutput.ReadToEndAsync()
        $stderr = $process.StandardError.ReadToEndAsync()
        if (-not $process.WaitForExit(30000)) {
            $process.Kill()
            return @{ ExitCode = -1; Output = ''; HttpStatus = 0; Failure = 'Timeout' }
        }
        $errorText = $stderr.GetAwaiter().GetResult()
        $status = 0
        if ($errorText -match '\(HTTP (\d{3})\)') { $status = [int]$Matches[1] }
        return @{ ExitCode = $process.ExitCode; Output = $stdout.GetAwaiter().GetResult(); HttpStatus = $status; Failure = 'CommandFailed' }
    }
    catch { return @{ ExitCode = -1; Output = ''; HttpStatus = 0; Failure = 'ToolFailure' } }
    finally { $process.Dispose() }
}

function Get-PublisherCapability {
    param([hashtable]$Route)
    $result = [ordered]@{
        Schema = 'publisher-capability-v1'; Mode = $Route.Mode; Capability = 'FAIL'
        FullP0 = $false; OwnershipAssigned = $false; PublicationAuthority = $false
        FutureMutationGuaranteed = $false; Identity = $null; Facts = [ordered]@{}
        Failure = $null
        Limitations = @('Native gate/session/ownership and immutable Target require independent verification.',
            'Full rule parameters and target-ref traditional protection require independent Publisher inspection.',
            'Read-only API permissions do not prove credential write scopes or future mutation success.',
            'Git and API may use different credentials; this report is not reusable authority.')
    }
    $saved = @{}
    try {
        foreach ($field in @('ExpectedPrincipal', 'ExpectedSid', 'ExpectedOrigin', 'ExpectedRepository', 'ExpectedGitHubLogin')) {
            if ([string]::IsNullOrWhiteSpace($Route[$field])) { throw 'RouteIncomplete' }
        }
        if ($Route.Mode -notin @('Discovery', 'P0Capability') -or $Route.MergeMethod -notin @('merge', 'squash', 'rebase')) { throw 'RouteInvalid' }
        if ($Route.ExpectedRepository -notmatch '^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$' -or
            $Route.ExpectedGitHubLogin -notmatch '^[A-Za-z0-9-]+$' -or
            $Route.ExpectedSid -notmatch '^S-1-[0-9-]+$') { throw 'RouteInvalid' }
        $repoName = $Route.ExpectedRepository
        if ($Route.ExpectedOrigin -cne "https://github.com/$repoName.git" -and
            $Route.ExpectedOrigin -cne "git@github.com:$repoName.git") { throw 'OriginRouteInvalid' }
        $actual = Get-PublisherIdentity
        $result.Identity = $actual
        if ($actual.Principal -ine $Route.ExpectedPrincipal -or $actual.Sid -cne $Route.ExpectedSid) { throw 'ExecutionIdentityMismatch' }
        # Process-local prompt suppression is restored even after any failed probe.
        foreach ($name in @('GIT_TERMINAL_PROMPT', 'GCM_INTERACTIVE', 'GH_PROMPT_DISABLED', 'GIT_OPTIONAL_LOCKS', 'GIT_SSH_COMMAND', 'GIT_SSH_VARIANT')) {
            $saved[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
        }
        $env:GIT_TERMINAL_PROMPT = '0'; $env:GCM_INTERACTIVE = 'Never'
        $env:GH_PROMPT_DISABLED = '1'; $env:GIT_OPTIONAL_LOCKS = '0'
        if ($Route.ExpectedOrigin -ceq "git@github.com:$repoName.git") {
            $env:GIT_SSH_COMMAND = 'ssh -o BatchMode=yes'; $env:GIT_SSH_VARIANT = 'ssh'
        }
        $path = $Route.RepositoryPath
        $local = @{}
        foreach ($probe in @(
            @{ Name = 'Origin'; Args = @('remote', 'get-url', 'origin') },
            @{ Name = 'Branch'; Args = @('symbolic-ref', '--quiet', '--short', 'HEAD') },
            @{ Name = 'Head'; Args = @('rev-parse', '--verify', 'HEAD') },
            @{ Name = 'Status'; Args = @('status', '--porcelain=v1', '--untracked-files=all') }
        )) {
            $read = Invoke-PublisherRead 'git.exe' $probe.Args $path
            if ($read.ExitCode -ne 0) { throw ('GitLocalReadFailed:' + $probe.Name) }
            $local[$probe.Name] = $read.Output.Trim()
        }
        if ($local.Origin -cne $Route.ExpectedOrigin) { throw 'OriginMismatch' }
        if ($local.Head -notmatch '^[0-9a-f]{40}([0-9a-f]{24})?$' -or $local.Branch -notmatch '^[A-Za-z0-9_./-]+$') { throw 'LocalRefInvalid' }
        $result.Facts.Local = @{ Branch = $local.Branch; Head = $local.Head; Clean = ($local.Status.Length -eq 0); OriginMatched = $true }
        if ($Route.Mode -eq 'P0Capability' -and $local.Status.Length -ne 0) { throw 'RepositoryDirty' }
        $api = @{}
        foreach ($probe in @(
            @{ Name = 'User'; Endpoint = 'user' },
            @{ Name = 'Repository'; Endpoint = "repos/$repoName" }
        )) {
            $read = Invoke-PublisherRead 'gh.exe' @('api', '--hostname', 'github.com', '--method', 'GET', $probe.Endpoint) $path
            if ($read.ExitCode -ne 0) { throw ('GitHubReadFailed:' + $probe.Name + ':HTTP' + $read.HttpStatus) }
            try { $api[$probe.Name] = (ConvertFrom-Json ('{"value":' + $read.Output + '}') -ErrorAction Stop).value }
            catch { throw 'GitHubResponseInvalid' }
        }
        if ($api.User.login -cne $Route.ExpectedGitHubLogin) { throw 'GitHubIdentityMismatch' }
        $repo = $api.Repository
        if ($repo.full_name -cne $repoName -or $repo.html_url -cne "https://github.com/$repoName" -or $repo.default_branch -cne 'main') { throw 'RepositoryIdentityMismatch' }
        if ($repo.archived -ne $false -or $repo.disabled -ne $false) { throw 'RepositoryUnavailable' }
        if ($repo.permissions.pull -ne $true -or $repo.permissions.push -ne $true) { throw 'RepositoryPermissionUnavailable' }
        $mergeProperty = @{ merge = 'allow_merge_commit'; squash = 'allow_squash_merge'; rebase = 'allow_rebase_merge' }[$Route.MergeMethod]
        if ($repo.$mergeProperty -ne $true) { throw 'MergeMethodUnavailable' }
        $result.Facts.GitHub = @{ Login = $api.User.login; Repository = $repoName; DefaultBranch = 'main'; Pull = $true; Push = $true; MergeMethod = $Route.MergeMethod; Archived = $false; Disabled = $false }
        $read = Invoke-PublisherRead 'git.exe' @('ls-remote', '--exit-code', 'origin', 'refs/heads/main') $path
        if ($read.ExitCode -ne 0) { throw 'OriginReadFailed' }
        if ($read.Output.Trim() -notmatch '^([0-9a-f]{40}([0-9a-f]{24})?)\s+refs/heads/main$') { throw 'OriginRefInvalid' }
        $result.Facts.OriginMain = $Matches[1]
        if ($Route.ExpectedBranch -and $local.Branch -cne $Route.ExpectedBranch) { throw 'TargetBranchMismatch' }
        foreach ($probe in @(
            @{ Name = 'Branch'; Endpoint = "repos/$repoName/branches/main" },
            @{ Name = 'Rules'; Endpoint = "repos/$repoName/rules/branches/main" },
            @{ Name = 'Protection'; Endpoint = "repos/$repoName/branches/main/protection" }
        )) {
            $read = Invoke-PublisherRead 'gh.exe' @('api', '--hostname', 'github.com', '--method', 'GET', $probe.Endpoint) $path
            if ($read.ExitCode -ne 0) {
                if ($probe.Name -eq 'Protection' -and $read.HttpStatus -eq 404) {
                    # REST 404 alone is ambiguous. An explicit null exact-ref rule,
                    # successful GraphQL response and matching ref/OID prove absence.
                    $query = 'query($owner:String!,$name:String!,$qualifiedRef:String!){repository(owner:$owner,name:$name){nameWithOwner ref(qualifiedName:$qualifiedRef){prefix name target{oid} branchProtectionRule{id}}}}'
                    $parts = $repoName.Split('/')
                    $graph = Invoke-PublisherRead 'gh.exe' @('api', '--hostname', 'github.com', '--method', 'POST', 'graphql', '-f', ('query=' + $query), '-f', ('owner=' + $parts[0]), '-f', ('name=' + $parts[1]), '-f', 'qualifiedRef=refs/heads/main') $path
                    if ($graph.ExitCode -ne 0) { throw 'ProtectionVisibilityUnavailable' }
                    try { $proof = ConvertFrom-Json $graph.Output -ErrorAction Stop }
                    catch { throw 'ProtectionProofInvalid' }
                    $graphRepo = $proof.data.repository
                    $ref = $graphRepo.ref
                    if (@($proof.errors).Where({ $null -ne $_ }).Count -ne 0 -or $graphRepo.nameWithOwner -cne $repoName -or
                        $ref.prefix -cne 'refs/heads/' -or $ref.name -cne 'main' -or $ref.target.oid -cne $result.Facts.OriginMain -or
                        $null -eq $ref.PSObject.Properties['branchProtectionRule'] -or $null -ne $ref.branchProtectionRule) { throw 'ProtectionVisibilityUnavailable' }
                    $api.Protection = $null
                } else { throw ('PolicyVisibilityUnavailable:' + $probe.Name + ':HTTP' + $read.HttpStatus) }
            } else {
                try { $api[$probe.Name] = (ConvertFrom-Json ('{"value":' + $read.Output + '}') -ErrorAction Stop).value }
                catch { throw 'PolicyResponseInvalid' }
            }
        }
        if ($api.Branch.name -cne 'main' -or $api.Branch.protected -isnot [bool] -or $api.Branch.commit.sha -cne $result.Facts.OriginMain) { throw 'DefaultBranchMismatch' }
        if ($null -ne $api.Protection -and ($api.Protection -isnot [pscustomobject] -or
            $null -eq $api.Protection.PSObject.Properties['required_status_checks'])) { throw 'ProtectionResponseInvalid' }
        if ($null -eq $api.Rules -or ($api.Rules -isnot [array] -and $api.Rules -isnot [System.Collections.IList])) { throw 'RulesResponseInvalid' }
        $types = @($api.Rules | ForEach-Object { if ($_.type -notmatch '^[a-z_]+$') { throw 'RulesResponseInvalid' }; $_.type })
        $encodedBranch = [Uri]::EscapeDataString($local.Branch)
        $targetRead = Invoke-PublisherRead 'gh.exe' @('api', '--hostname', 'github.com', '--method', 'GET', "repos/$repoName/rules/branches/$encodedBranch") $path
        if ($targetRead.ExitCode -ne 0) { throw 'TargetPolicyVisibilityUnavailable' }
        try { $targetRules = (ConvertFrom-Json ('{"value":' + $targetRead.Output + '}') -ErrorAction Stop).value }
        catch { throw 'TargetRulesResponseInvalid' }
        if ($null -eq $targetRules -or ($targetRules -isnot [array] -and $targetRules -isnot [System.Collections.IList])) { throw 'TargetRulesResponseInvalid' }
        $targetTypes = @($targetRules | ForEach-Object { if ($_.type -notmatch '^[a-z_]+$') { throw 'TargetRulesResponseInvalid' }; $_.type })
        $result.Facts.Policy = @{ BranchProtected = $api.Branch.protected; EffectiveRuleTypes = $types; TargetRuleTypes = $targetTypes; TraditionalProtection = $(if ($null -eq $api.Protection) { 'NoTraditionalRuleObserved' } else { 'Visible' }) }
        $result.Capability = 'PASS'
    }
    catch {
        # Only our fixed failure codes are emitted; never exception/tool payloads.
        $code = $_.Exception.Message
        $result.Failure = $(if ($code -match '^[A-Za-z]+(:[A-Za-z0-9]+)*$') { $code } else { 'ProbeFailure' })
    }
    finally {
        foreach ($name in $saved.Keys) {
            # PowerShell can bind $null as empty string to the .NET overload;
            # remove explicitly to preserve absent versus empty on modern .NET.
            if ($null -eq $saved[$name]) { Remove-Item -LiteralPath ('Env:' + $name) -ErrorAction SilentlyContinue }
            else { [Environment]::SetEnvironmentVariable($name, $saved[$name], 'Process') }
        }
    }
    return [pscustomobject]$result
}

if ($MyInvocation.InvocationName -ne '.') {
    $route = @{
        ExpectedPrincipal = $ExpectedPrincipal; ExpectedSid = $ExpectedSid
        ExpectedOrigin = $ExpectedOrigin; ExpectedRepository = $ExpectedRepository
        ExpectedGitHubLogin = $ExpectedGitHubLogin; Mode = $Mode
        ExpectedBranch = $ExpectedBranch
        MergeMethod = $MergeMethod; RepositoryPath = $RepositoryPath
    }
    $report = Get-PublisherCapability $route
    $report | ConvertTo-Json -Depth 8
    if ($report.Capability -ne 'PASS') { exit 1 }
}
