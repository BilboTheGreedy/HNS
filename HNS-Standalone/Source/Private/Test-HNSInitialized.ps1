function Test-HNSInitialized {
    <#
    .SYNOPSIS
        Tests if the HNS environment is initialized
    .DESCRIPTION
        Checks if the HNS module has been properly initialized
    .PARAMETER Throw
        Throw an exception if not initialized
    #>
    [CmdletBinding()]
    param(
        [Parameter()]
        [switch]$Throw
    )
    
    $initialized = $null -ne $script:Configuration -and 
                   $null -ne $script:Templates -and 
                   $null -ne $script:Hostnames
    
    if (-not $initialized -and $Throw) {
        throw "HNS environment not initialized. Please run Initialize-HNSEnvironment first."
    }
    
    return $initialized
}