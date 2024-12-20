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
                alert('用户名或密码错误。');
            } else if (response.status === 500) {
                alert('服务器出现问题，请稍后再试。');
            }
        })
        .then(data => {
            if (data) {
                // 登录成功，存储 token 并跳转
                localStorage.setItem('token', data.token);
                localStorage.setItem('username', username); // 存储用户名
                alert('登录成功！');
                window.location.href = '/display'; // 登录成功后跳转页面
                
                // 清空输入框
                document.getElementById('username').value = '';
                document.getElementById('password').value = '';
            }
        })
        .catch(error => {
            console.log('登录请求失败：', error);
            alert('登录请求失败，请检查网络连接。');
        });
});

