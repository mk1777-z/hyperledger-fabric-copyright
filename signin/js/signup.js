async function hashPassword(password) {
    const encoder = new TextEncoder();
    const data = encoder.encode(password);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

document.getElementById('signup-btn').addEventListener('click', async function() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirm-password').value;

    // 输入验证
    if (!username || !password || !confirmPassword) {
        alert('所有字段均为必填项。');
        return;
    }

    if (password !== confirmPassword) {
        alert('两次输入的密码不一致。');
        return;
    }

    // 验证用户名是否符合真实姓名要求
    const realNamePattern = /^[a-zA-Z\s]+$/;
    if (!realNamePattern.test(username)) {
        alert('用户名必须是您的真实姓名，只能包含字母和空格。');
        return;
    }

    // 加密密码
    const encryptedPassword = await hashPassword(password);

    // 发送注册请求到后端
    fetch('/signup', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password: encryptedPassword })
    })
    .then(response => {
        if (response.status === 200) {
            alert('注册成功，请登录。');
            window.location.href = 'index.html'; // 注册成功后跳转到登录页面
        } else {
            console.log('注册失败。');
        }
    })
    .catch(error => {
        console.log('注册请求失败：', error);
    });
});

// 返回按钮功能：点击返回到登录页面
document.getElementById('back-btn').addEventListener('click', function() {
    window.location.href = 'signin.html'; // 跳转到登录页面
});
