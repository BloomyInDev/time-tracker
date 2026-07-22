const searchInput = document.getElementById("client-search");
const rows = Array.from(document.querySelectorAll("#client-rows tr"));

searchInput.addEventListener("input", () => {
	const query = searchInput.value.trim().toLowerCase();
	rows.forEach((row) => {
		const matches = row.dataset.name.includes(query);
		row.style.display = matches ? "" : "none";
	});
});
