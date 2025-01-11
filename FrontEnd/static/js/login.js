const signupForm = document.getElementById('signupForm');
const loginForm = document.getElementById('loginForm');
const toLogin = document.getElementById('toLogin');
const toSignUp = document.getElementById('toSignUp');
const leftSection = document.getElementById('leftSection');
const BASE_URL = 'http://localhost:8080'

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

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('closeButton').addEventListener('click', function () {
        window.location.href = '/';
    });
});

// Function to show the toast
function showToast(message, duration = 3500) {
    const toast = document.getElementById('toast');
    const toastMessage = document.getElementById('toastMessage');

    // Set the message
    toastMessage.textContent = message;

    // Show the toast
    toast.classList.add('show');

    // Hide the toast after the specified duration
    setTimeout(() => {
        toast.classList.remove('show');
    }, duration);
}

// Function to hide the toast
function hideToast() {
    const toast = document.getElementById('toast');
    toast.classList.remove('show');
}

// Function to handle the signup form submission
document.getElementById('signupForm').querySelector('button').addEventListener('click', function (event) {
    event.preventDefault(); // Prevent the default form submission
    hideToast(); // Hide any previous errors

    // Retrieve form data
    const username = document.getElementById('signupForm').querySelector('input[type="text"]').value;
    const email = document.getElementById('signupForm').querySelector('input[type="email"]').value;
    const password = document.getElementById('signupForm').querySelector('input[type="password"]').value;
    const confirmPassword = document.getElementById('signupForm').querySelectorAll('input[type="password"]')[1].value;

    // Validate the form data (optional)
    if (password !== confirmPassword) {
        showToast('Passwords do not match!');
        return;
    }

    // Create an object with the form data
    const signupData = {
        username: username,
        email: email,
        password: password
    };

    // Send the data to the backend using fetch
    fetch(`${BASE_URL}/register`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            // 'X-CSRF-Token': getCSRFToken() // Include the CSRF token in the headers
        },
        body: JSON.stringify(signupData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.error) {
            showToast(data.error); 
        } else {
            console.log('Success:', data);
            window.location.href = data.redirect
            alert('Signup successful!');
        }
    })
    .catch((error) => {
        console.error('Error:', error);
        showToast('Signup failed. Please try again.'); 
    });
});

// Function to handle the login form submission
document.getElementById('loginForm').querySelector('button').addEventListener('click', function (event) {
    event.preventDefault();
    hideToast(); 

    const username = document.getElementById('loginForm').querySelector('input[type="text"]').value;
    const password = document.getElementById('loginForm').querySelector('input[type="password"]').value;

    const loginData = {
        username: username,
        password: password
    };

    fetch(`${BASE_URL}/login`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(loginData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.error) {
            showToast(data.error); 
        } else {
            console.log('Success:', data);
            // Redirect to homepage after successful login
            window.location.href = data.redirect;
        }
    })
    .catch((error) => {
        console.error('Error:', error);
        showToast('Login failed. Please try again.'); 
    });
});