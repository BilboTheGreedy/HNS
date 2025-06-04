function Initialize-HNSEnvironment {
    <#
    .SYNOPSIS
        Initializes the HNS standalone environment
    .DESCRIPTION
        Sets up the necessary directory structure, configuration files, and loads data for the HNS module.
        This function must be called before using other HNS functions.
    .PARAMETER Path
        Custom data path for HNS files. If not specified, uses %APPDATA%\HNS-Standalone
    .PARAMETER Force
        Force reinitialization even if already initialized
    .PARAMETER LoadEpirocTemplate
        Automatically create the Epiroc VM Standard template
    .PARAMETER DNSServers
        DNS servers to use for hostname verification
    .PARAMETER CheckDNS
        Enable DNS availability checking (default: true)
    .PARAMETER CheckICMP
        Enable ICMP ping checking (default: true)
    .PARAMETER CheckActiveDirectory
        Enable Active Directory checking (default: true)
    .PARAMETER CheckServiceNow
        Enable ServiceNow CMDB checking (default: false)
    .PARAMETER ServiceNowConfig
        ServiceNow connection configuration
    .EXAMPLE
        Initialize-HNSEnvironment
        
        Initializes HNS with default settings
    .EXAMPLE
        Initialize-HNSEnvironment -Path "C:\HNS" -LoadEpirocTemplate -DNSServers @("8.8.8.8", "8.8.4.4")
        
        Initializes HNS with custom path, loads Epiroc template, and configures DNS servers
    .EXAMPLE
        Initialize-HNSEnvironment -CheckServiceNow -ServiceNowConfig @{Instance='dev12345'; Username='user'; Token='xyz123'}
    #>
    [CmdletBinding(SupportsShouldProcess)]
    param(
        [Parameter()]
        [ValidateScript({
            if ($_ -and -not (Test-Path (Split-Path $_ -Parent))) {
                throw "Parent directory does not exist: $_"
            }
            $true
        })]
        [string]$Path,
        
        [Parameter()]
        [switch]$Force,
        
        [Parameter()]
        [switch]$LoadEpirocTemplate,
        
        [Parameter()]
        [string[]]$DNSServers = @(),
        
        [Parameter()]
        [switch]$CheckDNS = $true,
        
        [Parameter()]
        [switch]$CheckICMP = $true,
        
        [Parameter()]
        [switch]$CheckActiveDirectory = $true,
        
        [Parameter()]
        [switch]$CheckServiceNow = $false,
        
        [Parameter()]
        [hashtable]$ServiceNowConfig = @{}
    )
    
    begin {
        Write-Verbose "Initializing HNS Standalone environment..."
    }
    
    process {
        try {
            # Check if already initialized
            if (-not $Force -and (Test-HNSInitialized)) {
                Write-Verbose "HNS environment already initialized"
                return $script:Configuration
            }
            
            # Create or load configuration
            if ($Path) {
                $script:Configuration = [HNSConfiguration]::new($Path)
            } else {
                $script:Configuration = [HNSConfiguration]::new()
            }
            
            # Set DNS servers if provided
            if ($DNSServers) {
                $script:Configuration.DNSServers = $DNSServers
            }
            
            # Configure availability checking settings
            $script:Configuration.AvailabilityChecks = @{
                CheckDNS = $CheckDNS
                CheckICMP = $CheckICMP
                CheckActiveDirectory = $CheckActiveDirectory
                CheckServiceNow = $CheckServiceNow
                ServiceNowConfig = $ServiceNowConfig
            }
            
            # Validate and create directory structure
            $script:Configuration.Validate()
            
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
            
            # Initialize or load configuration
            $configPath = $script:Configuration.GetConfigPath()
            if (-not (Test-Path $configPath) -or $Force) {
                if ($PSCmdlet.ShouldProcess($configPath, "Create configuration file")) {
                    Save-HNSData -Type Configuration
                    Write-Verbose "Created configuration file: $configPath"
                }
            } else {
                Load-HNSData -Type Configuration
            }
            
            # Initialize data collections
            $script:Templates = @()
            $script:Hostnames = @()
            
            # Initialize data files
            $templatesPath = $script:Configuration.GetTemplatesPath()
            if (-not (Test-Path $templatesPath)) {
                if ($PSCmdlet.ShouldProcess($templatesPath, "Create templates file")) {
                    Save-HNSData -Type Templates
                    Write-Verbose "Created templates file: $templatesPath"
                }
            } else {
                Load-HNSData -Type Templates
            }
            
            $hostnamesPath = $script:Configuration.GetHostnamesPath()
            if (-not (Test-Path $hostnamesPath)) {
                if ($PSCmdlet.ShouldProcess($hostnamesPath, "Create hostnames file")) {
                    Save-HNSData -Type Hostnames
                    Write-Verbose "Created hostnames file: $hostnamesPath"
                }
            } else {
                Load-HNSData -Type Hostnames
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
            
            # Load Epiroc template if requested
            if ($LoadEpirocTemplate -and -not ($script:Templates | Where-Object { $_.Name -eq "Epiroc VM Standard" })) {
                if ($PSCmdlet.ShouldProcess("Epiroc VM Standard", "Create template")) {
                    Import-EpirocTemplate
                    Write-Verbose "Loaded Epiroc VM Standard template"
                }
            }
            
            Write-AuditLog -Action "Initialize" -Details "Environment initialized at $($script:Configuration.DataPath)"
            Write-Information "HNS environment initialized successfully at: $($script:Configuration.DataPath)" -InformationAction Continue
            
            return $script:Configuration
        }
        catch {
            Write-Error "Failed to initialize HNS environment: $_"
            throw
        }
    }
}