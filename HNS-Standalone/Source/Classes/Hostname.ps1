class Hostname {
    [int] $ID
    [string] $Name
    [int] $TemplateID
    [string] $TemplateName
    [HostnameStatus] $Status
    [int] $SequenceNum
    [string] $ReservedBy
    [datetime] $ReservedAt
    [string] $CommittedBy
    [Nullable[datetime]] $CommittedAt
    [string] $ReleasedBy
    [Nullable[datetime]] $ReleasedAt
    [bool] $DNSVerified
    [Nullable[datetime]] $DNSVerifiedAt
    [string] $DNSVerifiedBy
    [datetime] $CreatedAt
    [datetime] $UpdatedAt
    [hashtable] $Metadata
    [string] $Notes
    
    # Default constructor
    Hostname() {
        $this.Status = [HostnameStatus]::Available
        $this.DNSVerified = $false
        $this.CreatedAt = Get-Date
        $this.UpdatedAt = Get-Date
        $this.Metadata = @{}
    }
    
    # Constructor with hashtable
    Hostname([hashtable]$Properties) {
        $this.Status = [HostnameStatus]::Available
        $this.DNSVerified = $false
        $this.CreatedAt = Get-Date
        $this.UpdatedAt = Get-Date
        $this.Metadata = @{}
        
        foreach ($key in $Properties.Keys) {
            if ($null -ne $this.PSObject.Properties[$key]) {
                $this.$key = $Properties[$key]
            }
        }
    }
    
    # Reserve the hostname
    [void] Reserve([string]$ReservedBy) {
        if ($this.Status -ne [HostnameStatus]::Available) {
            throw "Hostname can only be reserved from Available status. Current status: $($this.Status)"
        }
        
        $this.Status = [HostnameStatus]::Reserved
        $this.ReservedBy = $ReservedBy
        $this.ReservedAt = Get-Date
        $this.UpdatedAt = Get-Date
    }
    
    # Commit the hostname
    [void] Commit([string]$CommittedBy) {
        if ($this.Status -ne [HostnameStatus]::Reserved) {
            throw "Hostname can only be committed from Reserved status. Current status: $($this.Status)"
        }
        
        $this.Status = [HostnameStatus]::Committed
        $this.CommittedBy = $CommittedBy
        $this.CommittedAt = Get-Date
        $this.UpdatedAt = Get-Date
    }
    
    # Release the hostname
    [void] Release([string]$ReleasedBy) {
        if ($this.Status -ne [HostnameStatus]::Committed) {
            throw "Hostname can only be released from Committed status. Current status: $($this.Status)"
        }
        
        $this.Status = [HostnameStatus]::Released
        $this.ReleasedBy = $ReleasedBy
        $this.ReleasedAt = Get-Date
        $this.UpdatedAt = Get-Date
    }
    
    # Set DNS verification status
    [void] SetDNSVerified([bool]$Verified, [string]$VerifiedBy) {
        $this.DNSVerified = $Verified
        if ($Verified) {
            $this.DNSVerifiedAt = Get-Date
            $this.DNSVerifiedBy = $VerifiedBy
        }
        else {
            $this.DNSVerifiedAt = $null
            $this.DNSVerifiedBy = $null
        }
        $this.UpdatedAt = Get-Date
    }
    
    # Get status color for display
    [string] GetStatusColor() {
        switch ($this.Status) {
            ([HostnameStatus]::Available) { return 'Green' }
            ([HostnameStatus]::Reserved) { return 'Yellow' }
            ([HostnameStatus]::Committed) { return 'Blue' }
            ([HostnameStatus]::Released) { return 'Gray' }
            default { return 'White' }
        }
        return 'White'
    }
    
    # Get age in days
    [int] GetAgeInDays() {
        return [Math]::Floor((Get-Date) - $this.CreatedAt).TotalDays
    }
    
    # Get time in current status
    [timespan] GetTimeInStatus() {
        $lastStatusChange = $this.CreatedAt
        
        switch ($this.Status) {
            ([HostnameStatus]::Reserved) { $lastStatusChange = $this.ReservedAt }
            ([HostnameStatus]::Committed) { $lastStatusChange = $this.CommittedAt }
            ([HostnameStatus]::Released) { $lastStatusChange = $this.ReleasedAt }
        }
        
        return (Get-Date) - $lastStatusChange
    }
    
    # ToString override
    [string] ToString() {
        return "$($this.Name) [$($this.Status)] (Template: $($this.TemplateName), Seq: $($this.SequenceNum))"
    }
    
    # Format for display
    [string] ToDisplayString() {
        $statusInfo = switch ($this.Status) {
            ([HostnameStatus]::Available) { "Available" }
            ([HostnameStatus]::Reserved) { "Reserved by $($this.ReservedBy) on $($this.ReservedAt.ToString('yyyy-MM-dd'))" }
            ([HostnameStatus]::Committed) { "Committed by $($this.CommittedBy) on $($this.CommittedAt.ToString('yyyy-MM-dd'))" }
            ([HostnameStatus]::Released) { "Released by $($this.ReleasedBy) on $($this.ReleasedAt.ToString('yyyy-MM-dd'))" }
        }
        
        $dnsInfo = if ($this.DNSVerified) { " [DNS OK]" } else { "" }
        
        return "$($this.Name) - $statusInfo$dnsInfo"
    }
}