let currentCities = [];
let editingCityId = null;
let currentPage = 1;
const limit = 20;

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

init();
