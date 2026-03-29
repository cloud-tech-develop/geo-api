let currentCities = [];
let editingCityId = null;
let currentPage = 1;
const limit = 20;
let bulkData = [];
let bulkMetadataFields = [];

const searchInput = document.getElementById('search-input');
const countryFilter = document.getElementById('country-filter');
const stateFilter = document.getElementById('state-filter');
const cityList = document.getElementById('city-list');
const editModal = document.getElementById('edit-modal');
const metadataFields = document.getElementById('metadata-fields');

// Initial setup
async function init() {
    await fetchCountries();
    await fetchCities();
}

async function fetchCountries() {
    try {
        // Request a high limit to get all countries for the dropdown
        const response = await fetch('/countries?limit=300');
        const data = await response.json();
        const countries = data.data || data;
        countryFilter.innerHTML = '<option value="">All Countries</option>' + 
            countries.map(c => `<option value="${c.iso2}">${c.emoji} ${c.name}</option>`).join('');
    } catch (e) { console.error('Error fetching countries', e); }
}

async function fetchStates(countryIso2) {
    if (!countryIso2) {
        stateFilter.innerHTML = '<option value="">All States</option>';
        stateFilter.disabled = true;
        return;
    }
    try {
        // Request a high limit to get all states for the dropdown
        const response = await fetch(`/countries/${countryIso2}/states?limit=1000`);
        const data = await response.json();
        const states = data.data || data;
        stateFilter.innerHTML = '<option value="">All States</option>' + 
            states.map(s => `<option value="${s.id}">${s.name}</option>`).join('');
        stateFilter.disabled = false;
    } catch (e) { console.error('Error fetching states', e); }
}

async function fetchCities(search = '', page = 1) {
    try {
        const country = countryFilter.value;
        const state = stateFilter.value;
        
        let url = `/admin/cities?page=${page}&limit=${limit}`;
        if (search) url += `&search=${encodeURIComponent(search)}`;
        if (country) url += `&country=${country}`;
        if (state) url += `&state=${state}`;

        const response = await fetch(url);
        
        if (response.status === 401) return;

        const data = await response.json();
        currentCities = data.data;
        renderCities(data.data);
        renderPagination(data.total, page);
    } catch (error) {
        console.error('Error fetching cities:', error);
        cityList.innerHTML = '<div class="loading">Error loading cities. Please try again.</div>';
    }
}

function renderCities(cities) {
    if (cities.length === 0) {
        cityList.innerHTML = '<div class="loading">No cities found.</div>';
        return;
    }

    cityList.innerHTML = cities.map(city => `
        <div class="city-card">
            <div class="city-info">
                <h2>${city.name}</h2>
                <p>${city.state_code}, ${city.country_code} (ID: ${city.id})</p>
                <div class="metadata-tags">
                    ${city.metadata ? Object.keys(city.metadata).map(key => `
                        <span class="metadata-tag">${key}: ${city.metadata[key]}</span>
                    `).join('') : ''}
                </div>
            </div>
            <button class="btn btn-primary" onclick="openEditModal(${city.id})">Edit</button>
        </div>
    `).join('');
}

function renderPagination(total, page) {
    const pagination = document.getElementById('pagination');
    const totalPages = Math.ceil(total / limit);
    
    pagination.innerHTML = `
        <button class="btn btn-secondary" ${page <= 1 ? 'disabled' : ''} onclick="changePage(${page - 1})">Prev</button>
        <span style="align-self: center; font-size: 0.875rem; color: var(--text-muted);">Page ${page} of ${totalPages || 1}</span>
        <button class="btn btn-secondary" ${page >= totalPages ? 'disabled' : ''} onclick="changePage(${page + 1})">Next</button>
    `;
}

function changePage(page) {
    currentPage = page;
    fetchCities(searchInput.value, page);
}

// Event Listeners
searchInput.addEventListener('input', debounce(() => {
    currentPage = 1;
    fetchCities(searchInput.value, 1);
}, 300));

countryFilter.addEventListener('change', () => {
    currentPage = 1;
    fetchStates(countryFilter.value);
    fetchCities(searchInput.value, 1);
});

stateFilter.addEventListener('change', () => {
    currentPage = 1;
    fetchCities(searchInput.value, 1);
});

function debounce(func, wait) {
    let timeout;
    return (...args) => {
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(this, args), wait);
    };
}

function openEditModal(cityId) {
    const city = currentCities.find(c => c.id === cityId);
    if (!city) return;

    editingCityId = cityId;
    document.getElementById('modal-city-name').innerText = city.name;
    document.getElementById('modal-city-details').innerText = `${city.state_code}, ${city.country_code} (ID: ${city.id})`;
    
    metadataFields.innerHTML = '';
    if (city.metadata) {
        Object.entries(city.metadata).forEach(([key, value]) => {
            addMetadataField(key, value);
        });
    }

    editModal.classList.add('active');
}

function closeModal() {
    editModal.classList.remove('active');
    editingCityId = null;
}

function addMetadataField(key = '', value = '') {
    const row = document.createElement('div');
    row.className = 'field-row';
    row.innerHTML = `
        <input type="text" placeholder="Key" class="meta-key" value="${key}">
        <input type="text" placeholder="Value" class="meta-value" value="${value}">
        <button class="btn" style="background: #ef4444; color: white;" onclick="this.parentElement.remove()">×</button>
    `;
    metadataFields.appendChild(row);
}

async function saveMetadata() {
    const rows = metadataFields.querySelectorAll('.field-row');
    const metadata = {};
    rows.forEach(row => {
        const key = row.querySelector('.meta-key').value.trim();
        const value = row.querySelector('.meta-value').value.trim();
        if (key) {
            metadata[key] = value;
        }
    });

    try {
        const response = await fetch(`/admin/cities/${editingCityId}/metadata`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(metadata)
        });

        if (response.ok) {
            closeModal();
            fetchCities(searchInput.value, currentPage);
        } else {
            alert('Failed to save metadata');
        }
    } catch (error) {
        console.error('Error saving metadata:', error);
        alert('Error saving metadata');
    }
}

function downloadCSV() {
    const country = countryFilter.value;
    const state = stateFilter.value;
    const search = searchInput.value;
    
    let url = `/admin/cities/export`;
    const params = [];
    if (search) params.push(`search=${encodeURIComponent(search)}`);
    if (country) params.push(`country=${country}`);
    if (state) params.push(`state=${state}`);
    if (params.length > 0) url += '?' + params.join('&');
    
    window.open(url, '_blank');
}

function downloadMetadataJSON() {
    window.open('/admin/metadata/download', '_blank');
}

function handleCSVUpload(event) {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = function(e) {
        const content = e.target.result;
        parseCSV(content);
    };
    reader.readAsText(file);
}

function parseCSV(content) {
    const lines = content.split('\n').filter(line => line.trim());
    if (lines.length < 2) {
        alert('CSV file is empty or invalid');
        return;
    }

    const headers = parseCSVLine(lines[0]);
    bulkData = [];
    
    for (let i = 1; i < lines.length; i++) {
        const values = parseCSVLine(lines[i]);
        const row = {};
        headers.forEach((header, index) => {
            row[header.trim().toLowerCase()] = values[index] || '';
        });
        if (row.id) {
            bulkData.push(row);
        }
    }

    if (bulkData.length === 0) {
        alert('No valid data found in CSV. Make sure the CSV has an "id" column.');
        return;
    }

    detectMetadataFields();
    openBulkPreview();
}

function parseCSVLine(line) {
    const result = [];
    let current = '';
    let inQuotes = false;
    
    for (let i = 0; i < line.length; i++) {
        const char = line[i];
        if (char === '"') {
            if (inQuotes && line[i + 1] === '"') {
                current += '"';
                i++;
            } else {
                inQuotes = !inQuotes;
            }
        } else if (char === ',' && !inQuotes) {
            result.push(current);
            current = '';
        } else {
            current += char;
        }
    }
    result.push(current);
    return result;
}

function detectMetadataFields() {
    const excludeFields = ['id', 'name', 'country_code', 'state_code', 'latitude', 'longitude', 'countryid', 'stateid', 'state_id'];
    bulkMetadataFields = [];
    
    for (const row of bulkData) {
        for (const key of Object.keys(row)) {
            if (!excludeFields.includes(key.toLowerCase()) && row[key].trim()) {
                if (!bulkMetadataFields.includes(key)) {
                    bulkMetadataFields.push(key);
                }
            }
        }
    }
}

function openBulkPreview() {
    if (bulkData.length === 0) {
        alert('No data to preview. Please upload a CSV file first.');
        return;
    }

    const tbody = document.getElementById('bulk-preview-body');
    tbody.innerHTML = '';

    bulkData.forEach((row, index) => {
        const metadata = {};
        bulkMetadataFields.forEach(field => {
            if (row[field] && row[field].trim()) {
                metadata[field] = row[field].trim();
            }
        });

        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td style="padding: 0.5rem; border-bottom: 1px solid var(--border);">${row.id}</td>
            <td style="padding: 0.5rem; border-bottom: 1px solid var(--border);">${row.name || 'N/A'}</td>
            <td style="padding: 0.5rem; border-bottom: 1px solid var(--border); font-size: 0.75rem;">
                ${Object.keys(metadata).map(k => `<span style="background: #312e81; color: #c7d2fe; padding: 2px 4px; border-radius: 3px; margin-right: 4px;">${k}: ${metadata[k]}</span>`).join('')}
                ${Object.keys(metadata).length === 0 ? '<span style="color: var(--text-muted);">No metadata</span>' : ''}
            </td>
            <td style="padding: 0.5rem; border-bottom: 1px solid var(--border);" id="bulk-status-${index}">
                <span style="color: var(--text-muted);">Pending</span>
            </td>
        `;
        tbody.appendChild(tr);
    });

    const validCount = bulkData.filter(row => {
        return bulkMetadataFields.some(field => row[field] && row[field].trim());
    }).length;

    document.getElementById('bulk-summary').innerHTML = `
        <strong>Summary:</strong> ${validCount} of ${bulkData.length} cities will be updated.<br>
        <span style="font-size: 0.75rem; color: var(--text-muted);">
            Detected metadata fields: ${bulkMetadataFields.join(', ') || 'None'}
        </span>
    `;

    document.getElementById('bulk-modal').classList.add('active');
}

function closeBulkModal() {
    document.getElementById('bulk-modal').classList.remove('active');
    bulkData = [];
    bulkMetadataFields = [];
    document.getElementById('csv-upload').value = '';
}

async function uploadBulkMetadata() {
    const btn = document.getElementById('bulk-upload-btn');
    btn.disabled = true;
    btn.textContent = 'Uploading...';

    const updates = [];
    bulkData.forEach((row, index) => {
        const metadata = {};
        let hasMetadata = false;
        
        bulkMetadataFields.forEach(field => {
            if (row[field] && row[field].trim()) {
                metadata[field] = row[field].trim();
                hasMetadata = true;
            }
        });

        if (hasMetadata && row.id) {
            updates.push({
                id: parseInt(row.id),
                metadata: metadata
            });
            document.getElementById(`bulk-status-${index}`).innerHTML = '<span style="color: #facc15;">Updating...</span>';
        } else {
            document.getElementById(`bulk-status-${index}`).innerHTML = '<span style="color: var(--text-muted);">Skipped</span>';
        }
    });

    if (updates.length === 0) {
        alert('No valid updates to upload');
        btn.disabled = false;
        btn.textContent = 'Upload Changes';
        return;
    }

    try {
        const response = await fetch('/admin/cities/bulk-metadata', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ updates })
        });

        const result = await response.json();
        
        let successCount = 0;
        let failCount = 0;
        
        updates.forEach((update, index) => {
            const statusEl = document.getElementById(`bulk-status-${index}`);
            const wasUpdated = !result.errors || !result.errors.some(e => e.id === update.id);
            
            if (wasUpdated) {
                successCount++;
                statusEl.innerHTML = '<span class="bulk-success">Success</span>';
            } else {
                failCount++;
                const error = result.errors?.find(e => e.id === update.id);
                statusEl.innerHTML = `<span class="bulk-error" title="${error?.error || 'Error'}">Failed</span>`;
            }
        });

        document.getElementById('bulk-summary').innerHTML = `
            <strong>Upload Complete:</strong><br>
            <span style="color: #4ade80;">${successCount} updated successfully</span><br>
            ${failCount > 0 ? `<span style="color: #ef4444;">${failCount} failed</span>` : ''}
        `;

        if (successCount > 0) {
            setTimeout(() => {
                closeBulkModal();
                fetchCities(searchInput.value, currentPage);
            }, 1500);
        }
    } catch (error) {
        console.error('Error uploading bulk metadata:', error);
        alert('Failed to upload metadata');
    }

    btn.disabled = false;
    btn.textContent = 'Upload Changes';
}

init();
