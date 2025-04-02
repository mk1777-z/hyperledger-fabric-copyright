document.addEventListener('DOMContentLoaded', function () {
    // 检查登录状态
    const token = localStorage.getItem('token');
    const username = localStorage.getItem('username');

    console.log('审核页面加载，当前用户:', username);

    if (!token || username !== '监管者') {
        console.log('未登录或非监管者用户，重定向到登录页');
        window.location.href = '/signin.html'; // 修改为正确的登录页路径
        return;
    }

    // 显示用户信息
    document.getElementById('userInfo').textContent = username;

    // 初始化Bootstrap标签页
    const tabs = document.querySelectorAll('[data-bs-toggle="tab"]');
    tabs.forEach(tab => {
        tab.addEventListener('click', function (event) {
            event.preventDefault();
            const targetId = this.getAttribute('href');

            // 移除所有活动标签
            document.querySelectorAll('.tab-pane').forEach(pane => {
                pane.classList.remove('show', 'active');
            });
            document.querySelectorAll('.nav-link').forEach(link => {
                link.classList.remove('active');
            });

            // 激活当前标签
            document.querySelector(targetId).classList.add('show', 'active');
            this.classList.add('active');
        });
    });

    // 退出登录
    document.getElementById('logoutBtn').addEventListener('click', function () {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        console.log('登出成功，重定向到登录页');
        window.location.href = '/signin.html'; // 确保退出后跳转到正确的登录页
    });

    // 刷新数据按钮
    document.getElementById('refreshBtn').addEventListener('click', loadAllItems);

    // 加载所有版权项目
    loadAllItems();

    // 审核表单处理
    const approveBtn = document.getElementById('approveBtn');
    const rejectBtn = document.getElementById('rejectBtn');

    approveBtn.addEventListener('click', function () {
        submitAudit('APPROVE');
    });

    rejectBtn.addEventListener('click', function () {
        submitAudit('REJECT');
    });
});

// 加载所有版权项目
function loadAllItems() {
    showLoading();

    fetch('/api/audit/categorized-items', {  // 使用新的分类API
        method: 'GET',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        }
    })
        .then(response => {
            console.log('获取项目列表响应状态:', response.status);
            return response.json();
        })
        .then(data => {
            console.log('获取到分类项目数据:', data);

            // 直接从响应中获取已分类的项目数据
            renderItems('pendingItems', data.pendingItems || [], 'pending');
            renderItems('approvedItems', data.approvedItems || [], 'approved');
            renderItems('rejectedItems', data.rejectedItems || [], 'rejected');

            // 检查是否有内容
            toggleEmptyState('pendingEmpty', !data.pendingItems || data.pendingItems.length === 0);
            toggleEmptyState('approvedEmpty', !data.approvedItems || data.approvedItems.length === 0);
            toggleEmptyState('rejectedEmpty', !data.rejectedItems || data.rejectedItems.length === 0);
        })
        .catch(error => {
            console.error('获取项目列表错误：', error);
            alert('获取版权项目失败：' + error.message);
        })
        .finally(() => {
            hideLoading();
        });
}

// 处理项目并分类
async function processItems(items) {
    // 清空现有项目
    document.getElementById('pendingItems').innerHTML = '';
    document.getElementById('approvedItems').innerHTML = '';
    document.getElementById('rejectedItems').innerHTML = '';

    const pendingItems = [];
    const approvedItems = [];
    const rejectedItems = [];

    // 处理每个项目
    for (const item of items) {
        // 如果项目没有transID，则跳过
        if (!item.transID) {
            continue;
        }

        // 获取第一个交易ID（如果有多个用空格分隔）
        const transIDs = item.transID.split(' ');
        const firstTransID = transIDs.length > 0 ? transIDs[0] : '';

        if (!firstTransID) {
            continue;
        }

        try {
            // 查询审核历史
            const status = await checkAuditStatus(firstTransID);

            // 根据状态分类
            switch (status) {
                case 'PENDING':
                    pendingItems.push({ ...item, firstTransID });
                    break;
                case 'APPROVE':
                    approvedItems.push({ ...item, firstTransID });
                    break;
                case 'REJECT':
                    rejectedItems.push({ ...item, firstTransID });
                    break;
            }
        } catch (error) {
            console.error(`检查项目 ${item.name} 的审核状态失败:`, error);
            // 出错的项目默认为待审核
            pendingItems.push({ ...item, firstTransID });
        }
    }

    // 更新界面
    renderItems('pendingItems', pendingItems, 'pending');
    renderItems('approvedItems', approvedItems, 'approved');
    renderItems('rejectedItems', rejectedItems, 'rejected');

    // 检查是否有内容
    toggleEmptyState('pendingEmpty', pendingItems.length === 0);
    toggleEmptyState('approvedEmpty', approvedItems.length === 0);
    toggleEmptyState('rejectedEmpty', rejectedItems.length === 0);
}

// 检查审核状态
async function checkAuditStatus(tradeID) {
    try {
        const response = await fetch(`/api/audit/history?tradeId=${tradeID}`, {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            }
        });

        const data = await response.json();

        if (data.message === "交易不存在") {
            return 'PENDING'; // 未审核状态
        }

        if (data.records && data.records.length > 0) {
            // 取最新的一条审核记录
            const latestRecord = data.records[data.records.length - 1];
            return latestRecord.decision; // APPROVE 或 REJECT
        }

        return 'PENDING'; // 默认为未审核状态
    } catch (error) {
        console.error('获取审核状态失败:', error);
        throw error;
    }
}

// 渲染项目列表 - 使用列表形式而非卡片
function renderItems(containerId, items, type) {
    const container = document.getElementById(containerId);

    // 如果没有项目，显示空状态
    if (!items || items.length === 0) {
        return;
    }

    // 创建表格视图
    const table = document.createElement('table');
    table.className = 'table table-hover table-striped';

    // 创建表头
    const thead = document.createElement('thead');
    thead.innerHTML = `
        <tr>
            <th scope="col">状态</th>
            <th scope="col">名称</th>
            <th scope="col">所有者</th>
            <th scope="col">上传时间</th>
            <th scope="col">交易ID</th>
            <th scope="col">操作</th>
        </tr>
    `;
    table.appendChild(thead);

    // 创建表格主体
    const tbody = document.createElement('tbody');

    items.forEach(item => {
        const tr = document.createElement('tr');

        // 获取状态标签
        const statusBadge = getStatusBadge(type);

        // 设置行内容
        tr.innerHTML = `
            <td>${statusBadge}</td>
            <td><div class="fw-bold">${item.name}</div>
                <small class="text-muted">${item.simple_dsc || '暂无描述'}</small></td>
            <td>${item.owner || '未知'}</td>
            <td>${formatTime(item.start_time) || '未知'}</td>
            <td><small class="text-truncate d-inline-block" style="max-width:150px;" 
                title="${item.firstTransID}">${item.firstTransID}</small></td>
            <td>
                <button class="btn btn-info btn-sm view-details" data-id="${item.id}">查看详情</button>
                ${type === 'pending' ? `<button class="btn btn-primary btn-sm ms-1 audit-item" 
                data-id="${item.id}" data-trans-id="${item.firstTransID}">审核</button>` : ''}
            </td>
        `;

        tbody.appendChild(tr);
    });

    table.appendChild(tbody);
    container.appendChild(table);

    // 添加事件监听
    container.querySelectorAll('.view-details').forEach(btn => {
        btn.addEventListener('click', function () {
            const itemId = this.getAttribute('data-id');
            viewItemDetails(itemId);
        });
    });

    container.querySelectorAll('.audit-item').forEach(btn => {
        btn.addEventListener('click', function () {
            const itemId = this.getAttribute('data-id');
            const transId = this.getAttribute('data-trans-id');
            showAuditModal(itemId, transId);
        });
    });
}

// 获取状态标签 - 改为文本标签
function getStatusBadge(type) {
    switch (type) {
        case 'pending':
            return '<span class="badge bg-warning">待审核</span>';
        case 'approved':
            return '<span class="badge bg-success">已通过</span>';
        case 'rejected':
            return '<span class="badge bg-danger">未通过</span>';
        default:
            return '';
    }
}

// 显示审核弹窗
function showAuditModal(itemId, transId) {
    // 获取项目详情
    fetch(`/api/audit/info?tradeId=${transId}`, {
        method: 'GET',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        }
    })
        .then(response => response.json())
        .then(data => {
            if (data.message && data.itemDetails) {
                const item = data.itemDetails;

                // 填充弹窗内容
                document.getElementById('modalItemId').value = transId;
                document.getElementById('modalItemName').textContent = item.name;
                document.getElementById('modalItemDesc').textContent = item.dsc || item.simple_dsc || '暂无描述';
                document.getElementById('modalItemOwner').textContent = item.owner || '未知';
                document.getElementById('modalItemPrice').textContent = `¥${item.price || 0}`;
                document.getElementById('modalItemTransID').textContent = transId;
                document.getElementById('modalItemTime').textContent = formatTime(item.start_time);

                // 设置图片
                const imgElement = document.getElementById('modalItemImage');
                if (item.img) {
                    imgElement.src = item.img;
                    imgElement.style.display = 'block';
                } else {
                    imgElement.style.display = 'none';
                }

                // 清空表单
                document.getElementById('auditComment').value = '';
                document.getElementById('regulatorPassword').value = '';

                // 显示弹窗
                const modal = new bootstrap.Modal(document.getElementById('auditModal'));
                modal.show();
            } else {
                alert('获取项目详情失败');
            }
        })
        .catch(error => {
            console.error('获取项目详情错误:', error);
            alert('获取项目详情失败: ' + error.message);
        });
}

// 提交审核决定
function submitAudit(decision) {
    const tradeId = document.getElementById('modalItemId').value;
    const comment = document.getElementById('auditComment').value;
    const password = document.getElementById('regulatorPassword').value;

    if (!comment) {
        alert('请输入审核意见');
        return;
    }

    if (!password) {
        alert('请输入监管者密码');
        return;
    }

    showLoading();

    fetch('/api/audit', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        },
        body: JSON.stringify({
            tradeId: tradeId,
            decision: decision,
            comment: comment,
            password: password
        })
    })
        .then(response => response.json())
        .then(data => {
            if (data.message === '审核成功') {
                // 关闭弹窗
                bootstrap.Modal.getInstance(document.getElementById('auditModal')).hide();

                // 重新加载数据
                loadAllItems();

                // 显示成功消息
                alert(`版权审核${decision === 'APPROVE' ? '通过' : '拒绝'}成功！`);
            } else {
                alert('审核失败: ' + (data.message || '未知错误'));
            }
        })
        .catch(error => {
            console.error('审核提交错误:', error);
            alert('审核提交失败: ' + error.message);
        })
        .finally(() => {
            hideLoading();
        });
}

// 查看项目详情
function viewItemDetails(itemId) {
    // 项目详情查看逻辑，可以跳转到项目详情页或显示详情弹窗
    window.open(`/item.html?id=${itemId}`, '_blank');
}

// 切换空状态显示
function toggleEmptyState(elementId, isEmpty) {
    const element = document.getElementById(elementId);
    if (isEmpty) {
        element.classList.remove('d-none');
    } else {
        element.classList.add('d-none');
    }
}

// 格式化时间
function formatTime(timestamp) {
    if (!timestamp) return '未知时间';

    try {
        const date = new Date(timestamp);
        return date.toLocaleString('zh-CN');
    } catch (e) {
        return timestamp;
    }
}

// 显示加载动画
function showLoading() {
    let spinner = document.querySelector('.spinner-overlay');

    if (!spinner) {
        spinner = document.createElement('div');
        spinner.className = 'spinner-overlay';
        spinner.innerHTML = `
            <div class="spinner-border text-primary" role="status">
                <span class="visually-hidden">加载中...</span>
            </div>
        `;
        document.body.appendChild(spinner);
    }

    spinner.style.display = 'flex';
}

// 隐藏加载动画
function hideLoading() {
    const spinner = document.querySelector('.spinner-overlay');
    if (spinner) {
        spinner.style.display = 'none';
    }
}
