// 使用js-sha256库进行较安全的加密
function hashPassword(password) {
    return sha256(password); // 使用js-sha256进行哈希
}

document.getElementById('login-btn').addEventListener('click', async function () {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    // 输入验证
    if (!username || !password) {
        layer.msg('请输入用户名和密码。', { icon: 0, time: 2000 }); // 使用layui的msg方法显示提示
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
        .then(async response => {
            if (response.status === 200) {
                return response.json(); // 返回JSON数据
            } else if (response.status === 400) {
                // 明确提示用户名或密码错误
                const errorText = await response.text();
                layer.msg(errorText || '用户名或密码错误。', { icon: 2, time: 2000 }); // 错误提示
                return null;
            } else if (response.status === 500) {
                layer.msg('服务器出现问题，请稍后再试。', { icon: 2, time: 2000 }); // 错误提示
                return null;
            } else {
                layer.msg('用户名或密码错误。', { icon: 2, time: 2000 });
                return null;
            }
        })
        .then(data => {
            if (data) {
                // 登录成功，存储 token 并跳转
                localStorage.setItem('token', data.token);
                localStorage.setItem('username', username); // 存储用户名
                layer.msg('登录成功！', { icon: 1, time: 1000 }); // 成功提示
                setTimeout(() => {
                    window.location.href = '/display'; // 登录成功后跳转页面
                }, 1000);

                // 清空输入框
                document.getElementById('username').value = '';
                document.getElementById('password').value = '';
            }
        })
        .catch(error => {
            console.error('登录请求失败：', error);
            layer.msg('登录请求失败，请检查网络连接。', { icon: 2, time: 2000 }); // 错误提示
        });
});