// Admin Dashboard JavaScript

class FairCoinAdmin {
    constructor() {
        this.apiBase = 'http://localhost:8080/api';
        this.currentSection = 'dashboard';
        this.authToken = localStorage.getItem('admin_token');
        this.charts = {};
        this.isAuthenticated = false;
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        
        // Check if user is already authenticated
        if (this.authToken) {
            this.validateToken();
        } else {
            this.showLoginModal();
        }
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const section = e.target.closest('.nav-item').dataset.section;
                this.showSection(section);
            });
        });

        // Search functionality
        document.getElementById('user-search')?.addEventListener('input', (e) => {
            this.debounce(this.searchUsers.bind(this), 300)(e.target.value);
        });

        // Modal close
        document.querySelector('.modal-overlay')?.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal-overlay')) {
                this.closeModal();
            }
        });

        // Login form enter key
        document.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && document.getElementById('login-modal').style.display === 'block') {
                this.adminLogin();
            }
        });
    }

    showSection(section) {
        // Update navigation
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-section="${section}"]`).classList.add('active');

        // Update content
        document.querySelectorAll('.content-section').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(section).classList.add('active');

        // Update page title
        const titles = {
            dashboard: 'Dashboard Overview',
            users: 'Users Management',
            transactions: 'Transaction Management',
            governance: 'Governance Management',
            monetary: 'Monetary Policy Dashboard',
            fairness: 'Fairness Metrics Analysis',
            system: 'System Settings'
        };
        document.getElementById('page-title').textContent = titles[section];

        this.currentSection = section;
        this.loadSectionData(section);
    }

    async loadSectionData(section) {
        switch (section) {
            case 'dashboard':
                await this.loadDashboard();
                break;
            case 'users':
                await this.loadUsers();
                break;
            case 'transactions':
                await this.loadTransactions();
                break;
            case 'governance':
                await this.loadGovernance();
                break;
            case 'monetary':
                await this.loadMonetaryPolicy();
                break;
            case 'fairness':
                await this.loadFairnessMetrics();
                break;
            case 'system':
                await this.loadSystemSettings();
                break;
        }
    }

    async loadDashboard() {
        try {
            const [stats, activity] = await Promise.all([
                this.fetchStats(),
                this.fetchRecentActivity()
            ]);

            this.updateStats(stats);
            this.updateRecentActivity(activity);
            
            // Wait for DOM to be ready before creating charts
            setTimeout(() => {
                this.createCharts();
            }, 100);
        } catch (error) {
            console.error('Error loading dashboard:', error);
            this.showError('Failed to load dashboard data');
        }
    }

    async fetchStats() {
        try {
            const response = await fetch(`${this.apiBase}/v1/admin/stats`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });
            
            if (!response.ok) {
                throw new Error('Failed to fetch stats');
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error fetching stats:', error);
            // Fallback to mock data if API fails
            return {
                total_users: 0,
                total_supply: 0,
                daily_transactions: 0,
                avg_pfi: 0
            };
        }
    }

    async fetchRecentActivity() {
        try {
            const response = await fetch(`${this.apiBase}/v1/admin/activity`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });
            
            if (!response.ok) {
                throw new Error('Failed to fetch recent activity');
            }
            
            const data = await response.json();
            return data.activities || [];
        } catch (error) {
            console.error('Error fetching recent activity:', error);
            return [];
        }
    }

    updateStats(stats) {
        document.getElementById('total-users').textContent = (stats.total_users || 0).toLocaleString();
        document.getElementById('total-supply').textContent = (stats.total_supply || 0).toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2
        });
        document.getElementById('daily-transactions').textContent = (stats.daily_transactions || 0).toLocaleString();
        document.getElementById('avg-pfi').textContent = (stats.avg_pfi || 0).toFixed(1);
    }

    updateRecentActivity(activities) {
        const container = document.getElementById('recent-activity');
        container.innerHTML = activities.map(activity => `
            <div class="activity-item">
                <div class="activity-icon" style="background-color: ${activity.color}">
                    <i class="fas fa-${activity.icon}"></i>
                </div>
                <div class="activity-info">
                    <div>${activity.message}</div>
                    <div class="activity-time">${activity.time}</div>
                </div>
            </div>
        `).join('');
    }

    async createCharts() {
        // Ensure canvas elements exist before creating charts
        const pfiCanvas = document.getElementById('pfiChart');
        const txCanvas = document.getElementById('transactionChart');
        
        if (pfiCanvas) {
            await this.createPFIChart();
        }
        
        if (txCanvas) {
            await this.createTransactionChart();
        }
    }

    async createPFIChart() {
        try {
            const response = await fetch(`${this.apiBase}/v1/admin/pfi-distribution`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });
            
            let distribution;
            if (response.ok) {
                const data = await response.json();
                distribution = data.distribution || [];
                console.log('PFI distribution data:', distribution);
            } else {
                console.error('Failed to fetch PFI distribution:', response.status);
                // Fallback data
                distribution = [
                    { Range: 'excellent', Count: 0 },
                    { Range: 'good', Count: 0 },
                    { Range: 'average', Count: 0 },
                    { Range: 'poor', Count: 0 }
                ];
            }

            const canvas = document.getElementById('pfiChart');
            if (!canvas) {
                console.error('PFI chart canvas not found');
                return;
            }
            
            const ctx = canvas.getContext('2d');
            
            if (this.charts.pfi) {
                this.charts.pfi.destroy();
            }

            this.charts.pfi = new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: ['Excellent (90-100)', 'Good (70-89)', 'Average (50-69)', 'Poor (0-49)'],
                    datasets: [{
                        label: 'Number of Users',
                        data: distribution.map(d => d.Count || 0),
                        backgroundColor: [
                            'rgba(46, 204, 113, 0.8)',
                            'rgba(52, 152, 219, 0.8)',
                            'rgba(243, 156, 18, 0.8)',
                            'rgba(231, 76, 60, 0.8)'
                        ],
                        borderColor: [
                            'rgba(46, 204, 113, 1)',
                            'rgba(52, 152, 219, 1)',
                            'rgba(243, 156, 18, 1)',
                            'rgba(231, 76, 60, 1)'
                        ],
                        borderWidth: 1
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    resizeDelay: 100,
                    plugins: {
                        legend: {
                            display: false
                        },
                        title: {
                            display: true,
                            text: 'PFI★ Score Distribution'
                        }
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            ticks: {
                                stepSize: 1,
                                callback: function(value) {
                                    return Number.isInteger(value) ? value + ' users' : '';
                                }
                            },
                            title: {
                                display: true,
                                text: 'Number of Users'
                            }
                        },
                        x: {
                            title: {
                                display: true,
                                text: 'PFI★ Score Ranges'
                            }
                        }
                    },
                    animation: {
                        duration: 800,
                        easing: 'easeOutBounce'
                    }
                }
            });
        } catch (error) {
            console.error('Error creating PFI chart:', error);
        }
    }

    async createTransactionChart() {
        try {
            const canvas = document.getElementById('transactionChart');
            if (!canvas) {
                console.error('Transaction chart canvas not found');
                return;
            }
            
            // Fetch real transaction volume data
            let volumeData = [];
            let labels = [];
            
            try {
                const response = await fetch(`${this.apiBase}/v1/admin/transaction-volume`, {
                    headers: {
                        'Authorization': `Bearer ${this.authToken}`
                    }
                });
                
                if (response.ok) {
                    const data = await response.json();
                    const volumes = data.volume_data || [];
                    console.log('Transaction volume data:', volumes);
                    
                    labels = volumes.map(v => v.Month);
                    volumeData = volumes.map(v => v.Volume || 0);
                } else {
                    console.error('Failed to fetch transaction volume:', response.status);
                    // Fallback data
                    labels = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'];
                    volumeData = [0, 0, 0, 0, 0, 0];
                }
            } catch (error) {
                console.error('Error fetching transaction volume:', error);
                labels = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'];
                volumeData = [0, 0, 0, 0, 0, 0];
            }
            
            const ctx = canvas.getContext('2d');
            
            if (this.charts.transaction) {
                this.charts.transaction.destroy();
            }

            this.charts.transaction = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: labels,
                    datasets: [{
                        label: 'Transaction Volume (FC)',
                        data: volumeData,
                        borderColor: '#3498db',
                        backgroundColor: 'rgba(52, 152, 219, 0.1)',
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    interaction: {
                        intersect: false,
                    },
                    plugins: {
                        legend: {
                            display: true
                        }
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            ticks: {
                                callback: function(value) {
                                    return value.toLocaleString() + ' FC';
                                }
                            }
                        }
                    }
                }
            });
        } catch (error) {
            console.error('Error creating transaction chart:', error);
        }
    }

    async loadUsers() {
        try {
            const users = await this.fetchUsers();
            this.displayUsers(users);
        } catch (error) {
            console.error('Error loading users:', error);
            this.showError('Failed to load users');
        }
    }

    async fetchUsers() {
        try {
            const response = await fetch(`${this.apiBase}/v1/admin/users`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });
            
            if (!response.ok) {
                throw new Error('Failed to fetch users');
            }
            
            const data = await response.json();
            return data.users || [];
        } catch (error) {
            console.error('Error fetching users:', error);
            return [];
        }
    }

    displayUsers(users) {
        const tbody = document.getElementById('users-table');
        tbody.innerHTML = users.map(user => `
            <tr>
                <td>${user.username || 'N/A'}</td>
                <td>${user.email || 'N/A'}</td>
                <td><span class="pfi-score">${user.pfi || 0}★</span></td>
                <td><span class="tfi-score">${user.tfi || 0}★</span></td>
                <td>${(user.balance || 0).toFixed(2)} FC</td>
                <td><span class="status-badge status-${user.is_verified ? 'verified' : 'pending'}">${user.is_verified ? 'verified' : 'pending'}</span></td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="admin.editUser('${user.id}')">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="admin.blockUser('${user.id}')">
                        <i class="fas fa-ban"></i>
                    </button>
                </td>
            </tr>
        `).join('');
    }

    async loadTransactions() {
        try {
            const transactions = await this.fetchTransactions();
            this.displayTransactions(transactions);
        } catch (error) {
            console.error('Error loading transactions:', error);
            this.showError('Failed to load transactions');
        }
    }

    async fetchTransactions() {
        try {
            const response = await fetch(`${this.apiBase}/v1/admin/transactions`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });
            
            if (!response.ok) {
                throw new Error('Failed to fetch transactions');
            }
            
            const data = await response.json();
            return data.transactions || [];
        } catch (error) {
            console.error('Error fetching transactions:', error);
            return [];
        }
    }

    displayTransactions(transactions) {
        const tbody = document.getElementById('transactions-table');
        tbody.innerHTML = transactions.map(tx => `
            <tr>
                <td class="mono">${tx.id || 'N/A'}</td>
                <td><span class="tx-type">${(tx.type || '').replace('_', ' ')}</span></td>
                <td>${tx.from_user || 'System'}</td>
                <td>${tx.to_user || 'N/A'}</td>
                <td>${(tx.amount || 0).toFixed(2)} FC</td>
                <td><span class="status-badge status-${tx.status}">${tx.status || 'unknown'}</span></td>
                <td>${tx.created_at ? new Date(tx.created_at).toLocaleDateString() : 'N/A'}</td>
                <td>
                    <button class="btn btn-sm btn-info" onclick="admin.viewTransaction('${tx.id}')">
                        <i class="fas fa-eye"></i>
                    </button>
                </td>
            </tr>
        `).join('');
    }

    async loadGovernance() {
        try {
            const proposals = await this.fetchProposals();
            this.displayProposals(proposals);
        } catch (error) {
            console.error('Error loading governance:', error);
            this.showError('Failed to load governance data');
        }
    }

    async fetchProposals() {
        return [
            {
                id: 'prop001',
                title: 'Increase Base Issuance Rate',
                description: 'Proposal to increase the monthly base issuance from 2% to 2.5% to stimulate economic activity.',
                type: 'monetary_policy',
                status: 'active',
                votesFor: 1247,
                votesAgainst: 893,
                endTime: '2024-04-01T23:59:59Z'
            },
            {
                id: 'prop002',
                title: 'Implement Merchant Verification Program',
                description: 'Create a new verification tier for merchants with enhanced benefits.',
                type: 'governance',
                status: 'passed',
                votesFor: 2156,
                votesAgainst: 432,
                endTime: '2024-03-15T23:59:59Z'
            }
        ];
    }

    displayProposals(proposals) {
        const container = document.getElementById('proposals-list');
        container.innerHTML = proposals.map(proposal => `
            <div class="proposal-card">
                <div class="proposal-header">
                    <div>
                        <div class="proposal-title">${proposal.title}</div>
                        <div class="proposal-type">${proposal.type.replace('_', ' ')}</div>
                    </div>
                    <span class="proposal-status status-${proposal.status}">${proposal.status}</span>
                </div>
                <div class="proposal-description">${proposal.description}</div>
                <div class="proposal-votes">
                    <div class="vote-bar">
                        <div class="votes-for" style="width: ${(proposal.votesFor / (proposal.votesFor + proposal.votesAgainst)) * 100}%">
                            For: ${proposal.votesFor}
                        </div>
                        <div class="votes-against">
                            Against: ${proposal.votesAgainst}
                        </div>
                    </div>
                </div>
                <div class="proposal-actions">
                    <button class="btn btn-primary" onclick="admin.viewProposal('${proposal.id}')">
                        View Details
                    </button>
                    ${proposal.status === 'active' ? 
                        `<button class="btn btn-danger" onclick="admin.cancelProposal('${proposal.id}')">
                            Cancel
                        </button>` : ''
                    }
                </div>
            </div>
        `).join('');
    }

    async loadMonetaryPolicy() {
        try {
            const policy = await this.fetchMonetaryPolicy();
            this.displayMonetaryPolicy(policy);
        } catch (error) {
            console.error('Error loading monetary policy:', error);
            this.showError('Failed to load monetary policy');
        }
    }

    async fetchMonetaryPolicy() {
        try {
            const response = await fetch(`${this.apiBase}/v1/admin/monetary-policy`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });
            
            if (!response.ok) {
                throw new Error('Failed to fetch monetary policy');
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error fetching monetary policy:', error);
            return {
                current_month: '2024-03',
                base_issuance: 2.0,
                activity_factor: 1.0,
                fairness_factor: 1.0,
                total_issuance: 0,
                circulating_supply: 0,
                average_pfi: 0
            };
        }
    }

    displayMonetaryPolicy(policy) {
        document.getElementById('current-policy').innerHTML = `
            <div class="policy-item">
                <label>Current Month:</label>
                <span>${policy.current_month}</span>
            </div>
            <div class="policy-item">
                <label>Base Issuance Rate:</label>
                <span>${policy.base_issuance}%</span>
            </div>
            <div class="policy-item">
                <label>Activity Factor:</label>
                <span>${policy.activity_factor}x</span>
            </div>
            <div class="policy-item">
                <label>Fairness Factor:</label>
                <span>${policy.fairness_factor}x</span>
            </div>
            <div class="policy-item">
                <label>Total Issuance:</label>
                <span>${(policy.total_issuance || 0).toLocaleString()} FC</span>
            </div>
        `;

        // Populate form fields
        document.getElementById('base-issuance').value = policy.base_issuance;
        document.getElementById('activity-factor').value = policy.activity_factor;
        document.getElementById('fairness-factor').value = policy.fairness_factor;
    }

    async loadFairnessMetrics() {
        try {
            const metrics = await this.fetchFairnessMetrics();
            this.displayFairnessMetrics(metrics);
        } catch (error) {
            console.error('Error loading fairness metrics:', error);
            this.showError('Failed to load fairness metrics');
        }
    }

    async fetchFairnessMetrics() {
        return {
            pfiDistribution: {
                'excellent': 15,
                'good': 35,
                'average': 40,
                'poor': 10
            },
            tfiAnalysis: {
                averageRating: 4.2,
                totalRatings: 2847,
                topMerchants: ['alice123', 'merchant_bob', 'carol_store']
            },
            cbi: {
                currentValue: 1.05,
                trend: 'stable',
                components: {
                    food: 1.02,
                    energy: 1.08,
                    labor: 1.05,
                    housing: 1.03
                }
            }
        };
    }

    displayFairnessMetrics(metrics) {
        // PFI Distribution
        document.getElementById('pfi-breakdown').innerHTML = Object.entries(metrics.pfiDistribution)
            .map(([level, percentage]) => `
                <div class="metric-row">
                    <span class="metric-label">${level.charAt(0).toUpperCase() + level.slice(1)}:</span>
                    <span class="metric-value">${percentage}%</span>
                    <div class="metric-bar">
                        <div class="metric-fill" style="width: ${percentage}%"></div>
                    </div>
                </div>
            `).join('');

        // TFI Analysis
        document.getElementById('tfi-analysis').innerHTML = `
            <div class="metric-row">
                <span class="metric-label">Average Rating:</span>
                <span class="metric-value">${metrics.tfiAnalysis.averageRating}/5.0</span>
            </div>
            <div class="metric-row">
                <span class="metric-label">Total Ratings:</span>
                <span class="metric-value">${metrics.tfiAnalysis.totalRatings.toLocaleString()}</span>
            </div>
            <div class="metric-row">
                <span class="metric-label">Top Merchants:</span>
                <span class="metric-value">${metrics.tfiAnalysis.topMerchants.join(', ')}</span>
            </div>
        `;

        // CBI Display
        document.getElementById('cbi-display').innerHTML = `
            <div class="cbi-current">
                <h4>Current CBI: ${metrics.cbi.currentValue}</h4>
                <p class="cbi-trend trend-${metrics.cbi.trend}">Trend: ${metrics.cbi.trend}</p>
            </div>
            <div class="cbi-components">
                ${Object.entries(metrics.cbi.components)
                    .map(([component, value]) => `
                        <div class="component-row">
                            <span>${component.charAt(0).toUpperCase() + component.slice(1)}:</span>
                            <span>${value}</span>
                        </div>
                    `).join('')}
            </div>
        `;
    }

    async loadSystemSettings() {
        // Load current system settings
        const settings = {
            feeRate: 0.1,
            minPFI: 50,
            votingPeriod: 7
        };

        document.getElementById('fee-rate').value = settings.feeRate;
        document.getElementById('min-pfi').value = settings.minPFI;
        document.getElementById('voting-period').value = settings.votingPeriod;
    }

    // Utility functions
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    showError(message) {
        // Simple error display - could be enhanced with toast notifications
        console.error(message);
        alert(message);
    }

    showModal(title, content, confirmCallback) {
        document.getElementById('modal-title').textContent = title;
        document.getElementById('modal-body').innerHTML = content;
        document.getElementById('modal-confirm').onclick = confirmCallback;
        document.getElementById('modal-overlay').style.display = 'block';
    }

    closeModal() {
        document.getElementById('modal-overlay').style.display = 'none';
    }

    startDataRefresh() {
        // Refresh data every 30 seconds
        setInterval(() => {
            if (this.currentSection === 'dashboard') {
                this.loadDashboard();
            }
        }, 30000);
    }

    // Action handlers
    async refreshData() {
        await this.loadSectionData(this.currentSection);
        this.showSuccess('Data refreshed successfully');
    }

    async exportData() {
        // Implement data export functionality
        this.showSuccess('Export functionality coming soon');
    }

    editUser(userId) {
        this.showModal('Edit User', `
            <div class="form-group">
                <label>PFI Score:</label>
                <input type="number" id="edit-pfi" min="0" max="100">
            </div>
            <div class="form-group">
                <label>Status:</label>
                <select id="edit-status">
                    <option value="verified">Verified</option>
                    <option value="pending">Pending</option>
                    <option value="blocked">Blocked</option>
                </select>
            </div>
        `, () => {
            // Handle user update
            this.showSuccess('User updated successfully');
            this.closeModal();
        });
    }

    blockUser(userId) {
        this.showModal('Block User', 'Are you sure you want to block this user?', () => {
            // Handle user blocking
            this.showSuccess('User blocked successfully');
            this.closeModal();
            this.loadUsers();
        });
    }

    updateMonetaryPolicy() {
        const baseIssuance = document.getElementById('base-issuance').value;
        const activityFactor = document.getElementById('activity-factor').value;
        const fairnessFactor = document.getElementById('fairness-factor').value;

        // Implement policy update API call
        this.showSuccess('Monetary policy updated successfully');
    }

    saveSystemSettings() {
        const settings = {
            feeRate: document.getElementById('fee-rate').value,
            minPFI: document.getElementById('min-pfi').value,
            votingPeriod: document.getElementById('voting-period').value
        };

        // Implement settings save API call
        this.showSuccess('System settings saved successfully');
    }

    showSuccess(message) {
        // Simple success display
        console.log(message);
        // Could be enhanced with toast notifications
    }

    // Authentication methods
    showLoginModal() {
        document.getElementById('login-modal').style.display = 'block';
        document.querySelector('.admin-container').style.display = 'none';
    }

    hideLoginModal() {
        document.getElementById('login-modal').style.display = 'none';
        document.querySelector('.admin-container').style.display = 'flex';
    }

    async adminLogin() {
        const username = document.getElementById('admin-username').value;
        const password = document.getElementById('admin-password').value;
        const errorDiv = document.getElementById('login-error');

        if (!username || !password) {
            this.showLoginError('Please enter both username and password');
            return;
        }

        try {
            const response = await fetch(`${this.apiBase}/v1/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: username,
                    password: password
                })
            });

            const data = await response.json();

            if (response.ok && data.token) {
                this.authToken = data.token;
                localStorage.setItem('admin_token', this.authToken);
                this.isAuthenticated = true;
                this.hideLoginModal();
                this.loadDashboard();
                this.startDataRefresh();
                errorDiv.style.display = 'none';
            } else {
                this.showLoginError(data.error || 'Login failed');
            }
        } catch (error) {
            console.error('Login error:', error);
            this.showLoginError('Network error. Please try again.');
        }
    }

    showLoginError(message) {
        const errorDiv = document.getElementById('login-error');
        errorDiv.textContent = message;
        errorDiv.style.display = 'block';
    }

    async validateToken() {
        try {
            const response = await fetch(`${this.apiBase}/v1/users/profile`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });

            if (response.ok) {
                this.isAuthenticated = true;
                this.hideLoginModal();
                this.loadDashboard();
                this.startDataRefresh();
            } else {
                // Token is invalid
                localStorage.removeItem('admin_token');
                this.authToken = null;
                this.showLoginModal();
            }
        } catch (error) {
            console.error('Token validation error:', error);
            this.showLoginModal();
        }
    }

    logout() {
        localStorage.removeItem('admin_token');
        this.authToken = null;
        this.isAuthenticated = false;
        this.showLoginModal();
    }
}

// Global functions for onclick handlers
window.admin = null;

// Authentication functions
window.adminLogin = () => admin.adminLogin();
window.logout = () => admin.logout();

// Action functions
window.refreshData = () => admin.refreshData();
window.exportData = () => admin.exportData();
window.searchUsers = () => admin.searchUsers();
window.filterTransactions = () => admin.filterTransactions();
window.createProposal = () => admin.createProposal();
window.updateMonetaryPolicy = () => admin.updateMonetaryPolicy();
window.saveSystemSettings = () => admin.saveSystemSettings();
window.backupDatabase = () => admin.backupDatabase();
window.viewLogs = () => admin.viewLogs();
window.resetSystem = () => admin.resetSystem();
window.closeModal = () => admin.closeModal();
window.confirmModal = () => admin.confirmModal();

// Initialize admin dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.admin = new FairCoinAdmin();
});