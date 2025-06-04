function Initialize-HNSEnvironment {
    <#
    .SYNOPSIS
        Initializes the HNS module environment
    .DESCRIPTION
        Creates necessary directories and files for the HNS module to function
    .PARAMETER Path
        Custom data path (optional)
    .PARAMETER Force
        Force initialization even if already initialized
    #>
    [CmdletBinding(SupportsShouldProcess)]
    param(
        [Parameter()]
        [string]$Path,
        
        [Parameter()]
        [switch]$Force
    )
    
    begin {
        Write-Verbose "Initializing HNS environment..."
    }
    
    process {
        try {
            # Load or create configuration
            if ($Path) {
                $script:Configuration = [HNSConfiguration]::new($Path)
            } else {
                $script:Configuration = [HNSConfiguration]::new()
            }
            
            # Validate configuration
            $script:Configuration.Validate()
            
            # Create directory structure
            $directories = @(
                $script:Configuration.DataPath
                $script:Configuration.GetBackupPath()
                (Join-Path $script:Configuration.DataPath 'logs')
                (Join-Path $script:Configuration.DataPath 'exports')
                (Join-Path $script:Configuration.DataPath 'archive')
            )
            
            foreach ($dir in $directories) {
                if (-not (Test-Path $dir)) {
                    if ($PSCmdlet.ShouldProcess($dir, "Create directory")) {
                        New-Item -Path $dir -ItemType Directory -Force | Out-Null
                        Write-Verbose "Created directory: $dir"
                    }
                }
            }
            
            # Initialize data files if they don't exist
            $configPath = $script:Configuration.GetConfigPath()
            if (-not (Test-Path $configPath) -or $Force) {
                if ($PSCmdlet.ShouldProcess($configPath, "Create configuration file")) {
                    $script:Configuration | ConvertTo-Json -Depth 10 | Set-Content -Path $configPath -Encoding UTF8
                    Write-Verbose "Created configuration file: $configPath"
                }
            } else {
                # Load existing configuration
                $existingConfig = Get-Content -Path $configPath -Raw | ConvertFrom-Json
                $script:Configuration = [HNSConfiguration]::new($existingConfig)
            }
            
            # Initialize templates file
            $templatesPath = $script:Configuration.GetTemplatesPath()
            if (-not (Test-Path $templatesPath)) {
                if ($PSCmdlet.ShouldProcess($templatesPath, "Create templates file")) {
                    @() | ConvertTo-Json | Set-Content -Path $templatesPath -Encoding UTF8
                    Write-Verbose "Created templates file: $templatesPath"
                }
            }
            
            # Initialize hostnames file
            $hostnamesPath = $script:Configuration.GetHostnamesPath()
            if (-not (Test-Path $hostnamesPath)) {
                if ($PSCmdlet.ShouldProcess($hostnamesPath, "Create hostnames file")) {
                    @() | ConvertTo-Json | Set-Content -Path $hostnamesPath -Encoding UTF8
                    Write-Verbose "Created hostnames file: $hostnamesPath"
                }
            }
            
            # Initialize audit log
            if ($script:Configuration.EnableAuditLog) {
                $auditPath = $script:Configuration.GetAuditLogPath()
                if (-not (Test-Path $auditPath)) {
                    if ($PSCmdlet.ShouldProcess($auditPath, "Create audit log")) {
                        $header = "Timestamp,Action,User,Details"
                        Set-Content -Path $auditPath -Value $header -Encoding UTF8
                        Write-Verbose "Created audit log: $auditPath"
                    }
                }
            }
            
            Write-Verbose "HNS environment initialized successfully"
            Write-AuditLog -Action "Initialize" -Details "Environment initialized at $($script:Configuration.DataPath)"
            
            return $script:Configuration
        }
        catch {
            Write-Error "Failed to initialize HNS environment: $_"
            throw
        }
    }
}