const signupForm = document.getElementById('signupForm');
const loginForm = document.getElementById('loginForm');
const toLogin = document.getElementById('toLogin');
const toSignUp = document.getElementById('toSignUp');
const BASE_URL = 'http://localhost:8080'

// Toggle to Log In Form
toLogin.addEventListener('click', (e) => {
    e.preventDefault();
    signupForm.classList.remove('active');
    loginForm.classList.add('active');
});

// Toggle to Sign Up Form
toSignUp.addEventListener('click', (e) => {
    e.preventDefault();
    loginForm.classList.remove('active');
    signupForm.classList.add('active');
});

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('closeButton').addEventListener('click', function () {
        window.location.href = '/';
    });
});

function togglePassword(inputId, iconId) {
    const passwordInput = document.getElementById(inputId);
    const icon = document.getElementById(iconId);
    
    if (passwordInput.type === "password") {
        passwordInput.type = "text";
        icon.classList.remove("fa-eye");
        icon.classList.add("fa-eye-slash");
    } else {
        passwordInput.type = "password";
        icon.classList.remove("fa-eye-slash");
        icon.classList.add("fa-eye");
    }
}


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

    // Get form elements using more specific selectors
    const usernameInput = document.getElementById('signupUsername');
    const emailInput = document.getElementById('signupEmail');
    const passwordInput = document.getElementById('signupPassword');
    const confirmPasswordInput = document.getElementById('signupConfirmPassword');
    

    // Check if all elements exist
    if (!usernameInput || !emailInput || !passwordInput || !confirmPasswordInput) {
        console.error('One or more form elements not found');
        showToast('Form error: Missing elements');
        return;
    }

    // Retrieve form data
    const username = usernameInput.value;
    const email = emailInput.value;
    const password = passwordInput.value;
    const confirmPassword = confirmPasswordInput.value;

    // Validate the form data
    if (!username || !email || !password || !confirmPassword) {
        showToast('Please fill in all fields');
        return;
    }

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
            'Accept': 'application/json',
            'X-Requested-With': 'XMLHttpRequest'
        },
        body: JSON.stringify(signupData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.error) {
            showToast(data.error); 
        } else {
            console.log('Success:', data);
            window.location.href = data.redirect;
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
            'Accept': 'application/json',
            'X-Requested-With': 'XMLHttpRequest'
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

document.querySelector(".google-button").addEventListener("click", () => {
    window.location.href = "/googleLogin";
});

document.querySelector(".github-button").addEventListener("click", () => {
    window.location.href = "/githubLogin";
});

// Add this function before the window.onload handler
function checkLoginStatus() {
    return fetch('/checkLoginStatus', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: 'include' // Important for sending cookies
    })
    .then(response => response.json())
    .then(data => {
        return data.loggedIn;
    })
    .catch(error => {
        console.error('Error checking login status:', error);
        return false;
    });
}

// The existing window.onload code remains the same
window.onload = function() {
    checkLoginStatus().then(isLoggedIn => {
        if (isLoggedIn) {
            window.location.replace('/');
        }
    });
};

