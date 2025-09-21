// FairCoin Frontend JavaScript

class FairCoinApp {
    constructor() {
        this.apiUrl = 'http://localhost:8080/api/v1';
        this.token = localStorage.getItem('faircoin_token');
        this.user = null;
        
        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.setupNavigation();
        
        // Check if user is logged in
        if (this.token) {
            await this.loadUserData();
        }
        
        // Load initial data
        await this.loadCommunityStats();
        await this.loadMerchants();
        
        // Update UI based on auth status
        this.updateAuthUI();
    }

    setupEventListeners() {
        // Navigation
        document.getElementById('navToggle').addEventListener('click', this.toggleMobileNav);
        
        // Auth buttons
        document.getElementById('loginBtn').addEventListener('click', () => this.showModal('loginModal'));
        document.getElementById('registerBtn').addEventListener('click', () => this.showModal('registerModal'));
        document.getElementById('logoutBtn').addEventListener('click', () => this.logout());
        
        // Modal controls
        document.getElementById('loginModalClose').addEventListener('click', () => this.hideModal('loginModal'));
        document.getElementById('registerModalClose').addEventListener('click', () => this.hideModal('registerModal'));
        document.getElementById('sendModalClose').addEventListener('click', () => this.hideModal('sendModal'));
        document.getElementById('createProposalModalClose').addEventListener('click', () => this.hideModal('createProposalModal'));
        
        // Modal switching
        document.getElementById('showRegisterModal').addEventListener('click', (e) => {
            e.preventDefault();
            this.hideModal('loginModal');
            this.showModal('registerModal');
        });
        document.getElementById('showLoginModal').addEventListener('click', (e) => {
            e.preventDefault();
            this.hideModal('registerModal');
            this.showModal('loginModal');
        });
        
        // Forms
        document.getElementById('loginForm').addEventListener('submit', (e) => this.handleLogin(e));
        document.getElementById('registerForm').addEventListener('submit', (e) => this.handleRegister(e));
        document.getElementById('sendForm').addEventListener('submit', (e) => this.handleSend(e));
        document.getElementById('createProposalForm').addEventListener('submit', (e) => this.handleCreateProposal(e));
        
        // Wallet actions
        document.getElementById('sendBtn').addEventListener('click', () => this.showModal('sendModal'));
        document.getElementById('historyBtn').addEventListener('click', () => this.loadTransactionHistory());
        
        // Merchant actions
        document.getElementById('becomeMerchantBtn').addEventListener('click', () => this.becomeMerchant());
        document.getElementById('merchantSearch').addEventListener('input', (e) => this.searchMerchants(e.target.value));
        
        // Governance actions
        document.getElementById('createProposalBtn').addEventListener('click', () => this.showModal('createProposalModal'));
        
        // Auth-required buttons
        document.getElementById('walletLoginBtn').addEventListener('click', () => this.showModal('loginModal'));
        document.getElementById('governanceLoginBtn').addEventListener('click', () => this.showModal('loginModal'));
        
        // Get started button
        document.getElementById('getStartedBtn').addEventListener('click', () => {
            if (this.token) {
                this.showSection('wallet');
            } else {
                this.showModal('registerModal');
            }
        });
        
        // Close modals on outside click
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal')) {
                this.hideModal(e.target.id);
            }
        });
    }

    setupNavigation() {
        const navLinks = document.querySelectorAll('.nav-link');
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = link.getAttribute('data-section');
                this.showSection(section);
                
                // Update active nav link
                navLinks.forEach(l => l.classList.remove('active'));
                link.classList.add('active');
                
                // Close mobile nav
                document.getElementById('navMenu').classList.remove('active');
            });
        });
    }

    showSection(sectionId) {
        // Hide all sections
        document.querySelectorAll('.section').forEach(section => {
            section.classList.remove('active');
        });
        
        // Show target section
        document.getElementById(sectionId).classList.add('active');
        
        // Load section-specific data
        this.loadSectionData(sectionId);
    }

    async loadSectionData(sectionId) {
        switch (sectionId) {
            case 'wallet':
                if (this.token) {
                    await this.loadWalletData();
                }
                break;
            case 'merchants':
                await this.loadMerchants();
                break;
            case 'fairness':
                await this.setupFairnessMetrics();
                break;
            case 'governance':
                if (this.token) {
                    await this.loadGovernanceData();
                }
                break;
            case 'community':
                await this.loadCommunityStats();
                await this.loadCommunityBasketIndex();
                break;
        }
    }

    toggleMobileNav() {
        document.getElementById('navMenu').classList.toggle('active');
    }

    showModal(modalId) {
        document.getElementById(modalId).classList.add('active');
    }

    hideModal(modalId) {
        document.getElementById(modalId).classList.remove('active');
    }

    showLoading() {
        document.getElementById('loadingOverlay').classList.add('active');
    }

    hideLoading() {
        document.getElementById('loadingOverlay').classList.remove('active');
    }

    showToast(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.innerHTML = `
            <div>${message}</div>
        `;
        
        document.getElementById('toastContainer').appendChild(toast);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            toast.remove();
        }, 5000);
    }

    async apiCall(endpoint, options = {}) {
        const url = `${this.apiUrl}${endpoint}`;
        const config = {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            }
        };

        if (this.token) {
            config.headers.Authorization = `Bearer ${this.token}`;
        }

        try {
            const response = await fetch(url, config);
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'API request failed');
            }
            
            return data;
        } catch (error) {
            console.error('API Error:', error);
            throw error;
        }
    }

    async handleLogin(e) {
        e.preventDefault();
        
        const username = document.getElementById('loginUsername').value;
        const password = document.getElementById('loginPassword').value;
        
        try {
            this.showLoading();
            
            const response = await this.apiCall('/auth/login', {
                method: 'POST',
                body: JSON.stringify({ username, password })
            });
            
            this.token = response.token;
            this.user = response.user;
            localStorage.setItem('faircoin_token', this.token);
            
            this.hideModal('loginModal');
            this.updateAuthUI();
            this.showToast('Login successful!', 'success');
            
            // Load user data
            await this.loadUserData();
            
        } catch (error) {
            this.showToast(error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    async handleRegister(e) {
        e.preventDefault();
        
        const firstName = document.getElementById('registerFirstName').value;
        const lastName = document.getElementById('registerLastName').value;
        const username = document.getElementById('registerUsername').value;
        const email = document.getElementById('registerEmail').value;
        const password = document.getElementById('registerPassword').value;
        const passwordConfirm = document.getElementById('registerPasswordConfirm').value;
        
        if (password !== passwordConfirm) {
            this.showToast('Passwords do not match', 'error');
            return;
        }
        
        try {
            this.showLoading();
            
            const response = await this.apiCall('/auth/register', {
                method: 'POST',
                body: JSON.stringify({
                    first_name: firstName,
                    last_name: lastName,
                    username,
                    email,
                    password
                })
            });
            
            this.token = response.token;
            this.user = response.user;
            localStorage.setItem('faircoin_token', this.token);
            
            this.hideModal('registerModal');
            this.updateAuthUI();
            this.showToast('Registration successful! Welcome to FairCoin!', 'success');
            
            // Show wallet section
            this.showSection('wallet');
            
        } catch (error) {
            this.showToast(error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    async handleSend(e) {
        e.preventDefault();
        
        const toUsername = document.getElementById('sendToUsername').value;
        const amount = parseFloat(document.getElementById('sendAmount').value);
        const description = document.getElementById('sendDescription').value;
        
        try {
            this.showLoading();
            
            await this.apiCall('/wallet/send', {
                method: 'POST',
                body: JSON.stringify({
                    to_username: toUsername,
                    amount,
                    description
                })
            });
            
            this.hideModal('sendModal');
            this.showToast('FairCoins sent successfully!', 'success');
            
            // Reload wallet data
            await this.loadWalletData();
            
            // Clear form
            document.getElementById('sendForm').reset();
            
        } catch (error) {
            this.showToast(error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    async handleCreateProposal(e) {
        e.preventDefault();
        
        const title = document.getElementById('proposalTitle').value;
        const type = document.getElementById('proposalType').value;
        const description = document.getElementById('proposalDescription').value;
        
        try {
            this.showLoading();
            
            await this.apiCall('/governance/proposals', {
                method: 'POST',
                body: JSON.stringify({
                    title,
                    type,
                    description
                })
            });
            
            this.hideModal('createProposalModal');
            this.showToast('Proposal created successfully!', 'success');
            
            // Reload governance data
            await this.loadGovernanceData();
            
            // Clear form
            document.getElementById('createProposalForm').reset();
            
        } catch (error) {
            this.showToast(error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    logout() {
        this.token = null;
        this.user = null;
        localStorage.removeItem('faircoin_token');
        
        this.updateAuthUI();
        this.showSection('home');
        this.showToast('Logged out successfully', 'info');
    }

    updateAuthUI() {
        const isLoggedIn = !!this.token;
        
        // Show/hide auth buttons
        document.getElementById('loginBtn').style.display = isLoggedIn ? 'none' : 'inline-flex';
        document.getElementById('registerBtn').style.display = isLoggedIn ? 'none' : 'inline-flex';
        document.getElementById('userMenu').style.display = isLoggedIn ? 'flex' : 'none';
        
        // Update user name
        if (isLoggedIn && this.user) {
            document.getElementById('userName').textContent = this.user.username;
        }
        
        // Show/hide auth-required content
        document.querySelectorAll('.auth-required').forEach(el => {
            el.style.display = isLoggedIn ? 'none' : 'flex';
        });
        
        document.querySelectorAll('.auth-only').forEach(el => {
            el.classList.toggle('show', isLoggedIn);
        });
        
        // Show/hide wallet and governance content
        const walletContent = document.getElementById('walletContent');
        const governanceContent = document.getElementById('governanceContent');
        
        if (walletContent) walletContent.style.display = isLoggedIn ? 'block' : 'none';
        if (governanceContent) governanceContent.style.display = isLoggedIn ? 'block' : 'none';
    }

    async loadUserData() {
        if (!this.token) return;
        
        try {
            const response = await this.apiCall('/users/profile');
            this.user = response.user;
        } catch (error) {
            console.error('Failed to load user data:', error);
            this.logout();
        }
    }

    async loadWalletData() {
        if (!this.token) return;
        
        try {
            // Load balance
            const balanceResponse = await this.apiCall('/wallet/balance');
            document.getElementById('walletBalance').textContent = `${balanceResponse.balance.toFixed(2)} FC`;
            document.getElementById('walletLocked').textContent = `${balanceResponse.locked_fc.toFixed(2)} FC`;
            
            // Load PFI
            const pfiResponse = await this.apiCall('/users/pfi');
            this.updatePFIDisplay(pfiResponse);
            
            // Load transaction history
            await this.loadTransactionHistory();
            
        } catch (error) {
            console.error('Failed to load wallet data:', error);
            this.showToast('Failed to load wallet data', 'error');
        }
    }

    updatePFIDisplay(pfiData) {
        document.getElementById('pfiScore').textContent = pfiData.current_pfi || 0;
        document.getElementById('communityServiceHours').textContent = `${pfiData.community_service_hours || 0} hours`;
        document.getElementById('attestationCount').textContent = pfiData.total_attestations || 0;
        document.getElementById('accountAge').textContent = `${pfiData.account_age_days || 0} days`;
    }

    async loadTransactionHistory() {
        if (!this.token) return;
        
        try {
            const response = await this.apiCall('/wallet/history?limit=10');
            const transactionList = document.getElementById('transactionList');
            
            if (response.transactions.length === 0) {
                transactionList.innerHTML = '<div class="text-center" style="padding: 2rem;">No transactions yet</div>';
                return;
            }
            
            transactionList.innerHTML = response.transactions.map(tx => this.renderTransaction(tx)).join('');
            
        } catch (error) {
            console.error('Failed to load transaction history:', error);
        }
    }

    renderTransaction(tx) {
        const isReceived = tx.to_user_id === this.user.id;
        const iconClass = isReceived ? 'received' : 'sent';
        const iconSymbol = isReceived ? '↓' : '↑';
        const amountClass = isReceived ? 'positive' : 'negative';
        const amountPrefix = isReceived ? '+' : '-';
        
        return `
            <div class="transaction-item">
                <div class="transaction-info">
                    <div class="transaction-icon ${iconClass}">
                        ${iconSymbol}
                    </div>
                    <div class="transaction-details">
                        <h4>${tx.description || (isReceived ? 'Received' : 'Sent')}</h4>
                        <p>${new Date(tx.created_at).toLocaleDateString()}</p>
                    </div>
                </div>
                <div class="transaction-amount">
                    <div class="amount ${amountClass}">${amountPrefix}${tx.amount.toFixed(2)} FC</div>
                    ${tx.fee > 0 ? `<div class="fee">Fee: ${tx.fee.toFixed(2)} FC</div>` : ''}
                </div>
            </div>
        `;
    }

    async loadMerchants() {
        try {
            const response = await this.apiCall('/public/merchants');
            const merchantsGrid = document.getElementById('merchantsGrid');
            
            if (response.merchants.length === 0) {
                merchantsGrid.innerHTML = '<div class="col-span-full text-center">No merchants found</div>';
                return;
            }
            
            merchantsGrid.innerHTML = response.merchants.map(merchant => this.renderMerchant(merchant)).join('');
            
        } catch (error) {
            console.error('Failed to load merchants:', error);
        }
    }

    renderMerchant(merchant) {
        const initials = `${merchant.first_name[0]}${merchant.last_name[0]}`.toUpperCase();
        
        return `
            <div class="merchant-card">
                <div class="merchant-avatar">${initials}</div>
                <div class="merchant-name">${merchant.first_name} ${merchant.last_name}</div>
                <div class="merchant-username">@${merchant.username}</div>
                <div class="tfi-score">
                    <span>TFI★</span>
                    <span class="tfi-number">${merchant.tfi}</span>
                    <span>/100</span>
                </div>
                <div class="merchant-actions">
                    <button class="btn btn-primary btn-sm" onclick="app.contactMerchant('${merchant.id}')">
                        <i class="fas fa-message"></i> Contact
                    </button>
                    <button class="btn btn-outline btn-sm" onclick="app.rateMerchant('${merchant.id}')">
                        <i class="fas fa-star"></i> Rate
                    </button>
                </div>
            </div>
        `;
    }

    async loadGovernanceData() {
        if (!this.token) return;
        
        try {
            // Load council members
            const councilResponse = await this.apiCall('/governance/council');
            this.renderCouncilMembers(councilResponse.council_members);
            
            // Load proposals
            const proposalsResponse = await this.apiCall('/governance/proposals');
            this.renderProposals(proposalsResponse.proposals);
            
        } catch (error) {
            console.error('Failed to load governance data:', error);
        }
    }

    renderCouncilMembers(members) {
        const councilContainer = document.getElementById('councilMembers');
        
        councilContainer.innerHTML = members.map(member => {
            const initials = `${member.first_name[0]}${member.last_name[0]}`.toUpperCase();
            return `
                <div class="council-member">
                    <div class="member-avatar">${initials}</div>
                    <div class="member-name">${member.first_name} ${member.last_name}</div>
                    <div class="member-pfi">PFI: ${member.pfi}</div>
                </div>
            `;
        }).join('');
    }

    renderProposals(proposals) {
        const proposalsList = document.getElementById('proposalsList');
        
        if (proposals.length === 0) {
            proposalsList.innerHTML = '<div class="text-center">No active proposals</div>';
            return;
        }
        
        proposalsList.innerHTML = proposals.map(proposal => this.renderProposal(proposal)).join('');
    }

    renderProposal(proposal) {
        return `
            <div class="proposal-card">
                <div class="proposal-header">
                    <div>
                        <div class="proposal-title">${proposal.title}</div>
                        <div class="proposal-meta">
                            By ${proposal.proposer.username} • 
                            ${new Date(proposal.created_at).toLocaleDateString()}
                        </div>
                    </div>
                    <div class="proposal-type">${proposal.type.replace('_', ' ')}</div>
                </div>
                <div class="proposal-description">${proposal.description}</div>
                <div class="proposal-voting">
                    <div class="vote-counts">
                        <div class="vote-count">
                            <i class="fas fa-thumbs-up text-success"></i>
                            <span>${proposal.votes_for}</span>
                        </div>
                        <div class="vote-count">
                            <i class="fas fa-thumbs-down text-danger"></i>
                            <span>${proposal.votes_against}</span>
                        </div>
                    </div>
                    <div class="vote-actions">
                        <button class="btn btn-vote-for btn-sm" onclick="app.voteOnProposal('${proposal.id}', true)">
                            <i class="fas fa-thumbs-up"></i> For
                        </button>
                        <button class="btn btn-vote-against btn-sm" onclick="app.voteOnProposal('${proposal.id}', false)">
                            <i class="fas fa-thumbs-down"></i> Against
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    async loadCommunityStats() {
        try {
            const response = await this.apiCall('/public/stats');
            
            // Update hero stats
            document.getElementById('totalUsers').textContent = this.formatNumber(response.total_users || 0);
            document.getElementById('totalMerchants').textContent = this.formatNumber(response.total_merchants || 0);
            document.getElementById('circulatingSupply').textContent = this.formatNumber(response.circulating_supply || 0);
            document.getElementById('averagePFI').textContent = (response.average_pfi || 0).toFixed(1);
            
            // Update community section stats
            document.getElementById('statCirculatingSupply').textContent = `${this.formatNumber(response.circulating_supply || 0)} FC`;
            document.getElementById('statTransactionVolume').textContent = `${this.formatNumber(response.transaction_volume_30d || 0)} FC`;
            document.getElementById('statAveragePFI').textContent = (response.average_pfi || 0).toFixed(1);
            
        } catch (error) {
            console.error('Failed to load community stats:', error);
        }
    }

    async loadCommunityBasketIndex() {
        try {
            const response = await this.apiCall('/public/cbi');
            const cbi = response.cbi;
            
            document.getElementById('statCBI').textContent = cbi.value.toFixed(2);
            
            // Update CBI chart
            this.updateCBIChart(cbi);
            
        } catch (error) {
            console.error('Failed to load CBI:', error);
        }
    }

    updateCBIChart(cbi) {
        const items = [
            { key: 'food', value: cbi.food_index, element: 'cbiFoodValue', bar: 'cbiFoodBar' },
            { key: 'labor', value: cbi.labor_index, element: 'cbiLaborValue', bar: 'cbiLaborBar' },
            { key: 'energy', value: cbi.energy_index, element: 'cbiEnergyValue', bar: 'cbiEnergyBar' },
            { key: 'housing', value: cbi.housing_index, element: 'cbiHousingValue', bar: 'cbiHousingBar' }
        ];
        
        items.forEach(item => {
            document.getElementById(item.element).textContent = item.value.toFixed(1);
            document.getElementById(item.bar).style.width = `${(item.value / 150) * 100}%`;
        });
    }

    async becomeMerchant() {
        if (!this.token) {
            this.showModal('loginModal');
            return;
        }
        
        try {
            this.showLoading();
            
            await this.apiCall('/merchants/register', {
                method: 'POST'
            });
            
            this.showToast('Successfully registered as merchant!', 'success');
            await this.loadUserData();
            
        } catch (error) {
            this.showToast(error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    async voteOnProposal(proposalId, vote) {
        if (!this.token) {
            this.showModal('loginModal');
            return;
        }
        
        try {
            this.showLoading();
            
            await this.apiCall(`/governance/proposals/${proposalId}/vote`, {
                method: 'POST',
                body: JSON.stringify({ vote })
            });
            
            this.showToast(`Vote ${vote ? 'for' : 'against'} recorded!`, 'success');
            await this.loadGovernanceData();
            
        } catch (error) {
            this.showToast(error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    searchMerchants(query) {
        const merchantCards = document.querySelectorAll('.merchant-card');
        
        merchantCards.forEach(card => {
            const name = card.querySelector('.merchant-name').textContent.toLowerCase();
            const username = card.querySelector('.merchant-username').textContent.toLowerCase();
            
            if (name.includes(query.toLowerCase()) || username.includes(query.toLowerCase())) {
                card.style.display = 'block';
            } else {
                card.style.display = 'none';
            }
        });
    }

    contactMerchant(merchantId) {
        this.showToast('Messaging feature coming soon!', 'info');
    }

    rateMerchant(merchantId) {
        this.showToast('Rating feature coming soon!', 'info');
    }

    formatNumber(num) {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        }
        if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toString();
    }

    // ===============================
    // FAIRNESS METRICS FUNCTIONALITY
    // ===============================

    async setupFairnessMetrics() {
        // Setup event listeners for fairness metrics
        const refreshBtn = document.getElementById('refreshMetricsBtn');
        const timeRangeSelect = document.getElementById('metricsTimeRange');
        
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.refreshFairnessMetrics());
        }
        
        if (timeRangeSelect) {
            timeRangeSelect.addEventListener('change', () => this.loadMetricsHistory());
        }

        // Load initial fairness metrics data
        await this.loadFairnessMetrics();
        await this.loadMetricsHistory();
    }

    async loadFairnessMetrics() {
        try {
            this.showLoading();
            
            // Load comprehensive fairness metrics
            const response = await fetch(`${this.apiUrl}/metrics/fairness`, {
                headers: this.getAuthHeaders()
            });
            
            if (!response.ok) {
                throw new Error('Failed to load fairness metrics');
            }
            
            const data = await response.json();
            
            // Update PFI distribution
            this.updatePFIDistribution(data.pfi_distribution);
            
            // Update TFI analysis
            this.updateTFIAnalysis(data.tfi_analysis);
            
            // Update top merchants
            this.updateTopMerchants(data.top_merchants);
            
            // Update CBI data
            this.updateCBIData(data.cbi);
            
            // Update alerts
            this.updateFairnessAlerts(data.alerts);
            
        } catch (error) {
            console.error('Error loading fairness metrics:', error);
            this.showToast('Failed to load fairness metrics', 'error');
        } finally {
            this.hideLoading();
        }
    }

    updatePFIDistribution(distribution) {
        if (!distribution) return;
        
        const total = distribution.total_users || 1;
        
        // Update excellent segment
        const excellent = distribution.excellent || { count: 0, percentage: 0 };
        document.getElementById('pfiExcellentPct').textContent = `${excellent.percentage}%`;
        document.getElementById('pfiExcellentCount').textContent = `${excellent.count} users`;
        
        // Update good segment
        const good = distribution.good || { count: 0, percentage: 0 };
        document.getElementById('pfiGoodPct').textContent = `${good.percentage}%`;
        document.getElementById('pfiGoodCount').textContent = `${good.count} users`;
        
        // Update average segment
        const average = distribution.average || { count: 0, percentage: 0 };
        document.getElementById('pfiAveragePct').textContent = `${average.percentage}%`;
        document.getElementById('pfiAverageCount').textContent = `${average.count} users`;
        
        // Update poor segment
        const poor = distribution.poor || { count: 0, percentage: 0 };
        document.getElementById('pfiPoorPct').textContent = `${poor.percentage}%`;
        document.getElementById('pfiPoorCount').textContent = `${poor.count} users`;
    }

    updateTFIAnalysis(analysis) {
        if (!analysis) return;
        
        document.getElementById('tfiAverageRating').textContent = `${analysis.average_rating || 0}/5.0`;
        document.getElementById('tfiTotalRatings').textContent = this.formatNumber(analysis.total_ratings || 0);
        document.getElementById('tfiTotalMerchants').textContent = this.formatNumber(analysis.total_merchants || 0);
    }

    updateTopMerchants(merchants) {
        const tbody = document.getElementById('topMerchantsBody');
        if (!tbody || !merchants) return;
        
        tbody.innerHTML = '';
        
        if (merchants.length === 0) {
            tbody.innerHTML = `
                <div class="no-merchants">
                    <p>No merchant data available</p>
                </div>
            `;
            return;
        }
        
        merchants.forEach(merchant => {
            const row = document.createElement('div');
            row.className = 'merchant-row';
            
            // Generate star rating
            const stars = this.generateStarRating(merchant.average_rating);
            
            row.innerHTML = `
                <div class="td rank">#${merchant.rank}</div>
                <div class="td merchant">
                    <div class="merchant-name">${merchant.first_name} ${merchant.last_name}</div>
                    <div class="merchant-username">@${merchant.username}</div>
                </div>
                <div class="td"><span class="tfi-score">${merchant.tfi}</span></div>
                <div class="td">
                    <div class="rating-stars">${stars}</div>
                    <div class="rating-value">${merchant.average_rating.toFixed(1)}</div>
                </div>
                <div class="td">${this.formatNumber(merchant.total_ratings)}</div>
            `;
            
            tbody.appendChild(row);
        });
    }

    generateStarRating(rating) {
        const fullStars = Math.floor(rating);
        const hasHalfStar = rating % 1 >= 0.5;
        let stars = '';
        
        for (let i = 0; i < fullStars; i++) {
            stars += '<i class="fas fa-star"></i>';
        }
        
        if (hasHalfStar) {
            stars += '<i class="fas fa-star-half-alt"></i>';
        }
        
        const emptyStars = 5 - Math.ceil(rating);
        for (let i = 0; i < emptyStars; i++) {
            stars += '<i class="far fa-star"></i>';
        }
        
        return stars;
    }

    updateCBIData(cbi) {
        if (!cbi) return;
        
        // Update main CBI value
        document.getElementById('cbiCurrentValue').textContent = cbi.current_cbi || 100;
        
        // Update trend
        const trendElement = document.getElementById('cbiTrend');
        const trend = cbi.trend || 'stable';
        
        trendElement.className = `cbi-trend ${trend}`;
        trendElement.innerHTML = this.getTrendIcon(trend) + `<span>${trend}</span>`;
        
        // Update components
        const components = cbi.components || {};
        this.updateCBIComponent('food', components.food || 100);
        this.updateCBIComponent('energy', components.energy || 100);
        this.updateCBIComponent('labor', components.labor || 100);
        this.updateCBIComponent('housing', components.housing || 100);
    }

    getTrendIcon(trend) {
        switch (trend) {
            case 'increasing':
                return '<i class="fas fa-arrow-up"></i>';
            case 'decreasing':
                return '<i class="fas fa-arrow-down"></i>';
            default:
                return '<i class="fas fa-minus"></i>';
        }
    }

    updateCBIComponent(component, value) {
        const fillElement = document.getElementById(`cbi${component.charAt(0).toUpperCase() + component.slice(1)}Fill`);
        const valueElement = document.getElementById(`cbi${component.charAt(0).toUpperCase() + component.slice(1)}Val`);
        
        if (fillElement) {
            fillElement.style.width = `${Math.min(100, Math.max(0, value))}%`;
        }
        
        if (valueElement) {
            valueElement.textContent = value.toFixed(0);
        }
    }

    updateFairnessAlerts(alerts) {
        const alertsContainer = document.getElementById('fairnessAlerts');
        if (!alertsContainer) return;
        
        alertsContainer.innerHTML = '';
        
        if (!alerts || alerts.length === 0) {
            alertsContainer.innerHTML = `
                <div class="no-alerts">
                    <i class="fas fa-check-circle"></i>
                    <h4>All Systems Normal</h4>
                    <p>No fairness alerts at this time. The system is operating smoothly.</p>
                </div>
            `;
            return;
        }
        
        alerts.forEach(alert => {
            const alertElement = document.createElement('div');
            alertElement.className = `alert alert-${alert.severity}`;
            
            const icon = this.getAlertIcon(alert.type, alert.severity);
            
            alertElement.innerHTML = `
                <div class="alert-icon">${icon}</div>
                <div class="alert-content">
                    <div class="alert-title">${alert.title}</div>
                    <div class="alert-description">${alert.description}</div>
                </div>
                <div class="alert-actions">
                    <button class="alert-btn" onclick="app.markAlertAsRead('${alert.id}')">Mark Read</button>
                </div>
            `;
            
            alertsContainer.appendChild(alertElement);
        });
    }

    getAlertIcon(type, severity) {
        const icons = {
            pfi_decline: 'fas fa-user-times',
            tfi_decline: 'fas fa-store-slash',
            cbi_change: 'fas fa-chart-line'
        };
        
        const severityIcons = {
            low: 'fas fa-info-circle',
            medium: 'fas fa-exclamation-triangle',
            high: 'fas fa-exclamation-circle',
            critical: 'fas fa-times-circle'
        };
        
        return `<i class="${icons[type] || severityIcons[severity] || 'fas fa-bell'}"></i>`;
    }

    async markAlertAsRead(alertId) {
        try {
            const response = await fetch(`${this.apiUrl}/metrics/alerts/${alertId}/read`, {
                method: 'POST',
                headers: this.getAuthHeaders()
            });
            
            if (response.ok) {
                this.showToast('Alert marked as read', 'success');
                // Reload alerts
                await this.loadFairnessMetrics();
            } else {
                throw new Error('Failed to mark alert as read');
            }
        } catch (error) {
            console.error('Error marking alert as read:', error);
            this.showToast('Failed to mark alert as read', 'error');
        }
    }

    async loadMetricsHistory() {
        try {
            const timeRange = document.getElementById('metricsTimeRange').value || '30';
            
            const response = await fetch(`${this.apiUrl}/metrics/history?days=${timeRange}`, {
                headers: this.getAuthHeaders()
            });
            
            if (!response.ok) {
                throw new Error('Failed to load metrics history');
            }
            
            const data = await response.json();
            this.renderMetricsChart(data.history);
            
        } catch (error) {
            console.error('Error loading metrics history:', error);
            this.showToast('Failed to load metrics history', 'error');
        }
    }

    renderMetricsChart(historyData) {
        const canvas = document.getElementById('metricsChartCanvas');
        const container = document.getElementById('metricsChart');
        
        if (!canvas || !historyData || historyData.length === 0) {
            container.innerHTML = `
                <div style="padding: 2rem; text-align: center; color: #7f8c8d;">
                    <i class="fas fa-chart-line" style="font-size: 3rem; margin-bottom: 1rem; opacity: 0.3;"></i>
                    <p>No historical data available</p>
                </div>
            `;
            return;
        }
        
        // Simple chart rendering (you might want to use Chart.js for better charts)
        const ctx = canvas.getContext('2d');
        const width = canvas.width;
        const height = canvas.height;
        
        // Clear canvas
        ctx.clearRect(0, 0, width, height);
        
        // Draw a simple line chart placeholder
        ctx.strokeStyle = '#667eea';
        ctx.lineWidth = 2;
        ctx.beginPath();
        
        // Sample line drawing
        for (let i = 0; i < historyData.length; i++) {
            const x = (i / (historyData.length - 1)) * width;
            const y = height - (historyData[i].average_tfi / 100) * height;
            
            if (i === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        }
        
        ctx.stroke();
        
        // Add chart.js message
        ctx.fillStyle = '#7f8c8d';
        ctx.font = '14px Arial';
        ctx.textAlign = 'center';
        ctx.fillText('Chart visualization requires Chart.js library', width / 2, height / 2);
    }

    async refreshFairnessMetrics() {
        await this.loadFairnessMetrics();
        await this.loadMetricsHistory();
        this.showToast('Fairness metrics refreshed', 'success');
    }

    getAuthHeaders() {
        const headers = {
            'Content-Type': 'application/json'
        };
        
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        return headers;
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new FairCoinApp();
});