const form = document.getElementById("task-form");
const clientSelect = document.getElementById("task-client-select");
const taskTypeSelect = document.getElementById("task-type-select");

const byClient = {};
form.dataset.clientTaskTypes.split(";").filter(Boolean).forEach((entry) => {
	const [clientID, ids] = entry.split(":");
	byClient[clientID] = ids ? ids.split(",") : [];
});

const allOptions = Array.from(taskTypeSelect.options);

function applyFilter() {
	const allowed = byClient[clientSelect.value];
	const options = allowed && allowed.length > 0
		? allOptions.filter((opt) => allowed.includes(opt.value))
		: allOptions;

	taskTypeSelect.innerHTML = "";
	options.forEach((opt) => taskTypeSelect.appendChild(opt.cloneNode(true)));
}

clientSelect.addEventListener("change", applyFilter);
applyFilter();
