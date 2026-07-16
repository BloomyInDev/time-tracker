// Hide days that hit their target exactly; the "show all" checkbox
// reveals them.
(function () {
	const toggle = document.getElementById("show-on-target");
	if (!toggle) return;

	const rows = document.querySelectorAll("tr.is-on-target");
	function apply() {
		for (const row of rows) {
			row.style.display = toggle.checked ? "" : "none";
		}
	}

	toggle.addEventListener("change", apply);
	apply();
})();
