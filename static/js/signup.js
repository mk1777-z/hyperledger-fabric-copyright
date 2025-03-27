// 使用 layui.use 确保在 layui 完全加载后执行代码
layui.use(['layer'], function() {
    var layer = layui.layer;
    
    // 绑定注册按钮点击事件
    document.getElementById('signup-btn').addEventListener('click', async function() {
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        const confirmPassword = document.getElementById('confirm-password').value;
        
        // 输入验证
        if (!username || !password || !confirmPassword) {
            layer.msg('所有字段均为必填项。', {icon: 2, time: 2000});
            return;
        }
        
        if (password !== confirmPassword) {
            layer.msg('两次输入的密码不一致。', {icon: 2, time: 2000});
            return;
        }
        
        // 验证用户名是否符合真实姓名要求
        const realNamePattern = /^[a-zA-Z\s]+$/;
        if (!realNamePattern.test(username)) {
            layer.msg('用户名必须是您的真实姓名，只能包含字母和空格。', {icon: 2, time: 2000});
            return;
        }
        
        // 加密密码
        const encryptedPassword = hashPassword(password);
        
        // 发送注册请求到后端
        fetch('/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password: encryptedPassword })
        })
        .then(response => {
            if (response.status === 200) {
                layer.msg('注册成功，请登录。', {icon: 1, time: 2000});
                setTimeout(() => {
                    window.location.href = '/signin'; // 注册成功后跳转到登录页面
                }, 2000);
            }
            else if (response.status === 400) {
                response.json().then(data => {
                    layer.msg(data.error, {icon: 2, time: 2000}); // 显示后端返回的错误信息
                });
            } else {
                layer.msg('注册失败，请稍后重试。', {icon: 2, time: 2000});
            }
        })
        .catch(error => {
            console.error('注册请求失败：', error);
            alert('网络错误，请稍后重试'); // 使用原生alert作为备用
        });
    });
});

// 使用js-sha256库进行较安全的加密
function hashPassword(password) {
    return sha256(password + 'copyright_salt'); // 添加固定盐值增强安全性
}
