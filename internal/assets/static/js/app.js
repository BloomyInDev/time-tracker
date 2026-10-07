// Keyboard shortcuts: an element with data-click-on="ctrl+s" is clicked when
// that combination is pressed. Only "ctrl+<key>" is supported; Cmd counts as
// Ctrl so the same shortcut works on macOS.
document.addEventListener("keydown", function (event) {
	if (!(event.ctrlKey || event.metaKey) || event.altKey || event.shiftKey) return;
	const combo = "ctrl+" + event.key.toLowerCase();
	for (const el of document.querySelectorAll("[data-click-on]")) {
		if (el.dataset.clickOn.toLowerCase() === combo && !el.disabled) {
			event.preventDefault();
			el.click();
			return;
		}
	}
});

// Mobile navbar: a .navbar-burger toggles the menu named by its data-target.
for (const burger of document.querySelectorAll(".navbar-burger")) {
	burger.addEventListener("click", function () {
		const menu = document.getElementById(burger.dataset.target);
		const open = burger.classList.toggle("is-active");
		menu.classList.toggle("is-active", open);
		burger.setAttribute("aria-expanded", String(open));
	});
}
