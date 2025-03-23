# HNS-DeleteTemplate.ps1
# Standalone script to delete a specific template by ID

param(
    [Parameter(Mandatory = $true)]
    [int]$TemplateId,
    
    [Parameter(Mandatory = $false)]
    [string]$HnsUrl = "http://localhost:8080",
    
    [Parameter(Mandatory = $false)]
    [string]$Username = "test",
    
    [Parameter(Mandatory = $false)]
    [string]$Password = "Logon123!",
    
    [Parameter(Mandatory = $false)]
    [switch]$Force = $false
)

# Authentication and get JWT token
Write-Host "Authenticating with HNS server at $HnsUrl..." -ForegroundColor Cyan

$loginBody = @{
    username = $Username
    password = $Password
} | ConvertTo-Json

try {
    $authResponse = Invoke-RestMethod -Uri "$HnsUrl/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $jwtToken = $authResponse.token
    
    Write-Host "Authentication successful." -ForegroundColor Green
    
    # Set up headers with token
    $headers = @{
        "Content-Type" = "application/json"
        "Accept" = "application/json"
        "Authorization" = "Bearer $jwtToken"
    }
    
    # Confirm deletion if not forced
    if (-not $Force) {
        # Get template info first
        try {
            $template = Invoke-RestMethod -Uri "$HnsUrl/api/templates/$TemplateId" -Method Get -Headers $headers
            Write-Host "Found template: $($template.name) (ID: $TemplateId)" -ForegroundColor Green
            
            $confirmation = Read-Host "Are you sure you want to delete this template? (y/n)"
            if ($confirmation -ne 'y') {
                Write-Host "Operation cancelled." -ForegroundColor Yellow
                return
            }
        }
        catch {
            Write-Host "Template with ID $TemplateId not found or cannot be accessed." -ForegroundColor Red
            return
        }
    }
    
    # Delete the template
    try {
        # First, try to make it inactive rather than delete
        $templateUpdateBody = @{
            is_active = $false
        } | ConvertTo-Json
        
        try {
            Invoke-RestMethod -Uri "$HnsUrl/api/templates/$TemplateId" -Method Put -Headers $headers -Body $templateUpdateBody
            Write-Host "Template deactivated successfully." -ForegroundColor Green
        }
        catch {
            Write-Host "Could not deactivate template. Attempting direct deletion..." -ForegroundColor Yellow
        }
        
        # Try direct DELETE request
        $response = Invoke-RestMethod -Uri "$HnsUrl/api/templates/$TemplateId" -Method Delete -Headers $headers
        Write-Host "Template with ID $TemplateId deleted successfully." -ForegroundColor Green
    }
    catch {
        $errorDetails = $_.Exception.Response
        if ($errorDetails) {
            try {
                $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
                $reader.BaseStream.Position = 0
                $reader.DiscardBufferedData()
                $responseBody = $reader.ReadToEnd()
                Write-Host "Failed to delete template. Error: $responseBody" -ForegroundColor Red
            }
            catch {
                Write-Host "Failed to delete template. Error: $($_.Exception.Message)" -ForegroundColor Red
            }
        }
        else {
            Write-Host "Failed to delete template. Error: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}
catch {
    Write-Host "Authentication failed: $($_.Exception.Message)" -ForegroundColor Red
}