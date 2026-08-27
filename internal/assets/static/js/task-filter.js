const form = document.getElementById("task-form");
const clientSelect = document.getElementById("task-client-select");
const taskTypeSelect = document.getElementById("task-type-select");

const byClient = {};
form.dataset.clientTaskTypes.split(";").filter(Boolean).forEach((entry) => {
	const [clientID, ids] = entry.split(":");
	byClient[clientID] = ids ? ids.split(",") : [];
});

// The create form ships the full option list in a <template>, since its
// select is server-rendered already narrowed to the preselected client.
// The edit form has no template and renders every type, so read from the
// select itself there.
const optionSource = document.getElementById("task-type-options");
const allOptions = Array.from(
	optionSource ? optionSource.content.querySelectorAll("option") : taskTypeSelect.options,
);

// On the edit form, a task whose type is no longer assigned to its client
// keeps that type as a choice, so opening the form doesn't silently
// reassign it. It's dropped as soon as another client is picked.
const keepClient = form.dataset.keepClient || "";
const keepTaskType = form.dataset.keepTaskType || "";

function applyFilter() {
	const allowed = byClient[clientSelect.value];
	const options = allowed && allowed.length > 0
		? allOptions.filter((opt) => allowed.includes(opt.value)
			|| (opt.value === keepTaskType && clientSelect.value === keepClient))
		: allOptions;

	const previous = taskTypeSelect.value;
	taskTypeSelect.innerHTML = "";
	options.forEach((opt) => taskTypeSelect.appendChild(opt.cloneNode(true)));
	if (options.some((opt) => opt.value === previous)) {
		taskTypeSelect.value = previous;
	}
}

clientSelect.addEventListener("change", applyFilter);
applyFilter();
