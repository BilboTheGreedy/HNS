#Requires -Version 5.1
<#
.SYNOPSIS
    Build script for HNS-Standalone module
.DESCRIPTION
    Builds, tests, and packages the HNS-Standalone PowerShell module
.PARAMETER Task
    The build task to run (Build, Test, Analyze, Package, Clean, Default)
.PARAMETER Configuration
    Build configuration (Debug, Release)
.EXAMPLE
    .\build.ps1 -Task Build
    
    Builds the module
.EXAMPLE
    .\build.ps1 -Task Test -Configuration Debug
    
    Runs tests in debug configuration
#>
[CmdletBinding()]
param(
    [Parameter()]
    [ValidateSet('Build', 'Test', 'Analyze', 'Package', 'Clean', 'Default')]
    [string]$Task = 'Default',
    
    [Parameter()]
    [ValidateSet('Debug', 'Release')]
    [string]$Configuration = 'Release'
)

# Script variables
$ModuleName = 'HNS-Standalone'
$SourcePath = Join-Path $PSScriptRoot '..\Source'
$OutputPath = Join-Path $PSScriptRoot '..\Output'
$TestsPath = Join-Path $PSScriptRoot '..\Tests'
$DocsPath = Join-Path $PSScriptRoot '..\docs'

# Helper functions
function Write-BuildHeader {
    param([string]$Message)
    Write-Host "`n===== $Message =====" -ForegroundColor Cyan
}

function Test-ModuleDependencies {
    Write-BuildHeader "Checking Dependencies"
    
    $dependencies = @(
        @{Name = 'Pester'; MinVersion = '5.0.0'},
        @{Name = 'PSScriptAnalyzer'; MinVersion = '1.19.0'},
        @{Name = 'platyPS'; MinVersion = '0.14.0'}
    )
    
    foreach ($dep in $dependencies) {
        $module = Get-Module -ListAvailable -Name $dep.Name | 
                  Where-Object { $_.Version -ge $dep.MinVersion } | 
                  Select-Object -First 1
        
        if (-not $module) {
            Write-Warning "$($dep.Name) $($dep.MinVersion) or higher not found. Installing..."
            Install-Module -Name $dep.Name -MinimumVersion $dep.MinVersion -Force -Scope CurrentUser
        } else {
            Write-Host "✓ $($dep.Name) $($module.Version)" -ForegroundColor Green
        }
    }
}

function Invoke-Build {
    Write-BuildHeader "Building Module"
    
    # Create output directory
    $ModuleOutput = Join-Path $OutputPath $ModuleName
    if (Test-Path $ModuleOutput) {
        Remove-Item $ModuleOutput -Recurse -Force
    }
    New-Item -Path $ModuleOutput -ItemType Directory -Force | Out-Null
    
    # Copy module files
    Copy-Item -Path "$PSScriptRoot\..\$ModuleName.psd1" -Destination $ModuleOutput
    Copy-Item -Path "$PSScriptRoot\..\$ModuleName.psm1" -Destination $ModuleOutput
    
    # Copy source files maintaining structure
    $sourceDirs = @('Classes', 'Enums', 'Private', 'Public', 'Resources')
    foreach ($dir in $sourceDirs) {
        $sourcePath = Join-Path $SourcePath $dir
        if (Test-Path $sourcePath) {
            $destPath = Join-Path $ModuleOutput "Source\$dir"
            New-Item -Path $destPath -ItemType Directory -Force | Out-Null
            Copy-Item -Path "$sourcePath\*" -Destination $destPath -Recurse
        }
    }
    
    Write-Host "✓ Module built to: $ModuleOutput" -ForegroundColor Green
}

function Invoke-Tests {
    Write-BuildHeader "Running Tests"
    
    Import-Module Pester -MinimumVersion 5.0.0
    
    $pesterConfig = New-PesterConfiguration
    $pesterConfig.Run.Path = $TestsPath
    $pesterConfig.Run.PassThru = $true
    $pesterConfig.Output.Verbosity = 'Detailed'
    $pesterConfig.TestResult.Enabled = $true
    $pesterConfig.TestResult.OutputPath = Join-Path $OutputPath 'TestResults.xml'
    $pesterConfig.TestResult.OutputFormat = 'NUnitXml'
    
    if ($Configuration -eq 'Debug') {
        $pesterConfig.Debug.WriteDebugMessages = $true
        $pesterConfig.Debug.WriteDebugMessagesFrom = 'Mock'
    }
    
    $results = Invoke-Pester -Configuration $pesterConfig
    
    if ($results.FailedCount -gt 0) {
        throw "Tests failed: $($results.FailedCount) failed, $($results.PassedCount) passed"
    } else {
        Write-Host "✓ All tests passed: $($results.PassedCount) tests" -ForegroundColor Green
    }
}

function Invoke-ScriptAnalyzer {
    Write-BuildHeader "Running PSScriptAnalyzer"
    
    Import-Module PSScriptAnalyzer
    
    $analyzerParams = @{
        Path = Join-Path $OutputPath $ModuleName
        Recurse = $true
        Settings = 'PSGallery'
        ReportSummary = $true
    }
    
    $results = Invoke-ScriptAnalyzer @analyzerParams
    
    if ($results) {
        $results | Format-Table -AutoSize
        
        $errors = $results | Where-Object { $_.Severity -eq 'Error' }
        if ($errors) {
            throw "Script analysis found $($errors.Count) error(s)"
        } else {
            Write-Warning "Script analysis found $($results.Count) warning(s)"
        }
    } else {
        Write-Host "✓ No issues found" -ForegroundColor Green
    }
}

function New-ModulePackage {
    Write-BuildHeader "Creating Module Package"
    
    $version = (Import-PowerShellDataFile "$PSScriptRoot\..\$ModuleName.psd1").ModuleVersion
    $packageName = "$ModuleName-v$version.zip"
    $packagePath = Join-Path $OutputPath $packageName
    
    # Create zip file
    Compress-Archive -Path (Join-Path $OutputPath $ModuleName) -DestinationPath $packagePath -Force
    
    # Generate file hash
    $hash = Get-FileHash -Path $packagePath -Algorithm SHA256
    @{
        Package = $packageName
        Version = $version
        SHA256 = $hash.Hash
        Size = (Get-Item $packagePath).Length
        Date = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
    } | ConvertTo-Json | Set-Content -Path "$packagePath.json"
    
    Write-Host "✓ Package created: $packagePath" -ForegroundColor Green
}

function Invoke-Clean {
    Write-BuildHeader "Cleaning Output"
    
    if (Test-Path $OutputPath) {
        Remove-Item $OutputPath -Recurse -Force
        Write-Host "✓ Output directory cleaned" -ForegroundColor Green
    }
}

# Main execution
try {
    $ErrorActionPreference = 'Stop'
    
    switch ($Task) {
        'Clean' {
            Invoke-Clean
        }
        'Build' {
            Test-ModuleDependencies
            Invoke-Build
        }
        'Test' {
            Invoke-Tests
        }
        'Analyze' {
            Invoke-ScriptAnalyzer
        }
        'Package' {
            Invoke-Build
            Invoke-Tests
            Invoke-ScriptAnalyzer
            New-ModulePackage
        }
        'Default' {
            Test-ModuleDependencies
            Invoke-Build
            Invoke-Tests
            Invoke-ScriptAnalyzer
        }
    }
    
    Write-Host "`n✓ Build completed successfully!" -ForegroundColor Green
}
catch {
    Write-Host "`n✗ Build failed: $_" -ForegroundColor Red
    exit 1
}