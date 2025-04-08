// Modal类 - 修复滚动问题
class Modal {
    constructor(element) {
        this.element = element;
        this.isVisible = false;
    }

    show() {
        this.element.classList.add('show');
        document.body.style.overflow = 'hidden'; // 防止背景滚动
        this.isVisible = true;
    }

    hide() {
        this.element.classList.remove('show');
        document.body.style.overflow = ''; // 恢复滚动
        this.isVisible = false;
    }

    toggle() {
        if (this.isVisible) {
            this.hide();
        } else {
            this.show();
        }
    }

    static getInstance(element) {
        return new Modal(element);
    }
}

// TabManager类 - 修改为使用nav-item类
class TabManager {
    constructor() {
        this.initTabs();
    }

    initTabs() {
        const tabs = document.querySelectorAll('[data-toggle="tab"]');
        tabs.forEach(tab => {
            tab.addEventListener('click', (event) => {
                event.preventDefault();
                this.activateTab(tab);
            });
        });
    }

    activateTab(tabElement) {
        const targetId = tabElement.getAttribute('href');

        // 移除所有活动标签
        document.querySelectorAll('.tab-pane').forEach(pane => {
            pane.classList.remove('active');
        });
        document.querySelectorAll('.nav-item').forEach(link => {
            link.classList.remove('active');
        });

        // 激活当前标签
        document.querySelector(targetId).classList.add('active');
        tabElement.classList.add('active');
    }
}

// 设置响应式导航栏
function setupMobileNav() {
    const toggler = document.getElementById('navbarToggler');
    const collapse = document.getElementById('navbarNav');

    if (toggler && collapse) {
        toggler.addEventListener('click', () => {
            collapse.classList.toggle('show');
        });
    }
}

// 初始化模态框事件监听
function initModalEvents() {
    // 为所有关闭按钮添加事件
    document.querySelectorAll('.modal-close').forEach(button => {
        const modalId = button.closest('.modal').id;
        button.addEventListener('click', () => {
            document.getElementById(modalId).classList.remove('show');
        });
    });

    // 点击模态框背景关闭
    document.querySelectorAll('.modal-backdrop').forEach(backdrop => {
        backdrop.addEventListener('click', (event) => {
            if (event.target === backdrop) {
                const modal = backdrop.closest('.modal');
                modal.classList.remove('show');
            }
        });
    });
}

// 显示自定义提示框
function showCustomAlert(message, type = 'success', duration = 2000) {
    const alert = document.createElement('div');
    alert.className = `custom-alert ${type}`;
    alert.innerHTML = `
        <div class="alert-icon-wrapper">
            <i class="${type === 'success' ? 'fas fa-check-circle' : 'fas fa-times-circle'}"></i>
        </div>
        <div>${message}</div>
    `;
    document.body.appendChild(alert);

    setTimeout(() => alert.classList.add('show'), 10);

    setTimeout(() => {
        alert.classList.remove('show');
        setTimeout(() => document.body.removeChild(alert), 300);
    }, duration);
}

// 修复关闭模态框时恢复滚动问题
function fixScrollAfterModalClose() {
    // 确保模态框关闭后恢复正常滚动
    document.querySelectorAll('.modal').forEach(modal => {
        const observer = new MutationObserver(mutations => {
            mutations.forEach(mutation => {
                if (mutation.attributeName === 'class' &&
                    !modal.classList.contains('show')) {
                    document.body.style.overflow = '';
                }
            });
        });

        observer.observe(modal, { attributes: true });
    });
}

// 缓存数据，避免重复请求
let cachedData = {
    pendingItems: [],
    approvedItems: [],
    rejectedItems: []
};

// 页面加载完成后执行
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

    // 初始化标签页管理
    new TabManager();

    // 设置移动导航
    setupMobileNav();

    // 初始化模态框事件
    initModalEvents();

    // 修复滚动问题
    fixScrollAfterModalClose();

    // 退出登录
    document.getElementById('logoutBtn').addEventListener('click', function () {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        console.log('登出成功，重定向到登录页');
        window.location.href = '/login'; // 确保退出后跳转到正确的登录页
    });

    // 刷新数据按钮
    document.getElementById('refreshBtn').addEventListener('click', function () {
        showCustomAlert('正在刷新数据...', 'success', 1000);
        loadAllItems();
    });

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

    // 关闭审核模态框的事件
    document.getElementById('closeAuditModal').addEventListener('click', function () {
        const modalElement = document.getElementById('auditModal');
        const modal = Modal.getInstance(modalElement);
        modal.hide();
    });
});

// 修改排序函数，默认按时间降序排序（最新的在最上面）
function sortAndRenderItems(status) {
    if (cachedData[`${status}Items`] && cachedData[`${status}Items`].length > 0) {
        const sortedItems = [...cachedData[`${status}Items`]].sort((a, b) => {
            // 获取审核时间
            const timeA = a.auditTime || 0;
            const timeB = b.auditTime || 0;

            // 默认降序排列（最新的在前）
            return timeB - timeA;
        });

        renderItems(`${status}Items`, sortedItems, status);
    }
}

// 加载所有版权项目
function loadAllItems() {
    showLoading();

    fetch('/api/audit/categorized-items', {  // 使用分类API
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

            // 缓存数据
            cachedData.pendingItems = data.pendingItems || [];
            cachedData.approvedItems = data.approvedItems || [];
            cachedData.rejectedItems = data.rejectedItems || [];

            // 渲染待审核项目（无需排序）
            renderItems('pendingItems', cachedData.pendingItems, 'pending');

            // 加载审核时间戳后再渲染其他项目
            loadAuditTimestamps('approved');
            loadAuditTimestamps('rejected');

            // 检查是否有内容
            toggleEmptyState('pendingEmpty', cachedData.pendingItems.length === 0);
            toggleEmptyState('approvedEmpty', cachedData.approvedItems.length === 0);
            toggleEmptyState('rejectedEmpty', cachedData.rejectedItems.length === 0);
        })
        .catch(error => {
            console.error('获取项目列表错误：', error);
            showCustomAlert('获取版权项目失败：' + error.message, 'error');
        })
        .finally(() => {
            hideLoading();
        });
}

// 加载指定类型项目的审核时间戳
function loadAuditTimestamps(status) {
    const items = cachedData[`${status}Items`];
    if (!items || items.length === 0) return;

    // 创建一个计数器跟踪加载完成的项目数
    let loadedCount = 0;
    const totalItems = items.length;

    items.forEach(item => {
        if (!item.firstTransID) {
            loadedCount++;
            return;
        }

        // 获取审核历史记录
        fetch(`/api/audit/history?tradeId=${item.firstTransID}`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
            .then(response => response.json())
            .then(data => {
                // 找到最新的审核记录
                if (data.records && data.records.length > 0) {
                    // 记录中的时间戳通常是秒，转换为毫秒
                    const latestRecord = data.records[0]; // 假设记录按时间降序排列
                    item.auditTime = latestRecord.timestamp * 1000; // 转为毫秒
                    item.auditDecision = latestRecord.decision;
                    item.auditComment = latestRecord.comment;
                }
            })
            .catch(error => {
                console.error(`获取项目 ${item.id} 的审核历史失败:`, error);
            })
            .finally(() => {
                loadedCount++;

                // 当所有项目都加载完成后，应用排序并渲染
                if (loadedCount === totalItems) {
                    sortAndRenderItems(status);
                }
            });
    });

    // 如果没有项目需要加载，立即渲染
    if (totalItems === 0) {
        renderItems(`${status}Items`, [], status);
    }
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
                ? `<button class="btn btn-primary btn-sm audit-item" data-id="${item.id}" data-trans-id="${item.firstTransID}"><i class="fas fa-gavel"></i> 审核</button>`
                : `<div class="btn-group">
                    <button class="btn btn-info btn-sm view-details" data-id="${item.id}" data-trans-id="${item.firstTransID}"><i class="fas fa-info-circle"></i> 详情</button>
                    <button class="btn btn-secondary btn-sm view-history" data-trans-id="${item.firstTransID}"><i class="fas fa-history"></i> 历史</button>
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
                    imgElement.src = 'https://via.placeholder.com/200x120/283250/a0aec0?text=暂无图片';
                    imgElement.style.display = 'block';
                }

                // 修改模态窗标题
                document.getElementById('auditModalTitle').textContent = '版权详情';

                // 隐藏审核表单和按钮
                document.getElementById('auditFormSection').style.display = 'none';
                document.getElementById('auditButtonsSection').style.display = 'none';

                // 显示弹窗
                const modalElement = document.getElementById('auditModal');
                const modal = Modal.getInstance(modalElement);
                modal.show();
            } else {
                showCustomAlert('获取项目详情失败', 'error');
            }
        })
        .catch(error => {
            hideLoading();
            console.error('获取项目详情错误:', error);
            showCustomAlert('获取项目详情失败: ' + error.message, 'error');
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
                    imgElement.src = 'https://via.placeholder.com/200x120/283250/a0aec0?text=暂无图片';
                    imgElement.style.display = 'block';
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
                const modalElement = document.getElementById('auditModal');
                const modal = Modal.getInstance(modalElement);
                modal.show();
            } else {
                showCustomAlert('获取项目详情失败', 'error');
            }
        })
        .catch(error => {
            hideLoading();
            console.error('获取项目详情错误:', error);
            showCustomAlert('获取项目详情失败: ' + error.message, 'error');
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
                        const decision = record.decision === 'APPROVE' ? '通过' : '拒绝';
                        const badgeClass = record.decision === 'APPROVE' ? 'approved' : 'rejected';

                        historyContent += `
                            <div class="list-group-item">
                                <div class="d-flex w-100 justify-content-between">
                                    <h5 class="mb-1">审核记录 #${index + 1}</h5>
                                    <small>${formattedDate}</small>
                                </div>
                                <p class="mb-1">${record.comment || '无审核意见'}</p>
                                <span class="status-badge ${badgeClass}">${decision}</span>
                            </div>
                        `;
                    });
                }

                historyContent += '</div>';

                // 创建并显示历史记录模态框
                const modalTitle = `交易 ${transId} 的审核历史`;
                showHistoryModal(modalTitle, historyContent);
            } else {
                showCustomAlert('获取审核历史失败', 'error');
            }
        })
        .catch(error => {
            hideLoading();
            console.error('获取审核历史错误:', error);
            showCustomAlert('获取审核历史失败: ' + error.message, 'error');
        });
}

// 显示审核历史模态框 - 修复滚动问题
function showHistoryModal(title, content) {
    // 如果已存在模态框，先移除
    let existingModal = document.getElementById('historyModal');
    if (existingModal) {
        existingModal.remove();
    }

    // 创建模态框
    const modalHTML = `
        <div class="modal" id="historyModal">
            <div class="modal-backdrop"></div>
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">${title}</h5>
                        <button type="button" class="modal-close" id="closeHistoryModal">&times;</button>
                    </div>
                    <div class="modal-body">
                        ${content}
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" id="closeHistoryBtn">
                            <i class="fas fa-times"></i> 关闭
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `;

    // 添加模态框到页面
    document.body.insertAdjacentHTML('beforeend', modalHTML);

    // 获取新创建的模态框元素
    const modalElement = document.getElementById('historyModal');

    // 添加关闭事件 - 修复滚动问题
    document.getElementById('closeHistoryModal').addEventListener('click', function () {
        modalElement.classList.remove('show');
        setTimeout(() => {
            document.body.style.overflow = '';
        }, 300);
    });

    document.getElementById('closeHistoryBtn').addEventListener('click', function () {
        modalElement.classList.remove('show');
        setTimeout(() => {
            document.body.style.overflow = '';
        }, 300);
    });

    // 点击背景关闭 - 修复滚动问题
    modalElement.querySelector('.modal-backdrop').addEventListener('click', function (event) {
        if (event.target === this) {
            modalElement.classList.remove('show');
            setTimeout(() => {
                document.body.style.overflow = '';
            }, 300);
        }
    });

    // 等待DOM更新后再显示模态框
    setTimeout(() => {
        try {
            modalElement.classList.add('show');
        } catch (error) {
            console.error('显示模态框出错:', error);
            showCustomAlert('无法显示审核历史，请刷新页面重试', 'error');
        }
    }, 50);
}

// 获取状态标签 - 改为文本标签
function getStatusBadge(type) {
    switch (type) {
        case 'pending':
            return '<span class="status-badge pending"><i class="fas fa-clock"></i> 待审核</span>';
        case 'approved':
            return '<span class="status-badge approved"><i class="fas fa-check-circle"></i> 已通过</span>';
        case 'rejected':
            return '<span class="status-badge rejected"><i class="fas fa-times-circle"></i> 未通过</span>';
        default:
            return '';
    }
}

// 提交审核决定 - 修复模态框关闭后的滚动问题
function submitAudit(decision) {
    const tradeId = document.getElementById('modalItemId').value;
    const comment = document.getElementById('auditComment').value;
    const password = document.getElementById('regulatorPassword').value;

    if (!comment) {
        showCustomAlert('请输入审核意见', 'error');
        return;
    }

    if (!password) {
        showCustomAlert('请输入监管者密码', 'error');
        return;
    }

    showLoading();

    fetch('/api/audit', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
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
                const modalElement = document.getElementById('auditModal');
                modalElement.classList.remove('show');

                // 恢复滚动
                document.body.style.overflow = '';

                // 显示成功消息
                showCustomAlert(`版权审核${decision === 'APPROVE' ? '通过' : '拒绝'}成功！`, 'success');

                // 重新加载数据
                setTimeout(() => {
                    loadAllItems();
                }, 500);
            } else {
                showCustomAlert('审核失败: ' + (data.message || '未知错误'), 'error');
            }
        })
        .catch(error => {
            console.error('审核提交错误:', error);
            showCustomAlert('审核提交失败: ' + error.message, 'error');
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
            <div class="spinner" role="status">
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

// 导出Modal类，供全局使用
window.Modal = Modal;
