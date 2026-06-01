# Demo Data for ExpenseOwl

This directory contains demo account and sample transaction data for testing and development purposes.

## Files

### `demo_credentials.json`
Contains the login credentials for the demo account:
- **Email**: `demo@expenseowl.com`
- **Password**: `Demo123456!`
- **User ID**: The unique identifier for the demo account in the database

**Security Note**: This file has restricted permissions (600) and should not be committed to version control for production environments.

### `demo_data.json`
Contains the transaction template and budgets for the demo account. This file includes:
- **Credentials**: Same as `demo_credentials.json`
- **Budgets**: 8 sample budget categories with IDR amounts
- **Transactions**: 30 sample transactions throughout the month covering various categories
- **GeneratedAt**: Timestamp when the data was created
- **Note**: Instructions for reusing the template

## Demo Account Details

- **Currency**: Indonesian Rupiah (IDR)
- **Transaction Period**: Full month (30 transactions)
- **Budget Categories**: 8 categories with realistic IDR amounts

### Budget Categories
1. Food & Dining: IDR 9,000,000
2. Transportation: IDR 3,000,000
3. Shopping: IDR 4,500,000
4. Entertainment: IDR 2,250,000
5. Utilities: IDR 3,750,000
6. Healthcare: IDR 1,500,000
7. Education: IDR 3,000,000
8. Personal Care: IDR 1,500,000

## Using the Demo Account

### First Time Login
1. Navigate to the ExpenseOwl login page
2. Use credentials from `demo_credentials.json`:
   - Email: `demo@expenseowl.com`
   - Password: `Demo123456!`
3. You'll have access to a full month of sample transactions and budgets

## Regenerating Demo Data

### For Next Month
To regenerate the demo data with the same structure but for a new month:

1. Run the seed script:
```bash
cd /opt/project/expenseOwl
STORAGE_TYPE=postgres \
STORAGE_URL=localhost:5432/expenseowldb \
STORAGE_USER=owluser \
STORAGE_PASS=owlpass \
STORAGE_SSL=disable \
JWT_SECRET=z8p4N9v2K7x5M1q6R3f8H0j4B2s7D1w5A9e0G6u3 \
JWT_EXPIRY_HOURS=24 \
go run scripts/seed_demo_data.go
```

2. The script will:
   - Clear existing demo account data
   - Create/update the demo account
   - Seed the same transactions and budgets with dates adjusted to the current month
   - Update `demo_credentials.json` and `demo_data.json` with the new timestamps

### Database Configuration
The seed script reads from environment variables:
- `STORAGE_TYPE`: Must be `postgres`
- `STORAGE_URL`: PostgreSQL connection URL (e.g., `localhost:5432/expenseowldb`)
- `STORAGE_USER`: Database user (e.g., `owluser`)
- `STORAGE_PASS`: Database password
- `STORAGE_SSL`: SSL mode (`disable`, `require`, etc.)

## Data Structure

### Transaction Fields
Each transaction in `demo_data.json` contains:
```json
{
  "name": "Transaction name",
  "amount": 187500,
  "category": "Food & Dining",
  "dayOfMonth": 1,
  "tags": ["breakfast"]
}
```

- **dayOfMonth**: (1-30) The day of the month when the transaction occurs
- The seed script automatically converts this to an actual date in the current month

### Budget Fields
Each budget contains:
```json
{
  "category": "Food & Dining",
  "amount": 9000000
}
```

## Note for Developers

The `dayOfMonth` field in transactions makes the template reusable across different months. When you run the seed script next month:
- Day 1 will be the 1st of the new month
- Day 15 will be the 15th of the new month
- And so on...

This allows you to maintain consistent transaction patterns without manually updating dates.

## Troubleshooting

### Demo account doesn't exist after running seed script
- Ensure PostgreSQL is running and accessible
- Check that the database `expenseowldb` exists with user `owluser`
- Verify environment variables are set correctly
- Check server logs for detailed error messages

### Currency or amount validation errors
- The seed script uses IDR (Indonesian Rupiah) as the currency
- Ensure the storage layer supports "idr" as a valid currency code
- Transaction amounts are in IDR (not cents)

## For Next Month Setup

To prepare demo data for a new month:

1. **Keep the template files** - You don't need to regenerate them manually
2. **Run the seed script** - It automatically uses the template with new dates
3. **The demo account will be updated** - Existing transactions will be replaced with new month's transactions

This approach allows quick setup of demo environments while maintaining consistent demo data structure across months.
