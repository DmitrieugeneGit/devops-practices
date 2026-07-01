const API = "/api/tasks";

let tasks = [];
let currentFilter = "all";

const form = document.getElementById("task-form");
const idInput = document.getElementById("task-id");
const titleInput = document.getElementById("title");
const descInput = document.getElementById("description");
const submitBtn = document.getElementById("submit-btn");
const cancelBtn = document.getElementById("cancel-btn");
const listEl = document.getElementById("task-list");
const emptyEl = document.getElementById("empty");
const counterEl = document.getElementById("counter");
const toastEl = document.getElementById("toast");

// --- Работа с API ---

async function apiRequest(url, options = {}) {
    const res = await fetch(url, {
        headers: { "Content-Type": "application/json" },
        ...options,
    });
    if (!res.ok) {
        let msg = `Ошибка ${res.status}`;
        try {
            const data = await res.json();
            if (data.error) msg = data.error;
        } catch (_) {}
        throw new Error(msg);
    }
    if (res.status === 204) return null;
    return res.json();
}

async function loadTasks() {
    try {
        tasks = await apiRequest(API);
        render();
    } catch (err) {
        showToast(err.message, true);
    }
}

async function saveTask(payload, id) {
    if (id) {
        return apiRequest(`${API}/${id}`, {
            method: "PUT",
            body: JSON.stringify(payload),
        });
    }
    return apiRequest(API, {
        method: "POST",
        body: JSON.stringify(payload),
    });
}

// --- Обработчики событий ---

form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const id = idInput.value;
    const payload = {
        title: titleInput.value.trim(),
        description: descInput.value.trim(),
        done: false,
    };
    if (!payload.title) return;

    // Сохраняем текущий статус done при редактировании.
    if (id) {
        const existing = tasks.find((t) => t.id == id);
        if (existing) payload.done = existing.done;
    }

    try {
        await saveTask(payload, id);
        resetForm();
        await loadTasks();
        showToast(id ? "Задача обновлена" : "Задача добавлена");
    } catch (err) {
        showToast(err.message, true);
    }
});

cancelBtn.addEventListener("click", resetForm);

document.querySelectorAll(".filter").forEach((btn) => {
    btn.addEventListener("click", () => {
        document.querySelectorAll(".filter").forEach((b) => b.classList.remove("active"));
        btn.classList.add("active");
        currentFilter = btn.dataset.filter;
        render();
    });
});

async function toggleDone(task) {
    try {
        await saveTask(
            { title: task.title, description: task.description, done: !task.done },
            task.id
        );
        await loadTasks();
    } catch (err) {
        showToast(err.message, true);
    }
}

async function deleteTask(id) {
    if (!confirm("Удалить задачу?")) return;
    try {
        await apiRequest(`${API}/${id}`, { method: "DELETE" });
        await loadTasks();
        showToast("Задача удалена");
    } catch (err) {
        showToast(err.message, true);
    }
}

function startEdit(task) {
    idInput.value = task.id;
    titleInput.value = task.title;
    descInput.value = task.description;
    submitBtn.textContent = "Сохранить";
    cancelBtn.classList.remove("hidden");
    titleInput.focus();
    window.scrollTo({ top: 0, behavior: "smooth" });
}

function resetForm() {
    form.reset();
    idInput.value = "";
    submitBtn.textContent = "Добавить";
    cancelBtn.classList.add("hidden");
}

// --- Отрисовка ---

function render() {
    const filtered = tasks.filter((t) => {
        if (currentFilter === "active") return !t.done;
        if (currentFilter === "done") return t.done;
        return true;
    });

    listEl.innerHTML = "";
    emptyEl.classList.toggle("hidden", filtered.length > 0);

    for (const task of filtered) {
        listEl.appendChild(renderTask(task));
    }

    const doneCount = tasks.filter((t) => t.done).length;
    counterEl.textContent = `Выполнено ${doneCount} из ${tasks.length}`;
}

function renderTask(task) {
    const li = document.createElement("li");
    li.className = "task" + (task.done ? " done" : "");

    const check = document.createElement("input");
    check.type = "checkbox";
    check.className = "task-check";
    check.checked = task.done;
    check.addEventListener("change", () => toggleDone(task));

    const body = document.createElement("div");
    body.className = "task-body";

    const title = document.createElement("div");
    title.className = "task-title";
    title.textContent = task.title;
    body.appendChild(title);

    if (task.description) {
        const desc = document.createElement("div");
        desc.className = "task-desc";
        desc.textContent = task.description;
        body.appendChild(desc);
    }

    const meta = document.createElement("div");
    meta.className = "task-meta";
    meta.textContent = "Создано: " + formatDate(task.created_at);
    body.appendChild(meta);

    const buttons = document.createElement("div");
    buttons.className = "task-buttons";

    const editBtn = document.createElement("button");
    editBtn.className = "icon-btn";
    editBtn.textContent = "Изм.";
    editBtn.addEventListener("click", () => startEdit(task));

    const delBtn = document.createElement("button");
    delBtn.className = "icon-btn delete";
    delBtn.textContent = "Удал.";
    delBtn.addEventListener("click", () => deleteTask(task.id));

    buttons.append(editBtn, delBtn);
    li.append(check, body, buttons);
    return li;
}

function formatDate(iso) {
    const d = new Date(iso);
    return d.toLocaleString("ru-RU", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
    });
}

let toastTimer;
function showToast(msg, isError = false) {
    clearTimeout(toastTimer);
    toastEl.textContent = msg;
    toastEl.className = "toast" + (isError ? " error" : "");
    toastTimer = setTimeout(() => toastEl.classList.add("hidden"), 3000);
}

// --- Старт ---
loadTasks();
