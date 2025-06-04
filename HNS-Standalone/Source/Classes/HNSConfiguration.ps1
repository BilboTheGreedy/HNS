class HNSConfiguration {
    [string] $DataPath
    [int] $NextTemplateID
    [int] $NextHostnameID
    [string] $DefaultUser
    [string[]] $DNSServers
    [int] $DNSTimeout
    [bool] $AutoBackup
    [int] $BackupRetentionDays
    [bool] $EnableAuditLog
    [string] $LogLevel
    [hashtable] $CustomSettings
    [hashtable] $AvailabilityChecks
    
    # Default constructor
    HNSConfiguration() {
        $this.DataPath = Join-Path $env:APPDATA 'HNS-Standalone'
        $this.NextTemplateID = 1
        $this.NextHostnameID = 1
        $this.DefaultUser = $env:USERNAME
        $this.DNSServers = @()
        $this.DNSTimeout = 5
        $this.AutoBackup = $true
        $this.BackupRetentionDays = 30
        $this.EnableAuditLog = $true
        $this.LogLevel = 'Information'
        $this.CustomSettings = @{}
        $this.AvailabilityChecks = @{
            CheckDNS = $true
            CheckICMP = $true
            CheckActiveDirectory = $true
            CheckServiceNow = $false
            ServiceNowConfig = @{}
        }
    }
    
    # Constructor with custom data path
    HNSConfiguration([string]$DataPath) {
        $this.DataPath = $DataPath
        $this.NextTemplateID = 1
        $this.NextHostnameID = 1
        $this.DefaultUser = $env:USERNAME
        $this.DNSServers = @()
        $this.DNSTimeout = 5
        $this.AutoBackup = $true
        $this.BackupRetentionDays = 30
        $this.EnableAuditLog = $true
        $this.LogLevel = 'Information'
        $this.CustomSettings = @{}
        $this.AvailabilityChecks = @{
            CheckDNS = $true
            CheckICMP = $true
            CheckActiveDirectory = $true
            CheckServiceNow = $false
            ServiceNowConfig = @{}
        }
    }
    
    # Constructor with hashtable
    HNSConfiguration([hashtable]$Properties) {
        # Set defaults first
        $this.DataPath = Join-Path $env:APPDATA 'HNS-Standalone'
        $this.NextTemplateID = 1
        $this.NextHostnameID = 1
        $this.DefaultUser = $env:USERNAME
        $this.DNSServers = @()
        $this.DNSTimeout = 5
        $this.AutoBackup = $true
        $this.BackupRetentionDays = 30
        $this.EnableAuditLog = $true
        $this.LogLevel = 'Information'
        $this.CustomSettings = @{}
        $this.AvailabilityChecks = @{
            CheckDNS = $true
            CheckICMP = $true
            CheckActiveDirectory = $true
            CheckServiceNow = $false
            ServiceNowConfig = @{}
        }
        
        # Override with provided values
        foreach ($key in $Properties.Keys) {
            if ($null -ne $this.PSObject.Properties[$key]) {
                $this.$key = $Properties[$key]
            }
        }
    }
    
    # Get the templates file path
    [string] GetTemplatesPath() {
        return Join-Path $this.DataPath 'templates.json'
    }
    
    # Get the hostnames file path
    [string] GetHostnamesPath() {
        return Join-Path $this.DataPath 'hostnames.json'
    }
    
    # Get the configuration file path
    [string] GetConfigPath() {
        return Join-Path $this.DataPath 'config.json'
    }
    
    # Get the audit log path
    [string] GetAuditLogPath() {
        return Join-Path $this.DataPath 'audit.log'
    }
    
    # Get the backup directory path
    [string] GetBackupPath() {
        return Join-Path $this.DataPath 'backups'
    }
    
    # Get next ID and increment
    [int] GetNextTemplateID() {
        $id = $this.NextTemplateID
        $this.NextTemplateID++
        return $id
    }
    
    [int] GetNextHostnameID() {
        $id = $this.NextHostnameID
        $this.NextHostnameID++
        return $id
    }
    
    # Validate configuration
    [bool] Validate() {
        # Check if data path exists or can be created
        if (-not (Test-Path $this.DataPath)) {
            try {
                New-Item -Path $this.DataPath -ItemType Directory -Force | Out-Null
            }
            catch {
                throw "Cannot create data path: $($this.DataPath)"
            }
        }
        
        # Validate DNS timeout
        if ($this.DNSTimeout -lt 1 -or $this.DNSTimeout -gt 60) {
            throw "DNSTimeout must be between 1 and 60 seconds"
        }
        
        # Validate backup retention
        if ($this.BackupRetentionDays -lt 0 -or $this.BackupRetentionDays -gt 365) {
            throw "BackupRetentionDays must be between 0 and 365"
        }
        
        # Validate log level
        $validLogLevels = @('Verbose', 'Debug', 'Information', 'Warning', 'Error')
        if ($this.LogLevel -notin $validLogLevels) {
            throw "LogLevel must be one of: $($validLogLevels -join ', ')"
        }
        
        return $true
    }
    
    # ToString override
    [string] ToString() {
        return "HNS Configuration (DataPath: $($this.DataPath), AutoBackup: $($this.AutoBackup), LogLevel: $($this.LogLevel))"
    }
}