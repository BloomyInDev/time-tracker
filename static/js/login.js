document.getElementById("login-form").addEventListener("submit", async (e) => {
	e.preventDefault();

	const errorBox = document.getElementById("login-error");
	errorBox.classList.add("is-hidden");

	const form = e.target;
	const body = {
		email: form.email.value,
		password: form.password.value,
	};

	const res = await fetch("/login", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(body),
	});

	if (!res.ok) {
		errorBox.textContent = res.status === 401 ? "Invalid credentials" : "Something went wrong";
		errorBox.classList.remove("is-hidden");
		return;
	}

	const { token } = await res.json();
	localStorage.setItem("token", token);
	window.location.href = "/";
});
