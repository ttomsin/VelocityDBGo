const token = localStorage.getItem('token');
let currentProjectId = null;
let currentProjectApiKey = null;
let currentCollectionName = null;

// Initialization
document.addEventListener('DOMContentLoaded', () => {
    fetchProjects();
});

// Auth
function logout() {
    localStorage.removeItem('token');
    window.location.href = '/auth';
}

// API Helpers
async function api(url, options = {}) {
    options.headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        ...options.headers
    };
    const res = await fetch(url, options);
    const data = await res.json();
    if (!res.ok) {
        if (res.status === 401) logout();
        throw new Error(data.error || 'API Error');
    }
    return data;
}

// UI State Management
function openModal(id) { document.getElementById(id).classList.remove('hidden'); }
function closeModal(id) { document.getElementById(id).classList.add('hidden'); }

// Projects
async function fetchProjects() {
    try {
        const projects = await api('/api/projects');
        const list = document.getElementById('projects-list');
        list.innerHTML = '';
        
        if(projects.length > 0 && !currentProjectId) {
            selectProject(projects[0].id, projects[0].name, projects[0].apiKey);
        }

        projects.forEach(p => {
            const isActive = p.id === currentProjectId;
            list.innerHTML += `
                <li>
                    <button onclick="selectProject(${p.id}, '${p.name}', '${p.apiKey}')" 
                        class="w-full text-left px-4 py-2 rounded-lg text-sm font-medium transition flex items-center gap-3
                        ${isActive ? 'bg-indigo-600 text-white' : 'text-gray-300 hover:bg-gray-800 hover:text-white'}">
                        <i class="fa-solid fa-folder${isActive ? '-open' : ''}"></i> ${p.name}
                    </button>
                </li>
            `;
        });
    } catch (e) {
        alert(e.message);
    }
}

async function createProject(e) {
    e.preventDefault();
    const name = document.getElementById('proj-name').value;
    const desc = document.getElementById('proj-desc').value;
    try {
        await api('/api/projects', { method: 'POST', body: JSON.stringify({ name, description: desc }) });
        closeModal('projectModal');
        e.target.reset();
        fetchProjects();
    } catch (err) { alert(err.message); }
}

function selectProject(id, name, apiKey) {
    currentProjectId = id;
    currentProjectApiKey = apiKey;
    
    document.getElementById('header-title').innerText = name;
    document.getElementById('api-key-box').classList.remove('hidden');
    document.getElementById('api-key-display').innerText = apiKey;
    
    fetchProjects(); // Update active state in sidebar
    showCollections();
}

function copyApiKey() {
    navigator.clipboard.writeText(currentProjectApiKey);
    alert('API Key copied to clipboard!');
}

// Collections
async function showCollections() {
    document.getElementById('empty-state').classList.add('hidden');
    document.getElementById('data-view').classList.add('hidden');
    document.getElementById('collections-view').classList.remove('hidden');

    try {
        const collections = await api(`/api/projects/${currentProjectId}/collections`);
        const grid = document.getElementById('collections-grid');
        grid.innerHTML = '';
        
        if (collections.length === 0) {
            grid.innerHTML = `<div class="col-span-full text-center py-10 text-gray-400 border-2 border-dashed border-gray-200 rounded-xl">No collections yet. Create one!</div>`;
            return;
        }

        collections.forEach(c => {
            grid.innerHTML += `
                <div class="bg-white border border-gray-200 rounded-xl p-5 hover:shadow-md cursor-pointer transition flex items-center justify-between" onclick="viewData('${c.name}')">
                    <div class="flex items-center gap-3">
                        <div class="bg-indigo-50 text-indigo-600 p-3 rounded-lg"><i class="fa-solid fa-table"></i></div>
                        <div>
                            <h4 class="font-bold text-gray-800">${c.name}</h4>
                            <p class="text-xs text-gray-500">Virtual Table</p>
                        </div>
                    </div>
                    <i class="fa-solid fa-chevron-right text-gray-300"></i>
                </div>
            `;
        });
    } catch (e) { alert(e.message); }
}

async function createCollection(e) {
    e.preventDefault();
    const name = document.getElementById('coll-name').value;
    try {
        await api(`/api/projects/${currentProjectId}/collections`, { method: 'POST', body: JSON.stringify({ name }) });
        closeModal('collectionModal');
        e.target.reset();
        showCollections();
    } catch (err) { alert(err.message); }
}

// Data View
async function viewData(collectionName) {
    currentCollectionName = collectionName;
    document.getElementById('collections-view').classList.add('hidden');
    document.getElementById('data-view').classList.remove('hidden');
    document.getElementById('current-collection-name').innerText = collectionName;
    
    // Reset filters
    clearFilter();
}

function applyFilter() {
    const field = document.getElementById('filter-field').value.trim();
    const operator = document.getElementById('filter-operator').value;
    const value = document.getElementById('filter-value').value.trim();

    if (!field || !value) {
        alert("Please enter both a Field Name and a Value to filter.");
        return;
    }

    const queryStr = `?${encodeURIComponent(field)}=${encodeURIComponent(operator + value)}&sort=_id:desc`;
    fetchData(queryStr);
}

function clearFilter() {
    document.getElementById('filter-field').value = '';
    document.getElementById('filter-operator').value = 'eq:';
    document.getElementById('filter-value').value = '';
    fetchData();
}

function copyJsonQuery() {
    const field = document.getElementById('filter-field').value.trim();
    const operator = document.getElementById('filter-operator').value;
    const value = document.getElementById('filter-value').value.trim();

    const queryPayload = {
        filter: {},
        sort: "_id:desc",
        limit: 10,
        offset: 0
    };

    if (field && value) {
        queryPayload.filter[field] = operator + value;
    }

    const jsonString = JSON.stringify(queryPayload, null, 2);
    
    navigator.clipboard.writeText(jsonString).then(() => {
        alert("JSON Query Payload copied to clipboard!\n\n" + jsonString);
    }).catch(err => {
        alert("Failed to copy JSON query. Please check your browser permissions.");
        console.error(err);
    });
}

async function fetchData(queryStr = '?sort=_id:desc') {
    try {
        const data = await api(`/api/projects/${currentProjectId}/data/${currentCollectionName}${queryStr}`);
        const tbody = document.getElementById('data-table-body');
        tbody.innerHTML = '';

        if (data.length === 0) {
            tbody.innerHTML = `<tr><td colspan="3" class="text-center py-8 text-gray-400">Collection is empty or no matches found.</td></tr>`;
            return;
        }

        data.forEach(doc => {
            const id = doc._id;
            delete doc._id; // Remove from display
            const jsonStr = JSON.stringify(doc, null, 2);
            
            tbody.innerHTML += `
                <tr class="hover:bg-gray-50 transition group">
                    <td class="px-6 py-4 font-mono text-gray-500">#${id}</td>
                    <td class="px-6 py-4">
                        <pre class="bg-gray-100 p-3 rounded-lg text-xs font-mono text-gray-700 overflow-x-auto max-h-32 overflow-y-auto">${jsonStr}</pre>
                    </td>
                    <td class="px-6 py-4 text-right">
                        <button onclick="deleteDocument(${id})" class="text-red-400 hover:text-red-600 opacity-0 group-hover:opacity-100 transition" title="Delete">
                            <i class="fa-solid fa-trash"></i>
                        </button>
                    </td>
                </tr>
            `;
        });
    } catch (e) { alert(e.message); }
}

async function insertDocument(e) {
    e.preventDefault();
    const jsonStr = document.getElementById('doc-json').value;
    const errorBox = document.getElementById('json-error');
    
    try {
        const payload = JSON.parse(jsonStr);
        errorBox.classList.add('hidden');
        
        await api(`/api/projects/${currentProjectId}/data/${currentCollectionName}`, {
            method: 'POST',
            body: JSON.stringify(payload)
        });
        
        closeModal('documentModal');
        e.target.reset();
        fetchData();
    } catch (err) {
        if (err instanceof SyntaxError) {
            errorBox.classList.remove('hidden');
        } else {
            alert(err.message);
        }
    }
}

async function deleteDocument(docId) {
    if(!confirm("Are you sure you want to delete this document?")) return;
    try {
        await api(`/api/projects/${currentProjectId}/data/${currentCollectionName}/${docId}`, { method: 'DELETE' });
        fetchData();
    } catch (e) { alert(e.message); }
}

function copyAiPrompt() {
    const promptText = document.getElementById('ai-prompt-content').value;
    navigator.clipboard.writeText(promptText).then(() => {
        alert("AI Prompt copied to clipboard! Paste it into ChatGPT or Claude.");
    }).catch(err => {
        console.error('Failed to copy text: ', err);
        alert("Failed to copy text. Please select it manually.");
    });
}