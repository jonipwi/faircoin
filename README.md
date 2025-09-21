# FairCoin - A Community-Driven Fair Transaction System

FairCoin is a blockchain-inspired monetary system designed to address the inequality issues of Bitcoin and the inflation problems of stablecoins. It uses Personal Fairness Index (PFI★) and Trade Fairness Index (TFI★) to create a more equitable economic system.

## 🎯 Key Features

- **PFI★ (Personal Fairness Index)**: Community reputation system (0-100) based on contributions and verified acts of service
- **TFI★ (Trade Fairness Index)**: Merchant quality score (0-100) reflecting delivery, transparency, and environmental factors
- **Fair Monetary Policy**: Supply grows with community productive capacity and fairness levels
- **Anti-Wealth Concentration**: Progressive fees and holding caps prevent hoarding
- **Community Governance**: PFI-weighted voting for protocol changes
- **Stable Value**: Anchored to Community Basket Index (CBI) of staple goods and services

## 🏗️ Architecture

```
faircoin/
├── backend/          # Go backend API server
│   ├── cmd/         # Application entry points
│   ├── internal/    # Internal packages
│   ├── api/         # REST API handlers
│   ├── models/      # Data models
│   ├── services/    # Business logic
│   └── database/    # Database utilities
├── frontend/        # HTML/CSS/JS frontend
│   ├── assets/      # Static assets
│   ├── components/  # Reusable UI components
│   └── pages/       # Application pages
└── docs/           # Documentation
```

## 🚀 Quick Start

### Backend (Go)
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

### Frontend
Open `frontend/index.html` in your browser or serve with a local HTTP server.

## 💡 Core Principles

1. **Fairness Over Speculation**: Value tied to community contribution, not market speculation
2. **Transparency**: All rules are algorithmic and publicly verifiable
3. **Community Control**: Governance by PFI-weighted voting
4. **Economic Stability**: CBI-anchored value with controlled supply growth
5. **Anti-Inequality**: Built-in mechanisms to prevent wealth concentration

## 🛠️ Technology Stack

- **Backend**: Go (Golang) with Gin web framework
- **Database**: SQLite for development, PostgreSQL for production
- **Frontend**: Vanilla HTML/CSS/JavaScript (Progressive Web App)
- **Security**: JWT authentication, HTTPS, input validation

## 📊 Fairness Metrics

### PFI★ Scoring Factors:
- Community service hours
- Peer ratings and attestations
- Dispute resolution participation
- Identity verification level
- On-chain positive behavior

### TFI★ Scoring Factors:
- Delivery success rate
- Customer satisfaction
- Price transparency
- Environmental impact
- Dispute frequency

## 🏛️ Governance Model

- **Community Council**: Rotating elected representatives
- **Voting Weight**: 60% stake + 40% PFI score
- **Proposal System**: Any high-PFI member can propose changes
- **Transparency**: All votes and decisions are public

## 📈 Monetary Policy

- **Base Issuance**: Monthly minting based on community activity and fairness
- **Distribution**: 50% liquidity, 25% fairness rewards, 15% merchant incentives, 10% maintenance
- **Stability**: CBI-anchored value with algorithmic stabilizer fund
- **Anti-Inflation**: Controlled growth tied to real economic activity

## 🤝 Contributing

We welcome contributions from developers, economists, and community members who share our vision of fair economic systems.

## 📄 License

MIT License - see LICENSE file for details

---
*"Building money that serves the people, not the other way around."*