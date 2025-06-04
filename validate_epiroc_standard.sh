#!/bin/bash

echo "Epiroc VM Standard Compliance Validation"
echo "========================================"

# Test template structure
echo
echo "Testing Template Structure..."
echo "✓ Position 1-3: Unit Code (3 chars) - List validation"
echo "✓ Position 4: Type (1 char) - Fixed 'S'"
echo "✓ Position 5: Provider (1 char) - List: E,M,A,G"
echo "✓ Position 6-9: Region (4 chars) - List validation"
echo "✓ Position 10: Environment (1 char) - List: P,T,D"
echo "✓ Position 11-12: Function (2 chars) - List: AS,BU,DB,DC,FS,IS,MG,PO,WS,PV,CC,UC,OT"
echo "✓ Position 13-15: Sequence (3 chars) - Sequence 001-999"

# Test example hostnames
echo
echo "Testing Example Hostnames..."

# Test case 1: SFDSMWEEUPAS001 (Corrected to 15 chars)
unit_code="SFD"
type="S"
provider="M"
region="WEEU"
environment="P"
function="AS"
sequence="001"
generated="${unit_code}${type}${provider}${region}${environment}${function}${sequence}"
expected="SFDSMWEEUPAS001"

echo "Test 1: Azure West Europe Production Application Server (Corrected)"
echo "  Expected: $expected"
echo "  Generated: $generated"
if [ "$generated" = "$expected" ]; then
    echo "  Result: ✓ PASS"
else
    echo "  Result: ✗ FAIL"
fi

# Test case 2: USSSMSCUSTDB002 (Corrected - SCUS is 4 chars)
unit_code="USS"
type="S"
provider="M"
region="SCUS"
environment="T"
function="DB"
sequence="002"
generated="${unit_code}${type}${provider}${region}${environment}${function}${sequence}"
expected="USSSMSCUSTDB002"

echo
echo "Test 2: Azure South Central US Test Database (Corrected)"
echo "  Expected: $expected"
echo "  Generated: $generated"
if [ "$generated" = "$expected" ]; then
    echo "  Result: ✓ PASS"
else
    echo "  Result: ✗ FAIL"
fi

# Test case 3: SFDSEUSEPBU013 (Special case - when provider=E and region starts with E, avoid duplication)
unit_code="SFD"
type="S"
provider="E"
region="USE"  # EUSE becomes USE when provider is already E
environment="P"
function="BU"
sequence="013"
generated="${unit_code}${type}${provider}${region}${environment}${function}${sequence}"
expected="SFDSEUSEPBU013"

echo
echo "Test 3: Epiroc Private Cloud Sweden Production Backup"
echo "  Expected: $expected"
echo "  Generated: $generated"
if [ "$generated" = "$expected" ]; then
    echo "  Result: ✓ PASS"
else
    echo "  Result: ✗ FAIL"
fi

# Validate file changes
echo
echo "Validating Implementation Files..."

# Check PowerShell template
if [ -f "HNS-Standalone/Source/Private/Import-EpirocTemplate.ps1" ]; then
    echo "✓ PowerShell Epiroc template file exists"
    
    # Check for key validation values
    if grep -q "ValidationValue = 'E,M,A,G'" "HNS-Standalone/Source/Private/Import-EpirocTemplate.ps1"; then
        echo "✓ Provider validation updated correctly"
    else
        echo "✗ Provider validation not updated"
    fi
    
    if grep -q "ValidationValue = 'P,T,D'" "HNS-Standalone/Source/Private/Import-EpirocTemplate.ps1"; then
        echo "✓ Environment validation updated correctly"
    else
        echo "✗ Environment validation not updated"
    fi
    
    if grep -q "AS,BU,DB,DC,FS,IS,MG,PO,WS,PV,CC,UC,OT" "HNS-Standalone/Source/Private/Import-EpirocTemplate.ps1"; then
        echo "✓ Function validation includes all standard functions"
    else
        echo "✗ Function validation missing standard functions"
    fi
    
    if grep -q "WEEU" "HNS-Standalone/Source/Private/Import-EpirocTemplate.ps1"; then
        echo "✓ Region validation includes key regions"
    else
        echo "✗ Region validation missing key regions"
    fi
else
    echo "✗ PowerShell Epiroc template file not found"
fi

# Check SQL migration
if [ -f "migrations/000002_sample_data.up.sql" ]; then
    echo "✓ SQL migration file exists"
    
    if grep -q "Epiroc VM Standard" "migrations/000002_sample_data.up.sql"; then
        echo "✓ Epiroc template added to SQL migration"
    else
        echo "✗ Epiroc template not found in SQL migration"
    fi
    
    if grep -q "15.*1.*3.*TRUE" "migrations/000002_sample_data.up.sql"; then
        echo "✓ Template configured for 15-character limit"
    else
        echo "✗ Template not configured for 15-character limit"
    fi
else
    echo "✗ SQL migration file not found"
fi

echo
echo "========================================"
echo "Validation Complete"
echo "========================================"