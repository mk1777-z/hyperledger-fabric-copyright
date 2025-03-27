// 使用 layui.use 确保在 layui 完全加载后执行代码
layui.use(['layer'], function() {
    var layer = layui.layer;
    
    // 登录函数
    function login(username, password) {
        // 密码加密
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
            if (response.ok) {
                return response.json();
            } else if (response.status === 401) {
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
            // 替换为原生的alert，避免layer未定义的风险
            alert('登录请求失败，请检查网络连接。');
        });
    }
    
    // 绑定登录按钮点击事件
    document.getElementById('login-btn').addEventListener('click', function() {
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        
        if (!username || !password) {
            layer.msg('用户名和密码不能为空', { icon: 2, time: 2000 });
            return;
        }
        
        login(username, password);
    });
});

// 使用js-sha256库进行较安全的加密
function hashPassword(password) {
    return sha256(password + 'copyright_salt'); // 添加固定盐值增强安全性
}
