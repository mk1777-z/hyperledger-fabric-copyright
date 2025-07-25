// 使用js-sha256库进行较安全的加密
function hashPassword(password) {
    return sha256(password); // 使用js-sha256进行哈希
}

document.addEventListener('DOMContentLoaded', function () {
    // 检查是否已登录为监管者，如果是则自动跳转到审核页面
    const token = localStorage.getItem('token');
    const username = localStorage.getItem('username');

    if (token && username === '监管者' &&
        (window.location.pathname.includes('/signin.html') || window.location.pathname === '/')) {
        window.location.href = '/audit.html';
        return;
    }

    // 获取登录按钮并添加事件监听
    const loginBtn = document.getElementById('login-btn');
    if (loginBtn) {
        loginBtn.addEventListener('click', handleLogin);
    }

    // 获取登录表单并添加提交事件监听
    const signinForm = document.querySelector('form');
    if (signinForm) {
        signinForm.addEventListener('submit', function (e) {
            e.preventDefault();
            handleLogin();
        });
    }
});

// 自定义消息提示函数
function showAlert(message, type = 'success', duration = 2000) {
    const alert = document.getElementById('customAlert');
    const icon = document.getElementById('alertIcon');
    const messageEl = document.getElementById('alertMessage');

    // 设置内容
    messageEl.textContent = message;
    alert.className = `custom-alert ${type}`;
    icon.className = type === 'success'
        ? 'fas fa-check-circle'
        : 'fas fa-times-circle';

    // 显示弹窗（带动画）
    alert.style.display = 'flex';
    setTimeout(() => alert.classList.add('show'), 10);

    // 自动隐藏
    setTimeout(() => {
        alert.classList.remove('show');
        setTimeout(() => alert.style.display = 'none', 400);
    }, duration);
}

// 统一的登录处理函数
function handleLogin() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password');
    const passwordValue = password ? password.value : '';

    // 输入验证
    if (!username || !passwordValue) {
        showAlert('请输入用户名和密码', 'error', 2000);
        return;
    }

    // 判断是否为监管者登录
    if (username === '监管者') {
        // 监管者登录逻辑 - 不进行密码哈希，直接使用明文密码
        handleRegulatorLogin(username, passwordValue);
    } else {
        // 普通用户登录逻辑
        handleUserLogin(username, passwordValue);
    }
}

// 处理监管者登录
function handleRegulatorLogin(username, password) {
    // 添加调试日志
    console.log('开始监管者登录请求');
    const encryptedPassword = hashPassword(password);

    fetch('/api/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, password: encryptedPassword })
    })
        .then(response => {
            console.log('监管者登录响应状态:', response.status);
            return response.json();
        })
        .then(data => {
            console.log('监管者登录响应数据:', data);
            if (data.token) {
                // 保存登录信息
                localStorage.setItem('token', data.token);
                localStorage.setItem('username', username);

                // 显示成功消息
                showAlert('监管者登录成功！', 'success', 1000);

                // 使用准确的HTML完整路径
                console.log('即将跳转到审核页面...');
                setTimeout(() => {
                    // 修改为正确的路径
                    window.location.href = '/audit';
                }, 1000);

                // 清空输入框
                clearInputs();
            } else {
                showAlert('监管者登录失败：' + (data.message || '验证失败'), 'error', 2000);
            }
        })
        .catch(error => {
            console.error('登录错误:', error);
            showAlert('监管者登录失败，请检查网络连接', 'error', 2000);
        });
}

// 处理普通用户登录
function handleUserLogin(username, password) {
    // 对普通用户的密码进行哈希
    const encryptedPassword = hashPassword(password);

    const urlParams = new URLSearchParams(window.location.search);
    const redirectUrl = urlParams.get('redirect') || '/display'; // 默认跳转到display页面
    fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password: encryptedPassword })
    })
        .then(async response => {
            if (response.status === 200) {
                return response.json();
            } else if (response.status === 400) {
                const errorText = await response.text();
                showAlert(errorText || '用户名或密码错误', 'error', 2000);
                return null;
            } else if (response.status === 500) {
                showAlert('服务器出现问题，请稍后再试', 'error', 2000);
                return null;
            } else {
                showAlert('用户名或密码错误', 'error', 2000);
                return null;
            }
        })
        .then(data => {
            if (data) {
                // 登录成功，存储信息
                localStorage.setItem('token', data.token);
                localStorage.setItem('username', username);

                showAlert('登录成功！', 'success', 1000);

                setTimeout(() => {
                    // window.location.href = '/display'; // 普通用户跳转页面
                    window.location.href = redirectUrl;
                }, 1000);

                // 清空输入框
                clearInputs();
            }
        })
        .catch(error => {
            console.error('登录请求失败：', error);
            showAlert('登录请求失败，请检查网络连接', 'error', 2000);
        });
}

// 清空输入框
function clearInputs() {
    const usernameInput = document.getElementById('username');
    const passwordInput = document.getElementById('password');

    if (usernameInput) usernameInput.value = '';
    if (passwordInput) passwordInput.value = '';
}