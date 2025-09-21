# FairCoin Development Setup

This document explains how to set up and run the FairCoin project on your local machine.

## Prerequisites

- Go 1.21 or later
- A modern web browser
- Git (optional, for version control)

## Quick Start

### 1. Build the Project

**On Windows:**
```cmd
build.bat
```

**On Linux/macOS:**
```bash
chmod +x build.sh
./build.sh
```

### 2. Configure Environment

The build script will create a `.env` file from `.env.example`. Update the configuration as needed:

```bash
cd backend
# Edit .env file with your preferred text editor
notepad .env  # Windows
nano .env     # Linux/macOS
```

### 3. Start the Server

**Development mode (recommended):**
```bash
cd backend
go run cmd/server/main.go
```

**Production mode:**
```bash
cd backend
./bin/faircoin      # Linux/macOS
./bin/faircoin.exe  # Windows
```

### 4. Access the Application

- **Frontend:** http://localhost:8080
- **API Health Check:** http://localhost:8080/health
- **API Base URL:** http://localhost:8080/api/v1

## Project Structure

```
faircoin/
├── backend/                 # Go backend server
│   ├── cmd/server/         # Application entry point
│   ├── internal/           # Internal packages
│   │   ├── api/           # REST API handlers
│   │   ├── config/        # Configuration management
│   │   ├── database/      # Database utilities
│   │   ├── models/        # Data models
│   │   └── services/      # Business logic services
│   ├── .env.example       # Environment configuration template
│   └── go.mod             # Go module definition
├── frontend/               # HTML/CSS/JS frontend
│   ├── assets/
│   │   ├── css/          # Stylesheets
│   │   └── js/           # JavaScript files
│   └── index.html        # Main HTML file
├── build.bat              # Windows build script
├── build.sh               # Linux/macOS build script
└── README.md              # Project documentation
```

## Key Features Implemented

### Backend Services

1. **User Management**
   - User registration and authentication
   - JWT-based session management
   - Profile management

2. **Wallet Services**
   - FairCoin balance tracking
   - Transaction processing
   - Fee calculation and distribution

3. **Fairness System**
   - Personal Fairness Index (PFI★) calculation
   - Trade Fairness Index (TFI★) for merchants
   - Community attestation system
   - Anti-gaming measures

4. **Monetary Policy**
   - Monthly FairCoin issuance
   - Community Basket Index (CBI) tracking
   - Fair distribution algorithms
   - Supply growth controls

5. **Governance System**
   - Community proposal creation
   - PFI-weighted voting
   - Council member election
   - Democratic decision making

### Frontend Features

1. **User Interface**
   - Responsive design for all devices
   - Modern, clean aesthetics
   - Intuitive navigation

2. **Wallet Management**
   - Balance display
   - Transaction history
   - Send/receive FairCoins
   - PFI score visualization

3. **Merchant Directory**
   - Browse verified merchants
   - TFI score display
   - Search functionality
   - Merchant registration

4. **Governance Dashboard**
   - View active proposals
   - Vote on community decisions
   - Council member display
   - Proposal creation

5. **Community Statistics**
   - Real-time network stats
   - CBI breakdown visualization
   - Transaction volume tracking
   - User growth metrics

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Token refresh

### User Management
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update profile
- `GET /api/v1/users/pfi` - Get PFI breakdown
- `POST /api/v1/users/attest` - Create attestation

### Wallet Operations
- `GET /api/v1/wallet/balance` - Get balance
- `GET /api/v1/wallet/history` - Transaction history
- `POST /api/v1/wallet/send` - Send FairCoins

### Merchants
- `GET /api/v1/merchants` - List merchants
- `POST /api/v1/merchants/register` - Register as merchant
- `GET /api/v1/merchants/:id/tfi` - Get TFI breakdown
- `POST /api/v1/merchants/:id/rate` - Rate merchant

### Governance
- `GET /api/v1/governance/proposals` - List proposals
- `POST /api/v1/governance/proposals` - Create proposal
- `POST /api/v1/governance/proposals/:id/vote` - Vote on proposal
- `GET /api/v1/governance/council` - Get council members

### Public Data
- `GET /api/v1/public/stats` - Community statistics
- `GET /api/v1/public/cbi` - Community Basket Index
- `GET /api/v1/public/merchants` - Public merchant list

## Configuration Options

### Database
- `DB_TYPE`: sqlite (development) or postgres (production)
- `DB_PATH`: SQLite database file path
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: PostgreSQL settings

### Security
- `JWT_SECRET`: Secret key for JWT tokens
- `BCRYPT_COST`: Password hashing cost (default: 12)

### Monetary Policy
- `BASE_MONTHLY_ISSUANCE`: Base monthly FairCoin issuance (default: 1000)
- `MAX_MONTHLY_GROWTH_RATE`: Maximum monthly supply growth (default: 0.02)
- `HOLDING_CAP_PERCENTAGE`: Holding cap as percentage of supply (default: 0.02)

### Fairness System
- `MIN_PFI_FOR_PROPOSALS`: Minimum PFI to create proposals (default: 50)
- `MIN_TFI_FOR_MERCHANT`: Minimum TFI for merchant status (default: 30)
- `ATTESTATION_REQUIRED_COUNT`: Required attestations for verification (default: 3)

## Development Tips

### Database
The application uses SQLite by default for development. The database file will be created automatically at `backend/faircoin.db`.

### Hot Reload
Use `go run cmd/server/main.go` for development to automatically restart on code changes (with tools like `air` or `entr`).

### Frontend Development
The frontend is served directly by the Go server. No separate build process is needed for HTML/CSS/JS files.

### Testing
Create test users and transactions to see the full functionality:

1. Register multiple users
2. Make transactions between users
3. Create attestations to increase PFI
4. Register as merchants
5. Create governance proposals

## Troubleshooting

### Common Issues

1. **Port already in use**
   - Change the `PORT` in `.env` file
   - Or stop other services using port 8080

2. **Database connection errors**
   - Check SQLite file permissions
   - Verify PostgreSQL connection settings

3. **JWT token errors**
   - Clear browser localStorage
   - Check `JWT_SECRET` configuration

4. **CORS errors**
   - Update `ALLOWED_ORIGINS` in `.env`
   - Check browser network tab for details

### Log Files
The application logs to console by default. Check terminal output for debugging information.

## Next Steps

### Enhancements to Consider

1. **Mobile App**: React Native or Flutter mobile application
2. **Smart Contracts**: Move to blockchain for full decentralization
3. **Advanced Analytics**: More detailed community metrics
4. **Messaging System**: In-app communication between users
5. **Multi-language**: Internationalization support
6. **API Rate Limiting**: Production-ready rate limiting
7. **Email Notifications**: Transaction and governance alerts
8. **Advanced Governance**: More complex voting mechanisms

### Production Deployment

1. Use PostgreSQL database
2. Set up HTTPS with TLS certificates
3. Configure reverse proxy (Nginx/Apache)
4. Set up monitoring and logging
5. Implement backup strategies
6. Configure load balancing if needed

## Contributing

This is a proof-of-concept implementation demonstrating the FairCoin principles. The code is designed to be educational and easily extensible.

Key areas for contribution:
- Security auditing
- Performance optimization
- UI/UX improvements
- Additional fairness algorithms
- Economic model refinement

## License

MIT License - feel free to use, modify, and distribute as needed.