// 使用js-sha256库进行较安全的加密
function hashPassword(password) {
    return sha256(password); // 使用js-sha256进行哈希
}

document.getElementById('login-btn').addEventListener('click', async function () {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    // 输入验证
    if (!username || !password) {
        alert('请输入用户名和密码。');
        return;
    }

    // 加密密码
    const encryptedPassword = hashPassword(password);

    // 发送登录请求到后端
    fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password: encryptedPassword })
    })
        .then(response => {
            if (response.status === 200) {
                return response.json();
            } else if (response.status === 400) {
                console.log('用户名或密码错误。');
            } else if (response.status === 500) {
                console.log('服务器出现问题，请稍后再试。');
            }
        })
        .then(data => {
            if (data) {
                localStorage.setItem('token', data.token);
                window.location.href = '/display'; // 登录成功后跳转页面
            }
        })
        .catch(error => {
            console.log('登录请求失败：', error);
        });
});

document.addEventListener('DOMContentLoaded', function () {
    // 获取DOM元素
    const loginBtn = document.getElementById('login-btn');
    const usernameInput = document.getElementById('username');
    const passwordInput = document.getElementById('password');

    // 假设的登录验证方法（可以替换成实际的登录验证）
    function login(username, password) {
        // 这里简单模拟登录逻辑
        if (username === 'admin' && password === 'password123') {
            return true; // 登录成功
        }
        return false; // 登录失败
    }

    // 处理登录按钮点击事件
    // 登录按钮点击事件处理
    loginBtn.addEventListener('click', function () {
        const username = usernameInput.value;
        const password = passwordInput.value;

        if (username && password) {
            const success = login(username, password);

            if (success) {
                // 登录成功，存储用户名到 localStorage
                localStorage.setItem('username', username);
                alert('登录成功！');
            } else {
                alert('用户名或密码错误！');
            }
        } else {
            alert('请输入用户名和密码！');
        }
    });

});
