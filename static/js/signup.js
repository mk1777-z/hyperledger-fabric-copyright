// 使用js-sha256库进行较安全的加密
function hashPassword(password) {
    return sha256(password); // 使用js-sha256进行哈希
}

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

document.getElementById('signup-btn').addEventListener('click', async function () {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirm-password').value;

    // 输入验证
    if (!username || !password || !confirmPassword) {
        showAlert('所有字段均为必填项', 'error', 2000);
        return;
    }

    if (password !== confirmPassword) {
        showAlert('两次输入的密码不一致', 'error', 2000);
        return;
    }

    // // 验证用户名是否符合真实姓名要求
    // const realNamePattern = /^[a-zA-Z\s]+$/;
    // if (!realNamePattern.test(username)) {
    //     alert('用户名必须是您的真实姓名，只能包含字母和空格。');
    //     return;
    // }

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
                showAlert('注册成功！', 'success', 5000);
                setTimeout(() => {
                    window.location.href = '/'; // 注册成功后跳转到登录页面
                }, 5000); // 延迟跳转5秒
            }
            else if (response.status === 400) {   // 用户名已存在
                response.json().then(data => {
                    showAlert(data.error, 'error', 2000); // 显示后端返回的错误信息
                });
            } else {
                console.log('注册失败...');
            }
        })
        .catch(error => {
            console.log('注册请求失败：', error);
        });
});

// 返回按钮功能：点击返回到登录页面
document.getElementById('back-btn').addEventListener('click', function () {
    window.location.href = 'signin.html'; // 跳转到登录页面
});
