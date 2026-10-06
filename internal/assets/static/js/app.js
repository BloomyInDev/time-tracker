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
