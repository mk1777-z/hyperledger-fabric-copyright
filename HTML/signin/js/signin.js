// 使用SHA-256加密
async function hashPassword(password) {
    const encoder = new TextEncoder();
    const data = encoder.encode(password);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

document.getElementById('login-btn').addEventListener('click', async function() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    // 输入验证
    if (!username || !password) {
        alert('请输入用户名和密码。');
        return;
    }

    // 验证用户名是否是字母和空格（要求使用真实姓名）
    const realNamePattern = /^[a-zA-Z\s]+$/;
    if (!realNamePattern.test(username)) {
        alert('用户名必须是您的真实姓名，只能包含字母和空格。');
        return;
    }

    // 防止SQL注入：禁止包含特殊字符
    const sqlPattern = /['"=;--]/;
    if (sqlPattern.test(username) || sqlPattern.test(password)) {
        alert('您的输入包含非法字符。');
        return;
    }

    // 加密密码
    const encryptedPassword = await hashPassword(password);

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
            window.location.href = 'display.html'; // 登录成功后跳转页面
        }
    })
    .catch(error => {
        console.log('登录请求失败：', error);
    });
});
