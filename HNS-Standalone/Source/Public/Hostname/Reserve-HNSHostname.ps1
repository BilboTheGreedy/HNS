function Reserve-HNSHostname {
    <#
    .SYNOPSIS
        Reserves an available hostname
    .DESCRIPTION
        Changes the status of a hostname from Available to Reserved, preventing others from using it
    .PARAMETER ID
        The ID of the hostname to reserve
    .PARAMETER Name
        The name of the hostname to reserve
    .PARAMETER User
        The user reserving the hostname (defaults to current user)
    .PARAMETER PassThru
        Return the updated hostname object
    .EXAMPLE
        Reserve-HNSHostname -Name "ABCS001"
        
        Reserves the hostname ABCS001
    .EXAMPLE
        Get-HNSHostname -Status Available | Select-Object -First 5 | Reserve-HNSHostname
        
        Reserves the first 5 available hostnames
    #>
    [CmdletBinding(DefaultParameterSetName = 'ByID', SupportsShouldProcess)]
    param(
        [Parameter(Mandatory, ParameterSetName = 'ByID', ValueFromPipelineByPropertyName)]
        [int]$ID,
        
        [Parameter(Mandatory, ParameterSetName = 'ByName', ValueFromPipelineByPropertyName)]
        [string]$Name,
        
        [Parameter()]
        [string]$User = $script:Configuration.DefaultUser,
        
        [Parameter()]
        [switch]$PassThru
    )
    
    begin {
        Test-HNSInitialized -Throw
        Write-Verbose "Reserving hostname(s)"
    }
    
    process {
        try {
            # Find hostname
            $hostname = if ($PSCmdlet.ParameterSetName -eq 'ByID') {
                $script:Hostnames | Where-Object { $_.ID -eq $ID }
            } else {
                $script:Hostnames | Where-Object { $_.Name -eq $Name }
            }
            
            if (-not $hostname) {
                throw "Hostname not found"
            }
            
            if ($hostname.Status -ne [HostnameStatus]::Available) {
                throw "Hostname '$($hostname.Name)' cannot be reserved. Current status: $($hostname.Status)"
            }
            
            if ($PSCmdlet.ShouldProcess($hostname.Name, "Reserve hostname")) {
                # Reserve the hostname
                $hostname.Reserve($User)
                
                # Save changes
                Save-HNSData -Type Hostnames
                
                # Audit log
                Write-AuditLog -Action "ReserveHostname" -Details "Reserved hostname: $($hostname.Name)" -User $User
                
                Write-Verbose "Reserved hostname: $($hostname.Name)"
                
                if ($PassThru) {
                    return $hostname
                }
            }
        }
        catch {
            Write-Error "Failed to reserve hostname: $_"
            throw
        }
    }
}