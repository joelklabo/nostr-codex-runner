const projectSelect = document.getElementById('project');
const epicTitle = document.getElementById('epicTitle');
const epicDesc = document.getElementById('epicDesc');
const issueType = document.getElementById('issueType');
const issueTitle = document.getElementById('issueTitle');
const issueDesc = document.getElementById('issueDesc');
const issueLabels = document.getElementById('issueLabels');
const issueParent = document.getElementById('issueParent');
const issuePriority = document.getElementById('issuePriority');
const editIssue = document.getElementById('editIssue');
const editStatus = document.getElementById('editStatus');
const editTitle = document.getElementById('editTitle');
const editDesc = document.getElementById('editDesc');
const editPriority = document.getElementById('editPriority');
const editAddLabels = document.getElementById('editAddLabels');
const editRemoveLabels = document.getElementById('editRemoveLabels');
const statusFilter = document.getElementById('statusFilter');
const editMeta = document.getElementById('editMeta');
const toast = document.getElementById('toast');

async function api(path, opts = {}) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  if (res.status === 204) return null;
  return res.json();
}

function currentProject() {
  return projectSelect.value;
}

function showToast(msg, ok = true) {
  toast.textContent = msg;
  toast.className = 'toast show';
  toast.style.borderColor = ok ? '#59f6ff' : '#ff8b8b';
  setTimeout(() => { toast.className = 'toast hidden'; }, 2500);
}

async function loadProjects() {
  const projects = await api('/api/projects');
  projectSelect.innerHTML = '';
  projects.forEach(p => {
    const opt = document.createElement('option');
    opt.value = p.id;
    opt.textContent = `${p.name} (${p.path})`;
    projectSelect.appendChild(opt);
  });
}

async function loadEpics() {
  const pid = currentProject();
  const epics = await api(`/api/projects/${pid}/issues?type=epic&status=open&limit=100`);
  issueParent.innerHTML = '<option value="">None</option>';
  epics.forEach(e => {
    const opt = document.createElement('option');
    opt.value = e.id;
    opt.textContent = `${e.id} — ${e.title}`;
    issueParent.appendChild(opt);
  });
}

async function loadIssues() {
  const pid = currentProject();
  const status = statusFilter.value;
  const qs = status ? `?status=${encodeURIComponent(status)}&limit=200` : '?limit=200';
  const issues = await api(`/api/projects/${pid}/issues${qs}`);
  editIssue.innerHTML = '';
  issues.forEach(i => {
    const opt = document.createElement('option');
    opt.value = i.id;
    opt.textContent = `${i.id} — ${i.title}`;
    editIssue.appendChild(opt);
  });
  if (issues.length) {
    populateIssueDetails(issues[0]);
  } else {
    editMeta.textContent = 'No open issues.';
  }
}

async function populateIssueDetails(issue) {
  editMeta.textContent = `${issue.issue_type} • ${issue.status} • priority P${issue.priority}`;
  editTitle.value = '';
  editDesc.value = '';
  editStatus.value = '';
  editPriority.value = '';
}

async function loadIssueDetails(id) {
  const pid = currentProject();
  const res = await api(`/api/projects/${pid}/issues/${id}`);
  const issue = Array.isArray(res) ? res[0] : res;
  if (issue) populateIssueDetails(issue);
}

async function createEpic() {
  const pid = currentProject();
  await api(`/api/projects/${pid}/issues`, {
    method: 'POST',
    body: JSON.stringify({
      type: 'epic',
      title: epicTitle.value,
      description: epicDesc.value,
    }),
  });
  showToast('Epic created');
  epicTitle.value = '';
  epicDesc.value = '';
  await loadEpics();
  await loadIssues();
}

async function createIssue() {
  const pid = currentProject();
  await api(`/api/projects/${pid}/issues`, {
    method: 'POST',
    body: JSON.stringify({
      type: issueType.value,
      title: issueTitle.value,
      description: issueDesc.value,
      parent: issueParent.value,
      priority: issuePriority.value,
      labels: splitCSV(issueLabels.value),
    }),
  });
  showToast('Issue created');
  issueTitle.value = '';
  issueDesc.value = '';
  issueLabels.value = '';
  await loadIssues();
}

async function updateIssue() {
  const pid = currentProject();
  const id = editIssue.value;
  await api(`/api/projects/${pid}/issues/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({
      title: editTitle.value || undefined,
      description: editDesc.value || undefined,
      status: editStatus.value || undefined,
      priority: editPriority.value || undefined,
      addLabels: splitCSV(editAddLabels.value),
      removeLabels: splitCSV(editRemoveLabels.value),
    }),
  });
  showToast('Issue updated');
  await loadIssues();
  if (id) await loadIssueDetails(id);
}

projectSelect.addEventListener('change', async () => {
  await loadEpics();
  await loadIssues();
});

document.getElementById('createEpic').onclick = () => createEpic().catch(err => showToast(err.message, false));
document.getElementById('createIssue').onclick = () => createIssue().catch(err => showToast(err.message, false));
document.getElementById('updateIssue').onclick = () => updateIssue().catch(err => showToast(err.message, false));
document.getElementById('reloadIssues').onclick = () => Promise.all([loadEpics(), loadIssues()]).catch(err => showToast(err.message, false));

editIssue.addEventListener('change', (e) => loadIssueDetails(e.target.value).catch(err => showToast(err.message, false)));
statusFilter.addEventListener('change', () => loadIssues().catch(err => showToast(err.message, false)));

function splitCSV(val) {
  return val
    .split(',')
    .map(s => s.trim())
    .filter(Boolean);
}

(async function init() {
  try {
    await loadProjects();
    await loadEpics();
    await loadIssues();
  } catch (e) {
    showToast(e.message, false);
  }
})();
