// Main JavaScript file for HNS

document.addEventListener('DOMContentLoaded', function() {
    // Initialize tooltips
    const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });

    // Copy to clipboard functionality
    document.querySelectorAll('.copy-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const targetId = this.getAttribute('data-copy-target');
            const targetElement = document.getElementById(targetId);
            
            if (targetElement) {
                // Create a temporary textarea element
                const textarea = document.createElement('textarea');
                textarea.value = targetElement.textContent;
                document.body.appendChild(textarea);
                
                // Select and copy the text
                textarea.select();
                document.execCommand('copy');
                
                // Remove the textarea
                document.body.removeChild(textarea);
                
                // Show copied notification
                const originalText = this.innerHTML;
                this.innerHTML = '<i class="fas fa-check"></i> Copied!';
                
                setTimeout(() => {
                    this.innerHTML = originalText;
                }, 2000);
            }
        });
    });

    // Hostname generator form handling
    const generatorForm = document.getElementById('hostname-generator-form');
    if (generatorForm) {
        generatorForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const jsonData = {};
            
            // Convert form data to JSON
            jsonData.template_id = parseInt(formData.get('template_id'));
            jsonData.params = {};
            
            // Get template parameters
            document.querySelectorAll('.template-param').forEach(param => {
                const paramName = param.getAttribute('name').replace('param_', '');
                jsonData.params[paramName] = param.value;
            });
            
            // Show loading indicator
            document.getElementById('generator-result').innerHTML = '<div class="loader"></div>';
            
            // Make the API request
            fetch('/api/hostnames/generate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + localStorage.getItem('token')
                },
                body: JSON.stringify(jsonData)
            })
            .then(response => response.json())
            .then(data => {
                let resultHtml = '';
                
                if (data.error) {
                    resultHtml = `<div class="alert alert-danger">${data.error}</div>`;
                } else {
                    resultHtml = `
                        <div class="card">
                            <div class="card-header">Generated Hostname</div>
                            <div class="card-body">
                                <h3 class="text-center mb-3">${data.hostname}</h3>
                                <div class="row">
                                    <div class="col-md-6">
                                        <p><strong>Template:</strong> ${document.querySelector('#template_id option:checked').textContent}</p>
                                        <p><strong>Sequence Number:</strong> <span class="sequence-num">${data.sequence_num}</span></p>
                                    </div>
                                    <div class="col-md-6">
                                        <div class="d-grid gap-2">
                                            <button id="reserve-btn" class="btn btn-primary" data-hostname="${data.hostname}" data-template-id="${data.template_id}" data-sequence-num="${data.sequence_num}">
                                                <i class="fas fa-bookmark"></i> Reserve Hostname
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    `;
                    
                    // If DNS check result exists
                    if (data.dns_check) {
                        let dnsStatusClass = data.dns_check.exists ? 'dns-warning' : 'dns-success';
                        let dnsStatusIcon = data.dns_check.exists ? 'fa-exclamation-triangle' : 'fa-check-circle';
                        let dnsStatusText = data.dns_check.exists ? 'Already exists in DNS' : 'Available in DNS';
                        
                        resultHtml += `
                            <div class="card mt-3">
                                <div class="card-header">DNS Check</div>
                                <div class="card-body">
                                    <div class="text-center ${dnsStatusClass}">
                                        <i class="fas ${dnsStatusIcon} fa-2x mb-2"></i>
                                        <h4>${dnsStatusText}</h4>
                                        ${data.dns_check.exists ? `<p>IP Address: ${data.dns_check.ip_address || 'N/A'}</p>` : ''}
                                        <p>Verified at: ${new Date(data.dns_check.verified_at).toLocaleString()}</p>
                                    </div>
                                </div>
                            </div>
                        `;
                    }
                }
                
                document.getElementById('generator-result').innerHTML = resultHtml;
                
                // Add event listener to reserve button
                const reserveBtn = document.getElementById('reserve-btn');
                if (reserveBtn) {
                    reserveBtn.addEventListener('click', reserveHostname);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                document.getElementById('generator-result').innerHTML = `
                    <div class="alert alert-danger">
                        Failed to generate hostname. Please try again.
                    </div>
                `;
            });
        });
    }

    // Template form dynamic groups
    const addGroupBtn = document.getElementById('add-group-btn');
    if (addGroupBtn) {
        let groupCounter = document.querySelectorAll('.template-group').length;
        
        addGroupBtn.addEventListener('click', function() {
            groupCounter++;
            
            const newGroup = document.createElement('div');
            newGroup.className = 'template-group';
            newGroup.innerHTML = `
                <div class="row">
                    <div class="col-md-6 mb-3">
                        <label for="group_name_${groupCounter}" class="form-label form-required">Group Name</label>
                        <input type="text" class="form-control" id="group_name_${groupCounter}" name="groups[${groupCounter-1}][name]" required>
                    </div>
                    <div class="col-md-3 mb-3">
                        <label for="group_length_${groupCounter}" class="form-label form-required">Length</label>
                        <input type="number" class="form-control" id="group_length_${groupCounter}" name="groups[${groupCounter-1}][length]" min="1" required>
                    </div>
                    <div class="col-md-3 mb-3">
                        <label for="group_required_${groupCounter}" class="form-label">Required</label>
                        <div class="form-check mt-2">
                            <input class="form-check-input" type="checkbox" id="group_required_${groupCounter}" name="groups[${groupCounter-1}][is_required]" checked>
                            <label class="form-check-label" for="group_required_${groupCounter}">
                                Required
                            </label>
                        </div>
                    </div>
                </div>
                <div class="row">
                    <div class="col-md-6 mb-3">
                        <label for="group_validation_type_${groupCounter}" class="form-label form-required">Validation Type</label>
                        <select class="form-select validation-type-select" id="group_validation_type_${groupCounter}" name="groups[${groupCounter-1}][validation_type]" required>
                            <option value="fixed">Fixed Value</option>
                            <option value="list">List of Values</option>
                            <option value="regex">Regular Expression</option>
                            <option value="sequence">Sequence Number</option>
                        </select>
                    </div>
                    <div class="col-md-6 mb-3 validation-value-container">
                        <label for="group_validation_value_${groupCounter}" class="form-label validation-value-label form-required">Value</label>
                        <input type="text" class="form-control" id="group_validation_value_${groupCounter}" name="groups[${groupCounter-1}][validation_value]" required>
                        <small class="validation-info text-muted">
                            For fixed values, enter the exact value. For lists, enter comma-separated values. For regex, enter a valid regex pattern.
                        </small>
                    </div>
                </div>
                <button type="button" class="btn btn-sm btn-outline-danger remove-group-btn">
                    <i class="fas fa-trash"></i> Remove Group
                </button>
                <hr>
            `;
            
            document.getElementById('template-groups-container').appendChild(newGroup);
            
            // Add event listener to remove button
            newGroup.querySelector('.remove-group-btn').addEventListener('click', function() {
                newGroup.remove();
            });
            
            // Add event listener to validation type select
            newGroup.querySelector('.validation-type-select').addEventListener('change', handleValidationTypeChange);
        });
    }
    
    // Handle existing validation type selects
    document.querySelectorAll('.validation-type-select').forEach(select => {
        select.addEventListener('change', handleValidationTypeChange);
    });
    
    // Handle existing remove group buttons
    document.querySelectorAll('.remove-group-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            this.closest('.template-group').remove();
        });
    });

    // Template select change event
    const templateSelect = document.getElementById('template_id');
    if (templateSelect) {
        templateSelect.addEventListener('change', function() {
            const templateId = this.value;
            
            if (templateId) {
                // Show loading indicator
                document.getElementById('template-params-container').innerHTML = '<div class="loader"></div>';
                
                // Fetch template details
                fetch(`/api/templates/${templateId}`, {
                    headers: {
                        'Authorization': 'Bearer ' + localStorage.getItem('token')
                    }
                })
                .then(response => response.json())
                .then(data => {
                    if (data.error) {
                        document.getElementById('template-params-container').innerHTML = `
                            <div class="alert alert-danger">${data.error}</div>
                        `;
                    } else {
                        let paramsHtml = '';
                        
                        // Generate form fields for each group
                        data.groups.forEach(group => {
                            // Skip sequence groups
                            if (group.validation_type === 'sequence') {
                                return;
                            }
                            
                            let inputType = 'text';
                            let inputHtml = '';
                            
                            if (group.validation_type === 'list') {
                                // Create select dropdown for list type
                                const options = group.validation_value.split(',').map(value => {
                                    value = value.trim();
                                    return `<option value="${value}">${value}</option>`;
                                }).join('');
                                
                                inputHtml = `
                                    <select class="form-select template-param" id="param_${group.name}" name="param_${group.name}" ${group.is_required ? 'required' : ''}>
                                        <option value="">Select ${group.name}</option>
                                        ${options}
                                    </select>
                                `;
                            } else {
                                // Create text input for other types
                                inputHtml = `
                                    <input type="${inputType}" class="form-control template-param" id="param_${group.name}" 
                                        name="param_${group.name}" maxlength="${group.length}" ${group.is_required ? 'required' : ''}>
                                `;
                            }
                            
                            paramsHtml += `
                                <div class="mb-3">
                                    <label for="param_${group.name}" class="form-label ${group.is_required ? 'form-required' : ''}">
                                        ${group.name.charAt(0).toUpperCase() + group.name.slice(1)} (${group.length} chars)
                                    </label>
                                    ${inputHtml}
                                </div>
                            `;
                        });
                        
                        document.getElementById('template-params-container').innerHTML = paramsHtml;
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    document.getElementById('template-params-container').innerHTML = `
                        <div class="alert alert-danger">
                            Failed to fetch template details. Please try again.
                        </div>
                    `;
                });
            } else {
                // Clear params if no template is selected
                document.getElementById('template-params-container').innerHTML = '';
            }
        });
    }

    // DNS scan form
    const dnsScanForm = document.getElementById('dns-scan-form');
    if (dnsScanForm) {
        dnsScanForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const jsonData = {
                template_id: parseInt(formData.get('template_id')),
                start_seq: parseInt(formData.get('start_seq')),
                end_seq: parseInt(formData.get('end_seq')),
                max_concurrent: parseInt(formData.get('max_concurrent') || 10),
                params: {}
            };
            
            // Get template parameters
            document.querySelectorAll('.scan-param').forEach(param => {
                const paramName = param.getAttribute('name').replace('param_', '');
                jsonData.params[paramName] = param.value;
            });
            
            // Show loading indicator
            document.getElementById('scan-results-container').innerHTML = '<div class="loader"></div>';
            
            // Make the API request
            fetch('/api/dns/scan', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + localStorage.getItem('token')
                },
                body: JSON.stringify(jsonData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.error) {
                    document.getElementById('scan-results-container').innerHTML = `
                        <div class="alert alert-danger">${data.error}</div>
                    `;
                } else {
                    // Build results table
                    let resultsHtml = `
                        <div class="card">
                            <div class="card-header">Scan Results</div>
                            <div class="card-body">
                                <div class="row mb-4">
                                    <div class="col-md-4">
                                        <div class="card stats-card">
                                            <div class="card-body">
                                                <h5 class="card-title">Total Hostnames</h5>
                                                <div class="stats-value">${data.total_hostnames}</div>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="col-md-4">
                                        <div class="card stats-card">
                                            <div class="card-body">
                                                <h5 class="card-title">Existing</h5>
                                                <div class="stats-value text-warning">${data.existing_hostnames}</div>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="col-md-4">
                                        <div class="card stats-card">
                                            <div class="card-body">
                                                <h5 class="card-title">Available</h5>
                                                <div class="stats-value text-success">${data.total_hostnames - data.existing_hostnames}</div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                
                                <div class="table-responsive">
                                    <table class="table table-striped table-hover">
                                        <thead>
                                            <tr>
                                                <th>Hostname</th>
                                                <th>Status</th>
                                                <th>IP Address</th>
                                                <th>Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                    `;
                    
                    data.results.forEach(result => {
                        const statusClass = result.exists ? 'dns-result-exists' : 'dns-result-available';
                        const statusText = result.exists ? 'Exists' : 'Available';
                        const statusBadgeClass = result.exists ? 'bg-warning' : 'bg-success';
                        
                        resultsHtml += `
                            <tr class="${statusClass}">
                                <td>${result.hostname}</td>
                                <td><span class="badge ${statusBadgeClass}">${statusText}</span></td>
                                <td>${result.ip_address || 'N/A'}</td>
                                <td>
                                    ${!result.exists ? `
                                        <button class="btn btn-sm btn-primary reserve-from-scan-btn" 
                                            data-hostname="${result.hostname}" 
                                            data-template-id="${jsonData.template_id}">
                                            <i class="fas fa-bookmark"></i> Reserve
                                        </button>
                                    ` : ''}
                                </td>
                            </tr>
                        `;
                    });
                    
                    resultsHtml += `
                                        </tbody>
                                    </table>
                                </div>
                                <div class="text-muted mt-2">
                                    Scan completed in ${data.scan_duration}
                                </div>
                            </div>
                        </div>
                    `;
                    
                    document.getElementById('scan-results-container').innerHTML = resultsHtml;
                    
                    // Add event listeners to reserve buttons
                    document.querySelectorAll('.reserve-from-scan-btn').forEach(btn => {
                        btn.addEventListener('click', function() {
                            const hostname = this.getAttribute('data-hostname');
                            const templateId = this.getAttribute('data-template-id');
                            
                            // Implement hostname reservation logic
                            // This will depend on your API structure
                        });
                    });
                }
            })
            .catch(error => {
                console.error('Error:', error);
                document.getElementById('scan-results-container').innerHTML = `
                    <div class="alert alert-danger">
                        Failed to scan DNS. Please try again.
                    </div>
                `;
            });
        });
    }
});

// Helper function to handle validation type change
function handleValidationTypeChange() {
    const validationType = this.value;
    const container = this.closest('.template-group').querySelector('.validation-value-container');
    const label = container.querySelector('.validation-value-label');
    const input = container.querySelector('input');
    const info = container.querySelector('.validation-info');
    
    if (validationType === 'sequence') {
        // Hide validation value for sequence type
        container.style.display = 'none';
        input.removeAttribute('required');
    } else {
        container.style.display = 'block';
        input.setAttribute('required', 'required');
        
        // Update label and info text based on validation type
        if (validationType === 'fixed') {
            label.textContent = 'Fixed Value';
            info.textContent = 'Enter the exact value for this group.';
        } else if (validationType === 'list') {
            label.textContent = 'Allowed Values';
            info.textContent = 'Enter comma-separated list of allowed values (e.g., US,UK,DE,FR).';
        } else if (validationType === 'regex') {
            label.textContent = 'Regex Pattern';
            info.textContent = 'Enter a valid regular expression pattern.';
        }
    }
}

// Function to reserve a hostname
function reserveHostname() {
    const hostname = this.getAttribute('data-hostname');
    const templateId = this.getAttribute('data-template-id');
    
    const jsonData = {
        template_id: parseInt(templateId),
        params: {} // We would need to extract params if needed
    };
    
    // Disable the button and show loading state
    this.disabled = true;
    this.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Reserving...';
    
    // Make the API request
    fetch('/api/hostnames/reserve', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        },
        body: JSON.stringify(jsonData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.error) {
            alert('Failed to reserve hostname: ' + data.error);
            // Reset button
            this.disabled = false;
            this.innerHTML = '<i class="fas fa-bookmark"></i> Reserve Hostname';
        } else {
            // Show success message and redirect to hostnames page
            alert('Hostname reserved successfully!');
            window.location.href = '/hostnames';
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert('Failed to reserve hostname. Please try again.');
        // Reset button
        this.disabled = false;
        this.innerHTML = '<i class="fas fa-bookmark"></i> Reserve Hostname';
    });
}