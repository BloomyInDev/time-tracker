// Hide days after today; the "show future" checkbox reveals them.
(function () {
	const toggle = document.getElementById("show-future-tasks");
	if (!toggle) return;

	const rows = document.querySelectorAll("tr.is-future");

	function apply() {
		for (const row of rows) {
			row.style.display = toggle.checked ? "" : "none";
		}
	}

	toggle.addEventListener("change", apply);
	apply();
})();
