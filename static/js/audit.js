document.addEventListener('DOMContentLoaded', function () {
    // 检查登录状态
    const token = localStorage.getItem('token');
    const username = localStorage.getItem('username');

    console.log('审核页面加载，当前用户:', username);

    if (!token || username !== '监管者') {
        console.log('未登录或非监管者用户，重定向到登录页');
        window.location.href = '/login'; // 修改为正确的登录页路径
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
        window.location.href = '/login'; // 确保退出后跳转到正确的登录页
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

// 渲染项目列表 - 使用列表形式而非卡片
function renderItems(containerId, items, type) {
    const container = document.getElementById(containerId);
    container.innerHTML = ''; // 清空容器

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
                ${type === 'pending'
                ? `<button class="btn btn-primary btn-sm audit-item" data-id="${item.id}" data-trans-id="${item.firstTransID}">审核</button>`
                : `<div class="btn-group">
                    <button class="btn btn-info btn-sm view-details" data-id="${item.id}" data-trans-id="${item.firstTransID}">查看详情</button>
                    <button class="btn btn-secondary btn-sm view-history" data-trans-id="${item.firstTransID}">审核历史</button>
                   </div>`
            }
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
            const transId = this.getAttribute('data-trans-id');
            viewItemDetails(itemId, transId);
        });
    });

    container.querySelectorAll('.audit-item').forEach(btn => {
        btn.addEventListener('click', function () {
            const itemId = this.getAttribute('data-id');
            const transId = this.getAttribute('data-trans-id');
            showAuditModal(itemId, transId);
        });
    });

    // 添加审核历史按钮事件监听
    container.querySelectorAll('.view-history').forEach(btn => {
        btn.addEventListener('click', function () {
            const transId = this.getAttribute('data-trans-id');
            showAuditHistory(transId);
        });
    });
}

// 查看项目详情 - 修改为使用弹窗显示而非页面跳转
function viewItemDetails(itemId, transId) {
    showLoading();

    // 获取项目详情
    fetch(`/api/audit/info?tradeId=${transId}`, {
        method: 'GET',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        }
    })
        .then(response => response.json())
        .then(data => {
            hideLoading();

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

                // 修改模态窗标题
                document.getElementById('auditModalTitle').textContent = '版权详情';

                // 隐藏审核表单和按钮
                document.getElementById('auditFormSection').style.display = 'none';
                document.getElementById('auditButtonsSection').style.display = 'none';

                // 显示弹窗
                const modal = new bootstrap.Modal(document.getElementById('auditModal'));
                modal.show();
            } else {
                alert('获取项目详情失败');
            }
        })
        .catch(error => {
            hideLoading();
            console.error('获取项目详情错误:', error);
            alert('获取项目详情失败: ' + error.message);
        });
}

// 显示审核弹窗
function showAuditModal(itemId, transId) {
    showLoading();

    // 获取项目详情
    fetch(`/api/audit/info?tradeId=${transId}`, {
        method: 'GET',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        }
    })
        .then(response => response.json())
        .then(data => {
            hideLoading();

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

                // 修改模态窗标题
                document.getElementById('auditModalTitle').textContent = '版权审核';

                // 清空表单
                document.getElementById('auditComment').value = '';
                document.getElementById('regulatorPassword').value = '';

                // 显示审核表单和按钮
                document.getElementById('auditFormSection').style.display = 'block';
                document.getElementById('auditButtonsSection').style.display = 'flex';

                // 显示弹窗
                const modal = new bootstrap.Modal(document.getElementById('auditModal'));
                modal.show();
            } else {
                alert('获取项目详情失败');
            }
        })
        .catch(error => {
            hideLoading();
            console.error('获取项目详情错误:', error);
            alert('获取项目详情失败: ' + error.message);
        });
}

// 显示审核历史
function showAuditHistory(transId) {
    showLoading();

    fetch(`/api/audit/history?tradeId=${transId}`, {
        method: 'GET',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        }
    })
        .then(response => response.json())
        .then(data => {
            hideLoading();

            if (data.records && Array.isArray(data.records)) {
                // 创建审核历史内容
                let historyContent = '<div class="list-group">';

                if (data.records.length === 0) {
                    historyContent += '<div class="text-center py-3 text-muted">暂无审核记录</div>';
                } else {
                    data.records.forEach((record, index) => {
                        const recordDate = new Date(record.timestamp * 1000); // 转换为毫秒
                        // 修改为精确到秒的完整时间格式
                        const formattedDate = recordDate.toLocaleString('zh-CN');
                        const badgeClass = record.decision === 'APPROVE' ? 'bg-success' : 'bg-danger';
                        const decision = record.decision === 'APPROVE' ? '通过' : '拒绝';

                        historyContent += `
                            <div class="list-group-item">
                                <div class="d-flex w-100 justify-content-between">
                                    <h5 class="mb-1">审核记录 #${index + 1}</h5>
                                    <small>${formattedDate}</small>
                                </div>
                                <p class="mb-1">${record.comment || '无审核意见'}</p>
                                <span class="badge ${badgeClass}">${decision}</span>
                            </div>
                        `;
                    });
                }

                historyContent += '</div>';

                // 使用Bootstrap模态框显示审核历史
                const modalTitle = `交易 ${transId} 的审核历史`;
                showHistoryModal(modalTitle, historyContent);
            } else {
                alert('获取审核历史失败');
            }
        })
        .catch(error => {
            hideLoading();
            console.error('获取审核历史错误:', error);
            alert('获取审核历史失败: ' + error.message);
        });
}

// 显示审核历史模态框 - 修复Bootstrap模态框实例化问题
function showHistoryModal(title, content) {
    // 如果已存在模态框，先移除
    let existingModal = document.getElementById('historyModal');
    if (existingModal) {
        const bsModal = bootstrap.Modal.getInstance(existingModal);
        if (bsModal) {
            bsModal.dispose();
        }
        existingModal.remove();
    }

    // 创建模态框
    const modalHTML = `
        <div class="modal fade" id="historyModal" tabindex="-1" aria-labelledby="historyModalLabel" aria-hidden="true">
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title" id="historyModalLabel">${title}</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        ${content}
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">关闭</button>
                    </div>
                </div>
            </div>
        </div>
    `;

    // 添加模态框到页面
    document.body.insertAdjacentHTML('beforeend', modalHTML);

    // 获取新创建的模态框元素
    const modalElement = document.getElementById('historyModal');

    // 等待DOM更新后再显示模态框
    setTimeout(() => {
        try {
            const modal = new bootstrap.Modal(modalElement);
            modal.show();
        } catch (error) {
            console.error('显示模态框出错:', error);
            alert('无法显示审核历史，请刷新页面重试');
        }
    }, 50);
}

// 获取状态标签 - 改为文本标签
function getStatusBadge(type) {
    switch (type) {
        case 'pending':
            return '<span class="status-badge pending">待审核</span>';
        case 'approved':
            return '<span class="status-badge approved">已通过</span>';
        case 'rejected':
            return '<span class="status-badge rejected">未通过</span>';
        default:
            return '';
    }
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

// 切换空状态显示
function toggleEmptyState(elementId, isEmpty) {
    const element = document.getElementById(elementId);
    if (isEmpty) {
        element.classList.remove('d-none');
    } else {
        element.classList.add('d-none');
    }
}

// 修改格式化时间函数，只显示年月日
function formatTime(timestamp) {
    if (!timestamp) return '未知时间';

    try {
        const date = new Date(timestamp);
        // 只返回年月日，不返回具体时间
        return date.toLocaleDateString('zh-CN');
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
