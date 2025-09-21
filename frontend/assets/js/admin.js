// Admin Dashboard JavaScript

class FairCoinAdmin {
    constructor() {
        this.apiBase = 'http://localhost:8080/api';
        this.currentSection = 'dashboard';
        this.authToken = localStorage.getItem('admin_token');
        this.charts = {};
        this.isAuthenticated = false;
        this.currentUser = null;
        this.refreshInterval = null;
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupGlobalErrorHandling();
        
        // Check if user is already authenticated
        if (this.authToken) {
            this.validateToken();
        } else {
            this.showLoginModal();
        }
    }

    setupGlobalErrorHandling() {
        // Intercept all fetch requests to handle 401 responses globally
        const originalFetch = window.fetch;
        window.fetch = async (...args) => {
            const response = await originalFetch(...args);
            
            // If we get a 401 response, auto-logout
            if (response.status === 401) {
                console.warn('Authentication failed - auto-logout triggered');
                this.handleAutoLogout('Session expired or invalid credentials');
            }
            
            return response;
        };
        
        // Handle browser navigation to protected routes
        window.addEventListener('load', () => {
            this.checkInitialAuth();
        });
        
        // Also check immediately
        this.checkInitialAuth();
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

        // Modal close - handle clicking outside modal content (main frontend system)
        document.addEventListener('click', (e) => {
            // Check if click is on a modal background (not modal-content)
            if (e.target.classList.contains('modal')) {
                e.target.classList.remove('active');
            }
        });

        // Login form enter key
        document.addEventListener('keypress', (e) => {
            const loginModal = document.getElementById('login-modal');
            if (e.key === 'Enter' && loginModal && loginModal.classList.contains('active')) {
                this.adminLogin();
            }
        });

        // Setup wallet event listeners
        this.setupWalletEventListeners();
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
            wallet: 'Wallet Management',
            'demo-report': '🌟 FairCoin Demo Success Report',
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
            case 'demo-report':
                await this.loadDemoReport();
                break;
            case 'wallet':
                await this.loadWallet();
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

    async makeAuthenticatedRequest(url, options = {}) {
        if (!this.authToken) {
            this.handleAutoLogout('No authentication token available');
            throw new Error('No authentication token');
        }

        const defaultOptions = {
            headers: {
                'Authorization': `Bearer ${this.authToken}`,
                'Content-Type': 'application/json',
                ...options.headers
            }
        };

        try {
            const response = await fetch(url, { ...options, headers: defaultOptions.headers });
            
            if (response.status === 401) {
                this.handleAutoLogout('Authentication failed - please login again');
                throw new Error('Unauthorized access');
            } else if (response.status === 403) {
                this.handleAutoLogout('Access denied - insufficient privileges');
                throw new Error('Forbidden access');
            }
            
            return response;
        } catch (error) {
            if (error.message.includes('Failed to fetch')) {
                this.showError('Network error - server may be unavailable');
            }
            throw error;
        }
    }

    async fetchStats() {
        try {
            const response = await this.makeAuthenticatedRequest(`${this.apiBase}/v1/admin/stats`);
            
            if (!response.ok) {
                throw new Error('Failed to fetch stats');
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error fetching stats:', error);
            // Return zero values if API fails
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
        tbody.innerHTML = users.map(user => {
            // Build roles display
            const roles = [];
            if (user.is_admin) roles.push('<span class="role-badge admin"><i class="fas fa-crown"></i> Admin</span>');
            if (user.is_merchant) roles.push('<span class="role-badge merchant"><i class="fas fa-store"></i> Merchant</span>');
            if (user.is_verified) roles.push('<span class="role-badge verified"><i class="fas fa-check-circle"></i> Verified</span>');
            
            const rolesHtml = roles.length > 0 ? roles.join(' ') : '<span class="role-badge regular">Regular User</span>';
            
            return `
                <tr>
                    <td>${user.username || 'N/A'}</td>
                    <td>${user.email || 'N/A'}</td>
                    <td><span class="pfi-score ${this.getPFICategoryClass(user.pfi)}">${user.pfi || 0}★</span></td>
                    <td><span class="tfi-score ${this.getTFICategoryClass(user.tfi)}">${user.tfi || 0}★</span></td>
                    <td>${(user.balance || 0).toFixed(2)} FC</td>
                    <td>${rolesHtml}</td>
                    <td><span class="status-badge status-${user.is_verified ? 'verified' : 'pending'}">${user.is_verified ? 'verified' : 'pending'}</span></td>
                    <td>
                        <button class="btn btn-sm btn-primary" onclick="admin.editUserRoles('${user.id}', ${JSON.stringify(user).replace(/"/g, '&quot;')})" title="Edit Roles">
                            <i class="fas fa-user-cog"></i>
                        </button>
                        <button class="btn btn-sm btn-info" onclick="admin.editUser('${user.id}')" title="Edit User">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn btn-sm btn-warning" onclick="admin.toggleUserStatus('${user.id}', ${!user.is_verified})" title="Toggle Verification">
                            <i class="fas fa-${user.is_verified ? 'times' : 'check'}"></i>
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
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
        try {
            const response = await fetch(`${this.apiBase}/v1/governance/proposals`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            return data.proposals || [];
        } catch (error) {
            console.error('Error fetching proposals:', error);
            this.showError('Failed to fetch proposals from database');
            return [];
        }
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
                        <div class="votes-for" style="width: ${(proposal.votes_for / (proposal.votes_for + proposal.votes_against)) * 100}%">
                            For: ${proposal.votes_for}
                        </div>
                        <div class="votes-against">
                            Against: ${proposal.votes_against}
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
            this.showError('Failed to fetch monetary policy from database');
            return {
                current_month: new Date().toISOString().slice(0, 7), // Current year-month
                base_issuance: 0,
                activity_factor: 0,
                fairness_factor: 0,
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
        try {
            // Fetch users data (required)
            const usersResponse = await fetch(`${this.apiBase}/v1/admin/users`, {
                headers: { 'Authorization': `Bearer ${this.authToken}` }
            });
            
            // Try to fetch CBI from the available endpoint
            let cbiResponse;
            
            try {
                cbiResponse = await fetch(`${this.apiBase}/v1/public/cbi`, {
                    headers: { 'Authorization': `Bearer ${this.authToken}` }
                });
            } catch (error) {
                console.warn('CBI endpoint not available:', error);
                cbiResponse = { ok: false };
            }
            
            // Note: Ratings endpoint doesn't exist in backend yet
            let ratingsResponse = { ok: false };
            
            let pfiDistribution = { 'excellent': 0, 'good': 0, 'average': 0, 'poor': 0 };
            let tfiAnalysis = { averageRating: 0, totalRatings: 0, topMerchants: [] };
            let cbi = { currentValue: 1.0, trend: 'stable', components: {} };
            
            // Calculate PFI distribution from real user data
            if (usersResponse.ok) {
                const userData = await usersResponse.json();
                const users = userData.users || [];
                const total = users.length;
                
                users.forEach(user => {
                    const pfi = user.pfi || 0;
                    if (pfi >= 90) pfiDistribution.excellent++;
                    else if (pfi >= 70) pfiDistribution.good++;
                    else if (pfi >= 50) pfiDistribution.average++;
                    else pfiDistribution.poor++;
                });
                
                // Convert to percentages
                if (total > 0) {
                    Object.keys(pfiDistribution).forEach(key => {
                        pfiDistribution[key] = Math.round((pfiDistribution[key] / total) * 100);
                    });
                }
            }
            
            // Calculate TFI analysis from real ratings data (if available)
            if (ratingsResponse && ratingsResponse.ok) {
                try {
                    const ratingsData = await ratingsResponse.json();
                    const ratings = ratingsData.ratings || [];
                    
                    if (ratings.length > 0) {
                        const totalRating = ratings.reduce((sum, rating) => sum + (rating.rating || 0), 0);
                        tfiAnalysis.averageRating = (totalRating / ratings.length).toFixed(1);
                        tfiAnalysis.totalRatings = ratings.length;
                        
                        // Get top merchants by rating
                        const merchantRatings = {};
                        ratings.forEach(rating => {
                            const merchant = rating.merchant_id || rating.to_user;
                            if (merchant) {
                                if (!merchantRatings[merchant]) {
                                    merchantRatings[merchant] = { total: 0, count: 0 };
                                }
                                merchantRatings[merchant].total += rating.rating || 0;
                                merchantRatings[merchant].count++;
                            }
                        });
                        
                        tfiAnalysis.topMerchants = Object.entries(merchantRatings)
                            .map(([merchant, data]) => ({
                                merchant,
                                avgRating: data.total / data.count
                            }))
                            .sort((a, b) => b.avgRating - a.avgRating)
                            .slice(0, 3)
                            .map(item => item.merchant);
                    }
                } catch (error) {
                    console.warn('Error parsing ratings data:', error);
                }
            } else {
                console.info('Ratings endpoint not available - TFI analysis will show zero values');
            }
            
            // Get CBI data (if available)
            if (cbiResponse && cbiResponse.ok) {
                try {
                    const cbiData = await cbiResponse.json();
                    const cbiInfo = cbiData.cbi || cbiData; // Handle both response formats
                    cbi = {
                        currentValue: cbiInfo.value || 1.0,
                        trend: cbiInfo.trend || 'stable',
                        components: {
                            food: cbiInfo.food_index || 1.0,
                            energy: cbiInfo.energy_index || 1.0,
                            labor: cbiInfo.labor_index || 1.0,
                            housing: cbiInfo.housing_index || 1.0
                        }
                    };
                } catch (error) {
                    console.warn('Error parsing CBI data:', error);
                }
            } else {
                console.info('CBI endpoint not available - using default values');
                cbi = {
                    currentValue: 1.0,
                    trend: 'stable',
                    components: {
                        food: 1.0,
                        energy: 1.0,
                        labor: 1.0,
                        housing: 1.0
                    }
                };
            }
            
            return { pfiDistribution, tfiAnalysis, cbi };
        } catch (error) {
            console.error('Error fetching fairness metrics:', error);
            this.showError('Failed to fetch fairness metrics from database');
            return {
                pfiDistribution: { 'excellent': 0, 'good': 0, 'average': 0, 'poor': 0 },
                tfiAnalysis: { averageRating: 0, totalRatings: 0, topMerchants: [] },
                cbi: { currentValue: 1.0, trend: 'stable', components: {} }
            };
        }
    }

    displayFairnessMetrics(metrics) {
        // PFI Distribution
        document.getElementById('pfi-breakdown').innerHTML = Object.entries(metrics.pfiDistribution)
            .map(([level, percentage]) => {
                // Map distribution levels to color classes
                const colorClass = level === 'excellent' ? 'pfi-excellent' : 
                                  level === 'good' ? 'pfi-good' : 
                                  level === 'developing' || level === 'average' ? 'pfi-developing' : 'pfi-poor';
                
                return `
                <div class="metric-row">
                    <span class="metric-label">${level.charAt(0).toUpperCase() + level.slice(1)}:</span>
                    <span class="metric-value">${percentage}%</span>
                    <div class="metric-bar">
                        <div class="metric-fill ${colorClass}" style="width: ${percentage}%"></div>
                    </div>
                </div>`;
            }).join('');

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

    async loadDemoReport() {
        try {
            const [users, stats, governance, community] = await Promise.all([
                this.fetchDemoUsers(),
                this.fetchDemoStats(),
                this.fetchDemoGovernance(),
                this.fetchDemoCommunity()
            ]);

            this.displayDemoUsers(users);
            this.displayDemoStats(stats);
            this.displayDemoGovernance(governance);
            this.displayDemoCommunity(community);
            
            // Create demo charts
            setTimeout(() => {
                this.createDemoCharts();
            }, 100);
        } catch (error) {
            console.error('Error loading demo report:', error);
            this.showError('Failed to load demo report data');
        }
    }

    async fetchDemoUsers() {
        try {
            const response = await fetch(`${this.apiBase}/v1/admin/users`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });
            
            if (response.ok) {
                const data = await response.json();
                return data.users || [];
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error fetching demo users:', error);
            this.showError('Failed to fetch user data from database');
            return [];
        }
    }

    async fetchDemoStats() {
        try {
            // Fetch transaction statistics from multiple endpoints
            const [transactionsResponse, statsResponse] = await Promise.all([
                fetch(`${this.apiBase}/v1/admin/transactions`, {
                    headers: { 'Authorization': `Bearer ${this.authToken}` }
                }),
                fetch(`${this.apiBase}/v1/admin/stats`, {
                    headers: { 'Authorization': `Bearer ${this.authToken}` }
                })
            ]);
            
            if (transactionsResponse.ok && statsResponse.ok) {
                const transactions = await transactionsResponse.json();
                const stats = await statsResponse.json();
                
                // Calculate statistics from actual transaction data
                const txData = transactions.transactions || [];
                const transferTxs = txData.filter(tx => tx.type === 'transfer');
                const issuanceTxs = txData.filter(tx => tx.type === 'monthly_issuance');
                const fairnessTxs = txData.filter(tx => tx.type === 'fairness_reward');
                const merchantTxs = txData.filter(tx => tx.type === 'merchant_incentive');
                
                return {
                    total_transactions: txData.length,
                    transfer_count: transferTxs.length,
                    transfer_volume: transferTxs.reduce((sum, tx) => sum + (tx.amount || 0), 0),
                    issuance_count: issuanceTxs.length,
                    issuance_volume: issuanceTxs.reduce((sum, tx) => sum + (tx.amount || 0), 0),
                    fairness_count: fairnessTxs.length,
                    fairness_volume: fairnessTxs.reduce((sum, tx) => sum + (tx.amount || 0), 0),
                    merchant_count: merchantTxs.length,
                    merchant_volume: merchantTxs.reduce((sum, tx) => sum + (tx.amount || 0), 0),
                    total_volume: txData.reduce((sum, tx) => sum + (tx.amount || 0), 0),
                    daily_transactions: stats.daily_transactions || 0
                };
            } else {
                throw new Error('Failed to fetch transaction statistics');
            }
        } catch (error) {
            console.error('Error fetching demo stats:', error);
            this.showError('Failed to fetch transaction statistics from database');
            return {
                total_transactions: 0,
                transfer_count: 0,
                transfer_volume: 0,
                issuance_count: 0,
                issuance_volume: 0,
                total_volume: 0
            };
        }
    }

    async fetchDemoGovernance() {
        try {
            // Fetch proposals from available endpoint
            const proposalsResponse = await fetch(`${this.apiBase}/v1/governance/proposals`, {
                headers: { 'Authorization': `Bearer ${this.authToken}` }
            });
            
            if (proposalsResponse.ok) {
                const proposals = await proposalsResponse.json();
                const proposalData = proposals.proposals || [];
                
                // Calculate vote statistics from proposal data (since votes endpoint doesn't exist)
                let totalVotes = 0;
                proposalData.forEach(proposal => {
                    if (proposal.votes_for) totalVotes += proposal.votes_for;
                    if (proposal.votes_against) totalVotes += proposal.votes_against;
                });
                
                const voteData = []; // Empty array since we don't have individual vote records
                
                const activeProposals = proposalData.filter(p => p.status === 'active').length;
                const totalVoters = new Set(voteData.map(v => v.user_id)).size;
                
                return {
                    total_proposals: proposalData.length,
                    total_votes: voteData.length,
                    active_proposals: activeProposals,
                    total_voters: totalVoters,
                    participation_rate: proposalData.length > 0 ? Math.round((totalVoters / proposalData.length) * 100) : 0
                };
            } else {
                throw new Error('Failed to fetch governance data');
            }
        } catch (error) {
            console.error('Error fetching demo governance:', error);
            this.showError('Failed to fetch governance data from database');
            return {
                total_proposals: 0,
                total_votes: 0,
                active_proposals: 0,
                participation_rate: 0
            };
        }
    }

    async fetchDemoCommunity() {
        try {
            // Fetch community data from available endpoints only
            const usersResponse = await fetch(`${this.apiBase}/v1/admin/users`, {
                headers: { 'Authorization': `Bearer ${this.authToken}` }
            });
            
            if (usersResponse.ok) {
                const users = await usersResponse.json();
                const userData = users.users || [];
                
                // Since attestations and ratings endpoints don't exist yet,
                // calculate community metrics from available user data
                const attestationData = []; // Empty for now
                const ratingData = []; // Empty for now
                
                // Calculate community hours from user data
                const totalCommunityHours = userData.reduce((sum, user) => sum + (user.community_service || 0), 0);
                
                // Calculate engagement score based on user participation
                const activeUsers = userData.filter(user => user.pfi > 0).length;
                const engagementScore = userData.length > 0 ? Math.round((activeUsers / userData.length) * 100) : 0;
                
                return {
                    total_attestations: attestationData.length,
                    total_ratings: ratingData.length,
                    community_hours: totalCommunityHours,
                    engagement_score: engagementScore,
                    active_users: activeUsers,
                    total_users: userData.length
                };
            } else {
                throw new Error('Failed to fetch community data');
            }
        } catch (error) {
            console.error('Error fetching demo community:', error);
            this.showError('Failed to fetch community data from database');
            return {
                total_attestations: 0,
                total_ratings: 0,
                community_hours: 0,
                engagement_score: 0
            };
        }
    }

    displayDemoUsers(users) {
        const container = document.getElementById('demo-users-list');
        if (!container) return;
        
        // Sort users by PFI score from highest to lowest
        const sortedUsers = [...users].sort((a, b) => (b.pfi || 0) - (a.pfi || 0));
        
        container.innerHTML = sortedUsers.map(user => {
            const pfiCategory = this.getPFICategory(user.pfi);
            const pfiCategoryClass = this.getPFICategoryClass(user.pfi);
            
            return `
                <div class="user-profile-item">
                    <div class="user-info">
                        <div class="user-name">${user.first_name} ${user.last_name}</div>
                        <div class="user-username">@${user.username}</div>
                    </div>
                    <div class="user-pfi">
                        <span class="pfi-score ${pfiCategoryClass}">${user.pfi}★</span>
                        <span class="pfi-category">(${pfiCategory})</span>
                    </div>
                </div>
            `;
        }).join('');
    }

    displayDemoStats(stats) {
        const container = document.getElementById('demo-transaction-stats');
        if (!container) return;
        
        container.innerHTML = `
            <div class="stat-row">
                <div class="stat-label">Total Transactions</div>
                <div class="stat-value">${(stats.total_transactions || 0).toLocaleString()}</div>
            </div>
            <div class="stat-row">
                <div class="stat-label">P2P Transfers</div>
                <div class="stat-value">${(stats.transfer_count || 0).toLocaleString()} (${((stats.transfer_volume || 0)).toLocaleString()} FC)</div>
            </div>
            <div class="stat-row">
                <div class="stat-label">Monthly Issuance</div>
                <div class="stat-value">${(stats.issuance_count || 0).toLocaleString()} (${((stats.issuance_volume || 0)).toLocaleString()} FC)</div>
            </div>
            <div class="stat-row">
                <div class="stat-label">Total Volume</div>
                <div class="stat-value">${((stats.total_volume || 0)).toLocaleString()} FC</div>
            </div>
        `;

        // Update performance metrics
        document.getElementById('demo-transactions-count').textContent = (stats.total_transactions || 0).toLocaleString();
        document.getElementById('demo-volume-total').textContent = ((stats.total_volume || 0)).toLocaleString();
    }

    displayDemoGovernance(governance) {
        const container = document.getElementById('demo-governance-stats');
        if (!container) return;
        
        container.innerHTML = `
            <div class="governance-item">
                <div class="governance-label">Total Proposals</div>
                <div class="governance-value">${governance.total_proposals || 0}</div>
            </div>
            <div class="governance-item">
                <div class="governance-label">Votes Cast</div>
                <div class="governance-value">${governance.total_votes || 0}</div>
            </div>
            <div class="governance-item">
                <div class="governance-label">Active Proposals</div>
                <div class="governance-value">${governance.active_proposals || 0}</div>
            </div>
            <div class="governance-item">
                <div class="governance-label">Participation Rate</div>
                <div class="governance-value">${governance.participation_rate || 0}%</div>
            </div>
        `;

        // Update performance metrics
        document.getElementById('demo-proposals-count').textContent = governance.total_proposals || 0;
        document.getElementById('demo-votes-count').textContent = governance.total_votes || 0;
    }

    displayDemoCommunity(community) {
        const container = document.getElementById('demo-community-stats');
        if (!container) return;
        
        container.innerHTML = `
            <div class="community-item">
                <div class="community-label">Peer Attestations</div>
                <div class="community-value">${community.total_attestations || 0}</div>
            </div>
            <div class="community-item">
                <div class="community-label">Merchant Ratings</div>
                <div class="community-value">${community.total_ratings || 0}</div>
            </div>
            <div class="community-item">
                <div class="community-label">Community Hours</div>
                <div class="community-value">${community.community_hours || 0}</div>
            </div>
            <div class="community-item">
                <div class="community-label">Engagement Score</div>
                <div class="community-value">${community.engagement_score || 0}%</div>
            </div>
        `;

        // Update performance metrics
        document.getElementById('demo-attestations-count').textContent = community.total_attestations || 0;
    }

    async createDemoCharts() {
        await this.createDemoPFIChart();
        await this.createDemoTransactionTypesChart();
        await this.createDemoVolumeChart();
        await this.createDemoGovernanceChart();
    }

    async createDemoPFIChart() {
        const canvas = document.getElementById('demoPfiChart');
        if (!canvas) return;
        
        try {
            // Fetch real PFI distribution data
            const users = await this.fetchDemoUsers();
            
            // Calculate PFI distribution from real user data
            let excellent = 0, good = 0, average = 0, developing = 0;
            
            users.forEach(user => {
                const pfi = user.pfi || 0;
                if (pfi >= 80) excellent++;
                else if (pfi >= 70) good++;
                else if (pfi >= 50) average++;
                else developing++;
            });
            
            const ctx = canvas.getContext('2d');
            
            if (this.charts.demoPfi) {
                this.charts.demoPfi.destroy();
            }

            this.charts.demoPfi = new Chart(ctx, {
                type: 'doughnut',
                data: {
                    labels: ['Excellent (80-100)', 'Good (70-79)', 'Average (50-69)', 'Developing (0-49)'],
                    datasets: [{
                        data: [excellent, good, average, developing],
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
                        borderWidth: 2
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'bottom'
                        },
                        tooltip: {
                            callbacks: {
                                label: function(context) {
                                    const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                    const percentage = ((context.parsed / total) * 100).toFixed(1);
                                    return `${context.label}: ${context.parsed} users (${percentage}%)`;
                                }
                            }
                        }
                    }
                }
            });
        } catch (error) {
            console.error('Error creating demo PFI chart:', error);
        }
    }

    async createDemoTransactionTypesChart() {
        const canvas = document.getElementById('demoTransactionTypesChart');
        if (!canvas) return;
        
        try {
            // Fetch real transaction statistics
            const stats = await this.fetchDemoStats();
            
            const ctx = canvas.getContext('2d');
            
            if (this.charts.demoTxTypes) {
                this.charts.demoTxTypes.destroy();
            }

            this.charts.demoTxTypes = new Chart(ctx, {
                type: 'pie',
                data: {
                    labels: ['P2P Transfers', 'Monthly Issuance', 'Fairness Rewards', 'Merchant Incentives'],
                    datasets: [{
                        data: [
                            stats.transfer_count || 0,
                            stats.issuance_count || 0,
                            stats.fairness_count || 0,
                            stats.merchant_count || 0
                        ],
                        backgroundColor: [
                            'rgba(52, 152, 219, 0.8)',
                            'rgba(46, 204, 113, 0.8)',
                            'rgba(155, 89, 182, 0.8)',
                            'rgba(243, 156, 18, 0.8)'
                        ],
                        borderWidth: 2
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'bottom'
                        },
                        tooltip: {
                            callbacks: {
                                label: function(context) {
                                    const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                    const percentage = total > 0 ? ((context.parsed / total) * 100).toFixed(1) : 0;
                                    return `${context.label}: ${context.parsed} transactions (${percentage}%)`;
                                }
                            }
                        }
                    }
                }
            });
        } catch (error) {
            console.error('Error creating demo transaction types chart:', error);
        }
    }

    async createDemoVolumeChart() {
        const canvas = document.getElementById('demoVolumeChart');
        if (!canvas) return;
        
        try {
            // Fetch real transaction volume data over time
            const response = await fetch(`${this.apiBase}/v1/admin/transaction-volume`, {
                headers: { 'Authorization': `Bearer ${this.authToken}` }
            });
            
            let volumeData = [];
            let labels = [];
            
            if (response.ok) {
                const data = await response.json();
                const volumes = data.volume_data || [];
                
                labels = volumes.map(v => v.Month || v.period);
                volumeData = volumes.map(v => v.Volume || v.volume || 0);
            } else {
                // If no volume endpoint, calculate from transaction dates
                const transactions = await fetch(`${this.apiBase}/v1/admin/transactions`, {
                    headers: { 'Authorization': `Bearer ${this.authToken}` }
                });
                
                if (transactions.ok) {
                    const txData = await transactions.json();
                    const txList = txData.transactions || [];
                    
                    // Group by month
                    const monthlyVolume = {};
                    txList.forEach(tx => {
                        if (tx.created_at && tx.amount) {
                            const date = new Date(tx.created_at);
                            const monthKey = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
                            monthlyVolume[monthKey] = (monthlyVolume[monthKey] || 0) + tx.amount;
                        }
                    });
                    
                    labels = Object.keys(monthlyVolume).sort();
                    volumeData = labels.map(month => monthlyVolume[month]);
                }
            }
            
            const ctx = canvas.getContext('2d');
            
            if (this.charts.demoVolume) {
                this.charts.demoVolume.destroy();
            }

            this.charts.demoVolume = new Chart(ctx, {
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
            console.error('Error creating demo volume chart:', error);
        }
    }

    async createDemoGovernanceChart() {
        const canvas = document.getElementById('demoGovernanceChart');
        if (!canvas) return;
        
        try {
            // Fetch real governance data from proposals and votes
            const proposalsResponse = await fetch(`${this.apiBase}/v1/governance/proposals`, {
                headers: { 'Authorization': `Bearer ${this.authToken}` }
            });
            
            let proposalsCount = 0;
            let votesCount = 0;
            let activeVoters = 0;
            
            if (proposalsResponse.ok) {
                const proposalsData = await proposalsResponse.json();
                const proposals = proposalsData.proposals || [];
                proposalsCount = proposals.length;
                
                // Count votes from proposals
                proposals.forEach(proposal => {
                    if (proposal.votes) {
                        votesCount += proposal.votes.length || 0;
                    } else if (proposal.total_votes) {
                        votesCount += proposal.total_votes;
                    }
                });
                
                // Count unique voters (active voters)
                const voters = new Set();
                proposals.forEach(proposal => {
                    if (proposal.votes) {
                        proposal.votes.forEach(vote => {
                            if (vote.user_id || vote.voter_id) {
                                voters.add(vote.user_id || vote.voter_id);
                            }
                        });
                    }
                });
                activeVoters = voters.size;
            }
            
            const ctx = canvas.getContext('2d');
            
            if (this.charts.demoGovernance) {
                this.charts.demoGovernance.destroy();
            }

            this.charts.demoGovernance = new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: ['Proposals Created', 'Votes Cast', 'Active Voters'],
                    datasets: [{
                        label: 'Count',
                        data: [proposalsCount, votesCount, activeVoters],
                        backgroundColor: [
                            'rgba(155, 89, 182, 0.8)',
                            'rgba(52, 152, 219, 0.8)',
                            'rgba(46, 204, 113, 0.8)'
                        ],
                        borderColor: [
                            'rgba(155, 89, 182, 1)',
                            'rgba(52, 152, 219, 1)',
                            'rgba(46, 204, 113, 1)'
                        ],
                        borderWidth: 1
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            display: false
                        }
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            ticks: {
                                stepSize: 1
                            }
                        }
                    }
                }
            });
        } catch (error) {
            console.error('Error creating demo governance chart:', error);
        }
    }

    getPFICategory(pfi) {
        if (pfi >= 80) return 'Excellent';
        if (pfi >= 70) return 'Good';
        if (pfi >= 50) return 'Average';
        return 'Developing';
    }

    // Wallet Management Methods
    async loadWallet() {
        try {
            await this.loadWalletBalance();
            await this.loadTransactionHistory();
            this.updateReceiveInfo();
        } catch (error) {
            console.error('Failed to load wallet:', error);
            this.showError('Failed to load wallet information');
        }
    }

    async loadWalletBalance() {
        try {
            const response = await fetch(`${this.apiBase}/v1/wallet/balance`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });

            if (!response.ok) throw new Error('Failed to fetch balance');

            const wallet = await response.json();
            const balanceElement = document.getElementById('wallet-balance');
            const balanceUsdElement = document.getElementById('wallet-balance-usd');
            const availableBalanceElement = document.getElementById('available-balance');

            if (balanceElement) {
                balanceElement.textContent = `${wallet.balance.toFixed(2)} FC`;
            }
            if (balanceUsdElement) {
                // Simple conversion rate for demo (1 FC = $0.85)
                const usdValue = (wallet.balance * 0.85).toFixed(2);
                balanceUsdElement.textContent = `≈ $${usdValue} USD`;
            }
            if (availableBalanceElement) {
                availableBalanceElement.textContent = wallet.balance.toFixed(2);
            }

            this.currentBalance = wallet.balance;
        } catch (error) {
            console.error('Failed to load wallet balance:', error);
            this.showError('Failed to load wallet balance');
        }
    }

    async refreshWalletBalance() {
        const refreshBtn = document.querySelector('.refresh-btn');
        if (refreshBtn) {
            refreshBtn.style.transform = 'rotate(360deg)';
            setTimeout(() => {
                refreshBtn.style.transform = '';
            }, 500);
        }
        await this.loadWalletBalance();
        this.showMessage('Balance refreshed!', 'success');
    }

    async loadTransactionHistory() {
        try {
            const limit = document.getElementById('tx-limit')?.value || 20;
            const typeFilter = document.getElementById('tx-filter-type')?.value || 'all';
            
            let url = `${this.apiBase}/v1/wallet/history?limit=${limit}`;
            
            const response = await fetch(url, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });

            if (!response.ok) throw new Error('Failed to fetch transaction history');

            const data = await response.json();
            this.renderTransactionHistory(data.transactions || []);
        } catch (error) {
            console.error('Failed to load transaction history:', error);
            this.showError('Failed to load transaction history');
        }
    }

    async refreshTransactionHistory() {
        const historyBtn = document.querySelector('.history-btn');
        if (historyBtn) {
            historyBtn.style.opacity = '0.7';
            setTimeout(() => {
                historyBtn.style.opacity = '';
            }, 500);
        }
        await this.loadTransactionHistory();
        this.showMessage('Transaction history refreshed!', 'success');
    }

    renderTransactionHistory(transactions) {
        const container = document.getElementById('transaction-history');
        
        if (!transactions || transactions.length === 0) {
            container.innerHTML = `
                <div class="loading-spinner">
                    <i class="fas fa-wallet"></i>
                    <p>No transactions found</p>
                    <small>Your transaction history will appear here</small>
                </div>
            `;
            return;
        }

        const transactionHtml = transactions.map(tx => {
            const isReceived = this.currentUser && tx.to_user_id === this.currentUser.id;
            const isSent = this.currentUser && tx.user_id === this.currentUser.id && tx.to_user_id;
            
            let txType = 'reward';
            let icon = 'fas fa-gift';
            let amountClass = 'positive';
            let amountPrefix = '+';
            let title = 'Reward';
            let description = tx.description || 'System reward';

            if (isSent) {
                txType = 'send';
                icon = 'fas fa-arrow-up';
                amountClass = 'negative';
                amountPrefix = '-';
                title = 'Sent';
                description = tx.description || `Sent to ${tx.to_user?.username || 'Unknown'}`;
            } else if (isReceived) {
                txType = 'receive';
                icon = 'fas fa-arrow-down';
                amountClass = 'positive';
                amountPrefix = '+';
                title = 'Received';
                description = tx.description || `Received from ${tx.user?.username || 'Unknown'}`;
            } else if (tx.type === 'fee') {
                txType = 'fee';
                icon = 'fas fa-minus';
                amountClass = 'negative';
                amountPrefix = '-';
                title = 'Fee';
            }

            const date = new Date(tx.created_at).toLocaleDateString();
            const time = new Date(tx.created_at).toLocaleTimeString();

            return `
                <div class="transaction-item">
                    <div class="tx-icon tx-${txType}">
                        <i class="${icon}"></i>
                    </div>
                    <div class="tx-details">
                        <div class="tx-title">${title}</div>
                        <div class="tx-description">${description}</div>
                        <div class="tx-date">${date} at ${time}</div>
                        <div class="tx-status ${tx.status || 'completed'}">${tx.status || 'completed'}</div>
                    </div>
                    <div class="tx-amount ${amountClass}">
                        ${amountPrefix}${tx.amount.toFixed(2)} FC
                    </div>
                </div>
            `;
        }).join('');

        container.innerHTML = transactionHtml;
    }

    showSendModal() {
        const modal = document.getElementById('send-modal');
        if (modal) {
            modal.classList.add('active');
            // Update available balance
            const availableElement = document.getElementById('available-balance');
            if (availableElement && this.currentBalance) {
                availableElement.textContent = this.currentBalance.toFixed(2);
            }
        }
    }

    hideSendModal() {
        const modal = document.getElementById('send-modal');
        if (modal) {
            modal.classList.remove('active');
            // Clear form
            document.getElementById('send-form').reset();
            document.getElementById('tx-preview').style.display = 'none';
        }
    }

    showReceiveModal() {
        const modal = document.getElementById('receive-modal');
        if (modal) {
            modal.classList.add('active');
        }
    }

    hideReceiveModal() {
        const modal = document.getElementById('receive-modal');
        if (modal) {
            modal.classList.remove('active');
        }
    }

    updateReceiveInfo() {
        if (this.currentUser) {
            const usernameElement = document.getElementById('receive-username');
            const walletIdElement = document.getElementById('receive-wallet-id');
            
            if (usernameElement) {
                usernameElement.textContent = this.currentUser.username;
            }
            if (walletIdElement) {
                walletIdElement.textContent = this.currentUser.id || 'Loading...';
            }
        }
    }

    copyToClipboard(elementId) {
        const element = document.getElementById(elementId);
        if (element) {
            const text = element.textContent;
            navigator.clipboard.writeText(text).then(() => {
                this.showMessage('Copied to clipboard!', 'success');
            }).catch(() => {
                this.showMessage('Failed to copy to clipboard', 'error');
            });
        }
    }

    async handleSendTransaction(event) {
        event.preventDefault();
        
        const username = document.getElementById('send-username').value.trim();
        const amount = parseFloat(document.getElementById('send-amount').value);
        const description = document.getElementById('send-description').value.trim();

        // Enhanced validation
        if (!username || username.length < 3) {
            this.showError('Please enter a valid username (minimum 3 characters)');
            return;
        }

        if (!amount || amount <= 0 || isNaN(amount)) {
            this.showError('Please enter a valid amount greater than 0');
            return;
        }

        if (amount < 0.01) {
            this.showError('Minimum transfer amount is 0.01 FC');
            return;
        }

        if (amount > this.currentBalance) {
            this.showError('Insufficient balance for this transaction');
            return;
        }

        // Calculate total with fee
        const fee = amount * 0.001;
        const total = amount + fee;
        
        if (total > this.currentBalance) {
            this.showError(`Insufficient balance. Total with fee: ${total.toFixed(4)} FC, Available: ${this.currentBalance.toFixed(2)} FC`);
            return;
        }

        // Confirmation for large amounts
        if (amount > 100) {
            if (!confirm(`You are about to send ${amount.toFixed(2)} FC to ${username}. Are you sure?`)) {
                return;
            }
        }

        // Prevent sending to self
        if (this.currentUser && username.toLowerCase() === this.currentUser.username.toLowerCase()) {
            this.showError('You cannot send FairCoins to yourself');
            return;
        }

        // Rate limiting check
        const now = Date.now();
        if (this.lastTransactionTime && (now - this.lastTransactionTime) < 5000) {
            this.showError('Please wait 5 seconds between transactions');
            return;
        }

        // Disable submit button during processing
        const submitBtn = event.target.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;
        submitBtn.disabled = true;
        submitBtn.textContent = 'Sending...';

        try {
            const response = await fetch(`${this.apiBase}/v1/wallet/send`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.authToken}`
                },
                body: JSON.stringify({
                    to_username: username,
                    amount: amount,
                    description: description
                })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Transfer failed');
            }

            const result = await response.json();
            this.lastTransactionTime = now;
            this.hideSendModal();
            this.showMessage(`Successfully sent ${amount.toFixed(2)} FC to ${username}!`, 'success');
            
            // Refresh balance and transaction history
            await this.loadWalletBalance();
            await this.loadTransactionHistory();
            
        } catch (error) {
            console.error('Send transaction failed:', error);
            let errorMessage = 'Transfer failed';
            
            if (error.message.includes('not found')) {
                errorMessage = `User "${username}" not found`;
            } else if (error.message.includes('balance')) {
                errorMessage = 'Insufficient balance';
            } else if (error.message.includes('network')) {
                errorMessage = 'Network error - please try again';
            } else {
                errorMessage = error.message;
            }
            
            this.showError(errorMessage);
        } finally {
            // Re-enable submit button
            submitBtn.disabled = false;
            submitBtn.textContent = originalText;
        }
    }

    setupWalletEventListeners() {
        // Send form submission
        const sendForm = document.getElementById('send-form');
        if (sendForm) {
            sendForm.addEventListener('submit', this.handleSendTransaction.bind(this));
        }

        // Amount input change for preview
        const amountInput = document.getElementById('send-amount');
        if (amountInput) {
            amountInput.addEventListener('input', this.updateTransactionPreview.bind(this));
        }

        // Transaction filter changes
        const typeFilter = document.getElementById('tx-filter-type');
        const limitInput = document.getElementById('tx-limit');
        
        if (typeFilter) {
            typeFilter.addEventListener('change', this.loadTransactionHistory.bind(this));
        }
        if (limitInput) {
            limitInput.addEventListener('change', this.loadTransactionHistory.bind(this));
        }

        // Close modals when clicking outside
        window.addEventListener('click', (event) => {
            const sendModal = document.getElementById('send-modal');
            const receiveModal = document.getElementById('receive-modal');
            
            if (event.target === sendModal) {
                this.hideSendModal();
            }
            if (event.target === receiveModal) {
                this.hideReceiveModal();
            }
        });
    }

    updateTransactionPreview() {
        const amountInput = document.getElementById('send-amount');
        const amount = parseFloat(amountInput.value);
        const preview = document.getElementById('tx-preview');
        
        // Clear any previous validation styling
        amountInput.classList.remove('error', 'warning');
        
        if (amount && amount > 0 && !isNaN(amount)) {
            const fee = amount * 0.001; // 0.1% fee
            const total = amount + fee;
            
            // Validation styling
            if (amount < 0.01) {
                amountInput.classList.add('error');
            } else if (total > this.currentBalance) {
                amountInput.classList.add('error');
            } else if (amount > this.currentBalance * 0.8) {
                amountInput.classList.add('warning');
            }
            
            document.getElementById('preview-amount').textContent = `${amount.toFixed(2)} FC`;
            document.getElementById('preview-fee').textContent = `${fee.toFixed(4)} FC`;
            document.getElementById('preview-total').textContent = `${total.toFixed(4)} FC`;
            
            // Show insufficient balance warning
            if (total > this.currentBalance) {
                document.getElementById('preview-total').innerHTML = `
                    ${total.toFixed(4)} FC 
                    <span style="color: #e74c3c; font-size: 12px;">(Insufficient balance)</span>
                `;
            }
            
            preview.style.display = 'block';
        } else {
            preview.style.display = 'none';
        }
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
        
        // Create a styled error notification
        const errorDiv = document.createElement('div');
        errorDiv.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: #dc3545;
            color: white;
            padding: 15px 20px;
            border-radius: 5px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            z-index: 10000;
            max-width: 400px;
            font-weight: 500;
        `;
        errorDiv.innerHTML = `<i class="fas fa-exclamation-circle"></i> ${message}`;
        document.body.appendChild(errorDiv);
        
        // Remove after 5 seconds
        setTimeout(() => {
            if (errorDiv.parentNode) {
                errorDiv.parentNode.removeChild(errorDiv);
            }
        }, 5000);
    }

    showMessage(message, type = 'info') {
        const colors = {
            success: '#28a745',
            error: '#dc3545',
            info: '#17a2b8',
            warning: '#ffc107'
        };
        
        const icons = {
            success: 'fas fa-check-circle',
            error: 'fas fa-exclamation-circle',
            info: 'fas fa-info-circle',
            warning: 'fas fa-exclamation-triangle'
        };
        
        const messageDiv = document.createElement('div');
        messageDiv.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${colors[type] || colors.info};
            color: ${type === 'warning' ? '#212529' : 'white'};
            padding: 15px 20px;
            border-radius: 5px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            z-index: 10000;
            max-width: 400px;
            font-weight: 500;
        `;
        messageDiv.innerHTML = `<i class="${icons[type] || icons.info}"></i> ${message}`;
        document.body.appendChild(messageDiv);
        
        // Remove after 3-5 seconds depending on type
        const timeout = type === 'error' ? 5000 : 3000;
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.parentNode.removeChild(messageDiv);
            }
        }, timeout);
    }

    handleUnauthorized() {
        this.showError('Session expired or insufficient privileges. Redirecting to login...');
        localStorage.removeItem('admin_token');
        this.authToken = null;
        this.currentUser = null;
        this.isAuthenticated = false;
        this.hideAdminInfo();
        this.showLoginModal();
    }

    handleAutoLogout(message = 'Session expired') {
        console.log('Auto-logout triggered:', message);
        
        // Clear authentication data
        localStorage.removeItem('admin_token');
        this.authToken = null;
        this.currentUser = null;
        this.isAuthenticated = false;
        
        // Stop all intervals
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
        if (this.authCheckInterval) {
            clearInterval(this.authCheckInterval);
            this.authCheckInterval = null;
        }
        
        this.hideAdminInfo();
        
        // Show user-friendly message
        this.showError(message + '. Please login again.');
        
        // Force show login modal
        this.showLoginModal();
        
        // If we're on a protected route, redirect to admin page
        if (window.location.pathname !== '/admin.html' && window.location.pathname !== '/') {
            setTimeout(() => {
                window.location.href = '/admin.html';
            }, 2000);
        }
    }

    showModal(title, content, confirmCallback) {
        document.getElementById('modal-title').textContent = title;
        document.getElementById('modal-body').innerHTML = content;
        document.getElementById('modal-confirm').onclick = confirmCallback;
        document.getElementById('modal-overlay').classList.add('show');
    }

    closeModal() {
        // Close any open modals using main frontend system
        const modals = document.querySelectorAll('.modal.active');
        modals.forEach(modal => {
            modal.classList.remove('active');
        });
        
        // Also handle legacy modal-overlay if it exists
        const modalOverlay = document.getElementById('modal-overlay');
        if (modalOverlay) {
            modalOverlay.classList.remove('show');
            // Clear any inline styles that might interfere
            modalOverlay.style.display = '';
        }
    }

    // Debug utility to check modal state
    debugModals() {
        console.log('=== Modal Debug Info ===');
        const allModals = document.querySelectorAll('.modal');
        allModals.forEach(modal => {
            const computedStyle = window.getComputedStyle(modal);
            console.log(`Modal ${modal.id}:`, {
                display: computedStyle.display,
                position: computedStyle.position,
                zIndex: computedStyle.zIndex,
                classes: modal.className,
                visible: modal.classList.contains('active'),
                rect: modal.getBoundingClientRect()
            });
        });
        console.log('=====================');
    }

    getPFICategoryClass(pfi) {
        // Convert PFI score to appropriate color category CSS class
        const score = parseInt(pfi) || 0;
        
        if (score >= 80) return 'pfi-excellent';      // Green: Excellent (80-100)
        if (score >= 60) return 'pfi-good';           // Blue: Good (60-79)
        if (score >= 40) return 'pfi-developing';     // Orange: Developing (40-59)
        return 'pfi-poor';                            // Red: Poor (0-39)
    }

    getTFICategoryClass(tfi) {
        // Convert TFI score to appropriate color category CSS class
        const score = parseInt(tfi) || 0;
        
        if (score >= 80) return 'tfi-excellent';      // Green: Excellent (80-100)
        if (score >= 60) return 'tfi-good';           // Blue: Good (60-79)
        if (score >= 40) return 'tfi-developing';     // Orange: Developing (40-59)
        return 'tfi-poor';                            // Red: Poor (0-39)
    }

    startDataRefresh() {
        // Clear any existing intervals
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }
        if (this.authCheckInterval) {
            clearInterval(this.authCheckInterval);
        }
        
        // Refresh data every 30 seconds
        this.refreshInterval = setInterval(() => {
            if (this.currentSection === 'dashboard' && this.isAuthenticated) {
                this.loadDashboard();
            }
        }, 30000);
        
        // Check authentication status every 60 seconds
        this.authCheckInterval = setInterval(() => {
            if (this.isAuthenticated && this.authToken) {
                this.validateTokenSilently();
            }
        }, 60000);
    }

    async validateTokenSilently() {
        try {
            const response = await fetch(`${this.apiBase}/v1/users/profile`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });

            if (!response.ok) {
                console.warn('Silent token validation failed - triggering auto-logout');
                this.handleAutoLogout('Session expired - please login again');
            }
        } catch (error) {
            console.warn('Silent token validation error:', error);
            // Don't auto-logout on network errors, just log it
        }
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

    async editUser(userId) {
        try {
            // Fetch detailed user data
            const userData = await this.fetchUserDetails(userId);
            if (!userData) {
                this.showError('Failed to fetch user details');
                return;
            }

            this.showModal('Edit User Details', this.generateUserEditForm(userData), () => {
                this.saveUserDetails(userId);
            });
        } catch (error) {
            console.error('Error opening user editor:', error);
            this.showError('Failed to open user editor');
        }
    }

    async fetchUserDetails(userId) {
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
            const users = data.users || [];
            return users.find(user => user.id === userId);
        } catch (error) {
            console.error('Error fetching user details:', error);
            return null;
        }
    }

    generateUserEditForm(user) {
        return `
            <div class="user-edit-form">
                <!-- Basic Information -->
                <div class="form-section">
                    <h4><i class="fas fa-user"></i> Basic Information</h4>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="edit-username">Username:</label>
                            <input type="text" id="edit-username" value="${user.username || ''}" readonly>
                            <small>Username cannot be changed</small>
                        </div>
                        <div class="form-group">
                            <label for="edit-email">Email:</label>
                            <input type="email" id="edit-email" value="${user.email || ''}" readonly>
                            <small>Email cannot be changed</small>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="edit-first-name">First Name:</label>
                            <input type="text" id="edit-first-name" value="${user.first_name || ''}" placeholder="Enter first name">
                        </div>
                        <div class="form-group">
                            <label for="edit-last-name">Last Name:</label>
                            <input type="text" id="edit-last-name" value="${user.last_name || ''}" placeholder="Enter last name">
                        </div>
                    </div>
                </div>

                <!-- Fairness Scores -->
                <div class="form-section">
                    <h4><i class="fas fa-star"></i> Fairness Scores</h4>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="edit-pfi">PFI★ Score:</label>
                            <input type="number" id="edit-pfi" min="0" max="100" value="${user.pfi || 0}">
                            <small>Personal Fairness Index (0-100)</small>
                        </div>
                        <div class="form-group">
                            <label for="edit-tfi">TFI★ Score:</label>
                            <input type="number" id="edit-tfi" min="0" max="100" value="${user.tfi || 0}">
                            <small>Trade Fairness Index (0-100)</small>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="edit-community-service">Community Service Hours:</label>
                            <input type="number" id="edit-community-service" min="0" value="${user.community_service || 0}">
                            <small>Hours of community service contributed</small>
                        </div>
                    </div>
                </div>

                <!-- Roles and Permissions -->
                <div class="form-section">
                    <h4><i class="fas fa-shield-alt"></i> Roles & Permissions</h4>
                    <div class="form-row">
                        <div class="form-group checkbox-group">
                            <label class="checkbox-label">
                                <input type="checkbox" id="edit-is-admin" ${user.is_admin ? 'checked' : ''}>
                                <span class="checkmark"></span>
                                <i class="fas fa-crown"></i> Administrator
                            </label>
                            <small>Full system access and management permissions</small>
                        </div>
                        <div class="form-group checkbox-group">
                            <label class="checkbox-label">
                                <input type="checkbox" id="edit-is-merchant" ${user.is_merchant ? 'checked' : ''}>
                                <span class="checkmark"></span>
                                <i class="fas fa-store"></i> Merchant
                            </label>
                            <small>Can sell goods and services on the platform</small>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group checkbox-group">
                            <label class="checkbox-label">
                                <input type="checkbox" id="edit-is-verified" ${user.is_verified ? 'checked' : ''}>
                                <span class="checkmark"></span>
                                <i class="fas fa-check-circle"></i> Verified User
                            </label>
                            <small>Identity and profile have been verified</small>
                        </div>
                    </div>
                </div>

                <!-- Account Information -->
                <div class="form-section">
                    <h4><i class="fas fa-info-circle"></i> Account Information</h4>
                    <div class="form-row">
                        <div class="form-group">
                            <label>User ID:</label>
                            <input type="text" value="${user.id}" readonly>
                            <small>Unique identifier</small>
                        </div>
                        <div class="form-group">
                            <label>Account Balance:</label>
                            <input type="text" value="${(user.balance || 0).toFixed(2)} FC" readonly>
                            <small>Current FairCoin balance</small>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label>Member Since:</label>
                            <input type="text" value="${user.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'}" readonly>
                            <small>Account creation date</small>
                        </div>
                    </div>
                </div>

                <!-- Quick Actions -->
                <div class="form-section">
                    <h4><i class="fas fa-bolt"></i> Quick Actions</h4>
                    <div class="quick-actions">
                        <button type="button" class="btn btn-outline btn-sm" onclick="admin.resetUserPassword('${user.id}')">
                            <i class="fas fa-key"></i> Reset Password
                        </button>
                        <button type="button" class="btn btn-outline btn-sm" onclick="admin.viewUserTransactions('${user.id}')">
                            <i class="fas fa-exchange-alt"></i> View Transactions
                        </button>
                        <button type="button" class="btn btn-outline btn-sm" onclick="admin.suspendUser('${user.id}')">
                            <i class="fas fa-user-slash"></i> Suspend Account
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    async saveUserDetails(userId) {
        try {
            const updateData = {
                first_name: document.getElementById('edit-first-name').value.trim(),
                last_name: document.getElementById('edit-last-name').value.trim(),
                pfi: parseInt(document.getElementById('edit-pfi').value) || 0,
                tfi: parseInt(document.getElementById('edit-tfi').value) || 0,
                community_service: parseInt(document.getElementById('edit-community-service').value) || 0,
                is_admin: document.getElementById('edit-is-admin').checked,
                is_merchant: document.getElementById('edit-is-merchant').checked,
                is_verified: document.getElementById('edit-is-verified').checked
            };

            const response = await fetch(`${this.apiBase}/v1/admin/users/${userId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.authToken}`
                },
                body: JSON.stringify(updateData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to update user');
            }

            const result = await response.json();
            this.showSuccess(`User ${result.user.username} updated successfully`);
            this.closeModal();
            this.loadUsers(); // Refresh the users list
        } catch (error) {
            console.error('Error saving user details:', error);
            this.showError(`Failed to update user: ${error.message}`);
        }
    }

    // User Role Management Functions
    async editUserRoles(userId, userData) {
        const user = typeof userData === 'string' ? JSON.parse(userData.replace(/&quot;/g, '"')) : userData;
        
        this.showModal('Manage User Roles', `
            <div class="role-management-form">
                <div class="user-info">
                    <h4><i class="fas fa-user"></i> ${user.username}</h4>
                    <p>${user.email}</p>
                    <p>PFI★: <span class="pfi-score ${this.getPFICategoryClass(user.pfi)}">${user.pfi || 0}★</span> | Balance: ${(user.balance || 0).toFixed(2)} FC</p>
                </div>
                
                <div class="roles-section">
                    <h5>Assign Roles</h5>
                    <div class="role-checkboxes">
                        <label class="role-checkbox admin">
                            <input type="checkbox" id="role-admin" ${user.is_admin ? 'checked' : ''}>
                            <span class="role-badge admin"><i class="fas fa-crown"></i> Administrator</span>
                            <small>Full system access and management permissions</small>
                        </label>
                        
                        <label class="role-checkbox merchant">
                            <input type="checkbox" id="role-merchant" ${user.is_merchant ? 'checked' : ''}>
                            <span class="role-badge merchant"><i class="fas fa-store"></i> Merchant</span>
                            <small>Can sell goods and services on the platform</small>
                        </label>
                        
                        <label class="role-checkbox verified">
                            <input type="checkbox" id="role-verified" ${user.is_verified ? 'checked' : ''}>
                            <span class="role-badge verified"><i class="fas fa-check-circle"></i> Verified User</span>
                            <small>Identity and profile have been verified</small>
                        </label>
                    </div>
                </div>
            </div>
        `, () => {
            this.saveUserRoles(userId);
        });
    }

    async saveUserRoles(userId) {
        try {
            const updateData = {
                is_admin: document.getElementById('role-admin').checked,
                is_merchant: document.getElementById('role-merchant').checked,
                is_verified: document.getElementById('role-verified').checked
            };

            const response = await fetch(`${this.apiBase}/v1/admin/users/${userId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.authToken}`
                },
                body: JSON.stringify(updateData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to update user roles');
            }

            const result = await response.json();
            this.showSuccess(`Roles updated for ${result.user.username}`);
            this.closeModal();
            this.loadUsers();
        } catch (error) {
            console.error('Error saving user roles:', error);
            this.showError(`Failed to update roles: ${error.message}`);
        }
    }

    async toggleUserStatus(userId, newStatus) {
        try {
            const updateData = {
                is_verified: newStatus
            };

            const response = await fetch(`${this.apiBase}/v1/admin/users/${userId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.authToken}`
                },
                body: JSON.stringify(updateData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to update user status');
            }

            const result = await response.json();
            this.showSuccess(`${result.user.username} is now ${newStatus ? 'verified' : 'unverified'}`);
            this.loadUsers();
        } catch (error) {
            console.error('Error toggling user status:', error);
            this.showError(`Failed to update status: ${error.message}`);
        }
    }

    // Quick Action Functions
    resetUserPassword(userId) {
        this.showModal('Reset Password', `
            <p>Are you sure you want to reset the password for this user?</p>
            <p><strong>Note:</strong> This will generate a temporary password that must be changed on first login.</p>
        `, () => {
            // TODO: Implement password reset API
            this.showSuccess('Password reset link sent to user email');
            this.closeModal();
        });
    }

    viewUserTransactions(userId) {
        // Switch to transactions tab and filter by user
        this.showSection('transactions');
        // TODO: Implement user-specific transaction filtering
        this.showInfo('Showing all transactions. User-specific filtering coming soon.');
    }

    suspendUser(userId) {
        this.showModal('Suspend User', `
            <p>Are you sure you want to suspend this user account?</p>
            <p><strong>This will:</strong></p>
            <ul>
                <li>Prevent the user from logging in</li>
                <li>Block all transactions</li>
                <li>Suspend trading activities</li>
            </ul>
            <div class="form-group">
                <label for="suspend-reason">Reason for suspension:</label>
                <textarea id="suspend-reason" placeholder="Enter reason for suspension..."></textarea>
            </div>
        `, () => {
            const reason = document.getElementById('suspend-reason').value;
            if (!reason.trim()) {
                this.showError('Please provide a reason for suspension');
                return;
            }
            // TODO: Implement user suspension API
            this.showSuccess('User account suspended successfully');
            this.closeModal();
            this.loadUsers();
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
        // Ensure DOM is ready before accessing elements
        setTimeout(() => {
            const loginModal = document.getElementById('login-modal');
            const adminContainer = document.querySelector('.admin-container');
            
            if (loginModal) {
                // Use main frontend modal system
                loginModal.classList.add('active');
                console.log('Login modal shown');
            } else {
                console.error('Login modal element not found!');
            }
            
            if (adminContainer) {
                adminContainer.style.display = 'none';
            }
        }, 100);
    }

    hideLoginModal() {
        const loginModal = document.getElementById('login-modal');
        const adminContainer = document.querySelector('.admin-container');
        
        if (loginModal) {
            // Use main frontend modal system
            loginModal.classList.remove('active');
        }
        
        if (adminContainer) {
            adminContainer.style.display = 'flex';
        }
    }

    showAdminInfo() {
        if (this.currentUser) {
            const adminInfo = document.getElementById('admin-info');
            const adminUsernameDisplay = document.getElementById('admin-username-display');
            const authStatus = document.getElementById('auth-status');
            
            if (adminInfo && adminUsernameDisplay) {
                adminUsernameDisplay.textContent = this.currentUser.username;
                adminInfo.style.display = 'block';
                
                if (authStatus) {
                    authStatus.className = '';
                    authStatus.innerHTML = '<i class="fas fa-circle" style="animation: pulse 2s infinite;"></i> AUTHENTICATED';
                    authStatus.style.background = '#28a745';
                }
            }
        }
    }

    hideAdminInfo() {
        const adminInfo = document.getElementById('admin-info');
        const authStatus = document.getElementById('auth-status');
        
        if (adminInfo) {
            adminInfo.style.display = 'none';
        }
        
        if (authStatus) {
            authStatus.className = 'expired';
            authStatus.innerHTML = '<i class="fas fa-times-circle"></i> EXPIRED';
            authStatus.style.background = '#dc3545';
        }
    }

    redirectToMainApp() {
        // Redirect non-admin users to the main application
        setTimeout(() => {
            window.location.href = '/index.html';
        }, 2000);
    }

    checkAuthStatus() {
        // Check if we're on a protected route
        const isProtectedRoute = window.location.pathname.includes('/admin');
        
        if (isProtectedRoute && !this.isAuthenticated && !this.authToken) {
            this.handleAutoLogout('Authentication required for admin access');
            return false;
        }
        
        return true;
    }

    checkInitialAuth() {
        const currentPath = window.location.pathname;
        const isAdminRoute = currentPath.includes('/admin');
        
        console.log('Checking initial auth - Path:', currentPath, 'Is Admin Route:', isAdminRoute, 'Has Token:', !!this.authToken);
        
        if (isAdminRoute && this.authToken) {
            // If we have a token, validate it (mark as initial check)
            this.validateToken(true);
        } else if (isAdminRoute && !this.authToken) {
            // If no token on admin route, just show login modal (don't trigger auto-logout)
            console.log('No token found - showing login modal');
            this.showLoginModal();
        }
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

            if (response.ok && data.token && data.user) {
                // Check if user has admin privileges
                if (!data.user.is_admin) {
                    this.showLoginError('Access denied: Administrator privileges required');
                    return;
                }
                
                this.authToken = data.token;
                this.currentUser = data.user;
                localStorage.setItem('admin_token', this.authToken);
                this.isAuthenticated = true;
                this.hideLoginModal();
                this.showAdminInfo();
                this.loadDashboard();
                this.startDataRefresh();
                errorDiv.style.display = 'none';
                
                // Hide make admin button if user is already admin
                const makeAdminBtn = document.getElementById('makeAdminBtn');
                if (makeAdminBtn) {
                    makeAdminBtn.style.display = 'none';
                }
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

    async validateToken(isInitialCheck = false) {
        if (!this.authToken) {
            if (isInitialCheck) {
                // On initial page load, just show login modal
                this.showLoginModal();
                return;
            } else {
                // On subsequent validations, trigger auto-logout
                this.handleAutoLogout('No authentication token found');
                return;
            }
        }

        try {
            const response = await fetch(`${this.apiBase}/v1/users/profile`, {
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });

            if (response.ok) {
                const userData = await response.json();
                const user = userData.user;
                
                // Check if user has admin privileges
                if (!user.is_admin) {
                    this.handleAutoLogout('Access denied: Administrator privileges required');
                    setTimeout(() => {
                        this.redirectToMainApp();
                    }, 2000);
                    return;
                }
                
                this.isAuthenticated = true;
                this.currentUser = user;
                this.hideLoginModal();
                this.showAdminInfo();
                this.loadDashboard();
                this.startDataRefresh();
                
                // Hide make admin button if user is already admin
                const makeAdminBtn = document.getElementById('makeAdminBtn');
                if (makeAdminBtn) {
                    makeAdminBtn.style.display = 'none';
                }
            } else if (response.status === 401) {
                if (isInitialCheck) {
                    // On initial check, clear invalid token and show login modal
                    localStorage.removeItem('admin_token');
                    this.authToken = null;
                    this.showLoginModal();
                } else {
                    this.handleAutoLogout('Authentication token expired');
                }
            } else {
                if (isInitialCheck) {
                    this.showLoginModal();
                } else {
                    this.handleAutoLogout('Failed to validate authentication');
                }
            }
        } catch (error) {
            console.error('Token validation error:', error);
            if (isInitialCheck) {
                this.showLoginModal();
            } else {
                this.handleAutoLogout('Network error during authentication');
            }
        }
    }

    async runNewDemo() {
        this.showModal('Run New Demo', 'This will clear all existing data and run a new demonstration. Are you sure you want to continue?', async () => {
            try {
                this.showSuccess('Starting new demo...');
                
                // Call the demo endpoint (assuming we have one)
                const response = await fetch(`${this.apiBase}/v1/admin/run-demo`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${this.authToken}`,
                        'Content-Type': 'application/json'
                    }
                });

                if (response.ok) {
                    this.showSuccess('New demo completed successfully!');
                    await this.loadDemoReport();
                } else {
                    this.showError('Failed to run new demo');
                }
            } catch (error) {
                console.error('Error running new demo:', error);
                this.showError('Error running new demo: ' + error.message);
            }
            
            this.closeModal();
        });
    }

    // User Role Management Functions
    async editUserRoles(userId, userData) {
        const user = typeof userData === 'string' ? JSON.parse(userData) : userData;
        
        const modalBody = `
            <div class="user-role-form">
                <h4>Manage Roles for: ${user.username}</h4>
                <div class="form-group">
                    <label>
                        <input type="checkbox" id="role-admin" ${user.is_admin ? 'checked' : ''}>
                        <i class="fas fa-crown"></i> Administrator
                    </label>
                    <small>Full system access and management privileges</small>
                </div>
                <div class="form-group">
                    <label>
                        <input type="checkbox" id="role-merchant" ${user.is_merchant ? 'checked' : ''}>
                        <i class="fas fa-store"></i> Merchant
                    </label>
                    <small>Can sell products and receive ratings</small>
                </div>
                <div class="form-group">
                    <label>
                        <input type="checkbox" id="role-verified" ${user.is_verified ? 'checked' : ''}>
                        <i class="fas fa-check-circle"></i> Verified User
                    </label>
                    <small>Account has been verified and trusted</small>
                </div>
                <div class="form-group">
                    <label for="user-pfi">Personal Fairness Index (PFI★):</label>
                    <input type="number" id="user-pfi" min="0" max="100" value="${user.pfi || 0}">
                    <small>User's fairness score (0-100)</small>
                </div>
            </div>
        `;
        
        this.showModal('Edit User Roles', modalBody, async () => {
            await this.updateUserRoles(userId);
        });
    }

    async updateUserRoles(userId) {
        try {
            const isAdmin = document.getElementById('role-admin').checked;
            const isMerchant = document.getElementById('role-merchant').checked;
            const isVerified = document.getElementById('role-verified').checked;
            const pfi = parseInt(document.getElementById('user-pfi').value) || 0;

            const response = await fetch(`${this.apiBase}/v1/admin/users/${userId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.authToken}`
                },
                body: JSON.stringify({
                    is_admin: isAdmin,
                    is_merchant: isMerchant,
                    is_verified: isVerified,
                    pfi: pfi
                })
            });

            const data = await response.json();
            
            if (response.ok) {
                this.showSuccess('User roles updated successfully!');
                await this.loadUsers(); // Refresh the users list
            } else {
                this.showError('Failed to update user roles: ' + (data.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Error updating user roles:', error);
            this.showError('Error updating user roles: ' + error.message);
        }
    }

    async toggleUserStatus(userId, newStatus) {
        try {
            const response = await fetch(`${this.apiBase}/v1/admin/users/${userId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.authToken}`
                },
                body: JSON.stringify({
                    is_verified: newStatus
                })
            });

            const data = await response.json();
            
            if (response.ok) {
                this.showSuccess(`User ${newStatus ? 'verified' : 'unverified'} successfully!`);
                await this.loadUsers(); // Refresh the users list
            } else {
                this.showError('Failed to update user status: ' + (data.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Error updating user status:', error);
            this.showError('Error updating user status: ' + error.message);
        }
    }

    async makeAdmin() {
        try {
            const response = await fetch(`${this.apiBase}/v1/admin/make-admin`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });

            const data = await response.json();
            
            if (response.ok) {
                alert('Success! You are now an admin. Please refresh the page.');
                document.getElementById('makeAdminBtn').style.display = 'none';
                location.reload();
            } else {
                alert('Failed to make admin: ' + data.error);
            }
        } catch (error) {
            console.error('Error making admin:', error);
            alert('Error making admin: ' + error.message);
        }
    }

    logout() {
        console.log('Manual logout triggered');
        
        // Clear all authentication data
        localStorage.removeItem('admin_token');
        this.authToken = null;
        this.currentUser = null;
        this.isAuthenticated = false;
        this.hideAdminInfo();
        
        // Clear any ongoing intervals/timers
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
        
        // Show logout message and login modal
        this.showSuccess('Logged out successfully');
        this.showLoginModal();
        
        // Redirect to main page after a short delay
        setTimeout(() => {
            window.location.href = '/';
        }, 1500);
    }
}

// Global functions for onclick handlers
window.admin = null;

// Authentication functions
window.adminLogin = () => admin?.adminLogin();
window.logout = () => admin?.logout();

// Wallet modal functions
window.showSendModal = () => admin?.showSendModal();
window.hideSendModal = () => admin?.hideSendModal();
window.showReceiveModal = () => admin?.showReceiveModal();
window.hideReceiveModal = () => admin?.hideReceiveModal();
window.refreshWalletBalance = () => admin?.refreshWalletBalance();
window.refreshTransactionHistory = () => admin?.refreshTransactionHistory();
window.copyToClipboard = (elementId) => admin?.copyToClipboard(elementId);
window.hideLoginModal = () => admin?.hideLoginModal();

// Action functions
window.refreshData = () => admin?.refreshData();
window.exportData = () => admin?.exportData();
window.searchUsers = () => admin?.searchUsers();
window.makeAdmin = () => admin?.makeAdmin();
window.editUser = (userId) => admin?.editUser(userId);
window.editUserRoles = (userId, userData) => admin?.editUserRoles(userId, userData);
window.toggleUserStatus = (userId, newStatus) => admin?.toggleUserStatus(userId, newStatus);
window.resetUserPassword = (userId) => admin?.resetUserPassword(userId);
window.viewUserTransactions = (userId) => admin?.viewUserTransactions(userId);
window.suspendUser = (userId) => admin?.suspendUser(userId);
window.filterTransactions = () => admin?.filterTransactions();
window.createProposal = () => admin?.createProposal();
window.updateMonetaryPolicy = () => admin?.updateMonetaryPolicy();
window.saveSystemSettings = () => admin?.saveSystemSettings();
window.backupDatabase = () => admin?.backupDatabase();
window.viewLogs = () => admin?.viewLogs();
window.resetSystem = () => admin?.resetSystem();
window.refreshDemoData = () => admin?.loadDemoReport();
window.runNewDemo = () => admin?.runNewDemo();
window.closeModal = () => admin?.closeModal();
window.confirmModal = () => admin.confirmModal();

// Initialize admin dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.admin = new FairCoinAdmin();
});