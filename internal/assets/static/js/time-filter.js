// Hide days that hit their target exactly; the "show all" checkbox reveals
// them and, in the same move, tells the PDF export to include them too.
(function () {
	const toggle = document.getElementById("show-on-target");
	if (!toggle) return;

	const rows = document.querySelectorAll("tr.is-on-target");
	const link = document.getElementById("export-link");

	function apply() {
		for (const row of rows) {
			row.style.display = toggle.checked ? "" : "none";
		}
		if (link) {
			const url = new URL(link.href, location.origin);
			if (toggle.checked) {
				url.searchParams.set("all", "1");
			} else {
				url.searchParams.delete("all");
			}
			link.href = url.pathname + "?" + url.searchParams.toString();
		}
	}

	toggle.addEventListener("change", apply);
	apply();
})();
