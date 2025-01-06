const signupForm = document.getElementById('signupForm');
const loginForm = document.getElementById('loginForm');
const toLogin = document.getElementById('toLogin');
const toSignUp = document.getElementById('toSignUp');
const leftSection = document.getElementById('leftSection');

// Toggle to Log In Form
toLogin.addEventListener('click', (e) => {
    e.preventDefault();
    signupForm.classList.remove('active');
    loginForm.classList.add('active');
    leftSection.innerHTML = '<h1>Login</h1><p>Welcome back!</p>';
});

// Toggle to Sign Up Form
toSignUp.addEventListener('click', (e) => {
    e.preventDefault();
    loginForm.classList.remove('active');
    signupForm.classList.add('active');
    leftSection.innerHTML = '<h1>Sign up</h1><p>to use all features of the application</p>';
});