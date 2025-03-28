document.addEventListener("DOMContentLoaded", function () {
    // 获取表单和错误消息元素
    const loginForm = document.getElementById('login-form');
    const errorMessage = document.getElementById('error-message');
    const loading = document.getElementById('loading');

    // 确保表单存在
    if (!loginForm) {
        console.error("未找到登录表单（#login-form）");
        return;
    }

    loginForm.addEventListener('submit', function (event) {
        event.preventDefault();

        // 获取用户名和密码
        const username = document.getElementById('username');
        const password = document.getElementById('password');
        const submitButton = loginForm.querySelector('button[type="submit"]');

        // 前端验证
        let hasError = false;
        if (!username.value.trim()) {
            username.classList.add('is-invalid');
            hasError = true;
        } else {
            username.classList.remove('is-invalid');
        }
        if (!password.value.trim()) {
            password.classList.add('is-invalid');
            hasError = true;
        } else {
            password.classList.remove('is-invalid');
        }
        if (hasError) {
            return;
        }

        // 显示加载状态
        submitButton.disabled = true;
        submitButton.innerText = '登录中...';
        loading.classList.add('show');
        errorMessage.style.display = 'none'; // 清空之前的错误消息

        // 获取 CSRF token（假设后端在页面中提供）
        const csrfToken = document.querySelector('meta[name="csrf-token"]')?.content || '';

        // 发送 AJAX 请求到后端登录接口
        fetch('/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'X-CSRF-Token': csrfToken // 添加 CSRF token
            },
            body: new URLSearchParams({
                username: username.value,
                password: password.value
            })
        })
            .then(response => {
                if (response.ok) {
                    // 登录成功，跳转到资产管理页面
                    window.location.href = '/assets';
                } else {
                    // 登录失败，显示错误信息
                    response.text().then(error => {
                        errorMessage.style.display = 'block';
                        errorMessage.innerText = error || '登录失败，请重试';
                        // 5秒后自动隐藏错误消息
                        setTimeout(() => {
                            errorMessage.style.display = 'none';
                        }, 5000);
                    });
                }
            })
            .catch(error => {
                console.error('登录请求失败:', error);
                errorMessage.style.display = 'block';
                errorMessage.innerText = '网络错误，请稍后重试';
                setTimeout(() => {
                    errorMessage.style.display = 'none';
                }, 5000);
            })
            .finally(() => {
                // 恢复按钮状态
                submitButton.disabled = false;
                submitButton.innerText = '登录';
                loading.classList.remove('show');
            });
    });
});