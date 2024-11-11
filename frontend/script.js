let isSignIn = false;
let authModal;

let ws;
let authToken;

document.addEventListener('DOMContentLoaded', () => {
    authModal = new bootstrap.Modal(document.getElementById('authModal'));
    checkAuth();

    document.getElementById('authForm').addEventListener('submit', handleAuth);
});

function checkAuth() {
    authToken = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!authToken) {
        showAuthModal();
    } else {
        const username = localStorage.getItem('username') || sessionStorage.getItem('username');
        updateUIForAuth(username);
    }
}

function updateUIForAuth(username) {
    document.getElementById('username').textContent = username;
    document.getElementById('userInfo').classList.remove('d-none');
    document.getElementById('playButton').onclick = startGame;
}

function updateUIForUnauth() {
    document.getElementById('userInfo').classList.add('d-none');
    document.getElementById('playButton').onclick = showAuthModal;
    showAuthModal();
}

function showAuthModal() {
    document.getElementById('authModalTitle').textContent = isSignIn ? 'SIGN IN' : 'CREATE ACCOUNT';
    authModal.show();
}

function toggleAuthMode() {
    isSignIn = !isSignIn;
    document.getElementById('authModalTitle').textContent = isSignIn ? 'SIGN IN' : 'CREATE ACCOUNT';
    const switchButton = document.querySelector('.switch-auth-button');
    switchButton.textContent = isSignIn ? 'New player? Create Account' : 'Already have an account? Sign In';
}

async function handleAuth(e) {
    e.preventDefault();

    const username = document.getElementById('usernameInput').value;
    const password = document.getElementById('passwordInput').value;
    const rememberMe = document.getElementById('rememberMe').checked;

    const endpoint = isSignIn ? 'signin' : 'signup';

    try {
        const response = await fetch(`${location.protocol}//localhost:8080/api/auth/${endpoint}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password }),
        });

        if (response.ok) {
            const {token} = await response.json();
            const storage = rememberMe ? localStorage : sessionStorage;

            authToken = token;
            storage.setItem('token', token);
            storage.setItem('username', username);

            updateUIForAuth(username);
            authModal.hide();
            document.getElementById('authForm').reset();
        } else {
            const message = isSignIn ? 'Invalid credentials' : 'Username already exists';
            showGameAlert(message);
        }
    } catch (error) {
        showGameAlert('Connection error! Try again.');
    }
}

function showGameAlert(message) {
    const alertDiv = document.createElement('div');
    alertDiv.className = 'game-alert';
    alertDiv.textContent = message;
    document.querySelector('.modal-body').prepend(alertDiv);

    setTimeout(() => alertDiv.remove(), 3000);
}

function signOut() {
    authToken = null;
    localStorage.removeItem('token');
    localStorage.removeItem('username');
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('username');
    updateUIForUnauth();
}


function startGame() {
    showChatPanel();
    connectWebSocket();
}

function showChatPanel() {
    document.getElementById('mainScreen').classList.add('d-none');
    document.getElementById('chatPanel').classList.remove('d-none');
}

function closeChatPanel() {
    document.getElementById('mainScreen').classList.remove('d-none');
    document.getElementById('chatPanel').classList.add('d-none');
}

function sendMessage() {
    const input = document.getElementById('chatInput');
    const message = input.value.trim();

    if (message && ws) {
        ws.send(JSON.stringify({
            type: 'message',
            content: message
        }));

        input.value = '';
    }
}


function connectWebSocket() {
    const wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//localhost:8080/api/ws?token=${authToken}`;

    const wsClient = new WebSocket(wsUrl);
    wsClient.onopen = () => {
        console.log('Connected to WebSocket');
    };

    ws = wsClient;

    ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        handleWebSocketMessage(message);
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        alert('WebSocket error: Press F12 for more info');
        closeChatPanel();
    };

    ws.onclose = (event) => {
        console.log('Disconnected from WebSocket:', event.code, event.reason);
        closeChatPanel();
        ws = null;
    };
}

function handleWebSocketMessage(message) {
    switch (message.type) {
        case 'player_joined':
            addSystemMessage(`${message.username} joined the room`);
            break;
        case 'player_left':
            addSystemMessage(`${message.username} left the room`);
            break;
        case 'message':
            addChatMessage(message.username, JSON.parse(message.content).content);
            break;
    }
}

function addSystemMessage(message) {
    const chatMessages = document.getElementById('chatMessages');
    const messageDiv = document.createElement('div');

    messageDiv.classList.add('system-message');
    messageDiv.textContent = message;

    chatMessages.appendChild(messageDiv);

    chatMessages.scrollTop = chatMessages.scrollHeight;
}

function addChatMessage(sender, message) {
    const chatMessages = document.getElementById('chatMessages');
    const messageDiv = document.createElement('div');

    messageDiv.classList.add('chat-message');
    messageDiv.innerHTML = `<strong>${sender}:</strong> ${message}`;

    chatMessages.appendChild(messageDiv);

    chatMessages.scrollTop = chatMessages.scrollHeight;
}