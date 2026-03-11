function switchTab(mode) {
    document.getElementById('auth-mode').value = mode;
    const loginBtn = document.getElementById('show-login');
    const signupBtn = document.getElementById('show-signup');
    const submitBtn = document.getElementById('submit-btn');

    if (mode === 'login') {
        loginBtn.className = "flex-1 py-2 text-sm font-medium rounded-md bg-white shadow-sm text-gray-800";
        signupBtn.className = "flex-1 py-2 text-sm font-medium rounded-md text-gray-500 hover:text-gray-800";
        submitBtn.innerText = "Login";
    } else {
        signupBtn.className = "flex-1 py-2 text-sm font-medium rounded-md bg-white shadow-sm text-gray-800";
        loginBtn.className = "flex-1 py-2 text-sm font-medium rounded-md text-gray-500 hover:text-gray-800";
        submitBtn.innerText = "Sign Up";
    }
}

async function handleAuth(event) {
    event.preventDefault();
    const mode = document.getElementById('auth-mode').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const errorBox = document.getElementById('error-box');

    errorBox.classList.add('hidden');

    try {
        const res = await fetch(`/api/auth/${mode}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await res.json();

        if (!res.ok) {
            throw new Error(data.error || "Authentication failed");
        }

        if (mode === 'login') {
            localStorage.setItem('token', data.token);
            window.location.href = '/dashboard';
        } else {
            // Auto login after signup
            switchTab('login');
            alert("Account created successfully! Please log in.");
            document.getElementById('password').value = "";
        }
    } catch (err) {
        errorBox.innerText = err.message;
        errorBox.classList.remove('hidden');
    }
}