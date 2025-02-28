package bank

import "fmt"

type Customer struct { // Corrected typo: Custumer -> Customer
	name    string
	balance int
}

type Bank struct {
	accounts []*Customer
	logs     string
}

func NewBank() *Bank {
	return &Bank{
		accounts: make([]*Customer, 0), // Initialize the accounts slice
	}
}

func (b *Bank) CreateBankAccount(accountHolder string, initialBalance int) error {
	if accountHolder == "" {
		return fmt.Errorf("invalid account holder")
	}
	if initialBalance < 0 {
		return fmt.Errorf("invalid amount")
	}
	for _, account := range b.accounts {
		if account.name == accountHolder {
			return fmt.Errorf("account already exists")
		}
	}

	b.accounts = append(b.accounts, &Customer{name: accountHolder, balance: initialBalance})
	return nil
}

func (b *Bank) Deposit(accountHolder string, amount int) error {
	for _, account := range b.accounts { // No need for index since we have pointers
		if account.name == accountHolder {
			if amount <= 0 {
				return fmt.Errorf("invalid amount")
			}
			account.balance += amount // Directly modify the Customer
			b.logs += fmt.Sprintf("Deposit %d from %s\n", amount, accountHolder)
			// b.logs = append(b.logs, fmt.Sprintf("Deposit %d to %s", amount, accountHolder))
			return nil
		}
	}
	return fmt.Errorf("account not found")
}

func (b *Bank) Withdraw(accountHolder string, amount int) error {
	for _, account := range b.accounts {
		if account.name == accountHolder {
			if amount <= 0 {
				return fmt.Errorf("invalid amount")
			}
			if account.balance < amount {
				return fmt.Errorf("insufficient funds")
			}
			account.balance -= amount
			b.logs += fmt.Sprintf("Withdraw %d from %s\n", amount, accountHolder)
			return nil
		}
	}
	return fmt.Errorf("account not found")
}

func (b *Bank) GetBalance(accountHolder string) (int, error) {
	for _, account := range b.accounts {
		if account.name == accountHolder {
			return account.balance, nil
		}
	}
	return 0, fmt.Errorf("account not found")
}

func (b *Bank) Transfer(from string, to string, amount int) error {
	fromIndex, toIndex := -1, -1

	for i := range b.accounts { // Use range with index to get proper indices
		if b.accounts[i].name == from {
			fromIndex = i
		}
		if b.accounts[i].name == to {
			toIndex = i
		}
	}

	if fromIndex == -1 || toIndex == -1 {
		return fmt.Errorf("one or both accounts not found")
	}

	if amount <= 0 {
		return fmt.Errorf("invalid amount")
	}

	if b.accounts[fromIndex].balance < amount {
		return fmt.Errorf("insufficient funds")
	}

	b.accounts[fromIndex].balance -= amount
	b.accounts[toIndex].balance += amount

	b.logs += fmt.Sprintf("Transfer %d from %s to %s", amount, from, to)
	return nil
}
func (b *Bank) TransactionLogs() string {
	if b == nil {
		return ""
	}
	return b.logs
}
