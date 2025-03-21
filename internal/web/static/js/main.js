// Main JavaScript for HNS application

document.addEventListener('DOMContentLoaded', function() {
    // Initialize tooltips
    const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]');
    if (tooltipTriggerList.length) {
        [...tooltipTriggerList].map(el => new bootstrap.Tooltip(el));
    }

    // Initialize popovers
    const popoverTriggerList = document.querySelectorAll('[data-bs-toggle="popover"]');
    if (popoverTriggerList.length) {
        [...popoverTriggerList].map(el => new bootstrap.Popover(el));
    }

    // Copy to clipboard functionality
    document.querySelectorAll('.copy-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const targetId = this.getAttribute('data-copy-target');
            const targetElement = document.getElementById(targetId);
            
            if (targetElement) {
                const textToCopy = targetElement.textContent.trim();
                
                // Use Clipboard API if available
                if (navigator.clipboard && window.isSecureContext) {
                    navigator.clipboard.writeText(textToCopy).then(() => {
                        showCopySuccess(this);
                    }).catch(err => {
                        console.error('Failed to copy text: ', err);
                    });
                } else {
                    // Fallback method for older browsers
                    const textarea = document.createElement('textarea');
                    textarea.value = textToCopy;
                    textarea.style.position = 'fixed';  // Avoid scrolling to bottom
                    document.body.appendChild(textarea);
                    textarea.select();
                    
                    try {
                        document.execCommand('copy');
                        showCopySuccess(this);
                    } catch (err) {
                        console.error('Failed to copy text: ', err);
                    } finally {
                        document.body.removeChild(textarea);
                    }
                }
            }
        });
    });

    // Helper function to show copy success
    function showCopySuccess(button) {
        const originalContent = button.innerHTML;
        button.innerHTML = '<i class="fas fa-check"></i> Copied!';
        button.disabled = true;
        
        setTimeout(() => {
            button.innerHTML = originalContent;
            button.disabled = false;
        }, 2000);
    }

    // Handle authentication token storage
    // When the user successfully logs in, store the JWT token
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            // Form submission is handled by the server
            // This is just for demonstration purposes
            
            // Store token from API response (normally done after successful login)
            // localStorage.setItem('auth_token', response.token);
        });
    }

    // Add token to API requests
    function getAuthToken() {
        return localStorage.getItem('auth_token');
    }

    // Example API request with authentication
    async function apiRequest(url, method = 'GET', data = null) {
        const headers = {
            'Content-Type': 'application/json'
        };
        
        const token = getAuthToken();
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }
        
        const options = {
            method,
            headers
        };
        
        if (data && (method === 'POST' || method === 'PUT')) {
            options.body = JSON.stringify(data);
        }
        
        try {
            const response = await fetch(url, options);
            const result = await response.json();
            
            if (!response.ok) {
                throw new Error(result.error || 'An error occurred');
            }
            
            return result;
        } catch (error) {
            console.error('API request error:', error);
            throw error;
        }
    }

    // Expose API functions globally
    window.hnsApi = {
        getTemplates: () => apiRequest('/api/templates'),
        getTemplate: (id) => apiRequest(`/api/templates/${id}`),
        getHostnames: (filters) => {
            const queryParams = new URLSearchParams(filters).toString();
            return apiRequest(`/api/hostnames?${queryParams}`);
        },
        generateHostname: (data) => apiRequest('/api/hostnames/generate', 'POST', data),
        reserveHostname: (data) => apiRequest('/api/hostnames/reserve', 'POST', data),
        checkDns: (hostname) => apiRequest(`/api/dns/check/${hostname}`)
    };

    // Template dynamic form handlers
    initTemplateForms();
    
    // Hostname generator and reservation form handlers
    initHostnameForms();
});

// Initialize template-related forms
function initTemplateForms() {
    // Add group button for template forms
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
                        <input type="text" class="form-control" id="group_name_${groupCounter}" 
                            name="groups[${groupCounter-1}][name]" required>
                    </div>
                    <div class="col-md-3 mb-3">
                        <label for="group_length_${groupCounter}" class="form-label form-required">Length</label>
                        <input type="number" class="form-control" id="group_length_${groupCounter}" 
                            name="groups[${groupCounter-1}][length]" min="1" value="2" required>
                    </div>
                    <div class="col-md-3 mb-3">
                        <label for="group_required_${groupCounter}" class="form-label">Required</label>
                        <div class="form-check mt-2">
                            <input class="form-check-input" type="checkbox" id="group_required_${groupCounter}" 
                                name="groups[${groupCounter-1}][is_required]" checked>
                            <label class="form-check-label" for="group_required_${groupCounter}">
                                Required
                            </label>
                        </div>
                    </div>
                </div>
                <div class="row">
                    <div class="col-md-6 mb-3">
                        <label for="group_validation_type_${groupCounter}" class="form-label form-required">Validation Type</label>
                        <select class="form-select validation-type-select" id="group_validation_type_${groupCounter}" 
                            name="groups[${groupCounter-1}][validation_type]" required>
                            <option value="fixed">Fixed Value</option>
                            <option value="list" selected>List of Values</option>
                            <option value="regex">Regular Expression</option>
                            <option value="sequence">Sequence Number</option>
                        </select>
                    </div>
                    <div class="col-md-6 mb-3 validation-value-container">
                        <label for="group_validation_value_${groupCounter}" class="form-label validation-value-label form-required">Value</label>
                        <input type="text" class="form-control" id="group_validation_value_${groupCounter}" 
                            name="groups[${groupCounter-1}][validation_value]" required>
                        <small class="validation-info text-muted">
                            For list values, enter comma-separated options (e.g., US,EU,AP)
                        </small>
                    </div>
                </div>
                <button type="button" class="btn btn-sm btn-outline-danger remove-group-btn">
                    <i class="fas fa-trash"></i> Remove Group
                </button>
                <hr>
            `;
            
            document.getElementById('template-groups-container').appendChild(newGroup);
            
            // Add event listener to new elements
            newGroup.querySelector('.remove-group-btn').addEventListener('click', function() {
                newGroup.remove();
            });
            
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
}

// Handle validation type change in template forms
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

// Initialize hostname-related forms
function initHostnameForms() {
    // Template select change event for hostname generator and reservation
    const templateSelect = document.getElementById('template_id');
    if (templateSelect) {
        templateSelect.addEventListener('change', function() {
            const templateId = this.value;
            
            if (!templateId) {
                document.getElementById('template-params-container').innerHTML = '';
                return;
            }
            
            // Show loading indicator
            document.getElementById('template-params-container').innerHTML = '<div class="loader"></div>';
            
            // Fetch template details
            window.hnsApi.getTemplate(templateId)
                .then(template => {
                    let paramsHtml = '';
                    
                    // Generate form fields for each group
                    template.groups.forEach(group => {
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
                                <select class="form-select template-param" id="param_${group.name}" 
                                    name="param_${group.name}" ${group.is_required ? 'required' : ''}>
                                    <option value="">Select ${group.name}</option>
                                    ${options}
                                </select>
                            `;
                        } else if (group.validation_type === 'fixed') {
                            // Fixed value should be readonly
                            inputHtml = `
                                <input type="text" class="form-control template-param" id="param_${group.name}" 
                                    name="param_${group.name}" value="${group.validation_value}" readonly>
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
                })
                .catch(error => {
                    console.error('Error fetching template:', error);
                    document.getElementById('template-params-container').innerHTML = `
                        <div class="alert alert-danger">
                            Error loading template parameters. Please try again.
                        </div>
                    `;
                });
        });
    }
    
    // Hostname generator form
    const generatorForm = document.getElementById('hostname-generator-form');
    if (generatorForm) {
        generatorForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const templateId = document.getElementById('template_id').value;
            if (!templateId) {
                alert('Please select a template');
                return;
            }
            
            // Collect parameters
            const params = {};
            document.querySelectorAll('.template-param').forEach(input => {
                const paramName = input.id.replace('param_', '');
                params[paramName] = input.value;
            });
            
            // Show loading indicator
            document.getElementById('generator-result').innerHTML = '<div class="loader"></div>';
            
            // Check if DNS check is enabled
            const checkDns = document.getElementById('check_dns')?.checked || false;
            
            // Generate hostname
            window.hnsApi.generateHostname({
                template_id: parseInt(templateId),
                params: params,
                check_dns: checkDns
            })
            .then(data => {
                // Format result in a card
                let resultHtml = `
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
                                        <button id="reserve-btn" class="btn btn-primary">
                                            <i class="fas fa-bookmark"></i> Reserve Hostname
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                `;
                
                // Add DNS check results if available
                if (data.dns_check) {
                    const dnsStatusClass = data.dns_check.exists ? 'text-warning' : 'text-success';
                    const dnsStatusIcon = data.dns_check.exists ? 'fa-exclamation-triangle' : 'fa-check-circle';
                    const dnsStatusText = data.dns_check.exists ? 'Hostname exists in DNS' : 'Hostname available in DNS';
                    
                    resultHtml += `
                        <div class="card mt-3">
                            <div class="card-header">DNS Check Result</div>
                            <div class="card-body text-center">
                                <i class="fas ${dnsStatusIcon} fa-3x ${dnsStatusClass} mb-2"></i>
                                <h4 class="${dnsStatusClass}">${dnsStatusText}</h4>
                                ${data.dns_check.exists ? `<p>IP: ${data.dns_check.ip_address || 'N/A'}</p>` : ''}
                            </div>
                        </div>
                    `;
                }
                
                document.getElementById('generator-result').innerHTML = resultHtml;
                
                // Add event listener to reserve button
                document.getElementById('reserve-btn').addEventListener('click', function() {
                    window.hnsApi.reserveHostname({
                        template_id: parseInt(templateId),
                        params: params
                    })
                    .then(reserveData => {
                        window.location.href = `/hostnames/${reserveData.id}`;
                    })
                    .catch(err => {
                        alert('Failed to reserve hostname: ' + err.message);
                    });
                });
            })
            .catch(error => {
                console.error('Error generating hostname:', error);
                document.getElementById('generator-result').innerHTML = `
                    <div class="alert alert-danger">
                        Error generating hostname: ${error.message}
                    </div>
                `;
            });
        });
    }
}